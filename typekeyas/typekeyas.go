package main

import (
	"os"

	"github.com/zetamatta/go-getch/consoleinput"
)

func writeRune(handle consoleinput.Handle,c rune) uint32 {
	records := []consoleinput.InputRecord{
		consoleinput.InputRecord{EventType: consoleinput.KEY_EVENT},
		consoleinput.InputRecord{EventType: consoleinput.KEY_EVENT},
	}
	keydown := records[0].KeyEvent()
	keydown.KeyDown = 1
	keydown.UnicodeChar = uint16(c)

	keyup := records[1].KeyEvent()
	keyup.UnicodeChar = uint16(c)

	return handle.Write(records[:])
}

func  writeString(handle consoleinput.Handle,s string) {
	for _, c := range s {
		writeRune(handle,c)
	}
}

func main() {
	console, err := consoleinput.New()
	if err != nil {
		println(err.Error())
		return
	}
	for _, s := range os.Args[1:] {
		writeString(console,s)
		writeRune(console,'\r')
	}
}
