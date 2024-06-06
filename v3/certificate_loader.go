package luddite

import (
	"crypto/tls"
	"fmt"
	"path"
	"time"

	"github.com/fsnotify/fsnotify"
	log "github.com/sirupsen/logrus"
)

const (
	dedupDelay = 2 * time.Second
	loggerName = "luddite.v3.certificate_loader"
)

type CertificateLoader interface {
	GetCertificate(*tls.ClientHelloInfo) (*tls.Certificate, error)
	Watch() error
	Close() error
}

type certLoader struct {
	cert         *tls.Certificate
	certFilePath string
	keyFilePath  string
	watcher      *fsnotify.Watcher
	log          *log.Entry
}

func NewCertificateLoader(config *ServiceConfig, logger *log.Logger) (CertificateLoader, error) {
	cl := &certLoader{
		certFilePath: config.Transport.CertFilePath,
		keyFilePath:  config.Transport.KeyFilePath,
		log:          logger.WithFields(log.Fields{"logger": loggerName}),
	}
	if err := cl.loadCertificate(); err != nil {
		return nil, err
	}
	if config.Transport.ReloadOnUpdate {
		if err := cl.Watch(); err != nil {
			return nil, err
		}
	}
	return cl, nil
}

func (l *certLoader) GetCertificate(_ *tls.ClientHelloInfo) (*tls.Certificate, error) {
	return l.cert, nil
}

func (l *certLoader) Watch() error {
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
	go l.fsWatch(l.watcher, []string{certFile, keyFile}, l.loadCertificate)

	return nil
}

func (l *certLoader) Close() error {
	if l.watcher != nil {
		return l.watcher.Close()
	}
	return nil
}

func (l *certLoader) fsWatch(watcher *fsnotify.Watcher, filenames []string, onUpdate func() error) {
	var (
		timer *time.Timer
	)
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
			if updated := handleFsEvents(event, filenames, l.log); updated {
				timer = setTimer(timer, onUpdate, l.log)
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			log.WithError(err).Error("certificate watcher error")
		}
	}
}

func (l *certLoader) loadCertificate() error {
	l.log.Debugf("loading cert: '%s', key: '%s'", l.certFilePath, l.keyFilePath)
	cert, err := tls.LoadX509KeyPair(l.certFilePath, l.keyFilePath)
	if err != nil {
		return fmt.Errorf("failed to load certificate '%s': '%s'", l.certFilePath, err)
	}
	l.cert = &cert
	return nil
}

func handleFsEvents(event fsnotify.Event, files []string, logEntry *log.Entry) bool {
	if event.Has(fsnotify.Write) || event.Has(fsnotify.Create) {
		for _, fn := range files {
			if path.Base(event.Name) == path.Base(fn) {
				logEntry.Debugf("file '%s' was updated", fn)
				return true
			}
		}
	}
	return false
}

func setTimer(timer *time.Timer, onUpdate func() error, logEntry *log.Entry) *time.Timer {
	if timer == nil {
		timer = time.AfterFunc(time.Hour, func() {
			if err := onUpdate(); err != nil {
				logEntry.WithError(err).Error("error updating certificate")
			}
		})
		timer.Stop()
	}
	timer.Reset(dedupDelay)
	return timer
}
