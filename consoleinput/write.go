package consoleinput

import (
	"unsafe"
)

var writeConsoleInput = kernel32.NewProc("WriteConsoleInputW")

func (handle Handle) Write(events []InputRecord) uint32 {
	var count uint32
	writeConsoleInput.Call(uintptr(handle), uintptr(unsafe.Pointer(&events[0])), uintptr(len(events)), uintptr(unsafe.Pointer(&count)))

	return count
}
