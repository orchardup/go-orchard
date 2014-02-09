package proxy

import "net"
import "io"

type Proxy struct {
	ErrorChannel chan error
	FromType     string
	FromAddr     string
	ToType       string
	ToAddr       string

	Listener *net.Listener
}

func New(fromType string, fromAddr string, toType string, toAddr string) *Proxy {
	p := new(Proxy)

	p.ErrorChannel = make(chan error)
	p.FromType = fromType
	p.FromAddr = fromAddr
	p.ToType = toType
	p.ToAddr = toAddr

	return p
}

func (p *Proxy) Start() {
	listener, err := net.Listen(p.FromType, p.FromAddr)
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
	serverConn, err := net.Dial(p.ToType, p.ToAddr)
	if err != nil {
		panic(err)
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
