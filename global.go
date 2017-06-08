package getch

import (
	"github.com/zetamatta/go-getch/consoleevent"
)

var hconin *consoleevent.Handle

func init() {
	var err error
	hconin, err = consoleevent.New()
	if err != nil {
		panic(err.Error())
	}
}

// Get all console-event (keyboard,resize,...)
func All() consoleevent.Event {
	return hconin.All()
}

// Get character as a Rune
func Rune() rune {
	return hconin.Rune()
}

func Count() (int, error) {
	return hconin.GetNumberOfEvent()
}

func Flush() error {
	return hconin.Flush()
}

// wait for keyboard event
func Wait(timeout_msec uintptr) (bool, error) {
	return hconin.Wait(timeout_msec)
}

func Within(msec uintptr) (consoleevent.Event, error) {
	return hconin.Within(msec)
}

func RuneWithin(msec uintptr) (rune, error) {
	return hconin.RuneWithin(msec)
}

func IsCtrlCPressed() bool {
	return hconin.IsCtrlCPressed()
}

func DisableCtrlC() {
	hconin.DisableCtrlC()
}
