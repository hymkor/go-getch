go-getch
=========

`go-getch` is a library to read the console-event 
(keyboard-hits or screen-resize),
for the programming language Go for Windows,

Example
-------

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

- `go-getch` supports the surrogate pair of Unicode.
- `go-getch` is used in Windows CUI Shell [NYAGOS](https://github.com/zetamatta/nyagos)

Types
-----

	type Event struct {
		Focus   *struct{}
		Key     *keyEvent // == KeyDown
		KeyDown *keyEvent
		KeyUp   *keyEvent
		Menu    *struct{}
		Mouse   *struct{}
		Resize  *resizeEvent
	}

	type keyEvent struct {
		Rune  rune
		Scan  uint16
		Shift uint32
	}

	type resizeEvent struct {
		Width  uint
		Height uint
	}

Functions
---------

### func All() Event

Get all keyboard events.

### func Rune() rune

Get a KeyDown event.

### func Within(msec uintptr) (Event, error)

Get all keyboard events with time-out.

### func RuneWithin(msec uintptr) (rune, error)

Get a KeyDown event with time-out.
