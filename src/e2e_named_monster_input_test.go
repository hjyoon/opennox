package opennox

import (
	"image"
	"reflect"
	"testing"

	"github.com/opennox/libs/client/seat"
)

func TestNamedMonsterInputUsesActualMovementThenCurrentCursorBeforeSword(t *testing.T) {
	var sc e2eScenario
	var events []seat.InputEvent
	var clicks []e2eNamedMonsterInputTarget
	target := e2eNamedMonsterInputTarget{wire: 17, point: image.Pt(610, 360), movePoint: image.Pt(710, 335), distance: 100}
	available := false
	sc.strikeNamedMonsterInput("original creature", func() (e2eNamedMonsterInputTarget, bool) {
		return target, available
	}, func(ev ...seat.InputEvent) { events = append(events, ev...) }, func(v e2eNamedMonsterInputTarget) { clicks = append(clicks, v) })
	if len(sc.steps) != 2 || sc.steps[0].ready == nil || sc.steps[0].waitTimeout != 600 || sc.steps[1].time-sc.steps[0].time != 1 {
		t.Fatal("actual input needs one bounded approach/cursor gate and a one-tick release")
	}
	gate := sc.steps[0]
	if gate.ready() || len(events) != 0 {
		t.Fatal("missing target must not receive a movement or attack")
	}
	available = true
	if gate.ready() {
		t.Fatal("distant creature must be approached through actual right-button movement")
	}
	target.movePoint = image.Pt(690, 300)
	if gate.ready() {
		t.Fatal("changed drawable/camera must be observed again")
	}
	target.distance, target.point = 25, image.Pt(535, 370)
	if gate.ready() {
		t.Fatal("distance alone must not substitute the client's actual cursor target")
	}
	target.hovered = true
	if !gate.ready() {
		t.Fatal("observed proximity and actual cursor should permit one sword input")
	}
	gate.fnc()
	sc.steps[1].fnc()
	want := []seat.InputEvent{
		&seat.MouseMoveEvent{Pos: image.Pt(710, 335), Relative: false},
		&seat.MouseButtonEvent{Button: seat.MouseButtonRight, Pressed: true},
		&seat.MouseMoveEvent{Pos: image.Pt(690, 300), Relative: false},
		&seat.MouseButtonEvent{Button: seat.MouseButtonRight, Pressed: false},
		&seat.MouseMoveEvent{Pos: image.Pt(535, 370), Relative: false},
		&seat.MouseButtonEvent{Button: seat.MouseButtonLeft, Pressed: true},
		&seat.MouseButtonEvent{Button: seat.MouseButtonLeft, Pressed: false},
	}
	if !reflect.DeepEqual(events, want) || !reflect.DeepEqual(clicks, []e2eNamedMonsterInputTarget{target}) {
		t.Fatalf("actual events=%#v observations=%+v", events, clicks)
	}
}

func TestNamedMonsterInputReleasesMovementOnDisappearanceAndDoesNotAttackDeadTarget(t *testing.T) {
	var sc e2eScenario
	var events []seat.InputEvent
	target := e2eNamedMonsterInputTarget{wire: 22, movePoint: image.Pt(700, 380), distance: 100}
	available := true
	var deadObserved bool
	sc.strikeNamedMonsterInput("original creature", func() (e2eNamedMonsterInputTarget, bool) {
		return target, available
	}, func(ev ...seat.InputEvent) { events = append(events, ev...) }, func(v e2eNamedMonsterInputTarget) { deadObserved = v.dead })
	gate := sc.steps[0]
	if gate.ready() {
		t.Fatal("distant target unexpectedly released input gate")
	}
	available = false
	target.hovered = true
	for i := 0; i < 3; i++ {
		if gate.ready() {
			t.Fatal("stale cursor must not replace a live target")
		}
	}
	if len(events) != 3 {
		t.Fatalf("vanished target must release the actual right button once: %#v", events)
	}
	if ev, ok := events[2].(*seat.MouseButtonEvent); !ok || ev.Button != seat.MouseButtonRight || ev.Pressed {
		t.Fatalf("missing movement release: %#v", events[2])
	}
	available, target.dead = true, true
	if !gate.ready() {
		t.Fatal("original target's observed death should finish without another attack")
	}
	gate.fnc()
	sc.steps[1].fnc()
	if !deadObserved || len(events) != 3 {
		t.Fatalf("dead target received extra input: dead=%t events=%#v", deadObserved, events)
	}
}
