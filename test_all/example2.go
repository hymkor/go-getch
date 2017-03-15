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
		fmt.Println(e.String())
	}
}
