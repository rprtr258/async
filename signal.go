package imhttp

import (
	"os"
	"os/signal"
)

func NotifySignals(signals ...os.Signal) Future[os.Signal] {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, signals...)
	return Future[os.Signal]{ch}
}
