package consoleoutput

import (
	"bytes"
	"fmt"
	"strings"
	"syscall"
	"unsafe"
)

var kernel32 = syscall.NewLazyDLL("kernel32")

type Handle syscall.Handle

func New() (Handle, error) {
	handle, err := syscall.Open("CONOUT$", syscall.O_RDWR, 0)
	if err != nil {
		return Handle(0), err
	}
	return Handle(handle), nil
}

func (handle Handle) Close() error {
	return syscall.Close(syscall.Handle(handle))
}

type CharInfoT struct {
	UnicodeChar uint16
	Attributes  uint16
}

type SmallRect struct {
	Left, Top, Right, Bottom int16
}

type Coord struct {
	X, Y int16
}

const (
	COMMON_LVB_LEADING_BYTE  = 0x0100
	COMMON_LVB_TRAILING_BYTE = 0x0200
)

var readConsoleOutput = kernel32.NewProc("ReadConsoleOutputW")

func (handle Handle) ReadConsoleOutput(buffer []CharInfoT, size Coord, coord Coord, read_region *SmallRect) error {

	sizeValue := *(*uintptr)(unsafe.Pointer(&size))
	coordValue := *(*uintptr)(unsafe.Pointer(&coord))

	status, _, err := readConsoleOutput.Call(
		uintptr(handle),
		uintptr(unsafe.Pointer(&buffer[0])),
		sizeValue,
		coordValue,
		uintptr(unsafe.Pointer(read_region)))
	if status == 0 {
		return err
	}
	return nil
}

type console_screen_buffer_info_t struct {
	Size              Coord
	CursorPosition    Coord
	Attributes        uint16
	Window            SmallRect
	MaximumWindowSize Coord
}

func (this *console_screen_buffer_info_t) Width() int {
	return int(this.Size.X)
}

func (this *console_screen_buffer_info_t) Height() int {
	return int(this.Size.Y)
}

func (this *console_screen_buffer_info_t) CursorX() int {
	return int(this.CursorPosition.X)
}

func (this *console_screen_buffer_info_t) CursorY() int {
	return int(this.CursorPosition.Y)
}

var getConsoleScreenBufferInfo = kernel32.NewProc("GetConsoleScreenBufferInfo")

func (handle Handle) GetScreenBufferInfo() (*console_screen_buffer_info_t, error) {
	var csbi console_screen_buffer_info_t
	status, _, err := getConsoleScreenBufferInfo.Call(
		uintptr(handle),
		uintptr(unsafe.Pointer(&csbi)))
	if status == 0 {
		return nil, fmt.Errorf("GetConsoleScreenBufferInfo: %s", err.Error())
	}
	return &csbi, nil
}

func (handle Handle) GetRecentOutput() (string, error) {
	screen, err := handle.GetScreenBufferInfo()
	if err != nil {
		return "", err
	}

	y := 0
	h := 1
	if screen.CursorY() >= 1 {
		y = screen.CursorY() - 1
		h++
	}

	region := &SmallRect{
		Left:   0,
		Top:    int16(y),
		Right:  screen.Size.X - 1,
		Bottom: int16(y + h - 1),
	}

	home := &Coord{X: 0, Y: 0}
	charinfo := make([]CharInfoT, screen.Width()*screen.Height())
	err = handle.ReadConsoleOutput(charinfo, screen.Size, *home, region)
	if err != nil {
		return "", err
	}

	var buffer bytes.Buffer
	for i := 0; i < screen.Height(); i++ {
		for j := 0; j < screen.Width(); j++ {
			p := &charinfo[i*screen.Width()+j]
			if (p.Attributes & COMMON_LVB_TRAILING_BYTE) != 0 {
				// right side of wide charactor

			} else if (p.Attributes & COMMON_LVB_LEADING_BYTE) != 0 {
				// left side of wide charactor
				if p.UnicodeChar != 0 {
					buffer.WriteRune(rune(p.UnicodeChar))
				}
			} else {
				// narrow charactor
				if p.UnicodeChar != 0 {
					buffer.WriteRune(rune(p.UnicodeChar & 0xFF))
				}
			}
		}
	}
	return strings.TrimSpace(buffer.String()), nil
}
