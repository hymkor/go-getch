package getch

import (
	"errors"
	"fmt"
	"strings"
	"unicode/utf16"

	"github.com/zetamatta/go-getch/consoleinput"
)

const (
	RIGHT_ALT_PRESSED  = 1
	LEFT_ALT_PRESSED   = 2
	RIGHT_CTRL_PRESSED = 4
	LEFT_CTRL_PRESSED  = 8
	CTRL_PRESSED       = RIGHT_CTRL_PRESSED | LEFT_CTRL_PRESSED
	ALT_PRESSED        = RIGHT_ALT_PRESSED | LEFT_ALT_PRESSED
)


var hConin consoleinput.Handle

func init() {
	var err error
	hConin, err = consoleinput.New()
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

const ( // Button
	FROM_LEFT_1ST_BUTTON_PRESSED = 0x0001
	FROM_LEFT_2ND_BUTTON_PRESSED = 0x0004
	FROM_LEFT_3RD_BUTTON_PRESSED = 0x0008
	FROM_LEFT_4TH_BUTTON_PRESSED = 0x0010
	RIGHTMOST_BUTTON_PRESSED     = 0x0002
)

type Event struct {
	Focus   *struct{} // MS says it should be ignored
	Key     *keyEvent // == KeyDown
	KeyDown *keyEvent
	KeyUp   *keyEvent
	Menu    *struct{}                      // MS says it should be ignored
	Mouse   *consoleinput.MouseEventRecord // not supported,yet
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

func readEvents(flag uint32) []Event {
	orgConMode := hConin.GetConsoleMode()
	hConin.SetConsoleMode(flag)
	defer hConin.SetConsoleMode(orgConMode)

	result := make([]Event, 0, 2)

	for len(result) <= 0 {
		var events [10]consoleinput.InputRecord
		numberOfEventsRead := hConin.Read(events[:])

		for i := uint32(0); i < numberOfEventsRead; i++ {
			e := events[i]
			var r Event
			switch e.EventType {
			case consoleinput.FOCUS_EVENT:
				r = Event{Focus: &struct{}{}}
			case consoleinput.KEY_EVENT:
				p := e.KeyEvent()
				k := &keyEvent{
					Rune:  rune(p.UnicodeChar),
					Scan:  p.VirtualKeyCode,
					Shift: p.ControlKeyState,
				}
				if p.KeyDown != 0 {
					r = Event{Key: k, KeyDown: k}
				} else {
					r = Event{KeyUp: k}
				}
			case consoleinput.MENU_EVENT:
				r = Event{Menu: &struct{}{}}
			case consoleinput.MOUSE_EVENT:
				p := e.MouseEvent()
				r = Event{
					Mouse: &consoleinput.MouseEventRecord{
						X:          p.X,
						Y:          p.Y,
						Button:     p.Button,
						ControlKey: p.ControlKey,
						Event:      p.Event,
					},
				}
			case consoleinput.WINDOW_BUFFER_SIZE_EVENT:
				width,height := e.ResizeEvent()
				r = Event{
					Resize: &resizeEvent{
						Width:  uint(width),
						Height: uint(height),
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

const ALL_EVENTS = consoleinput.ENABLE_WINDOW_INPUT | consoleinput.ENABLE_MOUSE_INPUT

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
	return hConin.GetNumberOfEvent()
}

func Flush() error {
	org := hConin.GetConsoleMode()
	hConin.SetConsoleMode(ALL_EVENTS)
	defer hConin.SetConsoleMode(org)

	eventBuffer = nil
	return hConin.FlushConsoleInputBuffer()
}

// wait for keyboard event
func Wait(timeout_msec uintptr) (bool, error) {
	status, err := hConin.WaitForSingleObject(timeout_msec)
	switch status {
	case consoleinput.WAIT_OBJECT_0:
		return true, nil
	case consoleinput.WAIT_TIMEOUT:
		return false, nil
	case consoleinput.WAIT_ABANDONED:
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
	orgConMode := hConin.GetConsoleMode()
	hConin.SetConsoleMode(ALL_EVENTS)
	defer hConin.SetConsoleMode(orgConMode)

	if ok, err := Wait(msec); err != nil || !ok {
		return Event{}, err
	}
	return All(), nil
}

const NUL = '\000'

func RuneWithin(msec uintptr) (rune, error) {
	orgConMode := hConin.GetConsoleMode()
	hConin.SetConsoleMode(IGNORE_RESIZE_EVENT)
	defer hConin.SetConsoleMode(orgConMode)

	if ok, err := Wait(msec); err != nil || !ok {
		return NUL, err
	}
	e := getEvent(IGNORE_RESIZE_EVENT)
	if e.Key != nil {
		return e.Key.Rune, nil
	}
	return NUL, nil
}
