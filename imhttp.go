package imhttp

import (
	"fmt"
	"net"
	"net/http"
	"os"
)

type Request struct {
	*http.Request
	Response http.ResponseWriter
	done     <-chan struct{}
}

func (c Request) Done() {
	<-c.done
}

func Run(addr string) <-chan Request {
	res := make(chan Request)
	go func() {
		err := func() error {
			ln, err := net.Listen("tcp", addr)
			if err != nil {
				return err
			}
			defer ln.Close()

			server := &http.Server{
				Addr: addr,
				Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					doneCh := make(chan struct{})
					res <- Request{
						Request:  r,
						Response: w,
						done:     doneCh,
					}
					doneCh <- struct{}{}
				}),
			}
			return server.Serve(ln)
		}()
		fmt.Fprintln(os.Stderr, err.Error())
		close(res)
	}()
	return res
}
