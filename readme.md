go-getch
=========

`go-getch` is a library to read the console-event 
(keyboard-hits or screen-resize),
for the programming language Go for Windows,

Example:

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
