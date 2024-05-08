package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"os"
	"strconv"

	"github.com/rprtr258/imhttp"
)

func main() {
	// no synchronization at all!
	counters := map[string]int{}
	addr := ":8080"
	fmt.Fprintln(os.Stderr, "listening on", addr)
	for req := range imhttp.Run(addr) {
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
		// don't forget to call Done or connection will leak
		req.Done()
	}
}
