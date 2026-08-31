package server

import (
	"fmt"
	"math"
	"reflect"
	"testing"
)

type warcryProximityTestObject4FC4C0 struct {
	name string
	x    float32
	y    float32
}

type warcryProximityTestPlayer4FC4C0 struct {
	name  string
	unit  *warcryProximityTestObject4FC4C0
	class uint8
	next  *warcryProximityTestPlayer4FC4C0
}

type warcryProximityTestWorld4FC4C0 struct {
	first     *warcryProximityTestPlayer4FC4C0
	target    *warcryProximityTestObject4FC4C0
	active    map[*warcryProximityTestObject4FC4C0]int32
	mapResult map[*warcryProximityTestObject4FC4C0]int32
	events    []string
	faultAt   int
	after     map[string]func()
}

func warcryProximityObjectName4FC4C0(object *warcryProximityTestObject4FC4C0) string {
	if object == nil {
		return "nil"
	}
	return object.name
}

func warcryProximityPlayerName4FC4C0(player *warcryProximityTestPlayer4FC4C0) string {
	if player == nil {
		return "nil"
	}
	return player.name
}

func (w *warcryProximityTestWorld4FC4C0) record(event string) {
	w.events = append(w.events, event)
	if w.faultAt != 0 && len(w.events) == w.faultAt {
		panic(event)
	}
	if after := w.after[event]; after != nil {
		after()
	}
}

func (w *warcryProximityTestWorld4FC4C0) hooks() warcryProximityScanHooks4FC4C0[
	*warcryProximityTestPlayer4FC4C0,
	*warcryProximityTestObject4FC4C0,
] {
	return warcryProximityScanHooks4FC4C0[
		*warcryProximityTestPlayer4FC4C0,
		*warcryProximityTestObject4FC4C0,
	]{
		firstPlayer: func() *warcryProximityTestPlayer4FC4C0 {
			player := w.first
			w.record("first=" + warcryProximityPlayerName4FC4C0(player))
			return player
		},
		loadTargetArg: func() *warcryProximityTestObject4FC4C0 {
			target := w.target
			w.record("target=" + warcryProximityObjectName4FC4C0(target))
			return target
		},
		loadPlayerUnit: func(player *warcryProximityTestPlayer4FC4C0) *warcryProximityTestObject4FC4C0 {
			unit := player.unit
			w.record("unit:" + player.name + "=" + warcryProximityObjectName4FC4C0(unit))
			return unit
		},
		loadPlayerClass: func(player *warcryProximityTestPlayer4FC4C0) uint8 {
			class := player.class
			w.record(fmt.Sprintf("class:%s=%d", player.name, class))
			return class
		},
		isAbilityActive: func(unit *warcryProximityTestObject4FC4C0, ability Ability) int32 {
			result := w.active[unit]
			w.record(fmt.Sprintf("active:%s:%d=%d", warcryProximityObjectName4FC4C0(unit), ability, result))
			return result
		},
		loadPosX: func(object *warcryProximityTestObject4FC4C0) float32 {
			if object == nil {
				w.record("x:nil")
				panic("nil object X")
			}
			value := object.x
			w.record(fmt.Sprintf("x:%s=%08x", object.name, math.Float32bits(value)))
			return value
		},
		loadPosY: func(object *warcryProximityTestObject4FC4C0) float32 {
			if object == nil {
				w.record("y:nil")
				panic("nil object Y")
			}
			value := object.y
			w.record(fmt.Sprintf("y:%s=%08x", object.name, math.Float32bits(value)))
			return value
		},
		mapCheck: func(unit, target *warcryProximityTestObject4FC4C0) int32 {
			result := w.mapResult[unit]
			w.record(fmt.Sprintf(
				"map:%s:%s=%d",
				warcryProximityObjectName4FC4C0(unit),
				warcryProximityObjectName4FC4C0(target),
				result,
			))
			return result
		},
		nextPlayer: func(player *warcryProximityTestPlayer4FC4C0) *warcryProximityTestPlayer4FC4C0 {
			next := player.next
			w.record("next:" + player.name + "=" + warcryProximityPlayerName4FC4C0(next))
			return next
		},
	}
}

