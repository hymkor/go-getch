package getch

import (
	"errors"
	"fmt"
	"strings"
	"syscall"
	"unicode/utf16"
	"unsafe"
)

var kernel32 = syscall.NewLazyDLL("kernel32")

const (
	RIGHT_ALT_PRESSED  = 1
	LEFT_ALT_PRESSED   = 2
	RIGHT_CTRL_PRESSED = 4
	LEFT_CTRL_PRESSED  = 8
	CTRL_PRESSED       = RIGHT_CTRL_PRESSED | LEFT_CTRL_PRESSED
	ALT_PRESSED        = RIGHT_ALT_PRESSED | LEFT_ALT_PRESSED
)

type inputRecordT struct {
	eventType uint16
	_         uint16
	// _KEY_EVENT_RECORD {
	bKeyDown         int32
	wRepeartCount    uint16
	wVirtualKeyCode  uint16
	wVirtualScanCode uint16
	unicodeChar      uint16
	// }
	dwControlKeyState uint32
}

var getConsoleMode = kernel32.NewProc("GetConsoleMode")
var setConsoleMode = kernel32.NewProc("SetConsoleMode")
var readConsoleInput = kernel32.NewProc("ReadConsoleInputW")
var getNumberOfConsoleInputEvents = kernel32.NewProc("GetNumberOfConsoleInputEvents")
var flushConsoleInputBuffer = kernel32.NewProc("FlushConsoleInputBuffer")
var waitForSingleObject = kernel32.NewProc("WaitForSingleObject")

var hConin syscall.Handle

func init() {
	var err error
	hConin, err = syscall.Open("CONIN$", syscall.O_RDWR, 0)
	if err != nil {
		panic(fmt.Sprintf("conio: %v", err))
	}
}

type keyEvent struct {
	Rune  rune
	Scan  uint16
	Shift uint32
}

func (k keyEvent) String() string {
	return fmt.Sprintf("Rune:%v,Scan=%d,Shift=%d", k.Rune, k.Scan, k.Shift)
}

type resizeEvent struct {
	Width  uint
	Height uint
}

func (r resizeEvent) String() string {
	return fmt.Sprintf("Width:%d,Height:%d", r.Width, r.Height)
}

type Event struct {
	Focus   *struct{}
	Key     *keyEvent // == KeyDown
	KeyDown *keyEvent
	KeyUp   *keyEvent
	Menu    *struct{}
	Mouse   *struct{}
	Resize  *resizeEvent
}

func (e Event) String() string {
	event := make([]string, 0, 7)
	if e.Focus != nil {
		event = append(event, "Focus")
	}
	if e.KeyDown != nil {
		event = append(event, "KeyDown("+e.KeyDown.String()+")")
	}
	if e.KeyUp != nil {
		event = append(event, "KeyUp("+e.KeyUp.String()+")")
	}
	if e.Menu != nil {
		event = append(event, "Menu")
	}
	if e.Mouse != nil {
		event = append(event, "Mouse")
	}
	if e.Resize != nil {
		event = append(event, "Resize("+e.Resize.String()+")")
	}
	if len(event) > 0 {
		return strings.Join(event, ",")
	} else {
		return "no events"
	}
}

