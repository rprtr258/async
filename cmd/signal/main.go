package main

import (
	"fmt"
	"syscall"

	imhttp "github.com/rprtr258/async"
)

func main() {
	for sigs := imhttp.NotifySignals(syscall.SIGINT, syscall.SIGTERM); ; {
		sig := sigs.Await()
		fmt.Println(sig.String(), int(sig.(syscall.Signal)))
	}
}
