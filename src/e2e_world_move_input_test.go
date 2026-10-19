package opennox

import (
	"image"
	"reflect"
	"testing"

	"github.com/opennox/libs/client/seat"
)

func TestWorldMoveInputRequiresObservedArrivalAndReleasesBeforeClientSettles(t *testing.T) {
	var sc e2eScenario
	var events []seat.InputEvent
	var arrivals []e2eWorldMoveInputTarget
	current := e2eWorldMoveInputTarget{point: image.Pt(700, 400), serverDistance: 100, clientDistance: 105}
	sc.moveToWorldInput("original route", func() (e2eWorldMoveInputTarget, bool) { return current, true },
		func(v ...seat.InputEvent) { events = append(events, v...) },
		func(v e2eWorldMoveInputTarget) { arrivals = append(arrivals, v) })
	if len(sc.steps) != 2 || sc.steps[0].waitTimeout != 600 || sc.steps[1].time-sc.steps[0].time != 5 {
		t.Fatal("movement needs bounded observed arrival and ordinary release settling")
	}
	gate := sc.steps[0]
	for i := 0; i < 20; i++ {
		if gate.ready() {
			t.Fatal("elapsed input frames must not declare a blocked route reached")
		}
	}
	current.point = image.Pt(430, 270)
	if gate.ready() {
		t.Fatal("new camera/position observation still has not reached the goal")
	}
	current.serverDistance, current.clientDistance = 8, 40
	if gate.ready() {
		t.Fatal("server arrival must wait for the client's actual observation")
	}
	if gate.ready() {
		t.Fatal("waiting for the client must not resume movement")
	}
	current.clientDistance = 10
	if !gate.ready() {
		t.Fatal("both current positions are within arrival tolerance")
	}
	gate.fnc()
	if !reflect.DeepEqual(arrivals, []e2eWorldMoveInputTarget{current}) {
		t.Fatal("arrival must report the current copied observations")
	}
	var buttons []*seat.MouseButtonEvent
	for _, event := range events {
		if v, ok := event.(*seat.MouseButtonEvent); ok {
			buttons = append(buttons, v)
		}
	}
	want := []*seat.MouseButtonEvent{{Button: seat.MouseButtonRight, Pressed: true}, {Button: seat.MouseButtonRight, Pressed: false}}
	if !reflect.DeepEqual(buttons, want) {
		t.Fatalf("movement must be pressed once and released once before arrival: %#v", buttons)
	}
	if move, ok := events[len(events)-2].(*seat.MouseMoveEvent); !ok || move.Pos != image.Pt(430, 270) {
		t.Fatalf("last movement did not use the latest observed direction: %#v", events)
	}
}

func TestWorldMoveInputReleasesOnUnavailablePlayerAndDoesNotUseStaleArrival(t *testing.T) {
	var sc e2eScenario
	var events []seat.InputEvent
	current := e2eWorldMoveInputTarget{point: image.Pt(700, 380), serverDistance: 100, clientDistance: 100}
	available := false
	arrived := false
	sc.moveToWorldInput("original route", func() (e2eWorldMoveInputTarget, bool) { return current, available },
		func(v ...seat.InputEvent) { events = append(events, v...) }, func(e2eWorldMoveInputTarget) { arrived = true })
	gate := sc.steps[0]
	if gate.ready() || len(events) != 0 {
		t.Fatal("missing player must not receive movement")
	}
	available = true
	if gate.ready() {
		t.Fatal("distant target should start actual movement")
	}
	available = false
	current.serverDistance, current.clientDistance = 0, 0
	for i := 0; i < 3; i++ {
		if gate.ready() {
			t.Fatal("stale copied distances must not substitute a live player")
		}
	}
	if len(events) != 3 || arrived {
		t.Fatalf("disappearance must release the button exactly once: %#v", events)
	}
	available = true
	if !gate.ready() {
		t.Fatal("a fresh valid arrival should complete without another movement")
	}
	gate.fnc()
	if !arrived || len(events) != 3 {
		t.Fatal("arrival must not add another input")
	}
}
