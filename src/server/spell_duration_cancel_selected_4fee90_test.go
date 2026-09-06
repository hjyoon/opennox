package server

import (
	"fmt"
	"math"
	"reflect"
	"testing"
)

const (
	spellDurationCancelSelectedRecordA4FEE90 = uint64(0x100000101)
	spellDurationCancelSelectedRecordB4FEE90 = uint64(0x200000202)
	spellDurationCancelSelectedRecordC4FEE90 = uint64(0x300000303)
	spellDurationCancelSelectedRecordD4FEE90 = uint64(0x400000404)
	spellDurationCancelSelectedCaster4FEE90  = uint64(0x1234567889abcdef)
	spellDurationCancelSelectedAlias4FEE90   = uint64(0x0000000089abcdef)
)

type spellDurationCancelSelectedRecordState4FEE90 struct {
	next   uint64
	caster uint64
	spell  uint32
}

type spellDurationCancelSelectedWorld4FEE90 struct {
	events    []string
	after     map[string]func()
	faultAt   int
	head      uint64
	casterArg uint64
	records   map[uint64]*spellDurationCancelSelectedRecordState4FEE90
}

func spellDurationCancelSelectedRecordName4FEE90(record uint64) string {
	switch record {
	case 0:
		return "nil"
	case spellDurationCancelSelectedRecordA4FEE90:
		return "A"
	case spellDurationCancelSelectedRecordB4FEE90:
		return "B"
	case spellDurationCancelSelectedRecordC4FEE90:
		return "C"
	case spellDurationCancelSelectedRecordD4FEE90:
		return "D"
	default:
		return fmt.Sprintf("%x", record)
	}
}

func (w *spellDurationCancelSelectedWorld4FEE90) observe(event string) {
	w.events = append(w.events, event)
	if w.faultAt != 0 && len(w.events) == w.faultAt {
		panic(event)
	}
	if after := w.after[event]; after != nil {
		after()
	}
}

func (w *spellDurationCancelSelectedWorld4FEE90) hooks() SpellDurationCancelSelectedHooks4FEE90[uint64, uint64] {
	return SpellDurationCancelSelectedHooks4FEE90[uint64, uint64]{
		LoadFirst: func() uint64 {
			value := w.head
			w.observe("head=" + spellDurationCancelSelectedRecordName4FEE90(value))
			return value
		},
		LoadCasterArg: func() uint64 {
			value := w.casterArg
			w.observe(fmt.Sprintf("caster-arg=%016x", value))
			return value
		},
		LoadNext: func(record uint64) uint64 {
			value := w.records[record].next
			w.observe(fmt.Sprintf("next:%s=%s", spellDurationCancelSelectedRecordName4FEE90(record), spellDurationCancelSelectedRecordName4FEE90(value)))
			return value
		},
		LoadCaster: func(record uint64) uint64 {
			value := w.records[record].caster
			w.observe(fmt.Sprintf("caster:%s=%016x", spellDurationCancelSelectedRecordName4FEE90(record), value))
			return value
		},
		LoadSpell: func(record uint64) uint32 {
			value := w.records[record].spell
			w.observe(fmt.Sprintf("spell:%s=%08x", spellDurationCancelSelectedRecordName4FEE90(record), value))
			return value
		},
		Cancel: func(record uint64) {
			w.observe("cancel:" + spellDurationCancelSelectedRecordName4FEE90(record))
		},
	}
}

