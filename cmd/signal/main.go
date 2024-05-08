package main

import (
	"fmt"
	"syscall"

	"github.com/rprtr258/imhttp/signal"
)

func main() {
	for sig := range signal.New(syscall.SIGINT, syscall.SIGTERM) {
		fmt.Println(sig.String(), int(sig.(syscall.Signal)))
	}
}