func TestWarcryProximityScan4FC4C0ExactTraversalAndFirstMatch(t *testing.T) {
	target := &warcryProximityTestObject4FC4C0{name: "target"}
	ignored := &warcryProximityTestObject4FC4C0{name: "ignored", x: 1}
	inactive := &warcryProximityTestObject4FC4C0{name: "inactive", x: 2}
	far := &warcryProximityTestObject4FC4C0{name: "far", x: 300}
	blocked := &warcryProximityTestObject4FC4C0{name: "blocked", x: 3, y: 4}
	match := &warcryProximityTestObject4FC4C0{name: "match", x: 6, y: 8}
	tailUnit := &warcryProximityTestObject4FC4C0{name: "tail", x: 1}

	tail := &warcryProximityTestPlayer4FC4C0{name: "tail", unit: tailUnit}
	matching := &warcryProximityTestPlayer4FC4C0{name: "matching", unit: match, next: tail}
	blockedPlayer := &warcryProximityTestPlayer4FC4C0{name: "blocked", unit: blocked, next: matching}
	farPlayer := &warcryProximityTestPlayer4FC4C0{name: "far", unit: far, next: blockedPlayer}
	inactivePlayer := &warcryProximityTestPlayer4FC4C0{name: "inactive", unit: inactive, next: farPlayer}
	nonWarrior := &warcryProximityTestPlayer4FC4C0{name: "non-warrior", unit: ignored, class: 1, next: inactivePlayer}
	nilUnit := &warcryProximityTestPlayer4FC4C0{name: "nil-unit", next: nonWarrior}

	w := warcryProximityTestWorld4FC4C0{
		first:  nilUnit,
		target: target,
		active: map[*warcryProximityTestObject4FC4C0]int32{
			inactive: 0,
			far:      1,
			blocked:  1,
			match:    math.MinInt32,
		},
		mapResult: map[*warcryProximityTestObject4FC4C0]int32{
			blocked: 0,
			match:   math.MinInt32,
		},
		after: make(map[string]func()),
	}

	if got := warcryProximityScan4FC4C0(w.hooks()); got != 1 {
		t.Fatalf("scan = %d, want canonical 1", got)
	}
	want := []string{
		"first=nil-unit", "target=target",
		"unit:nil-unit=nil", "next:nil-unit=non-warrior",
		"unit:non-warrior=ignored", "class:non-warrior=1", "next:non-warrior=inactive",
		"unit:inactive=inactive", "class:inactive=0", "active:inactive:2=0", "next:inactive=far",
		"unit:far=far", "class:far=0", "active:far:2=1", "unit:far=far",
		"x:target=00000000", "x:far=43960000", "y:target=00000000", "y:far=00000000", "next:far=blocked",
		"unit:blocked=blocked", "class:blocked=0", "active:blocked:2=1", "unit:blocked=blocked",
		"x:target=00000000", "x:blocked=40400000", "y:target=00000000", "y:blocked=40800000",
		"map:blocked:target=0", "next:blocked=matching",
		"unit:matching=match", "class:matching=0", "active:match:2=-2147483648", "unit:matching=match",
		"x:target=00000000", "x:match=40c00000", "y:target=00000000", "y:match=41000000",
		"map:match:target=-2147483648",
	}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %q, want %q", w.events, want)
	}
}

func TestWarcryProximityScan4FC4C0EmptyListDoesNotObserveTarget(t *testing.T) {
	w := warcryProximityTestWorld4FC4C0{
		target:    &warcryProximityTestObject4FC4C0{name: "target"},
		active:    make(map[*warcryProximityTestObject4FC4C0]int32),
		mapResult: make(map[*warcryProximityTestObject4FC4C0]int32),
		after:     make(map[string]func()),
	}
	if got := warcryProximityScan4FC4C0(w.hooks()); got != 0 {
		t.Fatalf("scan = %d, want canonical 0", got)
	}
	if want := []string{"first=nil"}; !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %q, want %q", w.events, want)
	}
}

func TestWarcryProximityScan4FC4C0CachesTargetAndReloadsUnitAfterAbility(t *testing.T) {
	target := &warcryProximityTestObject4FC4C0{name: "target", x: 10, y: 20}
	decoyTarget := &warcryProximityTestObject4FC4C0{name: "decoy-target", x: 1000, y: 1000}
	original := &warcryProximityTestObject4FC4C0{name: "original", x: 1000, y: 1000}
	replacement := &warcryProximityTestObject4FC4C0{name: "replacement", x: 13, y: 24}
	player := &warcryProximityTestPlayer4FC4C0{name: "player", unit: original}
	w := warcryProximityTestWorld4FC4C0{
		first:     player,
		target:    target,
		active:    map[*warcryProximityTestObject4FC4C0]int32{original: 1},
		mapResult: map[*warcryProximityTestObject4FC4C0]int32{replacement: 7},
		after:     make(map[string]func()),
	}
	w.after["active:original:2=1"] = func() {
		player.unit = replacement
		w.target = decoyTarget
	}

	if got := warcryProximityScan4FC4C0(w.hooks()); got != 1 {
		t.Fatalf("scan = %d, want canonical 1", got)
	}
	want := []string{
		"first=player", "target=target", "unit:player=original", "class:player=0", "active:original:2=1",
		"unit:player=replacement", "x:target=41200000", "x:replacement=41500000",
		"y:target=41a00000", "y:replacement=41c00000", "map:replacement:target=7",
	}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %q, want %q", w.events, want)
	}
}

