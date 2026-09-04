package server

import (
	"fmt"
	"math"
	"reflect"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"

	noxflags "github.com/opennox/opennox/v1/common/flags"
)

type spellManaPreflightTestWorld4FCEF0 struct {
	unit          uint64
	sequence      uint64
	count         int32
	godMode       int32
	classLow      uint8
	mana          uint16
	spells        map[int32]int32
	summonCosts   map[int32]int32
	ordinaryCosts map[int32]int32
	events        []string
	faultAt       int
	after         map[string]func()
}

func newSpellManaPreflightTestWorld4FCEF0() *spellManaPreflightTestWorld4FCEF0 {
	return &spellManaPreflightTestWorld4FCEF0{
		unit:          0x100001234,
		sequence:      0x200005678,
		count:         1,
		classLow:      uint8(object.ClassPlayer),
		mana:          100,
		spells:        map[int32]int32{0: 1},
		summonCosts:   make(map[int32]int32),
		ordinaryCosts: make(map[int32]int32),
		after:         make(map[string]func()),
	}
}

func (w *spellManaPreflightTestWorld4FCEF0) record(event string) {
	w.events = append(w.events, event)
	if w.faultAt != 0 && len(w.events) == w.faultAt {
		panic(event)
	}
	if after := w.after[event]; after != nil {
		after()
	}
}

func (w *spellManaPreflightTestWorld4FCEF0) hooks() spellManaPreflightHooks4FCEF0[uint64, uint64] {
	return spellManaPreflightHooks4FCEF0[uint64, uint64]{
		loadUnitArg: func() uint64 {
			value := w.unit
			w.record(fmt.Sprintf("unit=%016x", value))
			return value
		},
		loadSequenceArg: func() uint64 {
			value := w.sequence
			w.record(fmt.Sprintf("sequence=%016x", value))
			return value
		},
		loadCountArg: func() int32 {
			value := w.count
			w.record(fmt.Sprintf("count=%d", value))
			return value
		},
		loadGodMode: func() int32 {
			value := w.godMode
			w.record(fmt.Sprintf("god=%d", value))
			return value
		},
		loadClassLow: func(unit uint64) uint8 {
			value := w.classLow
			w.record(fmt.Sprintf("class:%016x=%02x", unit, value))
			return value
		},
		loadOldMana: func(unit uint64) uint16 {
			value := w.mana
			w.record(fmt.Sprintf("mana:%016x=%d", unit, value))
			return value
		},
		loadSpell: func(sequence uint64, index int32) int32 {
			value := w.spells[index]
			w.record(fmt.Sprintf("spell:%016x[%d]=%d", sequence, index, value))
			return value
		},
		summonCost: func(spellID int32, unit uint64) int32 {
			value := w.summonCosts[spellID]
			w.record(fmt.Sprintf("summon:%d:%016x=%d", spellID, unit, value))
			return value
		},
		spellManaCost: func(spellID, costType int32) int32 {
			value := w.ordinaryCosts[spellID]
			w.record(fmt.Sprintf("ordinary:%d:%d=%d", spellID, costType, value))
			return value
		},
	}
}

func TestSpellManaPreflight4FCEF0EntryGatesAndNegativeCount(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*spellManaPreflightTestWorld4FCEF0)
		want   int32
		events []string
	}{
		{
			name:   "nil unit",
			mutate: func(w *spellManaPreflightTestWorld4FCEF0) { w.unit = 0 },
			want:   0,
			events: []string{"unit=0000000000000000"},
		},
		{
			name:   "nil sequence",
			mutate: func(w *spellManaPreflightTestWorld4FCEF0) { w.sequence = 0 },
			want:   0,
			events: []string{"unit=0000000100001234", "sequence=0000000000000000"},
		},
		{
			name:   "zero count",
			mutate: func(w *spellManaPreflightTestWorld4FCEF0) { w.count = 0 },
			want:   0,
			events: []string{"unit=0000000100001234", "sequence=0000000200005678", "count=0"},
		},
		{
			name:   "god mode",
			mutate: func(w *spellManaPreflightTestWorld4FCEF0) { w.godMode = -1 },
			want:   1,
			events: []string{"unit=0000000100001234", "sequence=0000000200005678", "count=1", "god=-1"},
		},
		{
			name:   "monster",
			mutate: func(w *spellManaPreflightTestWorld4FCEF0) { w.classLow = 0x82 },
			want:   1,
			events: []string{"unit=0000000100001234", "sequence=0000000200005678", "count=1", "god=0", "class:0000000100001234=82"},
		},
		{
			name:   "negative count still reads mana",
			mutate: func(w *spellManaPreflightTestWorld4FCEF0) { w.count = math.MinInt32 },
			want:   1,
			events: []string{
				"unit=0000000100001234", "sequence=0000000200005678", "count=-2147483648",
				"god=0", "class:0000000100001234=04", "mana:0000000100001234=100",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			w := newSpellManaPreflightTestWorld4FCEF0()
			test.mutate(w)
			if got := spellManaPreflight4FCEF0(w.hooks()); got != test.want {
				t.Fatalf("result = %d, want canonical %d", got, test.want)
			}
			if !reflect.DeepEqual(w.events, test.events) {
				t.Fatalf("events = %q, want %q", w.events, test.events)
			}
		})
	}
}

