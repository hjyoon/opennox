package opennox

import (
	"image"
	"math"
	"reflect"
	"testing"

	"github.com/opennox/libs/client/seat"
)

func e2eTestNearbyMonsterCandidate(wire uint16, distance float32) e2eNearestMonsterInputCandidate {
	return e2eNearestMonsterInputCandidate{
		identity: e2eMonsterInputIdentity{wire: wire, extent: uint32(wire) + 100, scriptID: int32(wire) + 1000},
		typeID:   "Bat", hostile: true, available: true,
		target: e2eNamedMonsterInputTarget{wire: wire, distance: distance, health: 5, point: image.Pt(550, 330), movePoint: image.Pt(650, 350)},
	}
}

func TestNearestMonsterInputSelectsVisibleHostileTypeAndPinsIdentity(t *testing.T) {
	valid := e2eTestNearbyMonsterCandidate(17, 60)
	wrongType := e2eTestNearbyMonsterCandidate(18, 1)
	wrongType.typeID = "SmallAlbinoSpider"
	friendly := e2eTestNearbyMonsterCandidate(19, 2)
	friendly.hostile = false
	invisible := e2eTestNearbyMonsterCandidate(20, 3)
	invisible.available = false
	dead := e2eTestNearbyMonsterCandidate(21, 4)
	dead.target.dead, dead.target.health = true, 0
	candidates := []e2eNearestMonsterInputCandidate{
		wrongType, friendly, invisible, dead,
		e2eTestNearbyMonsterCandidate(0, 1),
		e2eTestNearbyMonsterCandidate(22, float32(math.NaN())),
		e2eTestNearbyMonsterCandidate(23, float32(math.Inf(1))),
		e2eTestNearbyMonsterCandidate(24, -1),
		e2eTestNearbyMonsterCandidate(25, 513),
		e2eTestNearbyMonsterCandidate(26, 100), valid,
	}
	var selection e2eNearestMonsterInputSelection
	got, ok := selection.observe("Bat", candidates)
	if !ok || got != valid.target || selection.identity != valid.identity {
		t.Fatalf("wrong original creature selected: target=%+v identity=%+v", got, selection.identity)
	}
	closer := e2eTestNearbyMonsterCandidate(27, 1)
	valid.target.distance, valid.target.point = 20, image.Pt(525, 340)
	got, ok = selection.observe("Bat", []e2eNearestMonsterInputCandidate{closer, valid})
	if !ok || got != valid.target {
		t.Fatalf("a nearer newcomer replaced the pinned creature: %+v", got)
	}
	valid.hostile = false
	got, ok = selection.observe("Bat", []e2eNearestMonsterInputCandidate{valid, closer})
	if !ok || !got.absent || got.dead || selection.identity != valid.identity {
		t.Fatal("a newly friendly original target must stop the action without an attack or a death claim")
	}
	got, ok = selection.observe("Bat", []e2eNearestMonsterInputCandidate{closer})
	if !ok || !got.absent || got.dead {
		t.Fatal("retirement must be reported as absence, without transferring the click")
	}
	selection = e2eNearestMonsterInputSelection{}
	valid.hostile = true
	equal := e2eTestNearbyMonsterCandidate(16, 20)
	got, ok = selection.observe("Bat", []e2eNearestMonsterInputCandidate{valid, equal})
	if !ok || got.wire != 16 {
		t.Fatal("equal-distance selection must be stable when world iteration order changes")
	}
	selection = e2eNearestMonsterInputSelection{}
	got, ok = selection.observe("Bat", nil)
	if !ok || !got.absent || got.dead || selection.pinned {
		t.Fatal("an empty visible set must not fabricate a creature or an observed death")
	}
}

