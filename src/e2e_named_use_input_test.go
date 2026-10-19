package opennox

import (
	"image"
	"reflect"
	"testing"

	"github.com/opennox/libs/client/seat"
	"github.com/opennox/opennox/v1/client/gui"
)

func TestNamedUseInputRequiresExactDrawableAndActualUseCursor(t *testing.T) {
	var sc e2eScenario
	var events []seat.InputEvent
	var clicks []e2eNamedUseInputTarget
	target := e2eNamedUseInputTarget{wire: 0x8011, point: image.Pt(480, 355), cursor: gui.CursorUse}
	available := false
	sc.useNamedObjectInput("original entrance switch", func() (e2eNamedUseInputTarget, bool) {
		return target, available
	}, func(evs ...seat.InputEvent) { events = append(events, evs...) }, func(v e2eNamedUseInputTarget) { clicks = append(clicks, v) })
	if len(sc.steps) != 2 || sc.steps[0].ready == nil || sc.steps[0].waitTimeout != 600 || sc.steps[1].time-sc.steps[0].time != 1 {
		t.Fatal("named use needs a bounded cursor gate and one-tick ordinary release")
	}
	gate := sc.steps[0]
	if gate.ready() || len(events) != 0 {
		t.Fatal("an unavailable object must not receive input")
	}
	available = true
	if gate.ready() {
		t.Fatal("Use cursor over another drawable must not select the named object")
	}
	target.hovered, target.cursor = true, gui.CursorTalk
	if gate.ready() {
		t.Fatal("the right drawable with a different cursor must not receive a use click")
	}
	target.point, target.hovered, target.cursor = image.Pt(492, 350), false, gui.CursorUse
	if gate.ready() {
		t.Fatal("a changed viewport point must be measured and normally selected again")
	}
	target.hovered = true
	if !gate.ready() {
		t.Fatal("exact drawable and real Use cursor did not release the input gate")
	}
	gate.fnc()
	sc.steps[1].fnc()
	want := []seat.InputEvent{
		&seat.MouseMoveEvent{Pos: image.Pt(480, 355), Relative: false},
		&seat.MouseMoveEvent{Pos: image.Pt(480, 355), Relative: false},
		&seat.MouseMoveEvent{Pos: image.Pt(492, 350), Relative: false},
		&seat.MouseButtonEvent{Button: seat.MouseButtonLeft, Pressed: true},
		&seat.MouseButtonEvent{Button: seat.MouseButtonLeft, Pressed: false},
	}
	if !reflect.DeepEqual(events, want) || !reflect.DeepEqual(clicks, []e2eNamedUseInputTarget{target}) {
		t.Fatalf("actual input=%#v observations=%+v", events, clicks)
	}
}

func TestNamedUseInputRejectsRetiredAndInvalidIdentityBeforeRetargeting(t *testing.T) {
	var sc e2eScenario
	var events []seat.InputEvent
	var clicks []e2eNamedUseInputTarget
	target := e2eNamedUseInputTarget{wire: 0x8022, point: image.Pt(470, 350)}
	available := true
	sc.useNamedObjectInput("original switch", func() (e2eNamedUseInputTarget, bool) {
		return target, available
	}, func(evs ...seat.InputEvent) { events = append(events, evs...) }, func(v e2eNamedUseInputTarget) { clicks = append(clicks, v) })
	gate := sc.steps[0]
	if gate.ready() {
		t.Fatal("unselected object unexpectedly received a use click")
	}
	available, target.hovered, target.cursor = false, true, gui.CursorUse
	for i := 0; i < 3; i++ {
		if gate.ready() {
			t.Fatal("stale cursor state must not replace a live object lookup")
		}
	}
	available, target.wire = true, 0
	if gate.ready() || len(events) != 1 || len(clicks) != 0 {
		t.Fatal("missing/zero identity must not supply input or hover confirmation")
	}
	target.wire, target.point, target.hovered = 0x8044, image.Pt(510, 340), false
	if gate.ready() {
		t.Fatal("replacement drawable must receive actual mouse selection first")
	}
	target.hovered = true
	if !gate.ready() {
		t.Fatal("new live identity with its own Use cursor did not release the gate")
	}
	gate.fnc()
	sc.steps[1].fnc()
	if len(events) != 4 || !reflect.DeepEqual(clicks, []e2eNamedUseInputTarget{target}) {
		t.Fatalf("retired/replacement target input=%#v observations=%+v", events, clicks)
	}
	if move, ok := events[1].(*seat.MouseMoveEvent); !ok || move.Pos != image.Pt(510, 340) {
		t.Fatalf("replacement drawable did not receive current mouse motion: %#v", events[1])
	}
}
