package luddite

import (
	"crypto/tls"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// Based on http://www.hydrogen18.com/blog/stop-listening-http-server-go.html,
// but stops on SIGINT instead of explicit Stop() call

type ListenerStoppedError struct{}

func (e *ListenerStoppedError) Error() string {
	return "listener stopped"
}

type StoppableTCPListener struct {
	*net.TCPListener
	stop       chan os.Signal
	keepalives bool
}

func (sl *StoppableTCPListener) Accept() (net.Conn, error) {
	for {
		// Wait up to one second for a new connection
		err := sl.TCPListener.SetDeadline(time.Now().Add(time.Second))
		if err != nil {
			return nil, err
		}
		newConn, err := sl.TCPListener.AcceptTCP()

		// Check for the channel being closed
		select {
		case <-sl.stop:
			return nil, &ListenerStoppedError{}
		default:
			// If nothing came in on the channel, continue as normal
		}

		if err != nil {
			// If this is a timeout, then continue to wait for new connections
			if e, ok := err.(net.Error); ok && e.Timeout() && e.Temporary() {
				continue
			}
			return nil, err
		}

		if sl.keepalives {
			_ = newConn.SetKeepAlive(true)
			_ = newConn.SetKeepAlivePeriod(3 * time.Minute)
		}
		return newConn, err
	}
}

func NewStoppableTCPListener(addr string, keepalives bool) (net.Listener, error) {
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}

	sl := &StoppableTCPListener{
		TCPListener: l.(*net.TCPListener),
		stop:        make(chan os.Signal, 1),
		keepalives:  keepalives,
	}
	signal.Notify(sl.stop, syscall.SIGINT)
	return sl, nil
}

func NewStoppableTLSListener(addr string, keepalives bool, certFile string, keyFile string) (net.Listener, error) {
	tlsConfig := &tls.Config{
		NextProtos:   []string{"http/1.1", "h2"},
		Certificates: make([]tls.Certificate, 1),
	}

	var err error
	if tlsConfig.Certificates[0], err = tls.LoadX509KeyPair(certFile, keyFile); err != nil {
		return nil, err
	}

	stl, err := NewStoppableTCPListener(addr, keepalives)
	if err != nil {
		return nil, err
	}
	return tls.NewListener(stl, tlsConfig), nil
}
