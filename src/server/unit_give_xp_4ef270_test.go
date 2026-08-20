package server

import (
	"fmt"
	"math"
	"reflect"
	"testing"
)

type unitGiveXPTestPlayer4EF270 struct {
	name  string
	token uint32
}

type unitGiveXPTestUpdate4EF270 struct {
	name   string
	player *unitGiveXPTestPlayer4EF270
}

type unitGiveXPTestObject4EF270 struct {
	name       string
	experience float32
	update     *unitGiveXPTestUpdate4EF270
}

type unitGiveXPTestWorld4EF270 struct {
	unit   *unitGiveXPTestObject4EF270
	target float32
	scale  float32
	one    float32
	zero   float32

	events          []string
	after           map[string]func()
	experienceLoads int
	targetLoads     int
	protectedToken  uint32
	protectedAward  float32
}

func unitGiveXPObjectName4EF270(unit *unitGiveXPTestObject4EF270) string {
	if unit == nil {
		return "nil"
	}
	return unit.name
}

func unitGiveXPUpdateName4EF270(update *unitGiveXPTestUpdate4EF270) string {
	if update == nil {
		return "nil"
	}
	return update.name
}

func unitGiveXPPlayerName4EF270(player *unitGiveXPTestPlayer4EF270) string {
	if player == nil {
		return "nil"
	}
	return player.name
}

func (w *unitGiveXPTestWorld4EF270) record(event string) {
	w.events = append(w.events, event)
	if after := w.after[event]; after != nil {
		delete(w.after, event)
		after()
	}
}

func (w *unitGiveXPTestWorld4EF270) hooks() unitGiveXPHooks4EF270[
	*unitGiveXPTestObject4EF270,
	*unitGiveXPTestUpdate4EF270,
	*unitGiveXPTestPlayer4EF270,
] {
	return unitGiveXPHooks4EF270[
		*unitGiveXPTestObject4EF270,
		*unitGiveXPTestUpdate4EF270,
		*unitGiveXPTestPlayer4EF270,
	]{
		loadUnitArg: func() *unitGiveXPTestObject4EF270 {
			unit := w.unit
			w.record("arg:" + unitGiveXPObjectName4EF270(unit))
			return unit
		},
		loadExperience: func(unit *unitGiveXPTestObject4EF270) float32 {
			w.experienceLoads++
			name := unitGiveXPObjectName4EF270(unit)
			if unit == nil {
				event := fmt.Sprintf("experience:%d:%s", w.experienceLoads, name)
				w.record(event)
				panic(event)
			}
			value := unit.experience
			w.record(fmt.Sprintf("experience:%d:%s=%08x", w.experienceLoads, name, math.Float32bits(value)))
			return value
		},
		loadTargetArg: func() float32 {
			w.targetLoads++
			value := w.target
			w.record(fmt.Sprintf("target:%d=%08x", w.targetLoads, math.Float32bits(value)))
			return value
		},
		loadUpdateData: func(unit *unitGiveXPTestObject4EF270) *unitGiveXPTestUpdate4EF270 {
			name := unitGiveXPObjectName4EF270(unit)
			if unit == nil {
				event := "update:" + name
				w.record(event)
				panic(event)
			}
			update := unit.update
			w.record("update:" + name + "=" + unitGiveXPUpdateName4EF270(update))
			return update
		},
		loadScale: func() float32 {
			value := w.scale
			w.record(fmt.Sprintf("scale=%08x", math.Float32bits(value)))
			return value
		},
		loadOne: func() float32 {
			value := w.one
			w.record(fmt.Sprintf("one=%08x", math.Float32bits(value)))
			return value
		},
		loadZero: func() float32 {
			value := w.zero
			w.record(fmt.Sprintf("zero=%08x", math.Float32bits(value)))
			return value
		},
		storeExperience: func(unit *unitGiveXPTestObject4EF270, value float32) {
			name := unitGiveXPObjectName4EF270(unit)
			event := fmt.Sprintf("store:%s=%08x", name, math.Float32bits(value))
			w.record(event)
			if unit == nil {
				panic(event)
			}
			unit.experience = value
		},
		loadPlayer: func(update *unitGiveXPTestUpdate4EF270) *unitGiveXPTestPlayer4EF270 {
			name := unitGiveXPUpdateName4EF270(update)
			if update == nil {
				event := "player:" + name
				w.record(event)
				panic(event)
			}
			player := update.player
			w.record("player:" + name + "=" + unitGiveXPPlayerName4EF270(player))
			return player
		},
		loadExperienceToken: func(player *unitGiveXPTestPlayer4EF270) uint32 {
			name := unitGiveXPPlayerName4EF270(player)
			if player == nil {
				event := "token:" + name
				w.record(event)
				panic(event)
			}
			token := player.token
			w.record(fmt.Sprintf("token:%s=%08x", name, token))
			return token
		},
		protectExperience: func(token uint32, award float32) {
			w.record(fmt.Sprintf("protect:%08x:%08x", token, math.Float32bits(award)))
			w.protectedToken, w.protectedAward = token, award
		},
		reportExperience: func(unit *unitGiveXPTestObject4EF270) {
			w.record("report:" + unitGiveXPObjectName4EF270(unit))
		},
		syncLevel: func(unit *unitGiveXPTestObject4EF270) {
			w.record("sync:" + unitGiveXPObjectName4EF270(unit))
		},
	}
}

