package main

import (
	"fmt"
	"github.com/zetamatta/go-getch/consoleoutput"
	"os"
)

func Main() error {
	console, err := consoleoutput.New()
	if err != nil {
		return err
	}
	defer console.Close()

	for _, arg1 := range os.Args {
		fmt.Println(arg1)
		output, err := console.GetRecentOutput()
		if err != nil {
			return err
		}
		fmt.Printf("-->[%s]\n", output)
		fmt.Print(arg1)
		output, err = console.GetRecentOutput()
		if err != nil {
			return err
		}
		fmt.Printf("\n-->[%s]\n", output)
	}
	return nil
}

func main() {
	if err := Main(); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
	os.Exit(0)
}
