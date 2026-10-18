package opennox

import (
	"image"
	"reflect"
	"testing"

	"github.com/opennox/libs/client/seat"
)

func TestNamedNPCInputTracksObservedMouseBeforeSingleClick(t *testing.T) {
	var sc e2eScenario
	var events []seat.InputEvent
	var clicked []e2eNamedNPCInputTarget
	target := e2eNamedNPCInputTarget{wire: 17, point: image.Pt(445, 320)}
	available := false
	sc.interactNamedNPCInput("stock NPC", func() (e2eNamedNPCInputTarget, bool) {
		return target, available
	}, func(evs ...seat.InputEvent) {
		events = append(events, evs...)
	}, func(v e2eNamedNPCInputTarget) { clicked = append(clicked, v) })
	if len(sc.steps) != 2 || sc.steps[0].ready == nil || sc.steps[0].waitTimeout != 600 || sc.steps[1].time-sc.steps[0].time != 1 {
		t.Fatal("real NPC input must have one bounded hover gate and a one-tick release")
	}
	gate := sc.steps[0]
	if gate.ready() || len(events) != 0 {
		t.Fatal("missing/offscreen NPC must not receive a click or substituted mouse target")
	}
	available = true
	if gate.ready() {
		t.Fatal("unselected NPC must wait for the actual client cursor")
	}
	// A naturally moving NPC or changed viewport must be measured again;
	// the old pixel must not be used to supply the interaction.
	target.point = image.Pt(500, 315)
	if gate.ready() {
		t.Fatal("a new screen position alone must not supply hover confirmation")
	}
	target.hovered = true
	if !gate.ready() {
		t.Fatal("actual cursor selection did not release the input gate")
	}
	gate.fnc()
	sc.steps[1].fnc()
	want := []seat.InputEvent{
		&seat.MouseMoveEvent{Pos: image.Pt(445, 320), Relative: false},
		&seat.MouseMoveEvent{Pos: image.Pt(500, 315), Relative: false},
		&seat.MouseButtonEvent{Button: seat.MouseButtonLeft, Pressed: true},
		&seat.MouseButtonEvent{Button: seat.MouseButtonLeft, Pressed: false},
	}
	if !reflect.DeepEqual(events, want) || !reflect.DeepEqual(clicked, []e2eNamedNPCInputTarget{target}) {
		t.Fatalf("input events=%#v observations=%+v", events, clicked)
	}
}

func TestNamedNPCInputDoesNotClickWhenTargetDisappears(t *testing.T) {
	var sc e2eScenario
	var events []seat.InputEvent
	target := e2eNamedNPCInputTarget{wire: 22, point: image.Pt(550, 335)}
	available := true
	sc.interactNamedNPCInput("stock NPC", func() (e2eNamedNPCInputTarget, bool) {
		return target, available
	}, func(evs ...seat.InputEvent) { events = append(events, evs...) }, func(e2eNamedNPCInputTarget) {
		t.Fatal("vanished target received a substituted click")
	})
	if sc.steps[0].ready() {
		t.Fatal("unselected target released input gate")
	}
	available = false
	target.hovered = true // stale cursor state must not replace a live lookup.
	for i := 0; i < 3; i++ {
		if sc.steps[0].ready() {
			t.Fatal("stale cursor released a vanished target")
		}
	}
	if len(events) != 1 {
		t.Fatalf("disappeared target queued extra input: %#v", events)
	}
	if _, ok := events[0].(*seat.MouseMoveEvent); !ok {
		t.Fatalf("first event is not ordinary mouse motion: %#v", events)
	}
}
