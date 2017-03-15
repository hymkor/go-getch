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
	info      [8]uint16
}

type keyEventRecord struct {
	bKeyDown          int32
	wRepeartCount     uint16
	wVirtualKeyCode   uint16
	wVirtualScanCode  uint16
	unicodeChar       uint16
	dwControlKeyState uint32
}

type windowBufferSizeRecord struct {
	X int16
	Y int16
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

type mouseEvent struct {
	X          int16
	Y          int16
	Button     uint32
	ControlKey uint32
	Event      uint32
}

const ( // Button
	FROM_LEFT_1ST_BUTTON_PRESSED = 0x0001
	FROM_LEFT_2ND_BUTTON_PRESSED = 0x0004
	FROM_LEFT_3RD_BUTTON_PRESSED = 0x0008
	FROM_LEFT_4TH_BUTTON_PRESSED = 0x0010
	RIGHTMOST_BUTTON_PRESSED     = 0x0002
)

func (m mouseEvent) String() string {
	return fmt.Sprintf("X:%,Y:%d,Button:%d,ControlKey:%d,Event:%d",
		m.X, m.Y, m.Button, m.ControlKey, m.Event)
}

type Event struct {
	Focus   *struct{} // MS says it should be ignored
	Key     *keyEvent // == KeyDown
	KeyDown *keyEvent
	KeyUp   *keyEvent
	Menu    *struct{}   // MS says it should be ignored
	Mouse   *mouseEvent // not supported,yet
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

func GetConsoleMode() uint32 {
	var mode uint32
	getConsoleMode.Call(uintptr(hConin), uintptr(unsafe.Pointer(&mode)))
	return mode
}

func SetConsoleMode(flag uint32) {
	setConsoleMode.Call(uintptr(hConin), uintptr(flag))
}

func readEvents(flag uint32) []Event {
	orgConMode := GetConsoleMode()
	SetConsoleMode(orgConMode)
	defer SetConsoleMode(orgConMode)

	result := make([]Event, 0, 2)

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
				p := (*keyEventRecord)(unsafe.Pointer(&e.info[0]))
				k := &keyEvent{
					Rune:  rune(p.unicodeChar),
					Scan:  p.wVirtualKeyCode,
					Shift: p.dwControlKeyState,
				}
				if p.bKeyDown != 0 {
					r = Event{Key: k, KeyDown: k}
				} else {
					r = Event{KeyUp: k}
				}
			case MENU_EVENT:
				r = Event{Menu: &struct{}{}}
			case MOUSE_EVENT:
				p := (*mouseEvent)(unsafe.Pointer(&e.info[0]))
				r = Event{
					Mouse: &mouseEvent{
						X:          p.X,
						Y:          p.Y,
						Button:     p.Button,
						ControlKey: p.ControlKey,
						Event:      p.Event,
					},
				}
			case WINDOW_BUFFER_SIZE_EVENT:
				p := (*windowBufferSizeRecord)(unsafe.Pointer(&e.info[0]))
				r = Event{
					Resize: &resizeEvent{
						Width:  uint(p.X),
						Height: uint(p.Y),
					},
				}
			default:
				continue
			}
			result = append(result, r)
		}
	}
	return result
}

var eventBuffer []Event
var eventBufferRead = 0

func bufReadEvent(flag uint32) Event {
	for eventBuffer == nil || eventBufferRead >= len(eventBuffer) {
		eventBuffer = readEvents(flag)
		eventBufferRead = 0
	}
	eventBufferRead++
	return eventBuffer[eventBufferRead-1]
}

var lastkey *keyEvent

// Get a event with concatinating a surrogate-pair of keyevents.
func getEvent(flag uint32) Event {
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

const ALL_EVENTS = ENABLE_WINDOW_INPUT | ENABLE_MOUSE_INPUT

// Get all console-event (keyboard,resize,...)
func All() Event {
	return getEvent(ALL_EVENTS)
}

const IGNORE_RESIZE_EVENT uint32 = 0

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
	org := GetConsoleMode()
	SetConsoleMode(ALL_EVENTS)
	defer SetConsoleMode(org)

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
	orgConMode := GetConsoleMode()
	SetConsoleMode(ALL_EVENTS)
	defer SetConsoleMode(orgConMode)

	if ok, err := Wait(msec); err != nil || !ok {
		return Event{}, err
	}
	return All(), nil
}

const NUL = '\000'

func RuneWithin(msec uintptr) (rune, error) {
	orgConMode := GetConsoleMode()
	SetConsoleMode(IGNORE_RESIZE_EVENT)
	defer SetConsoleMode(orgConMode)

	if ok, err := Wait(msec); err != nil || !ok {
		return NUL, err
	}
	e := getEvent(IGNORE_RESIZE_EVENT)
	if e.Key != nil {
		return e.Key.Rune, nil
	}
	return NUL, nil
}
