package getch

import (
	"os"
	"os/signal"
)

func ctrlCHandler(ch chan os.Signal) {
	for _ = range ch {
		event1 := Event{Key: &keyEvent{3, 0, LEFT_CTRL_PRESSED}}
		if eventBuffer == nil {
			eventBuffer = []Event{event1}
			eventBufferRead = 0
		} else {
			eventBuffer = append(eventBuffer, event1)
		}
	}
}

func IsCtrlCPressed() bool {
	if eventBuffer != nil {
		for _, p := range eventBuffer[eventBufferRead:] {
			if p.Key != nil && p.Key.Rune == rune(3) {
				return true
			}
		}
	}
	return false
}

func DisableCtrlC() {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)
	go ctrlCHandler(ch)
}
