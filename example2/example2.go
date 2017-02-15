package main

import (
	"fmt"

	"github.com/zetamatta/go-getch"
)

const COUNT = 5

func main() {
	for i := 0; i < COUNT; i++ {
		fmt.Printf("[%d/%d] ", i+1, COUNT)
		e := getch.All()
		if k := e.Key; k != nil {
			fmt.Printf("\n%c %08X %08X %08X\n",
				k.Rune, k.Rune, k.Scan, k.Shift)
		} else if r := e.Resize; r != nil {
			fmt.Printf("\nWidth=%d Height=%d\n", r.Width, r.Height)
		} else {
			fmt.Println("\n(unknown event)")
		}
	}
}
