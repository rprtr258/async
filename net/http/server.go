package http

import (
	"fmt"
	"net"
	"net/http"
	"os"

	. "github.com/rprtr258/async"
)

type Request struct {
	*http.Request
	Response http.ResponseWriter
	done     chan<- struct{}
}

func (r Request) Done() {
	r.done <- struct{}{}
}

func Run(addr string) Stream[Request] {
	res := make(chan Request)
	go func() {
		defer close(res)
		ln, err := net.Listen("tcp", addr)
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			return
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
				// wait till processing done
				<-doneCh
			}),
		}

		fmt.Fprintln(os.Stderr, server.Serve(ln).Error())
	}()
	return NewStream(res)
}

func RunIter(addr string) func(func(Request) bool) {
	res := make(chan Request)
	go func() {
		defer close(res)
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
					// wait till processing done
					<-doneCh
				}),
			}
			return server.Serve(ln)
		}()
		fmt.Fprintln(os.Stderr, err.Error())
	}()
	return func(yield func(Request) bool) {
		for req := range res {
			if !yield(req) {
				// TODO: close server
				return
			}
		}
	}
}
