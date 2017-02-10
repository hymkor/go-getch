package getch

// Deprecated: use All()
var OnWindowResize func(w, h uint)

// Deprecated: use All()
// Get all keyboard-event.
func Full() (code rune, scan uint16, shift uint32) {
	var flag uintptr = 0
	if OnWindowResize != nil {
		flag = ENABLE_WINDOW_INPUT
	}
	for {
		event := getEvent(flag)
		if e := event.Resize; e != nil {
			OnWindowResize(e.Width, e.Height)
		}
		if e := event.Key; e != nil {
			return e.Rune, e.Scan, e.Shift
		}
	}
}