func TestSpellManaPreflight4FCEF0CostDispatchAndBoundaries(t *testing.T) {
	w := newSpellManaPreflightTestWorld4FCEF0()
	w.count = 4
	w.mana = 100
	w.spells = map[int32]int32{0: 74, 1: 75, 2: 114, 3: 115}
	w.ordinaryCosts = map[int32]int32{74: 10, 115: 40}
	w.summonCosts = map[int32]int32{75: 20, 114: 30}

	if got := spellManaPreflight4FCEF0(w.hooks()); got != 1 {
		t.Fatalf("result = %d, want canonical 1", got)
	}
	want := []string{
		"unit=0000000100001234", "sequence=0000000200005678", "count=4",
		"god=0", "class:0000000100001234=04", "mana:0000000100001234=100",
		"spell:0000000200005678[0]=74", "ordinary:74:2=10",
		"spell:0000000200005678[1]=75", "summon:75:0000000100001234=20",
		"spell:0000000200005678[2]=114", "summon:114:0000000100001234=30",
		"spell:0000000200005678[3]=115", "ordinary:115:2=40",
	}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %q, want %q", w.events, want)
	}
}

func TestSpellManaPreflight4FCEF0StopsAtFirstInsufficientCost(t *testing.T) {
	w := newSpellManaPreflightTestWorld4FCEF0()
	w.count = 2
	w.mana = 10
	w.spells = map[int32]int32{0: 74, 1: 75}
	w.ordinaryCosts[74] = 11
	w.summonCosts[75] = 1

	if got := spellManaPreflight4FCEF0(w.hooks()); got != 0 {
		t.Fatalf("result = %d, want canonical 0", got)
	}
	wantTail := []string{"spell:0000000200005678[0]=74", "ordinary:74:2=11"}
	if got := w.events[len(w.events)-len(wantTail):]; !reflect.DeepEqual(got, wantTail) {
		t.Fatalf("event tail = %q, want %q", got, wantTail)
	}
}

func TestSpellManaPreflight4FCEF0UsesSignedWrappingRemainder(t *testing.T) {
	w := newSpellManaPreflightTestWorld4FCEF0()
	w.count = 2
	w.mana = 1
	w.spells = map[int32]int32{0: 75, 1: 76}
	w.summonCosts = map[int32]int32{75: math.MinInt32, 76: 0}

	// 1 - INT32_MIN wraps to -2147483647. The following zero cost is then
	// signed-greater than the remaining value and must fail.
	if got := spellManaPreflight4FCEF0(w.hooks()); got != 0 {
		t.Fatalf("result = %d, want canonical 0 after signed wrap", got)
	}
	wantTail := []string{
		"spell:0000000200005678[0]=75", "summon:75:0000000100001234=-2147483648",
		"spell:0000000200005678[1]=76", "summon:76:0000000100001234=0",
	}
	if got := w.events[len(w.events)-len(wantTail):]; !reflect.DeepEqual(got, wantTail) {
		t.Fatalf("event tail = %q, want %q", got, wantTail)
	}
}

func TestSpellManaPreflight4FCEF0HasNoFiveSpellCapAndLoadsEntriesLive(t *testing.T) {
	w := newSpellManaPreflightTestWorld4FCEF0()
	w.count = 6
	w.spells = map[int32]int32{0: 1, 1: 2, 2: 3, 3: 4, 4: 5, 5: 6}
	w.after["spell:0000000200005678[0]=1"] = func() { w.spells[5] = 115 }

	if got := spellManaPreflight4FCEF0(w.hooks()); got != 1 {
		t.Fatalf("result = %d, want canonical 1", got)
	}
	wantTail := []string{"spell:0000000200005678[5]=115", "ordinary:115:2=0"}
	if got := w.events[len(w.events)-len(wantTail):]; !reflect.DeepEqual(got, wantTail) {
		t.Fatalf("event tail = %q, want live sixth entry %q", got, wantTail)
	}
}

