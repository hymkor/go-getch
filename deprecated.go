package getch

import (
	"github.com/zetamatta/go-getch/consoleinput"
)

// Deprecated: use All()
var OnWindowResize func(w, h uint)

// Deprecated: use All()
// Get all keyboard-event.
func Full() (code rune, scan uint16, shift uint32) {
	var flag uint32 = 0
	if OnWindowResize != nil {
		flag = consoleinput.ENABLE_WINDOW_INPUT
	}
	for {
		event := hconin.GetEvent_(flag)
		if e := event.Resize; e != nil {
			OnWindowResize(e.Width, e.Height)
		}
		if e := event.Key; e != nil {
			return e.Rune, e.Scan, e.Shift
		}
	}
}