func newUnitGiveXPTestWorld4EF270() *unitGiveXPTestWorld4EF270 {
	player := &unitGiveXPTestPlayer4EF270{name: "player", token: 0x12345678}
	update := &unitGiveXPTestUpdate4EF270{name: "update", player: player}
	return &unitGiveXPTestWorld4EF270{
		unit:   &unitGiveXPTestObject4EF270{name: "unit", experience: 100, update: update},
		target: 200,
		scale:  math.Float32frombits(unitGiveXPScaleBits4EF270),
		one:    math.Float32frombits(unitGiveXPOneBits4EF270),
		zero:   math.Float32frombits(unitGiveXPZeroBits4EF270),
		after:  make(map[string]func()),
	}
}

func unitGiveXPMustPanic4EF270(t *testing.T, run func()) {
	t.Helper()
	panicked := false
	func() {
		defer func() { panicked = recover() != nil }()
		run()
	}()
	if !panicked {
		t.Fatal("call did not panic")
	}
}

func TestUnitGiveXP4EF270OrderedEarlyReturnIsPositiveZero(t *testing.T) {
	tests := []struct {
		name       string
		experience float32
		target     float32
	}{
		{name: "equal", experience: 100, target: 100},
		{name: "greater", experience: 101, target: 100},
		{name: "positive infinity", experience: float32(math.Inf(1)), target: math.MaxFloat32},
		{name: "signed zero equal", experience: math.Float32frombits(0x80000000), target: 0},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			w := newUnitGiveXPTestWorld4EF270()
			w.unit.experience, w.target = test.experience, test.target
			got := unitGiveXP4EF270(w.hooks())
			if math.Float64bits(got) != 0 {
				t.Fatalf("result bits = %016x, want positive zero", math.Float64bits(got))
			}
			want := []string{
				"arg:unit",
				fmt.Sprintf("experience:1:unit=%08x", math.Float32bits(test.experience)),
				fmt.Sprintf("target:1=%08x", math.Float32bits(test.target)),
				"zero=00000000",
			}
			if !reflect.DeepEqual(w.events, want) {
				t.Fatalf("events = %q, want %q", w.events, want)
			}
		})
	}
}

func TestUnitGiveXP4EF270ActivePathOrderCachingAndSavedAward(t *testing.T) {
	w := newUnitGiveXPTestWorld4EF270()
	entryUpdate := w.unit.update
	replacement := &unitGiveXPTestUpdate4EF270{
		name: "replacement", player: &unitGiveXPTestPlayer4EF270{name: "replacement-player", token: 9},
	}
	w.after["update:unit=update"] = func() { w.unit.update = replacement }
	w.after["protect:12345678:40000000"] = func() { w.unit.experience = 999 }

	got := unitGiveXP4EF270(w.hooks())
	if math.Float64bits(got) != math.Float64bits(2) {
		t.Fatalf("award bits = %016x, want 2.0", math.Float64bits(got))
	}
	if w.protectedToken != entryUpdate.player.token || math.Float32bits(w.protectedAward) != 0x40000000 {
		t.Fatalf("protected token/award = %08x/%08x", w.protectedToken, math.Float32bits(w.protectedAward))
	}
	want := []string{
		"arg:unit",
		"experience:1:unit=42c80000",
		"target:1=43480000",
		"target:2=43480000",
		"experience:2:unit=42c80000",
		"update:unit=update",
		"scale=3c23d70a",
		"one=3f800000",
		"experience:3:unit=42c80000",
		"store:unit=42cc0000",
		"player:update=player",
		"token:player=12345678",
		"protect:12345678:40000000",
		"report:unit",
		"sync:unit",
	}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %q, want %q", w.events, want)
	}
	if w.unit.update != replacement || w.unit.experience != 999 {
		t.Fatalf("callback mutations were lost: update=%p experience=%v", w.unit.update, w.unit.experience)
	}
}