func newSpellDurationCancelSelectedWorld4FEE90() *spellDurationCancelSelectedWorld4FEE90 {
	return &spellDurationCancelSelectedWorld4FEE90{
		after:     make(map[string]func()),
		head:      spellDurationCancelSelectedRecordA4FEE90,
		casterArg: spellDurationCancelSelectedCaster4FEE90,
		records: map[uint64]*spellDurationCancelSelectedRecordState4FEE90{
			spellDurationCancelSelectedRecordA4FEE90: {
				next: spellDurationCancelSelectedRecordB4FEE90, caster: spellDurationCancelSelectedAlias4FEE90, spell: 24,
			},
			spellDurationCancelSelectedRecordB4FEE90: {
				next: spellDurationCancelSelectedRecordC4FEE90, caster: spellDurationCancelSelectedCaster4FEE90, spell: 1,
			},
			spellDurationCancelSelectedRecordC4FEE90: {
				next: spellDurationCancelSelectedRecordD4FEE90, caster: spellDurationCancelSelectedCaster4FEE90, spell: 43,
			},
			spellDurationCancelSelectedRecordD4FEE90: {
				caster: spellDurationCancelSelectedCaster4FEE90, spell: 67,
			},
		},
	}
}

func TestSpellDurationCancelSelected4FEE90ExactTraceAndNativeTokens(t *testing.T) {
	w := newSpellDurationCancelSelectedWorld4FEE90()
	SpellDurationCancelSelected4FEE90(w.hooks())

	want := []string{
		"head=A", "caster-arg=1234567889abcdef",
		"next:A=B", "caster:A=0000000089abcdef",
		"next:B=C", "caster:B=1234567889abcdef", "spell:B=00000001",
		"next:C=D", "caster:C=1234567889abcdef", "spell:C=0000002b", "cancel:C",
		"next:D=nil", "caster:D=1234567889abcdef", "spell:D=00000043", "cancel:D",
	}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %q, want %q", w.events, want)
	}
}

func TestSpellDurationCancelSelected4FEE90EmptyAvoidsCasterArgument(t *testing.T) {
	w := newSpellDurationCancelSelectedWorld4FEE90()
	w.head = 0
	SpellDurationCancelSelected4FEE90(w.hooks())
	if !reflect.DeepEqual(w.events, []string{"head=nil"}) {
		t.Fatalf("events = %q, want only the head load", w.events)
	}
}

func TestSpellDurationCancelSelected4FEE90ExactSpellSet(t *testing.T) {
	tests := []struct {
		spell uint32
		want  bool
	}{
		{24, true}, {43, true}, {35, true}, {8, true}, {22, true}, {59, true}, {67, true},
		{0, false}, {7, false}, {9, false}, {21, false}, {23, false}, {25, false},
		{34, false}, {36, false}, {42, false}, {44, false}, {58, false}, {60, false},
		{66, false}, {68, false}, {math.MaxUint32, false},
	}
	for _, tc := range tests {
		t.Run(fmt.Sprintf("spell-%d", tc.spell), func(t *testing.T) {
			w := newSpellDurationCancelSelectedWorld4FEE90()
			w.records = map[uint64]*spellDurationCancelSelectedRecordState4FEE90{
				spellDurationCancelSelectedRecordA4FEE90: {
					caster: spellDurationCancelSelectedCaster4FEE90,
					spell:  tc.spell,
				},
			}
			SpellDurationCancelSelected4FEE90(w.hooks())
			got := w.events[len(w.events)-1] == "cancel:A"
			if got != tc.want {
				t.Fatalf("spell %d canceled = %t, want %t; events = %q", tc.spell, got, tc.want, w.events)
			}
		})
	}
}

