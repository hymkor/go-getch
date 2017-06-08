package consoleinput

import (
	"fmt"
	"unsafe"
)

var readConsoleInput = kernel32.NewProc("ReadConsoleInputW")

type InputRecord struct {
	EventType uint16
	_         uint16
	Info      [8]uint16
}

func (handle Handle) Read(events []InputRecord) uint32 {
	var n uint32
	readConsoleInput.Call(
		uintptr(handle),
		uintptr(unsafe.Pointer(&events[0])),
		uintptr(len(events)),
		uintptr(unsafe.Pointer(&n)))
	return n
}

type KeyEventRecord struct {
	KeyDown         int32
	RepeartCount    uint16
	VirtualKeyCode  uint16
	VirtualScanCode uint16
	UnicodeChar     uint16
	ControlKeyState uint32
}

func (e *InputRecord) KeyEvent() *KeyEventRecord {
	return (*KeyEventRecord)(unsafe.Pointer(&e.Info[0]))
}

type MouseEventRecord struct {
	X          int16
	Y          int16
	Button     uint32
	ControlKey uint32
	Event      uint32
}

func (e *InputRecord) MouseEvent() *MouseEventRecord {
	return (*MouseEventRecord)(unsafe.Pointer(&e.Info[0]))
}

func (m MouseEventRecord) String() string {
	return fmt.Sprintf("X:%,Y:%d,Button:%d,ControlKey:%d,Event:%d",
		m.X, m.Y, m.Button, m.ControlKey, m.Event)
}

type windowBufferSizeRecord struct {
	X int16
	Y int16
}

func (e *InputRecord) ResizeEvent() (int16, int16) {
	p := (*windowBufferSizeRecord)(unsafe.Pointer(&e.Info[0]))
	return p.X, p.Y
}
