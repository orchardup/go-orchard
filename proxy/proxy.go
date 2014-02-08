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
		incoming, err := listener.Accept()
		if err != nil {
			panic(err)
		}
		go p.ForwardConnection(incoming)
	}
}

func (p *Proxy) Stop() {
	if p.Listener != nil {
		(*p.Listener).Close()
	}
}

func (p *Proxy) ForwardConnection(incoming net.Conn) {
	outgoing, err := net.Dial(p.ToType, p.ToAddr)
	if err != nil {
		panic(err)
	}
	go io.Copy(incoming, outgoing)
	go io.Copy(outgoing, incoming)
}