func TestSpellManaPreflight4FCEF0AllFaultPrefixes(t *testing.T) {
	allEvents := []string{
		"unit=0000000100001234", "sequence=0000000200005678", "count=2",
		"god=0", "class:0000000100001234=04", "mana:0000000100001234=100",
		"spell:0000000200005678[0]=74", "ordinary:74:2=10",
		"spell:0000000200005678[1]=75", "summon:75:0000000100001234=20",
	}
	for faultAt := 1; faultAt <= len(allEvents); faultAt++ {
		t.Run(fmt.Sprintf("fault-%02d", faultAt), func(t *testing.T) {
			w := newSpellManaPreflightTestWorld4FCEF0()
			w.count = 2
			w.spells = map[int32]int32{0: 74, 1: 75}
			w.ordinaryCosts[74] = 10
			w.summonCosts[75] = 20
			w.faultAt = faultAt

			var recovered any
			func() {
				defer func() { recovered = recover() }()
				spellManaPreflight4FCEF0(w.hooks())
			}()
			if recovered != allEvents[faultAt-1] {
				t.Fatalf("recovered = %#v, want %q", recovered, allEvents[faultAt-1])
			}
			if want := allEvents[:faultAt]; !reflect.DeepEqual(w.events, want) {
				t.Fatalf("events = %q, want prefix %q", w.events, want)
			}
		})
	}
}

func TestSpellManaPreflightNative4FCEF0PreservesPointersAndWalksSixEntries(t *testing.T) {
	unit := &Object{ObjClass: object.ClassPlayer}
	sequence := [6]int32{74, 75, 114, 115, 0, 200}
	if unsafe.Sizeof(sequence[0]) != 4 {
		t.Fatalf("spell element size = %d, want 4", unsafe.Sizeof(sequence[0]))
	}
	if unsafe.Sizeof(uintptr(0)) == 8 {
		if uintptr(unsafe.Pointer(unit)) <= math.MaxUint32 {
			t.Fatalf("unit pointer = %p, want native address above 4 GiB", unit)
		}
		if uintptr(unsafe.Pointer(&sequence[0])) <= math.MaxUint32 {
			t.Fatalf("sequence pointer = %p, want native address above 4 GiB", &sequence[0])
		}
	}

	var gotIDs []int32
	got := spellManaPreflightNative4FCEF0(unit, &sequence[0], int32(len(sequence)), spellManaPreflightNativeDeps4FCEF0{
		loadGodMode: func() int32 { return 0 },
		loadOldMana: func(got *Object) uint16 {
			if got != unit {
				t.Fatalf("mana unit = %p, want %p", got, unit)
			}
			return 0
		},
		summonCost: func(id int32, got *Object) int32 {
			if got != unit {
				t.Fatalf("summon unit = %p, want %p", got, unit)
			}
			gotIDs = append(gotIDs, id)
			return 0
		},
		spellManaCost: func(id, costType int32) int32 {
			if costType != 2 {
				t.Fatalf("cost type = %d, want 2", costType)
			}
			gotIDs = append(gotIDs, id)
			return 0
		},
	})
	if got != 1 {
		t.Fatalf("result = %d, want canonical 1", got)
	}
	if want := sequence[:]; !reflect.DeepEqual(gotIDs, want) {
		t.Fatalf("spell IDs = %v, want %v", gotIDs, want)
	}
	runtime.KeepAlive(unit)
	runtime.KeepAlive(sequence)
}

func TestSpellManaPreflight4FCEF0ProductionBinding(t *testing.T) {
	wasGodMode := noxflags.HasEngine(noxflags.EngineGodMode)
	noxflags.UnsetEngine(noxflags.EngineGodMode)
	t.Cleanup(func() {
		if wasGodMode {
			noxflags.SetEngine(noxflags.EngineGodMode)
		} else {
			noxflags.UnsetEngine(noxflags.EngineGodMode)
		}
	})

	s := new(Server)
	unit := new(Object)
	sequence := [1]int32{0}
	if got := s.SpellManaPreflight4FCEF0(unit, &sequence[0], 1); got != 1 {
		t.Fatalf("ordinary zero-cost result = %d, want 1", got)
	}
	if got := s.SpellManaPreflight4FCEF0(unit, &sequence[0], math.MinInt32); got != 1 {
		t.Fatalf("negative-count result = %d, want 1", got)
	}
	if got := s.SpellManaPreflight4FCEF0(unit, &sequence[0], 0); got != 0 {
		t.Fatalf("zero-count result = %d, want 0", got)
	}
	if got := s.SpellManaPreflight4FCEF0(unit, nil, 1); got != 0 {
		t.Fatalf("nil-sequence result = %d, want 0", got)
	}
	unit.ObjClass = object.ClassMonster
	if got := s.SpellManaPreflight4FCEF0(unit, &sequence[0], 1); got != 1 {
		t.Fatalf("monster result = %d, want 1", got)
	}
}
