package server

import (
	"fmt"
	"math"
	"reflect"
	"testing"
)

type activeAbilityDeadlineTestRecord4FC070 struct {
	name     string
	unit     uint64
	ability  Ability
	deadline uint32
	active   uint32
	next     *activeAbilityDeadlineTestRecord4FC070
}

type activeAbilityDeadlineTestWorld4FC070 struct {
	head       *activeAbilityDeadlineTestRecord4FC070
	unitArg    uint64
	abilityArg Ability
	deltaArg   int32
	frame      uint32
	events     []string
	faultAt    int
	afterEvent map[string]func()
}

func activeAbilityDeadlineRecordName4FC070(record *activeAbilityDeadlineTestRecord4FC070) string {
	if record == nil {
		return "nil"
	}
	return record.name
}

func (w *activeAbilityDeadlineTestWorld4FC070) record(event string) {
	w.events = append(w.events, event)
	if w.faultAt != 0 && len(w.events) == w.faultAt {
		panic(event)
	}
	if after := w.afterEvent[event]; after != nil {
		after()
	}
}

func (w *activeAbilityDeadlineTestWorld4FC070) hooks() activeAbilityDeadlineHooks4FC070[
	*activeAbilityDeadlineTestRecord4FC070,
	uint64,
] {
	return activeAbilityDeadlineHooks4FC070[*activeAbilityDeadlineTestRecord4FC070, uint64]{
		loadHead: func() *activeAbilityDeadlineTestRecord4FC070 {
			head := w.head
			w.record("head=" + activeAbilityDeadlineRecordName4FC070(head))
			return head
		},
		loadAbilityArg: func() Ability {
			ability := w.abilityArg
			w.record(fmt.Sprintf("ability-arg=%d", ability))
			return ability
		},
		loadUnitArg: func() uint64 {
			unit := w.unitArg
			w.record(fmt.Sprintf("unit-arg=%016x", unit))
			return unit
		},
		loadUnit: func(record *activeAbilityDeadlineTestRecord4FC070) uint64 {
			unit := record.unit
			w.record(fmt.Sprintf("unit:%s=%016x", record.name, unit))
			return unit
		},
		loadAbility: func(record *activeAbilityDeadlineTestRecord4FC070) Ability {
			ability := record.ability
			w.record(fmt.Sprintf("ability:%s=%d", record.name, ability))
			return ability
		},
		loadNext: func(record *activeAbilityDeadlineTestRecord4FC070) *activeAbilityDeadlineTestRecord4FC070 {
			next := record.next
			w.record("next:" + record.name + "=" + activeAbilityDeadlineRecordName4FC070(next))
			return next
		},
		loadDeltaArg: func() int32 {
			delta := w.deltaArg
			w.record(fmt.Sprintf("delta-arg=%d", delta))
			return delta
		},
		loadFrame: func() uint32 {
			frame := w.frame
			w.record(fmt.Sprintf("frame=%08x", frame))
			return frame
		},
		storeDeadline: func(record *activeAbilityDeadlineTestRecord4FC070, deadline uint32) {
			w.record(fmt.Sprintf("store:%s=%08x", record.name, deadline))
			record.deadline = deadline
		},
	}
}

func TestActiveAbilityDeadline4FC070TraversalOrderNativeIdentityAndInactiveMatch(t *testing.T) {
	const (
		queryUnit = uint64(0x1234567889abcdef)
		lowAlias  = uint64(0x0000000089abcdef)
	)
	match := &activeAbilityDeadlineTestRecord4FC070{
		name: "match", unit: queryUnit, ability: Ability(-7), deadline: 99, active: 0,
	}
	wrongAbility := &activeAbilityDeadlineTestRecord4FC070{
		name: "wrong-ability", unit: queryUnit, ability: AbilityWarcry, deadline: 88,
		next: match,
	}
	low := &activeAbilityDeadlineTestRecord4FC070{
		name: "low-alias", unit: lowAlias, ability: Ability(-7), deadline: 77,
		next: wrongAbility,
	}
	w := activeAbilityDeadlineTestWorld4FC070{
		head: low, unitArg: queryUnit, abilityArg: Ability(-7), deltaArg: -2, frame: 1,
	}

	activeAbilityDeadline4FC070(w.hooks())

	want := []string{
		"head=low-alias",
		"ability-arg=-7",
		"unit-arg=1234567889abcdef",
		"unit:low-alias=0000000089abcdef",
		"next:low-alias=wrong-ability",
		"unit:wrong-ability=1234567889abcdef",
		"ability:wrong-ability=2",
		"next:wrong-ability=match",
		"unit:match=1234567889abcdef",
		"ability:match=-7",
		"delta-arg=-2",
		"frame=00000001",
		"store:match=ffffffff",
	}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %#v, want %#v", w.events, want)
	}
	if match.deadline != math.MaxUint32 {
		t.Fatalf("deadline = %08x, want ffffffff", match.deadline)
	}
	if low.deadline != 77 || wrongAbility.deadline != 88 || match.active != 0 {
		t.Fatal("deadline adjustment touched a skipped field or no longer covers an inactive match")
	}
}

