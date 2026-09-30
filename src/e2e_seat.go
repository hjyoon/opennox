package opennox

import (
	"fmt"
	"image"

	"github.com/opennox/libs/client/seat"
)

// Playback defaults to a deterministic headless seat. An explicit SDL seat
// exercises the production window, OpenGL presentation, and resize path while
// retaining the same scenario assertions. The option has no effect outside
// playback, including interactive E2E recording.
func e2eHeadlessPlayback(playback, backend string) (bool, error) {
	if playback == "" {
		return false, nil
	}
	switch backend {
	case "", "headless":
		return true, nil
	case "sdl":
		return false, nil
	default:
		return false, fmt.Errorf("NOX_E2E_SEAT must be headless or sdl (got %q)", backend)
	}
}

type e2eInputSpace uint8

const (
	e2eRawInputSpace e2eInputSpace = iota
	e2eCanvasInputSpace
	e2eScenarioInputSpace
)

type e2eQueuedInput struct {
	event seat.InputEvent
	space e2eInputSpace
}

// Existing scenario coordinates refer to the 1024x768 headless screen, not
// the canvas (menus use 640x480). Match the input handler's float32 scaling.
func e2eScenarioCanvasPos(p, canvas image.Point) image.Point {
	return image.Pt(
		int(float32(p.X)*(float32(canvas.X)/1024)),
		int(float32(p.Y)*(float32(canvas.Y)/768)),
	)
}

// Runtime helper clicks use canvas pixels, fixed scenarios use the reference
// screen, and recorded/real events already use logical window coordinates.
func e2ePlaybackInput(ev seat.InputEvent, space e2eInputSpace, canvasToWindow, scenarioToWindow func(image.Point) image.Point) seat.InputEvent {
	var toWindow func(image.Point) image.Point
	switch space {
	case e2eCanvasInputSpace:
		toWindow = canvasToWindow
	case e2eScenarioInputSpace:
		toWindow = scenarioToWindow
	}
	move, ok := ev.(*seat.MouseMoveEvent)
	if toWindow == nil || !ok || move == nil || move.Relative {
		return ev
	}
	out := *move
	out.Pos = toWindow(move.Pos)
	return &out
}

func e2eWrapSeat(s seat.Seat) seat.Seat {
	e2e.real = s
	s.OnInput(e2eRealInput)
	return &e2eSeat{s: s}
}

type e2eSeat struct {
	s seat.Seat
}

func (e *e2eSeat) ReplaceInputs(cfg seat.InputConfig) seat.InputConfig {
	return e.s.ReplaceInputs(cfg)
}

func (e *e2eSeat) ScreenSize() image.Point {
	return e.s.ScreenSize()
}

func (e *e2eSeat) ScreenMaxSize() image.Point {
	return image.Point{X: 1024, Y: 768}
}

func (e *e2eSeat) ResizeScreen(sz image.Point) {
	e.s.ResizeScreen(sz)
}

func (e *e2eSeat) SetScreenMode(mode seat.ScreenMode) {
	e.s.SetScreenMode(mode)
}

func (e *e2eSeat) SetGamma(v float32) {
	e.s.SetGamma(v)
}

func (e *e2eSeat) OnScreenResize(fnc func(sz image.Point)) {
	e.s.OnScreenResize(fnc)
}

func (e *e2eSeat) NewSurface(sz image.Point, filter bool) seat.Surface {
	return e.s.NewSurface(sz, filter)
}

func (e *e2eSeat) Clear() {
	e.s.Clear()
}

func (e *e2eSeat) Present() {
	e.s.Present()
}

func (e *e2eSeat) InputTick() {
	e.s.InputTick()
	e2eInputTick()
}

func (e *e2eSeat) OnInput(fnc func(ev seat.InputEvent)) {
	e2e.onInput = append(e2e.onInput, fnc)
}

func (e *e2eSeat) SetTextInput(enable bool) {
	if e2e.realEnable {
		e.s.SetTextInput(enable)
	}
}

func (e *e2eSeat) Close() error {
	return e.s.Close()
}
