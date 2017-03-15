package main

import (
	"fmt"
	"time"

	".."
)

func main() {
	getch.DisableCtrlC()

	for i := 5; i >= 0; i-- {
		fmt.Printf("%d\n", i)
		time.Sleep(time.Second)
	}
	if getch.IsCtrlCPressed() {
		fmt.Println("^C")
	}
}
