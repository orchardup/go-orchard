package proxy

import "net"
import "io"
import "fmt"

type Proxy struct {
	ErrorChannel chan error
	ListenFunc   func() (net.Listener, error)
	DialFunc     func() (net.Conn, error)

	Listener *net.Listener
}

func New(listenFunc func() (net.Listener, error), dialFunc func() (net.Conn, error)) *Proxy {
	p := new(Proxy)

	p.ErrorChannel = make(chan error)
	p.ListenFunc = listenFunc
	p.DialFunc = dialFunc

	return p
}

func (p *Proxy) Start() {
	listener, err := p.ListenFunc()
	p.Listener = &listener
	if err != nil {
		p.ErrorChannel <- err
		return
	}

	p.ErrorChannel <- nil

	for {
		clientConn, err := listener.Accept()
		if err != nil {
			panic(err)
		}
		go p.ForwardConnection(clientConn)
	}
}

func (p *Proxy) Stop() {
	if *p.Listener != nil {
		(*p.Listener).Close()
	}
}

func (p *Proxy) ForwardConnection(clientConn net.Conn) {
	defer clientConn.Close()
	serverConn, err := p.DialFunc()
	if err != nil {
		fmt.Printf("error connecting upstream: %s\n", err)
		return
	}
	defer serverConn.Close()
	complete := make(chan bool)
	go Copy(serverConn, clientConn, complete)
	go Copy(clientConn, serverConn, complete)
	<-complete
	<-complete
}

func Copy(to net.Conn, from net.Conn, complete chan bool) {
	io.Copy(to, from)
	CloseWrite(to)
	complete <- true
}

func CloseWrite(rwc net.Conn) {
	if tcpc, ok := rwc.(*net.TCPConn); ok {
		tcpc.CloseWrite()
	} else if unixc, ok := rwc.(*net.UnixConn); ok {
		unixc.CloseWrite()
	}
}
