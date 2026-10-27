package opennox

import (
	"image"
	"reflect"
	"testing"

	"github.com/opennox/libs/client/seat"
	"github.com/opennox/opennox/v1/client/gui"
)

func TestNamedPickupInputRequiresExactDrawableAndActualPickupCursor(t *testing.T) {
	var sc e2eScenario
	var events []seat.InputEvent
	var clicks []e2eNamedUseInputTarget
	target := e2eNamedUseInputTarget{wire: 0x8011, point: image.Pt(480, 355), cursor: gui.CursorPickup}
	available := false
	sc.clickNamedObjectCursorInput("original quest book", gui.CursorPickup,
		func() (e2eNamedUseInputTarget, bool) { return target, available },
		func(evs ...seat.InputEvent) { events = append(events, evs...) },
		func(v e2eNamedUseInputTarget) { clicks = append(clicks, v) })
	gate := sc.steps[0]
	if gate.ready() || len(events) != 0 {
		t.Fatal("unavailable named item must not receive input")
	}
	available, target.hovered = true, true
	for _, cursor := range []gui.Cursor{gui.CursorUse, gui.CursorTalk, gui.CursorPickupFar} {
		target.cursor = cursor
		if gate.ready() || len(clicks) != 0 {
			t.Fatalf("cursor %d must not authorize a pickup", cursor)
		}
	}
	target.cursor, target.hovered, target.point = gui.CursorPickup, false, image.Pt(492, 350)
	if gate.ready() {
		t.Fatal("Pickup cursor over another drawable must not authorize this item")
	}
	target.hovered = true
	if !gate.ready() || len(clicks) != 0 {
		t.Fatal("exact live item with Pickup cursor must release the gate without clicking yet")
	}
	gate.fnc()
	sc.steps[1].fnc()
	if !reflect.DeepEqual(clicks, []e2eNamedUseInputTarget{target}) {
		t.Fatalf("pickup observation = %+v", clicks)
	}
	want := []seat.InputEvent{
		&seat.MouseButtonEvent{Button: seat.MouseButtonLeft, Pressed: true},
		&seat.MouseButtonEvent{Button: seat.MouseButtonLeft, Pressed: false},
	}
	if !reflect.DeepEqual(events[len(events)-2:], want) || len(sc.steps) != 2 || gate.waitTimeout != 600 || sc.steps[1].time-gate.time != 1 {
		t.Fatalf("pickup must use one ordinary press/release with a bounded cursor gate: %#v", events)
	}
}