func TestWarcryProximityScan4FC4C0DistanceAndUnorderedComparison(t *testing.T) {
	tests := []struct {
		name       string
		targetX    float32
		unitX      float32
		want       int32
		wantMapHit bool
	}{
		{name: "below", unitX: 299, want: 1, wantMapHit: true},
		{name: "equal-distance", unitX: 300},
		{name: "positive-infinity", unitX: float32(math.Inf(1))},
		{name: "nan", unitX: float32(math.NaN()), want: 1, wantMapHit: true},
		{name: "infinity-minus-infinity", targetX: float32(math.Inf(1)), unitX: float32(math.Inf(1)), want: 1, wantMapHit: true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			target := &warcryProximityTestObject4FC4C0{name: "target", x: tc.targetX}
			unit := &warcryProximityTestObject4FC4C0{name: "unit", x: tc.unitX}
			player := &warcryProximityTestPlayer4FC4C0{name: "player", unit: unit}
			w := warcryProximityTestWorld4FC4C0{
				first:     player,
				target:    target,
				active:    map[*warcryProximityTestObject4FC4C0]int32{unit: math.MinInt32},
				mapResult: map[*warcryProximityTestObject4FC4C0]int32{unit: math.MinInt32},
				after:     make(map[string]func()),
			}
			if got := warcryProximityScan4FC4C0(w.hooks()); got != tc.want {
				t.Fatalf("scan = %d, want %d", got, tc.want)
			}
			mapHit := false
			for _, event := range w.events {
				if event == "map:unit:target=-2147483648" {
					mapHit = true
				}
			}
			if mapHit != tc.wantMapHit {
				t.Fatalf("map callback observed = %v, want %v; events: %q", mapHit, tc.wantMapHit, w.events)
			}
		})
	}
}

func TestWarcryProximityScan4FC4C0ReloadedNilUnitFaultsAtPosition(t *testing.T) {
	unit := &warcryProximityTestObject4FC4C0{name: "unit"}
	player := &warcryProximityTestPlayer4FC4C0{name: "player", unit: unit}
	w := warcryProximityTestWorld4FC4C0{
		first:     player,
		target:    &warcryProximityTestObject4FC4C0{name: "target"},
		active:    map[*warcryProximityTestObject4FC4C0]int32{unit: 1},
		mapResult: make(map[*warcryProximityTestObject4FC4C0]int32),
		after:     make(map[string]func()),
	}
	w.after["active:unit:2=1"] = func() { player.unit = nil }
	defer func() {
		if recover() == nil {
			t.Fatal("reloaded nil unit did not fault")
		}
		want := []string{
			"first=player", "target=target", "unit:player=unit", "class:player=0", "active:unit:2=1",
			"unit:player=nil", "x:target=00000000", "x:nil",
		}
		if !reflect.DeepEqual(w.events, want) {
			t.Fatalf("events = %q, want %q", w.events, want)
		}
	}()
	warcryProximityScan4FC4C0(w.hooks())
}

func TestWarcryProximityScan4FC4C0EveryObservationFaultPrefix(t *testing.T) {
	want := []string{
		"first=first", "target=target",
		"unit:first=first-unit", "class:first=0", "active:first-unit:2=1", "unit:first=first-unit",
		"x:target=00000000", "x:first-unit=3f800000", "y:target=00000000", "y:first-unit=40000000",
		"map:first-unit:target=0", "next:first=second",
		"unit:second=second-unit", "class:second=0", "active:second-unit:2=1", "unit:second=second-unit",
		"x:target=00000000", "x:second-unit=40400000", "y:target=00000000", "y:second-unit=40800000",
		"map:second-unit:target=1",
	}
	for faultAt := 1; faultAt <= len(want); faultAt++ {
		t.Run(fmt.Sprintf("fault-%02d", faultAt), func(t *testing.T) {
			target := &warcryProximityTestObject4FC4C0{name: "target"}
			firstUnit := &warcryProximityTestObject4FC4C0{name: "first-unit", x: 1, y: 2}
			secondUnit := &warcryProximityTestObject4FC4C0{name: "second-unit", x: 3, y: 4}
			second := &warcryProximityTestPlayer4FC4C0{name: "second", unit: secondUnit}
			first := &warcryProximityTestPlayer4FC4C0{name: "first", unit: firstUnit, next: second}
			w := warcryProximityTestWorld4FC4C0{
				first:     first,
				target:    target,
				active:    map[*warcryProximityTestObject4FC4C0]int32{firstUnit: 1, secondUnit: 1},
				mapResult: map[*warcryProximityTestObject4FC4C0]int32{firstUnit: 0, secondUnit: 1},
				faultAt:   faultAt,
				after:     make(map[string]func()),
			}
			defer func() {
				if got := recover(); got != want[faultAt-1] {
					t.Fatalf("panic = %v, want %q", got, want[faultAt-1])
				}
				if prefix := want[:faultAt]; !reflect.DeepEqual(w.events, prefix) {
					t.Fatalf("events = %q, want %q", w.events, prefix)
				}
			}()
			warcryProximityScan4FC4C0(w.hooks())
		})
	}
}
