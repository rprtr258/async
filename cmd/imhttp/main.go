package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"os"
	"strconv"
	"time"

	"github.com/rprtr258/imhttp"
)

func main() {
	const addr = ":8080"
	fmt.Fprintln(os.Stderr, "listening on", addr)
	appSrv := imhttp.Run(addr)
	metricsSrv := imhttp.Run(":9000")
	for {
		i, req := imhttp.Select(appSrv, metricsSrv)
		switch i {
		case 0: // app
			b, _ := httputil.DumpRequest(req.Request, true)
			log.Println("REQUEST:")
			log.Println(string(b))
			http.NotFound(req.Response, req.Request)
		case 1: // metric
			if req.URL.Path == "/" {
				log.Println("GET METRICS")
				fmt.Fprintln(req.Response, time.Now().String())
			}
		}
		// do not forget to finish processing request
		req.Done()
	}
}

func main2() {
	// no synchronization at all!
	counters := map[string]int{}
	addr := ":8080"
	fmt.Fprintln(os.Stderr, "listening on", addr)
	for reqs := imhttp.Run(addr); ; {
		req := reqs.Await()
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
			http.NotFound(req.Response, req.Request)
		}
		// do not forget to finish processing request
		req.Done()
	}
}
