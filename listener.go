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

type ListenerStoppedError struct {
}

func (e *ListenerStoppedError) Error() string {
	return "Listener stopped"
}

type StoppableTCPListener struct {
	*net.TCPListener
	stop       chan os.Signal
	keepalives bool
}

func (sl *StoppableTCPListener) Accept() (net.Conn, error) {

	for {
		//Wait up to one second for a new connection
		sl.TCPListener.SetDeadline(time.Now().Add(time.Second))
		newConn, err := sl.TCPListener.AcceptTCP()

		//Check for the channel being closed
		select {
		case <-sl.stop:
			return nil, &ListenerStoppedError{}
		default:
			//If nothing came in on the channel, continue as normal
		}

		if err != nil {
			netErr, ok := err.(net.Error)

			//If this is a timeout, then continue to wait for
			//new connections
			if ok && netErr.Timeout() && netErr.Temporary() {
				continue
			}
		}

		if sl.keepalives {
			newConn.SetKeepAlive(true)
			newConn.SetKeepAlivePeriod(3 * time.Minute)
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
		l.(*net.TCPListener),
		make(chan os.Signal, 1),
		keepalives,
	}
	signal.Notify(sl.stop, syscall.SIGINT)
	return sl, nil
}

func NewStoppableTLSListener(addr string, keepalives bool, certFile string, keyFile string) (net.Listener, error) {
	var err error
	tlsConfig := &tls.Config{}
	tlsConfig.NextProtos = []string{"http/1.1"}
	tlsConfig.Certificates = make([]tls.Certificate, 1)
	tlsConfig.Certificates[0], err = tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, err
	}

	stl, err := NewStoppableTCPListener(addr, keepalives)
	if err != nil {
		return nil, err
	}
	tlsListener := tls.NewListener(stl, tlsConfig)
	return tlsListener, nil
}
