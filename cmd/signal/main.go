package main

import (
	"fmt"
	"syscall"

	"github.com/rprtr258/imhttp"
)

func main() {
	for sigs := imhttp.NotifySignals(syscall.SIGINT, syscall.SIGTERM); ; {
		sig := sigs.Await()
		fmt.Println(sig.String(), int(sig.(syscall.Signal)))
	}
}
