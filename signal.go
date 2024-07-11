package imhttp

import (
	"os"
	"os/signal"
)

func NotifySignals(signals ...os.Signal) Promise[os.Signal] {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, signals...)
	return Promise[os.Signal]{ch}
}