func TestNearestMonsterInputRetirementAndUnavailableFramesDoNotClickStaleTarget(t *testing.T) {
	var sc e2eScenario
	var selection e2eNearestMonsterInputSelection
	target := e2eTestNearbyMonsterCandidate(17, 100)
	candidates := []e2eNearestMonsterInputCandidate{target}
	var events []seat.InputEvent
	var result e2eNamedMonsterInputTarget
	sc.strikeNamedMonsterInput("original unnamed bat", func() (e2eNamedMonsterInputTarget, bool) {
		return selection.observe("Bat", candidates)
	}, func(evs ...seat.InputEvent) { events = append(events, evs...) }, func(v e2eNamedMonsterInputTarget) { result = v })
	gate := sc.steps[0]
	if gate.ready() {
		t.Fatal("a distant original bat must first receive ordinary movement")
	}
	candidates[0].available, candidates[0].target.hovered = false, true
	for i := 0; i < 3; i++ {
		if gate.ready() {
			t.Fatal("a stale hover must not replace an unavailable drawable")
		}
	}
	// The native code is reused, but both original identities have changed.
	candidates[0] = e2eTestNearbyMonsterCandidate(17, 10)
	candidates[0].identity.extent++
	candidates[0].identity.scriptID++
	candidates[0].target.hovered = true
	if !gate.ready() {
		t.Fatal("retirement should complete without attacking the replacement")
	}
	gate.fnc()
	sc.steps[1].fnc()
	want := []seat.InputEvent{
		&seat.MouseMoveEvent{Pos: target.target.movePoint, Relative: false},
		&seat.MouseButtonEvent{Button: seat.MouseButtonRight, Pressed: true},
		&seat.MouseButtonEvent{Button: seat.MouseButtonRight, Pressed: false},
	}
	if !reflect.DeepEqual(events, want) || !result.absent || result.dead {
		t.Fatalf("retired creature received input: events=%#v result=%+v", events, result)
	}

	// A separate action can select a new live identity, after normal native hover.
	sc, selection = e2eScenario{}, e2eNearestMonsterInputSelection{}
	events = nil
	candidates[0].target.hovered = false
	sc.strikeNamedMonsterInput("new original bat action", func() (e2eNamedMonsterInputTarget, bool) {
		return selection.observe("Bat", candidates)
	}, func(evs ...seat.InputEvent) { events = append(events, evs...) }, func(v e2eNamedMonsterInputTarget) { result = v })
	gate = sc.steps[0]
	if gate.ready() {
		t.Fatal("the new target must receive actual mouse positioning before a click")
	}
	candidates[0].target.point, candidates[0].target.hovered = image.Pt(530, 335), true
	if !gate.ready() {
		t.Fatal("current proximity and the exact normal native hover should allow a click")
	}
	gate.fnc()
	sc.steps[1].fnc()
	want = []seat.InputEvent{
		&seat.MouseMoveEvent{Pos: target.target.point, Relative: false},
		&seat.MouseButtonEvent{Button: seat.MouseButtonLeft, Pressed: true},
		&seat.MouseButtonEvent{Button: seat.MouseButtonLeft, Pressed: false},
	}
	if !reflect.DeepEqual(events, want) || result.wire != 17 || result.point != image.Pt(530, 335) || result.absent {
		t.Fatalf("new action did not use current real hover: events=%#v result=%+v", events, result)
	}
}

func TestMonsterInputProjectionKeepsGroundApproachForAirborneTarget(t *testing.T) {
	player, ground, offset := image.Pt(512, 384), image.Pt(612, 484), image.Pt(0, -8)
	aim, move := e2eMonsterMousePoints(player, ground, offset, 50, true)
	if aim != image.Pt(612, 426) || move != image.Pt(653, 525) {
		t.Fatalf("airborne creature must be aimed at its sprite while movement follows its ground point: aim=%v move=%v", aim, move)
	}
	aim2, move2 := e2eMonsterMousePoints(player, ground, offset, 150, true)
	if aim2 != image.Pt(612, 326) || move2 != move {
		t.Fatal("a changing sprite altitude must not divert ordinary movement from the creature's ground position")
	}
	aim, move = e2eMonsterMousePoints(player, ground, offset, 0, false)
	if aim != image.Pt(612, 476) || move != image.Pt(659, 519) {
		t.Fatalf("existing named-creature projection changed: aim=%v move=%v", aim, move)
	}
}
