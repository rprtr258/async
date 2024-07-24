package net

import (
	"net"

	. "github.com/rprtr258/async"
)

type Conn struct{ net.Conn }

func (c Conn) Read(b []byte) Future[Result[int]] {
	return NewFuture(func() Result[int] {
		return NewResult(c.Conn.Read(b))
	})
}

func (c Conn) Write(b []byte) Future[Result[int]] {
	return NewFuture(func() Result[int] {
		return NewResult(c.Conn.Write(b))
	})
}

func Listen(network, addr string) Result[Stream[Conn]] {
	listener, err := net.Listen(network, addr)
	if err != nil {
		return Err[Stream[Conn]](err)
	}
	return Ok(NewGenerator(func() (Conn, bool) {
		conn, err := listener.Accept()
		return Conn{conn}, err == nil
	}))
}
