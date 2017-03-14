package getch

import (
	"fmt"
	"testing"
	"time"
)

func TestAll(t *testing.T) {
	if err := Flush(); err != nil {
		t.Error(err.Error())
		return
	}
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

func TestCount(t *testing.T) {
	var err error
	if err = Flush(); err != nil {
		t.Error(err.Error())
		return
	}
	var n int
	for {
		n, err = Count()
		if err != nil {
			t.Error(err.Error())
			return
		}
		if n > 0 {
			break
		}
		fmt.Println("sleep")
		time.Sleep(time.Second)
	}
	fmt.Printf("break(n=%d)\n", n)
	e := All()
	if e.Key != nil {
		fmt.Printf("'%c' typed.\n", e.Key.Rune)
	} else if e.Resize != nil {
		fmt.Printf("(%d,%d)\n", e.Resize.Width, e.Resize.Height)
	} else if e.Mouse != nil {
		fmt.Println("mouse event")
	} else if e.Menu != nil {
		fmt.Println("menu event")
	} else if e.Focus != nil {
		fmt.Println("focus event")
	} else if e.KeyUp != nil {
		fmt.Printf("keyup event '%c'\n",e.KeyUp.Rune)
	} else {
		fmt.Printf("Otherwise\n")
	}
}
