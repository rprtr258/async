package main

import (
	"fmt"
	"net"
	"os"
	"strconv"

	. "github.com/rprtr258/imhttp"
)

func run(host string, port int, path string) error {
	addrs, err := net.ResolveTCPAddr("tcp", host+":"+strconv.Itoa(port))
	if err != nil {
		return fmt.Errorf("Could not get address of %q: %w", host, err)
	}

	conn, err := net.DialTCP("tcp", nil, addrs)
	if err != nil {
		return fmt.Errorf("Could not connect to %s:%d: %w", host, port, err)
	}
	defer conn.Close()

	imhttp := New(conn)

	imhttp.Req(IMHTTP_GET, path).
		ReqHeaders(map[string]string{
			"Host":  host,
			"Foo":   "Bar",
			"Hello": "World",
		}).
		ReqBodyChunk("Hello, World\n").
		ReqBodyChunk("Test, test, test\n")

	imhttp.ResBegin()
	{
		fmt.Printf("Status Code: %d\n", imhttp.ResStatusCode())

		// Headers
		var name, value []byte
		for imhttp.ResNextHeader(&name, &value) {
			fmt.Printf("------------------------------\n")
			fmt.Printf("Header Name: %s\n", name)
			fmt.Printf("Header Value: %s\n", value)
		}
		fmt.Printf("------------------------------\n")

		// Body
		var chunk []byte
		for imhttp.ResNextBodyChunk(&chunk) {
			fmt.Printf("%s", chunk)
		}
	}
	imhttp.ResEnd()

	return nil
}

func main() {
	if err := run(
		// TODO: Sometimes http://anglesharp.azurewebsites.net/Chunked fires the asserts
		// "anglesharp.azurewebsites.net", 80, "/Chunked",
		"google.com", 80, "/",
	); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
