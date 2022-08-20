package signal

import (
	"os"
	"os/signal"
)

var Ch chan os.Signal

func Init() {
	Ch = make(chan os.Signal, 1)
	signal.Notify(Ch, os.Interrupt)
}
