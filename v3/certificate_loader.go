package luddite

import (
	"crypto/tls"
	"fmt"
	"path"
	"sync/atomic"
	"time"

	"github.com/fsnotify/fsnotify"
	log "github.com/sirupsen/logrus"
)

const (
	dedupDelay = 5 * time.Second
)

type CertificateLoader interface {
	GetCertificate(*tls.ClientHelloInfo) (*tls.Certificate, error)
	Close() error
}

type certLoader struct {
	cert         atomic.Pointer[tls.Certificate]
	certFilePath string
	keyFilePath  string
	watcher      *fsnotify.Watcher
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
	if config.Transport.ReloadOnUpdate {
		if err := cl.watch(); err != nil {
			return nil, err
		}
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

func (l *certLoader) Close() error {
	if l.watcher != nil {
		return l.watcher.Close()
	}
	return nil
}

func (l *certLoader) watch() error {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	l.watcher = w

	certFileDir := path.Dir(l.certFilePath)
	if err = l.watcher.Add(certFileDir); err != nil {
		return fmt.Errorf("error adding dir '%s' to watcher: %s", certFileDir, err.Error())
	}
	l.log.Debugf("cert directory '%s' added to watcher", certFileDir)

	keyFileDir := path.Dir(l.keyFilePath)
	if keyFileDir != certFileDir {
		if err = l.watcher.Add(keyFileDir); err != nil {
			return fmt.Errorf("error adding dir '%s' to watcher: %s", keyFileDir, err.Error())
		}
		l.log.Debugf("key directory '%s' added to watcher", keyFileDir)
	}
	certFile := path.Base(l.certFilePath)
	keyFile := path.Base(l.keyFilePath)
	go l.fsWatch(l.watcher, []string{certFile, keyFile})

	return nil
}

func (l *certLoader) fsWatch(watcher *fsnotify.Watcher, filenames []string) {
	var timer *time.Timer
	defer func() {
		if timer != nil {
			timer.Stop()
		}
	}()
	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			if updated := l.handleFsEvents(event, filenames); updated {
				// N.B. process the event after a delay to avoid duplicates when both files are written
				timer = l.setDeDupTimer(timer)
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			l.log.WithError(err).Error("certificate watcher error")
		}
	}
}

func (l *certLoader) handleFsEvents(event fsnotify.Event, files []string) bool {
	if event.Has(fsnotify.Write) || event.Has(fsnotify.Create) {
		for _, fn := range files {
			if path.Base(event.Name) == path.Base(fn) {
				l.log.Debugf("file '%s' was updated", fn)
				return true
			}
		}
	}
	return false
}

func (l *certLoader) setDeDupTimer(timer *time.Timer) *time.Timer {
	if timer == nil {
		timer = time.AfterFunc(time.Hour, func() {
			if err := l.storeCertificate(); err != nil {
				l.log.WithError(err).Error("error reloading certificate")
			}
		})
		timer.Stop()
	}
	timer.Reset(dedupDelay)
	return timer
}
