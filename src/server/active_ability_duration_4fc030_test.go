package server

import (
	"fmt"
	"reflect"
	"testing"
)

type activeAbilityDurationTestRecord4FC030 struct {
	name     string
	unit     uint64
	ability  Ability
	deadline uint32
	active   uint32
	next     *activeAbilityDurationTestRecord4FC030
}

type activeAbilityDurationTestWorld4FC030 struct {
	head       *activeAbilityDurationTestRecord4FC030
	frame      uint32
	events     []string
	faultAt    int
	afterEvent map[string]func()
}

func activeAbilityDurationRecordName4FC030(record *activeAbilityDurationTestRecord4FC030) string {
	if record == nil {
		return "nil"
	}
	return record.name
}

func (w *activeAbilityDurationTestWorld4FC030) record(event string) {
	w.events = append(w.events, event)
	if w.faultAt != 0 && len(w.events) == w.faultAt {
		panic(event)
	}
	if after := w.afterEvent[event]; after != nil {
		after()
	}
}

func (w *activeAbilityDurationTestWorld4FC030) hooks() activeAbilityDurationHooks4FC030[
	*activeAbilityDurationTestRecord4FC030,
	uint64,
] {
	return activeAbilityDurationHooks4FC030[*activeAbilityDurationTestRecord4FC030, uint64]{
		loadHead: func() *activeAbilityDurationTestRecord4FC030 {
			head := w.head
			w.record("head=" + activeAbilityDurationRecordName4FC030(head))
			return head
		},
		loadUnit: func(record *activeAbilityDurationTestRecord4FC030) uint64 {
			unit := record.unit
			w.record(fmt.Sprintf("unit:%s=%016x", record.name, unit))
			return unit
		},
		loadAbility: func(record *activeAbilityDurationTestRecord4FC030) Ability {
			ability := record.ability
			w.record(fmt.Sprintf("ability:%s=%d", record.name, ability))
			return ability
		},
		loadNext: func(record *activeAbilityDurationTestRecord4FC030) *activeAbilityDurationTestRecord4FC030 {
			next := record.next
			w.record("next:" + record.name + "=" + activeAbilityDurationRecordName4FC030(next))
			return next
		},
		loadDeadline: func(record *activeAbilityDurationTestRecord4FC030) uint32 {
			deadline := record.deadline
			w.record(fmt.Sprintf("deadline:%s=%08x", record.name, deadline))
			return deadline
		},
		loadFrame: func() uint32 {
			frame := w.frame
			w.record(fmt.Sprintf("frame=%08x", frame))
			return frame
		},
	}
}

func TestActiveAbilityDuration4FC030TraversalOrderNativeIdentityAndInactiveMatch(t *testing.T) {
	const (
		queryUnit = uint64(0x1234567889abcdef)
		lowAlias  = uint64(0x0000000089abcdef)
	)
	match := &activeAbilityDurationTestRecord4FC030{
		name: "match", unit: queryUnit, ability: AbilityTreadLightly,
		deadline: 0, active: 0,
	}
	wrongAbility := &activeAbilityDurationTestRecord4FC030{
		name: "wrong-ability", unit: queryUnit, ability: AbilityWarcry,
		next: match,
	}
	low := &activeAbilityDurationTestRecord4FC030{
		name: "low-alias", unit: lowAlias, ability: AbilityTreadLightly,
		next: wrongAbility,
	}
	w := activeAbilityDurationTestWorld4FC030{head: low, frame: 1}

	if got := activeAbilityDuration4FC030(queryUnit, AbilityTreadLightly, w.hooks()); got != -1 {
		t.Fatalf("duration = %d, want signed -1 wrap pattern", got)
	}
	want := []string{
		"head=low-alias",
		"unit:low-alias=0000000089abcdef",
		"next:low-alias=wrong-ability",
		"unit:wrong-ability=1234567889abcdef",
		"ability:wrong-ability=2",
		"next:wrong-ability=match",
		"unit:match=1234567889abcdef",
		"ability:match=4",
		"deadline:match=00000000",
		"frame=00000001",
	}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %#v, want %#v", w.events, want)
	}
	if match.active != 0 {
		t.Fatal("test setup no longer covers an inactive matching record")
	}
}

