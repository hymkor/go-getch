package getch

import (
	"fmt"
	"os"
	"os/signal"
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

type resizeEvent struct {
	Width  uint
	Height uint
}

type Event struct {
	Key    *keyEvent
	Resize *resizeEvent
}

func ctrlCHandler(ch chan os.Signal) {
	for _ = range ch {
		eventBuffer = append(eventBuffer, Event{
			Key: &keyEvent{3, 0, LEFT_CTRL_PRESSED},
		})
	}
}

func DisableCtrlC() {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)
	go ctrlCHandler(ch)
}

var OnWindowResize func(w, h uint)

func getEvents(flag uintptr) []Event {
	var numberOfEventsRead uint32
	var events [10]inputRecordT
	var orgConMode uint32

	result := make([]Event, 0, 0)

	getConsoleMode.Call(uintptr(hConin),
		uintptr(unsafe.Pointer(&orgConMode)))
	setConsoleMode.Call(uintptr(hConin), flag)
	for len(result) <= 0 {
		readConsoleInput.Call(
			uintptr(hConin),
			uintptr(unsafe.Pointer(&events[0])),
			uintptr(len(events)),
			uintptr(unsafe.Pointer(&numberOfEventsRead)))
		for i := uint32(0); i < numberOfEventsRead; i++ {
			if events[i].eventType == KEY_EVENT && events[i].bKeyDown != 0 {
				result = append(result, Event{
					Key: &keyEvent{
						rune(events[i].unicodeChar),
						events[i].wVirtualKeyCode,
						events[i].dwControlKeyState,
					},
				})
			} else if events[i].eventType == WINDOW_BUFFER_SIZE_EVENT {
				width := uint(events[i].bKeyDown & 0xFFFF)
				height := uint((events[i].bKeyDown >> 16) & 0xFFFF)
				result = append(result, Event{
					Resize: &resizeEvent{
						Width:  width,
						Height: height,
					},
				})
			}
		}
	}
	setConsoleMode.Call(uintptr(hConin), uintptr(orgConMode))
	return result
}

var eventBuffer []Event
var eventBufferRead = 0

func getRawEvent(flag uintptr) Event {
	for eventBuffer == nil || eventBufferRead >= len(eventBuffer) {
		eventBuffer = getEvents(flag)
		eventBufferRead = 0
	}
	eventBufferRead++
	return eventBuffer[eventBufferRead-1]
}

var lastkey *keyEvent

// Get a event with concatinating a surrogate-pair of keyevents.
func getEvent(flag uintptr) Event {
	for {
		event1 := getRawEvent(flag)
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

// (deprecated) Get all keyboard-event.
func Full() (code rune, scan uint16, shift uint32) {
	var flag uintptr = 0
	if OnWindowResize != nil {
		flag = ENABLE_WINDOW_INPUT
	}
	for {
		event := getEvent(flag)
		if e := event.Resize; e != nil {
			OnWindowResize(e.Width, e.Height)
		}
		if e := event.Key; e != nil {
			return e.Rune, e.Scan, e.Shift
		}
	}
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
