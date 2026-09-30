package opennox

import (
	"image"
	"testing"

	"github.com/opennox/libs/client/seat"
	"github.com/opennox/libs/types"
)

func TestE2EHeadlessPlayback(t *testing.T) {
	for _, tc := range []struct {
		name     string
		playback string
		backend  string
		want     bool
		wantErr  bool
	}{
		{name: "normal startup"},
		{name: "normal startup ignores headless", backend: "headless"},
		{name: "recording ignores invalid option", backend: "invalid"},
		{name: "default playback", playback: "scenario.yaml", want: true},
		{name: "default true playback", playback: "true", want: true},
		{name: "explicit headless", playback: "scenario.yaml", backend: "headless", want: true},
		{name: "explicit SDL", playback: "scenario.yaml", backend: "sdl"},
		{name: "unknown backend", playback: "scenario.yaml", backend: "invalid", wantErr: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := e2eHeadlessPlayback(tc.playback, tc.backend)
			if got != tc.want || (err != nil) != tc.wantErr {
				t.Fatalf("headless = %t, err = %v; want %t, error %t", got, err, tc.want, tc.wantErr)
			}
		})
	}
}

func TestE2EPlaybackInput(t *testing.T) {
	abs := &seat.MouseMoveEvent{Pos: image.Pt(240, 108)}
	rel := &seat.MouseMoveEvent{Relative: true, Rel: types.Pointf{X: 3, Y: 4}}
	key := &seat.KeyboardEvent{Pressed: true}
	toWindow := func(p image.Point) image.Point { return p.Mul(2).Add(image.Pt(98, 20)) }
	for _, tc := range []struct {
		name             string
		event            seat.InputEvent
		space            e2eInputSpace
		toWindow         func(image.Point) image.Point
		scenarioToWindow func(image.Point) image.Point
		want             image.Point
		mapped           bool
	}{
		{name: "SDL canvas mouse", event: abs, space: e2eCanvasInputSpace, toWindow: toWindow, want: image.Pt(578, 236), mapped: true},
		{name: "SDL scenario mouse", event: abs, space: e2eScenarioInputSpace, scenarioToWindow: toWindow, want: image.Pt(578, 236), mapped: true},
		{name: "headless canvas mouse", event: abs, space: e2eCanvasInputSpace},
		{name: "headless scenario mouse", event: abs, space: e2eScenarioInputSpace},
		{name: "recorded or real mouse", event: abs, toWindow: toWindow},
		{name: "relative mouse", event: rel, space: e2eCanvasInputSpace, toWindow: toWindow},
		{name: "keyboard", event: key, space: e2eCanvasInputSpace, toWindow: toWindow},
		{name: "window close", event: seat.WindowClosed, space: e2eCanvasInputSpace, toWindow: toWindow},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := e2ePlaybackInput(tc.event, tc.space, tc.toWindow, tc.scenarioToWindow)
			if !tc.mapped {
				if got != tc.event {
					t.Fatalf("non-canvas input was changed: %#v", got)
				}
				return
			}
			move := got.(*seat.MouseMoveEvent)
			if move == abs || move.Pos != tc.want || move.Relative {
				t.Fatalf("mapped mouse = %#v, want a copied absolute event at %v", move, tc.want)
			}
			if abs.Pos != image.Pt(240, 108) {
				t.Fatalf("original mouse input was mutated: %#v", abs)
			}
		})
	}
}

func TestE2EInputCoordinateSpaces(t *testing.T) {
	oldInput, oldToWindow, oldOnInput := e2e.input, e2e.inputToWindow, e2e.onInput
	oldScenarioToWindow := e2e.scenarioToWindow
	t.Cleanup(func() {
		e2e.input, e2e.inputToWindow, e2e.onInput = oldInput, oldToWindow, oldOnInput
		e2e.scenarioToWindow = oldScenarioToWindow
	})
	e2e.input = nil
	e2e.inputToWindow = func(p image.Point) image.Point { return p.Mul(2).Add(image.Pt(98, 20)) }
	e2e.scenarioToWindow = func(p image.Point) image.Point {
		return e2e.inputToWindow(e2eScenarioCanvasPos(p, image.Pt(640, 480)))
	}
	var got []seat.InputEvent
	e2e.onInput = []func(seat.InputEvent){func(ev seat.InputEvent) { got = append(got, ev) }}
	ev := &seat.MouseMoveEvent{Pos: image.Pt(240, 108)}
	e2eQueueInput(ev)
	e2eQueueRawInput(ev)
	var sc e2eScenario
	sc.Input(0, "", ev)
	sc.steps[0].fnc()
	e2eInputTick()
	if len(got) != 3 || len(e2e.input) != 0 {
		t.Fatalf("received %d events, %d remain queued", len(got), len(e2e.input))
	}
	if got[0].(*seat.MouseMoveEvent).Pos != image.Pt(578, 236) || got[1] != ev || got[2].(*seat.MouseMoveEvent).Pos != image.Pt(398, 154) {
		t.Fatalf("scripted and raw events lost their coordinate spaces: %#v", got)
	}
}

func TestE2EScenarioCanvasPos(t *testing.T) {
	p := image.Pt(448, 264)
	if got := e2eScenarioCanvasPos(p, image.Pt(640, 480)); got != image.Pt(280, 165) {
		t.Fatalf("reference menu position = %v, want (280,165)", got)
	}
	if got := e2eScenarioCanvasPos(p, image.Pt(1024, 768)); got != p {
		t.Fatalf("reference game position = %v, want %v", got, p)
	}
}
