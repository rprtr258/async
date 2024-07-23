package main

import (
	"fmt"
	"log"
	stdhttp "net/http"
	"net/http/httputil"
	"os"
	"strconv"
	"syscall"
	"time"

	. "github.com/rprtr258/async"
	"github.com/rprtr258/async/net/http"
)

func main() {
	terminate := NotifySignals(syscall.SIGTERM, syscall.SIGINT).Then(func(s os.Signal) {
		os.Exit(1)
	})

	const addr = ":8080"
	fmt.Fprintln(os.Stderr, "listening on", addr)
	appSrv := http.Run(addr)
	appHandler := func(req http.Request) Future[struct{}] {
		return NewFuture(func() struct{} {
			defer req.Done() // do not forget to finish processing request

			b, _ := httputil.DumpRequest(req.Request, true)
			log.Println("REQUEST:")
			log.Println(string(b))
			req.Response.Header().Set("Content-Type", "text/plain; charset=utf-8")
			req.Response.Header().Set("X-Content-Type-Options", "nosniff")
			req.Response.WriteHeader(stdhttp.StatusOK)
			req.Response.Write(b)
			return struct{}{}
		})
	}

	metricsSrv := http.Run(":9000")
	metricsHandler := func(req http.Request) Future[struct{}] {
		return NewFuture(func() struct{} {
			defer req.Done() // do not forget to finish processing request

			if req.URL.Path == "/" {
				log.Println("GET METRICS")
				fmt.Fprintln(req.Response, time.Now().String())
			}
			return struct{}{}
		})
	}

	s := NewFutureSet[struct{}]()
	// TODO: add timer example
	s.Push(terminate)
	s.PushStream(appSrv.Then(appHandler))
	s.PushStream(metricsSrv.Then(metricsHandler))
	ss := s.Stream()
	for {
		if _, ok := ss.Next().Await().Unpack(); !ok {
			break
		}
	}
}

func main2() {
	// no synchronization at all!
	counters := map[string]int{}
	addr := ":8080"
	fmt.Fprintln(os.Stderr, "listening on", addr)
	for reqs := http.Run(addr); ; {
		req := reqs.Next().Await().Unwrap()
		switch req.URL.Path {
		case "/get":
			name := req.URL.Query().Get("name")
			fmt.Fprintln(req.Response, counters[name])
		case "/set":
			name := req.URL.Query().Get("name")
			value, _ := strconv.Atoi(req.URL.Query().Get("value"))
			counters[name] = value
		case "/inc":
			name := req.URL.Query().Get("name")
			counters[name]++
		default:
			b, _ := httputil.DumpRequest(req.Request, true)
			log.Println("REQUEST:")
			log.Println(string(b))
			stdhttp.NotFound(req.Response, req.Request)
		}
		// do not forget to finish processing request
		req.Done()
	}
}