func TestUnitGiveXP4EF270UsesUnspilledAdjustedValueForExperience(t *testing.T) {
	w := newUnitGiveXPTestWorld4EF270()
	w.unit.experience = 4_903_687
	w.target = 92_600_616
	got := unitGiveXP4EF270(w.hooks())
	if bits := math.Float64bits(got); bits != math.Float64bits(float64(math.Float32frombits(0x49561aa4))) {
		t.Fatalf("award bits = %016x, want widened binary32 49561aa4", bits)
	}
	if bits := math.Float32bits(w.unit.experience); bits != 0x4ab06963 {
		t.Fatalf("stored experience bits = %08x, want retained-value result 4ab06963", bits)
	}
	if rounded := float32(w.protectedAward + float32(4_903_687)); math.Float32bits(rounded) != 0x4ab06962 {
		t.Fatalf("precision witness collapsed: rounded-award result = %08x, want 4ab06962", math.Float32bits(rounded))
	}
}

func TestUnitGiveXP4EF270UnorderedComparisonTakesAwardPath(t *testing.T) {
	for _, mutate := range []func(*unitGiveXPTestWorld4EF270){
		func(w *unitGiveXPTestWorld4EF270) { w.unit.experience = math.Float32frombits(0x7fc12345) },
		func(w *unitGiveXPTestWorld4EF270) { w.target = math.Float32frombits(0x7fc54321) },
	} {
		w := newUnitGiveXPTestWorld4EF270()
		mutate(w)
		got := unitGiveXP4EF270(w.hooks())
		if !math.IsNaN(got) || !math.IsNaN(float64(w.unit.experience)) || len(w.events) != 15 {
			t.Fatalf("unordered path result/experience/events = %v/%v/%q", got, w.unit.experience, w.events)
		}
		if w.events[len(w.events)-3] != fmt.Sprintf("protect:12345678:%08x", math.Float32bits(w.protectedAward)) ||
			w.events[len(w.events)-2] != "report:unit" || w.events[len(w.events)-1] != "sync:unit" {
			t.Fatalf("unordered callback tail = %q", w.events[len(w.events)-3:])
		}
	}
}

func TestUnitGiveXP4EF270FaultBoundaries(t *testing.T) {
	t.Run("nil unit faults before target", func(t *testing.T) {
		w := newUnitGiveXPTestWorld4EF270()
		w.unit = nil
		unitGiveXPMustPanic4EF270(t, func() { unitGiveXP4EF270(w.hooks()) })
		if want := []string{"arg:nil", "experience:1:nil"}; !reflect.DeepEqual(w.events, want) {
			t.Fatalf("events = %q, want %q", w.events, want)
		}
	})

	t.Run("nil update faults after experience store", func(t *testing.T) {
		w := newUnitGiveXPTestWorld4EF270()
		w.unit.update = nil
		unitGiveXPMustPanic4EF270(t, func() { unitGiveXP4EF270(w.hooks()) })
		if w.unit.experience != 102 || w.events[len(w.events)-1] != "player:nil" {
			t.Fatalf("experience/events = %v/%q", w.unit.experience, w.events)
		}
	})

	t.Run("nil player faults at token after experience store", func(t *testing.T) {
		w := newUnitGiveXPTestWorld4EF270()
		w.unit.update.player = nil
		unitGiveXPMustPanic4EF270(t, func() { unitGiveXP4EF270(w.hooks()) })
		if w.unit.experience != 102 || w.events[len(w.events)-1] != "token:nil" {
			t.Fatalf("experience/events = %v/%q", w.unit.experience, w.events)
		}
	})
}