func TestActiveAbilityDuration4FC030MissAndUint32ResultPatterns(t *testing.T) {
	t.Run("empty head", func(t *testing.T) {
		w := activeAbilityDurationTestWorld4FC030{frame: 0xfeedface}
		if got := activeAbilityDuration4FC030(uint64(7), Ability(-99), w.hooks()); got != -1 {
			t.Fatalf("miss = %d, want -1", got)
		}
		if want := []string{"head=nil"}; !reflect.DeepEqual(w.events, want) {
			t.Fatalf("events = %#v, want %#v", w.events, want)
		}
	})

	for _, tc := range []struct {
		name     string
		deadline uint32
		frame    uint32
		want     int32
	}{
		{name: "past-one-collides-with-miss", deadline: 0, frame: 1, want: -1},
		{name: "unsigned-wrap-to-one", deadline: 0, frame: ^uint32(0), want: 1},
		{name: "signed-boundary", deadline: 0x80000000, frame: 0, want: -0x80000000},
		{name: "all-ones-future", deadline: ^uint32(0), frame: 0, want: -1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			record := &activeAbilityDurationTestRecord4FC030{
				name: "record", unit: 0, ability: Ability(-7), deadline: tc.deadline,
			}
			w := activeAbilityDurationTestWorld4FC030{head: record, frame: tc.frame}
			if got := activeAbilityDuration4FC030(uint64(0), Ability(-7), w.hooks()); got != tc.want {
				t.Fatalf("duration = %d (%08x), want %d (%08x)", got, uint32(got), tc.want, uint32(tc.want))
			}
			wantEvents := []string{
				"head=record", "unit:record=0000000000000000", "ability:record=-7",
				fmt.Sprintf("deadline:record=%08x", tc.deadline),
				fmt.Sprintf("frame=%08x", tc.frame),
			}
			if !reflect.DeepEqual(w.events, wantEvents) {
				t.Fatalf("events = %#v, want %#v", w.events, wantEvents)
			}
		})
	}
}

func TestActiveAbilityDuration4FC030DeadlineThenLiveFrame(t *testing.T) {
	record := &activeAbilityDurationTestRecord4FC030{
		name: "record", unit: 9, ability: AbilityHarpoon, deadline: 10,
	}
	w := activeAbilityDurationTestWorld4FC030{
		head:       record,
		frame:      3,
		afterEvent: make(map[string]func()),
	}
	w.afterEvent["deadline:record=0000000a"] = func() {
		record.deadline = 1000
		w.frame = 8
	}

	if got := activeAbilityDuration4FC030(uint64(9), AbilityHarpoon, w.hooks()); got != 2 {
		t.Fatalf("duration = %d, want cached deadline 10 - live frame 8", got)
	}
	want := []string{
		"head=record", "unit:record=0000000000000009", "ability:record=3",
		"deadline:record=0000000a", "frame=00000008",
	}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %#v, want %#v", w.events, want)
	}
}

func TestActiveAbilityDuration4FC030FaultPrefixes(t *testing.T) {
	record := &activeAbilityDurationTestRecord4FC030{
		name: "record", unit: 7, ability: AbilityBerserk, deadline: 20,
	}
	all := []string{
		"head=record", "unit:record=0000000000000007", "ability:record=1",
		"deadline:record=00000014", "frame=00000005",
	}
	for faultAt := 1; faultAt <= len(all); faultAt++ {
		t.Run(fmt.Sprintf("fault-%d", faultAt), func(t *testing.T) {
			w := activeAbilityDurationTestWorld4FC030{head: record, frame: 5, faultAt: faultAt}
			defer func() {
				if recover() == nil {
					t.Fatal("fault was not propagated")
				}
				if want := all[:faultAt]; !reflect.DeepEqual(w.events, want) {
					t.Fatalf("events = %#v, want %#v", w.events, want)
				}
			}()
			activeAbilityDuration4FC030(uint64(7), AbilityBerserk, w.hooks())
		})
	}
}