func TestSpellDurationCancelSelected4FEE90CachesCasterAndUsesSavedSuccessor(t *testing.T) {
	w := newSpellDurationCancelSelectedWorld4FEE90()
	w.records = map[uint64]*spellDurationCancelSelectedRecordState4FEE90{
		spellDurationCancelSelectedRecordA4FEE90: {
			next: spellDurationCancelSelectedRecordB4FEE90, caster: spellDurationCancelSelectedCaster4FEE90, spell: 24,
		},
		spellDurationCancelSelectedRecordB4FEE90: {
			caster: spellDurationCancelSelectedCaster4FEE90, spell: 59,
		},
		spellDurationCancelSelectedRecordC4FEE90: {
			caster: spellDurationCancelSelectedAlias4FEE90, spell: 67,
		},
	}
	w.after["caster-arg=1234567889abcdef"] = func() {
		w.casterArg = spellDurationCancelSelectedAlias4FEE90
	}
	w.after["next:A=B"] = func() {
		w.records[spellDurationCancelSelectedRecordA4FEE90].next = spellDurationCancelSelectedRecordC4FEE90
	}
	w.after["cancel:A"] = func() {
		w.head = spellDurationCancelSelectedRecordC4FEE90
		w.records[spellDurationCancelSelectedRecordA4FEE90].next = 0
	}

	SpellDurationCancelSelected4FEE90(w.hooks())
	want := []string{
		"head=A", "caster-arg=1234567889abcdef",
		"next:A=B", "caster:A=1234567889abcdef", "spell:A=00000018", "cancel:A",
		"next:B=nil", "caster:B=1234567889abcdef", "spell:B=0000003b", "cancel:B",
	}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %q, want cached-caster/saved-successor trace %q", w.events, want)
	}
}

func TestSpellDurationCancelSelected4FEE90NilCasterIsValidIdentity(t *testing.T) {
	w := newSpellDurationCancelSelectedWorld4FEE90()
	w.casterArg = 0
	w.records = map[uint64]*spellDurationCancelSelectedRecordState4FEE90{
		spellDurationCancelSelectedRecordA4FEE90: {caster: 0, spell: 35},
	}
	SpellDurationCancelSelected4FEE90(w.hooks())
	if got := w.events[len(w.events)-1]; got != "cancel:A" {
		t.Fatalf("last event = %q, want nil-caster identity cancellation", got)
	}
}

func TestSpellDurationCancelSelected4FEE90DoesNotAddCycleGuard(t *testing.T) {
	w := newSpellDurationCancelSelectedWorld4FEE90()
	w.records = map[uint64]*spellDurationCancelSelectedRecordState4FEE90{
		spellDurationCancelSelectedRecordA4FEE90: {
			next: spellDurationCancelSelectedRecordA4FEE90, caster: spellDurationCancelSelectedAlias4FEE90,
		},
	}
	nextLoads := 0
	sentinel := &struct{}{}
	w.after["next:A=A"] = func() {
		nextLoads++
		if nextLoads == 3 {
			panic(sentinel)
		}
	}

	var recovered any
	func() {
		defer func() { recovered = recover() }()
		SpellDurationCancelSelected4FEE90(w.hooks())
	}()
	if recovered != sentinel {
		t.Fatalf("recovered = %#v, want third-next sentinel", recovered)
	}
	want := []string{
		"head=A", "caster-arg=1234567889abcdef",
		"next:A=A", "caster:A=0000000089abcdef",
		"next:A=A", "caster:A=0000000089abcdef",
		"next:A=A",
	}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %q, want repeated cycle prefix %q", w.events, want)
	}
}

func TestSpellDurationCancelSelected4FEE90FaultPrefixes(t *testing.T) {
	baseline := newSpellDurationCancelSelectedWorld4FEE90()
	SpellDurationCancelSelected4FEE90(baseline.hooks())
	want := append([]string(nil), baseline.events...)

	for faultAt := 1; faultAt <= len(want); faultAt++ {
		t.Run(fmt.Sprintf("fault-%d", faultAt), func(t *testing.T) {
			w := newSpellDurationCancelSelectedWorld4FEE90()
			w.faultAt = faultAt
			var recovered any
			func() {
				defer func() { recovered = recover() }()
				SpellDurationCancelSelected4FEE90(w.hooks())
			}()
			if recovered == nil {
				t.Fatal("fault sentinel was not recovered")
			}
			if prefix := want[:faultAt]; !reflect.DeepEqual(w.events, prefix) {
				t.Fatalf("events = %q, want fault prefix %q", w.events, prefix)
			}
		})
	}
}
