package main

import (
	"fmt"
	"os"

	"github.com/zetamatta/go-getch/consoleinput"
	"github.com/zetamatta/go-getch/typekeyas"
)

func main() {
	console, err := consoleinput.New()
	if err != nil {
		fmt.Fprintln(os.Stderr,err.Error())
		return
	}
	for _, s := range os.Args[1:] {
		typekeyas.String(console,s)
		typekeyas.Rune(console,'\r')
	}
	if err = console.Close() ; err != nil {
		fmt.Fprintln(os.Stderr,err.Error())
	}
}
