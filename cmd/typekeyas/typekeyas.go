package main

import (
	"os"

	"github.com/zetamatta/go-getch/consoleinput"
	"github.com/zetamatta/go-getch/typekeyas"
)

func main() {
	console, err := consoleinput.New()
	if err != nil {
		println(err.Error())
		return
	}
	for _, s := range os.Args[1:] {
		typekeyas.String(console,s)
		typekeyas.Rune(console,'\r')
	}
}
