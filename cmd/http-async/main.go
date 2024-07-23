package main

import (
	"bytes"
	"strconv"

	. "github.com/rprtr258/imhttp"
	"github.com/rprtr258/imhttp/net"
)

func unwrap[T any](res T, err error) T {
	if err != nil {
		panic(err)
	}
	return res
}

const (
	indexHTML = `<!DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="utf-8">
    <title>Hello!</title>
  </head>
  <body>
    <h1>Hello!</h1>
    <p>Hi from Rust</p>
  </body>
</html>
`
	notFoundHTML = `<!DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="utf-8">
    <title>Hello!</title>
  </head>
  <body>
    <h1>Oops!</h1>
    <p>Sorry, I don't know what you're asking for.</p>
  </body>
</html>
`
)

func handle_connection(stream net.Conn) Future[struct{}] {
	return NewFuture(func() struct{} {
		defer stream.Close()

		var buffer [1024]byte
		stream.Read(buffer[:]).Await().Unwrap()

		get := []byte("GET / HTTP/1.1\r\n")

		// Respond with greetings or a 404,
		// depending on the data in the request
		var statusLine, contents string
		if bytes.HasPrefix(buffer[:], get) {
			statusLine, contents = "200 OK", indexHTML
		} else {
			statusLine, contents = "404 NOT FOUND", notFoundHTML
		}

		// Write response back to the stream,
		// and flush the stream to ensure the response is sent back to the client
		stream.Write([]byte("HTTP/1.1 " + statusLine + "\r\n" +
			"Content-Type: text/html; charset=utf-8\r\n" +
			"Content-Length: " + strconv.Itoa(len(contents)) + "\r\n" +
			"\r\n" + contents,
		)).Await().Unwrap()
		return struct{}{}
	})
}

func main() {
	listener := net.Listen("tcp", "127.0.0.1:7878").Unwrap()
	listener.ForEachConcurrent(func(conn net.Conn) Future[struct{}] {
		return handle_connection(conn)
	}).Await()
}
