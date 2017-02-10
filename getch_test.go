package getch

import (
	"fmt"
	"testing"
)

func TestAll(t *testing.T) {
	for i := 0; i < 3; i++ {
		fmt.Printf("[%d/3] ", i+1)
		e := All()
		if k := e.Key; k != nil {
			fmt.Printf("key hit: code=%04X scan=%04X shift=%04X\n",
				k.Rune, k.Scan, k.Shift)
		}
		if r := e.Resize; r != nil {
			fmt.Printf("window resize: width=%d height=%d\n",
				r.Width, r.Height)
		}

	}
}
