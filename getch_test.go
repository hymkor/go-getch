package getch

import (
	"fmt"
	"testing"
)

func onWindowResize(w, h uint) {
	fmt.Printf("width=%d height=%d\n", w, h)
}

func TestFull(t *testing.T) {
	OnWindowResize = onWindowResize
	code, scan, shift := Full()
	fmt.Printf("code=%04X scan=%04X shift=%04X\n", code, scan, shift)
	OnWindowResize = nil
	code, scan, shift = Full()
	fmt.Printf("code=%04X scan=%04X shift=%04X\n", code, scan, shift)
}
