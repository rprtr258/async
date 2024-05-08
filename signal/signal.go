package signal

import (
	"os"
	"os/signal"
)

func New(signals ...os.Signal) <-chan os.Signal {
	ch := make(chan os.Signal)
	signal.Notify(ch, signals...)
	return ch
}
