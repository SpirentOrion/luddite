package luddite

import (
	"crypto/tls"
	"fmt"
	"os"
	"path/filepath"
	"sync/atomic"
	"time"

	log "github.com/sirupsen/logrus"
)

const (
	defaultWatcherScanMinutes = 5
)

type CertificateLoader interface {
	GetCertificate(*tls.ClientHelloInfo) (*tls.Certificate, error)
	Close()
}

type certLoader struct {
	cert         atomic.Pointer[tls.Certificate]
	certFilePath string
	keyFilePath  string
	watcher      Watcher
	log          *log.Logger
}

func NewCertificateLoader(config *ServiceConfig, logger *log.Logger) (CertificateLoader, error) {
	cl := &certLoader{
		certFilePath: config.Transport.CertFilePath,
		keyFilePath:  config.Transport.KeyFilePath,
		log:          logger,
	}
	if err := cl.storeCertificate(); err != nil {
		return nil, err
	}
	if !config.Transport.CertWatcher.Disabled {
		cl.watcher = NewWatcher(logger, cl.certFilePath, cl.keyFilePath)
		scanMinutes := config.Transport.CertWatcher.ScanMinutes
		if scanMinutes == 0 {
			scanMinutes = defaultWatcherScanMinutes
		}
		cl.watcher.Watch(cl.storeCertificate, time.Duration(scanMinutes)*time.Minute)
	}
	return cl, nil
}

func (l *certLoader) storeCertificate() error {
	l.log.Debugf("storing cert: '%s', key: '%s'", l.certFilePath, l.keyFilePath)
	cert, err := tls.LoadX509KeyPair(l.certFilePath, l.keyFilePath)
	if err != nil {
		return fmt.Errorf("failed to load certificate '%s': '%s'", l.certFilePath, err)
	}
	l.cert.Store(&cert)
	return nil
}

func (l *certLoader) GetCertificate(_ *tls.ClientHelloInfo) (*tls.Certificate, error) {
	return l.cert.Load(), nil
}

func (l *certLoader) Close() {
	if l.watcher != nil {
		l.watcher.Close()
	}
	return
}

type Watcher interface {
	Close()
	Watch(loadCertCallback func() error, frequency time.Duration)
}

type watcher struct {
	watchPaths WatchPaths
	done       chan interface{}
	log        *log.Logger
}

func NewWatcher(logger *log.Logger, paths ...string) Watcher {
	return &watcher{
		watchPaths: NewWatchPaths(logger, paths...),
		done:       make(chan interface{}),
		log:        logger,
	}
}

func (w *watcher) Close() {
	close(w.done)
}

func (w *watcher) Watch(loadCertCallback func() error, frequency time.Duration) {
	go func() {
		ticker := time.NewTicker(frequency)
		defer ticker.Stop()
		for {
			select {
			case <-w.done:
				return
			case <-ticker.C:
				if w.watchPaths.Update() {
					if err := loadCertCallback(); err != nil {
						w.log.WithError(err).Error("error reloading certificate")
					}
				}
			}
		}
	}()
}

type WatchPaths []WatchPath

func NewWatchPaths(logger *log.Logger, paths ...string) WatchPaths {
	wps := make(WatchPaths, len(paths))
	for i, fp := range paths {
		wps[i] = NewWatchPath(fp, logger)
	}
	return wps
}

func (wps WatchPaths) Update() (modified bool) {
	for _, wp := range wps {
		modified = modified || wp.Update()
	}
	return
}

type WatchPath interface {
	Update() bool
}

type watchPath struct {
	path    string
	modTime time.Time
	log     *log.Logger
}

func NewWatchPath(p string, logger *log.Logger) WatchPath {
	wp := &watchPath{
		path: p,
		log:  logger,
	}
	wp.Update()
	return wp
}

func (wp *watchPath) Update() (modified bool) {
	if latestModTime := wp.latestModTime(); !latestModTime.IsZero() {
		if modified = wp.modTime.IsZero() || !wp.modTime.Equal(latestModTime); modified {
			wp.modTime = latestModTime
			wp.log.Debugf("mod time stored for path '%s'", wp.path)
		}
	}
	return
}

func (wp *watchPath) latestModTime() (lmt time.Time) {
	f, err := filepath.EvalSymlinks(wp.path)
	if err != nil {
		wp.log.WithError(err).Errorf("failed to eval file path '%s'", wp.path)
		return
	}
	fi, err := os.Stat(f)
	if err != nil {
		wp.log.WithError(err).Errorf("failed to get file info '%s'", f)
		return
	}
	lmt = fi.ModTime().UTC()
	wp.log.Debugf("got file info '%s': '%s'", f, lmt)
	return
}