func TestActiveAbilityDeadline4FC070EmptyAndMissAvoidLateReads(t *testing.T) {
	t.Run("empty head", func(t *testing.T) {
		w := activeAbilityDeadlineTestWorld4FC070{
			unitArg: 7, abilityArg: Ability(-99), deltaArg: 123, frame: 0xfeedface,
		}
		activeAbilityDeadline4FC070(w.hooks())
		if want := []string{"head=nil"}; !reflect.DeepEqual(w.events, want) {
			t.Fatalf("events = %#v, want %#v", w.events, want)
		}
	})

	t.Run("nonempty miss", func(t *testing.T) {
		record := &activeAbilityDeadlineTestRecord4FC070{
			name: "record", unit: 7, ability: AbilityWarcry, deadline: 44,
		}
		w := activeAbilityDeadlineTestWorld4FC070{
			head: record, unitArg: 7, abilityArg: Ability(-99), deltaArg: 123, frame: 0xfeedface,
		}
		activeAbilityDeadline4FC070(w.hooks())
		want := []string{
			"head=record", "ability-arg=-99", "unit-arg=0000000000000007",
			"unit:record=0000000000000007", "ability:record=2", "next:record=nil",
		}
		if !reflect.DeepEqual(w.events, want) {
			t.Fatalf("events = %#v, want %#v", w.events, want)
		}
		if record.deadline != 44 {
			t.Fatalf("deadline = %d, want unchanged 44", record.deadline)
		}
	})
}

func TestActiveAbilityDeadline4FC070Uint32AdditionPatterns(t *testing.T) {
	for _, tc := range []struct {
		name  string
		frame uint32
		delta int32
		want  uint32
	}{
		{name: "zero", frame: 0, delta: 0, want: 0},
		{name: "negative-one", frame: 0, delta: -1, want: math.MaxUint32},
		{name: "minimum-signed", frame: 1, delta: math.MinInt32, want: 0x80000001},
		{name: "positive-overflow", frame: math.MaxUint32, delta: 2, want: 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			record := &activeAbilityDeadlineTestRecord4FC070{name: "record", unit: 0, ability: Ability(-7)}
			w := activeAbilityDeadlineTestWorld4FC070{
				head: record, unitArg: 0, abilityArg: Ability(-7), deltaArg: tc.delta, frame: tc.frame,
			}
			activeAbilityDeadline4FC070(w.hooks())
			if record.deadline != tc.want {
				t.Fatalf("deadline = %08x, want %08x", record.deadline, tc.want)
			}
		})
	}
}

func TestActiveAbilityDeadline4FC070DeltaThenLiveFrameThenStore(t *testing.T) {
	record := &activeAbilityDeadlineTestRecord4FC070{
		name: "record", unit: 9, ability: AbilityHarpoon, deadline: 100,
	}
	w := activeAbilityDeadlineTestWorld4FC070{
		head: record, unitArg: 9, abilityArg: AbilityHarpoon, deltaArg: 2, frame: 3,
		afterEvent: make(map[string]func()),
	}
	w.afterEvent["delta-arg=2"] = func() {
		w.deltaArg = 100
		w.frame = 8
	}
	w.afterEvent["frame=00000008"] = func() {
		w.frame = 1000
		record.deadline = 0xdeadbeef
	}

	activeAbilityDeadline4FC070(w.hooks())

	if record.deadline != 10 {
		t.Fatalf("deadline = %d, want cached delta 2 + live then cached frame 8", record.deadline)
	}
	want := []string{
		"head=record", "ability-arg=3", "unit-arg=0000000000000009",
		"unit:record=0000000000000009", "ability:record=3",
		"delta-arg=2", "frame=00000008", "store:record=0000000a",
	}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %#v, want %#v", w.events, want)
	}
}

func TestActiveAbilityDeadline4FC070FaultPrefixes(t *testing.T) {
	all := []string{
		"head=record", "ability-arg=1", "unit-arg=0000000000000007",
		"unit:record=0000000000000007", "ability:record=1",
		"delta-arg=-3", "frame=00000005", "store:record=00000002",
	}
	for faultAt := 1; faultAt <= len(all); faultAt++ {
		t.Run(fmt.Sprintf("fault-%d", faultAt), func(t *testing.T) {
			record := &activeAbilityDeadlineTestRecord4FC070{
				name: "record", unit: 7, ability: AbilityBerserk, deadline: 20,
			}
			w := activeAbilityDeadlineTestWorld4FC070{
				head: record, unitArg: 7, abilityArg: AbilityBerserk, deltaArg: -3,
				frame: 5, faultAt: faultAt,
			}
			defer func() {
				if recover() == nil {
					t.Fatal("fault was not propagated")
				}
				if want := all[:faultAt]; !reflect.DeepEqual(w.events, want) {
					t.Fatalf("events = %#v, want %#v", w.events, want)
				}
				if record.deadline != 20 {
					t.Fatalf("faulting operation changed deadline to %08x", record.deadline)
				}
			}()
			activeAbilityDeadline4FC070(w.hooks())
		})
	}
}