func readEvents(flag uintptr) []Event {
	result := make([]Event, 0, 2)

	var orgConMode uint32
	getConsoleMode.Call(uintptr(hConin),
		uintptr(unsafe.Pointer(&orgConMode)))
	setConsoleMode.Call(uintptr(hConin), flag)
	for len(result) <= 0 {
		var events [10]inputRecordT
		var numberOfEventsRead uint32

		readConsoleInput.Call(
			uintptr(hConin),
			uintptr(unsafe.Pointer(&events[0])),
			uintptr(len(events)),
			uintptr(unsafe.Pointer(&numberOfEventsRead)))
		for i := uint32(0); i < numberOfEventsRead; i++ {
			e := events[i]
			var r Event
			switch e.eventType {
			case FOCUS_EVENT:
				r = Event{Focus: &struct{}{}}
			case KEY_EVENT:
				k := &keyEvent{
					Rune:  rune(e.unicodeChar),
					Scan:  e.wVirtualKeyCode,
					Shift: e.dwControlKeyState,
				}
				if e.bKeyDown != 0 {
					r = Event{Key: k, KeyDown: k}
				} else {
					r = Event{KeyUp: k}
				}
			case MENU_EVENT:
				r = Event{Menu: &struct{}{}}
			case MOUSE_EVENT:
				r = Event{Mouse: &struct{}{}}
			case WINDOW_BUFFER_SIZE_EVENT:
				r = Event{
					Resize: &resizeEvent{
						Width:  uint(e.bKeyDown & 0xFFFF),
						Height: uint((e.bKeyDown >> 16) & 0xFFFF),
					},
				}
			default:
				continue
			}
			result = append(result, r)
		}
	}
	setConsoleMode.Call(uintptr(hConin), uintptr(orgConMode))
	return result
}

var eventBuffer []Event
var eventBufferRead = 0

func bufReadEvent(flag uintptr) Event {
	for eventBuffer == nil || eventBufferRead >= len(eventBuffer) {
		eventBuffer = readEvents(flag)
		eventBufferRead = 0
	}
	eventBufferRead++
	return eventBuffer[eventBufferRead-1]
}

var lastkey *keyEvent

// Get a event with concatinating a surrogate-pair of keyevents.
func getEvent(flag uintptr) Event {
	for {
		event1 := bufReadEvent(flag)
		if k := event1.Key; k != nil {
			if lastkey != nil {
				k.Rune = utf16.DecodeRune(lastkey.Rune, k.Rune)
				lastkey = nil
			} else if utf16.IsSurrogate(k.Rune) {
				lastkey = k
				continue
			}
		}
		return event1
	}
}

// Get all console-event (keyboard,resize,...)
func All() Event {
	return getEvent(ENABLE_WINDOW_INPUT)
}

const IGNORE_RESIZE_EVENT uintptr = 0

// Get character as a Rune
func Rune() rune {
	for {
		e := getEvent(IGNORE_RESIZE_EVENT)
		if e.Key != nil && e.Key.Rune != 0 {
			return e.Key.Rune
		}
	}
}

func Count() (int, error) {
	var count uint32 = 0

	status, _, err := getNumberOfConsoleInputEvents.Call(uintptr(hConin),
		uintptr(unsafe.Pointer(&count)))
	if status != 0 {
		return int(count), nil
	} else {
		return 0, err
	}
}

func Flush() error {
	eventBuffer = nil
	status, _, err := flushConsoleInputBuffer.Call(uintptr(hConin))
	if status != 0 {
		return nil
	} else {
		return err
	}
}

// wait for keyboard event
func Wait(timeout_msec uintptr) (bool, error) {
	status, _, err := waitForSingleObject.Call(uintptr(hConin), timeout_msec)
	switch status {
	case WAIT_OBJECT_0:
		return true, nil
	case WAIT_TIMEOUT:
		return false, nil
	case WAIT_ABANDONED:
		return false, errors.New("WAIT_ABANDONED")
	default: // including WAIT_FAILED:
		if err != nil {
			return false, err
		} else {
			return false, errors.New("WAIT_FAILED")
		}
	}
}

func Within(msec uintptr) (Event, error) {
	if ok, err := Wait(msec); err != nil || !ok {
		return Event{}, err
	}
	return All(), nil
}

const NUL = '\000'

func RuneWithin(msec uintptr) (rune, error) {
	if ok, err := Wait(msec); err != nil || !ok {
		return NUL, err
	}
	e := getEvent(ENABLE_WINDOW_INPUT)
	if e.Key != nil {
		return e.Key.Rune, nil
	}
	return NUL, nil
}
