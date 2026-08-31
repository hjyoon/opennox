package server

import (
	"fmt"
	"math"
	"reflect"
	"testing"
)

type activeAbilityDeactivateTestRecord4FC440 struct {
	name     string
	unit     uint64
	ability  Ability
	active   uint32
	deadline uint32
	next     *activeAbilityDeactivateTestRecord4FC440
	previous *activeAbilityDeactivateTestRecord4FC440
}

type activeAbilityDeactivateTestWorld4FC440 struct {
	unitArg       uint64
	abilityArg    Ability
	unitClasses   map[uint64]uint8
	updates       map[uint64]string
	players       map[string]string
	playerClasses map[string]uint8
	head          *activeAbilityDeactivateTestRecord4FC440
	events        []string
	faultAt       int
	after         map[string]func()
}

func activeAbilityDeactivateRecordName4FC440(record *activeAbilityDeactivateTestRecord4FC440) string {
	if record == nil {
		return "nil"
	}
	return record.name
}

func activeAbilityDeactivateOpaqueName4FC440(value string) string {
	if value == "" {
		return "nil"
	}
	return value
}

func (w *activeAbilityDeactivateTestWorld4FC440) record(event string) {
	w.events = append(w.events, event)
	if w.faultAt != 0 && len(w.events) == w.faultAt {
		panic(event)
	}
	if after := w.after[event]; after != nil {
		after()
	}
}

func (w *activeAbilityDeactivateTestWorld4FC440) hooks() activeAbilityDeactivateHooks4FC440[
	uint64,
	string,
	string,
	*activeAbilityDeactivateTestRecord4FC440,
] {
	return activeAbilityDeactivateHooks4FC440[
		uint64,
		string,
		string,
		*activeAbilityDeactivateTestRecord4FC440,
	]{
		loadUnitArg: func() uint64 {
			unit := w.unitArg
			w.record(fmt.Sprintf("unit-arg=%016x", unit))
			return unit
		},
		loadUnitClassLow: func(unit uint64) uint8 {
			if unit == 0 {
				w.record("unit-class:nil")
				panic("nil unit")
			}
			class := w.unitClasses[unit]
			w.record(fmt.Sprintf("unit-class:%016x=%02x", unit, class))
			return class
		},
		loadUpdateData: func(unit uint64) string {
			update := w.updates[unit]
			w.record(fmt.Sprintf("update:%016x=%s", unit, activeAbilityDeactivateOpaqueName4FC440(update)))
			return update
		},
		loadPlayer: func(update string) string {
			player := w.players[update]
			w.record("player:" + update + "=" + activeAbilityDeactivateOpaqueName4FC440(player))
			return player
		},
		loadPlayerClass: func(player string) uint8 {
			if player == "" {
				w.record("player-class:nil")
				panic("nil player")
			}
			class := w.playerClasses[player]
			w.record(fmt.Sprintf("player-class:%s=%d", player, class))
			return class
		},
		loadExecHead: func() *activeAbilityDeactivateTestRecord4FC440 {
			head := w.head
			w.record("head=" + activeAbilityDeactivateRecordName4FC440(head))
			return head
		},
		loadAbilityArg: func() Ability {
			ability := w.abilityArg
			w.record(fmt.Sprintf("ability-arg=%d", ability))
			return ability
		},
		loadExecUnit: func(record *activeAbilityDeactivateTestRecord4FC440) uint64 {
			unit := record.unit
			w.record(fmt.Sprintf("exec-unit:%s=%016x", record.name, unit))
			return unit
		},
		loadExecNext: func(record *activeAbilityDeactivateTestRecord4FC440) *activeAbilityDeactivateTestRecord4FC440 {
			next := record.next
			w.record("exec-next:" + record.name + "=" + activeAbilityDeactivateRecordName4FC440(next))
			return next
		},
		loadExecAbility: func(record *activeAbilityDeactivateTestRecord4FC440) Ability {
			ability := record.ability
			w.record(fmt.Sprintf("exec-ability:%s=%d", record.name, ability))
			return ability
		},
		storeExecActive: func(record *activeAbilityDeactivateTestRecord4FC440, active uint32) {
			w.record(fmt.Sprintf("exec-active:%s=%08x", record.name, active))
			record.active = active
		},
	}
}

func activeAbilityDeactivateWarriorWorld4FC440(unit uint64) activeAbilityDeactivateTestWorld4FC440 {
	return activeAbilityDeactivateTestWorld4FC440{
		unitArg:       unit,
		abilityArg:    AbilityBerserk,
		unitClasses:   map[uint64]uint8{unit: activeAbilityDeactivatePlayerClass4FC440},
		updates:       map[uint64]string{unit: "update"},
		players:       map[string]string{"update": "player"},
		playerClasses: map[string]uint8{"player": activeAbilityDeactivateWarrior4FC440},
		after:         make(map[string]func()),
	}
}

func TestActiveAbilityDeactivate4FC440TraversalOrderNativeIdentityAndFirstMatch(t *testing.T) {
	const (
		unit     = uint64(0x1234567889abcdef)
		lowAlias = uint64(0x0000000089abcdef)
	)
	tail := &activeAbilityDeactivateTestRecord4FC440{
		name: "tail", unit: unit, ability: Ability(math.MinInt32), active: 0x55667788,
	}
	match := &activeAbilityDeactivateTestRecord4FC440{
		name: "match", unit: unit, ability: Ability(math.MinInt32), active: math.MaxUint32,
		deadline: 0xfeedface, next: tail,
	}
	wrongAbility := &activeAbilityDeactivateTestRecord4FC440{
		name: "wrong-ability", unit: unit, ability: AbilityWarcry, active: 0x12345678, next: match,
	}
	low := &activeAbilityDeactivateTestRecord4FC440{
		name: "low-alias", unit: lowAlias, ability: Ability(math.MinInt32), active: 0x87654321, next: wrongAbility,
	}
	wrongAbility.previous = low
	match.previous = wrongAbility
	tail.previous = match
	w := activeAbilityDeactivateWarriorWorld4FC440(unit)
	w.abilityArg = Ability(math.MinInt32)
	w.head = low

	activeAbilityDeactivate4FC440(w.hooks())
	want := []string{
		"unit-arg=1234567889abcdef", "unit-class:1234567889abcdef=04",
		"update:1234567889abcdef=update", "player:update=player", "player-class:player=0",
		"head=low-alias", "ability-arg=-2147483648",
		"exec-unit:low-alias=0000000089abcdef", "exec-next:low-alias=wrong-ability",
		"exec-unit:wrong-ability=1234567889abcdef", "exec-next:wrong-ability=match", "exec-ability:wrong-ability=2",
		"exec-unit:match=1234567889abcdef", "exec-next:match=tail", "exec-ability:match=-2147483648",
		"exec-active:match=00000000",
	}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %q, want %q", w.events, want)
	}
	if match.active != 0 || tail.active != 0x55667788 || low.active != 0x87654321 ||
		wrongAbility.active != 0x12345678 || match.deadline != 0xfeedface ||
		low.next != wrongAbility || wrongAbility.next != match || match.next != tail ||
		wrongAbility.previous != low || match.previous != wrongAbility || tail.previous != match {
		t.Fatal("deactivation changed an ignored field, another record, or list topology")
	}
}

func TestActiveAbilityDeactivate4FC440GatesNilUpdateAndRequiredPointerFaults(t *testing.T) {
	t.Run("nil unit faults on class read", func(t *testing.T) {
		w := activeAbilityDeactivateWarriorWorld4FC440(0)
		defer func() {
			if recover() == nil {
				t.Fatal("nil unit did not fault")
			}
			want := []string{"unit-arg=0000000000000000", "unit-class:nil"}
			if !reflect.DeepEqual(w.events, want) {
				t.Fatalf("events = %q, want %q", w.events, want)
			}
		}()
		activeAbilityDeactivate4FC440(w.hooks())
	})

	t.Run("non-player avoids late reads", func(t *testing.T) {
		w := activeAbilityDeactivateWarriorWorld4FC440(1)
		w.unitClasses[1] = 0x82
		activeAbilityDeactivate4FC440(w.hooks())
		want := []string{"unit-arg=0000000000000001", "unit-class:0000000000000001=82"}
		if !reflect.DeepEqual(w.events, want) {
			t.Fatalf("events = %q, want %q", w.events, want)
		}
	})

	t.Run("nil UpdateData skips class gate", func(t *testing.T) {
		w := activeAbilityDeactivateWarriorWorld4FC440(2)
		delete(w.updates, 2)
		w.head = &activeAbilityDeactivateTestRecord4FC440{
			name: "match", unit: 2, ability: AbilityBerserk, active: math.MaxUint32,
		}
		activeAbilityDeactivate4FC440(w.hooks())
		want := []string{
			"unit-arg=0000000000000002", "unit-class:0000000000000002=04", "update:0000000000000002=nil",
			"head=match", "ability-arg=1", "exec-unit:match=0000000000000002", "exec-next:match=nil",
			"exec-ability:match=1", "exec-active:match=00000000",
		}
		if !reflect.DeepEqual(w.events, want) || w.head.active != 0 {
			t.Fatalf("events/active = %q/%08x, want %q/00000000", w.events, w.head.active, want)
		}
	})

	t.Run("nil Player faults on class read", func(t *testing.T) {
		w := activeAbilityDeactivateWarriorWorld4FC440(3)
		delete(w.players, "update")
		defer func() {
			if recover() == nil {
				t.Fatal("nil Player did not fault")
			}
			want := []string{
				"unit-arg=0000000000000003", "unit-class:0000000000000003=04",
				"update:0000000000000003=update", "player:update=nil", "player-class:nil",
			}
			if !reflect.DeepEqual(w.events, want) {
				t.Fatalf("events = %q, want %q", w.events, want)
			}
		}()
		activeAbilityDeactivate4FC440(w.hooks())
	})

	t.Run("non-Warrior avoids head and ability", func(t *testing.T) {
		w := activeAbilityDeactivateWarriorWorld4FC440(4)
		w.playerClasses["player"] = 2
		activeAbilityDeactivate4FC440(w.hooks())
		want := []string{
			"unit-arg=0000000000000004", "unit-class:0000000000000004=04",
			"update:0000000000000004=update", "player:update=player", "player-class:player=2",
		}
		if !reflect.DeepEqual(w.events, want) {
			t.Fatalf("events = %q, want %q", w.events, want)
		}
	})

	t.Run("empty head avoids ability", func(t *testing.T) {
		w := activeAbilityDeactivateWarriorWorld4FC440(5)
		w.abilityArg = Ability(math.MinInt32)
		activeAbilityDeactivate4FC440(w.hooks())
		want := []string{
			"unit-arg=0000000000000005", "unit-class:0000000000000005=04",
			"update:0000000000000005=update", "player:update=player", "player-class:player=0", "head=nil",
		}
		if !reflect.DeepEqual(w.events, want) {
			t.Fatalf("events = %q, want %q", w.events, want)
		}
	})
}

func TestActiveAbilityDeactivate4FC440CachedUnitNextAndLiveAbility(t *testing.T) {
	const unit = uint64(6)
	tail := &activeAbilityDeactivateTestRecord4FC440{name: "tail", unit: unit, ability: AbilityHarpoon, active: 3}
	head := &activeAbilityDeactivateTestRecord4FC440{name: "head", unit: unit, ability: AbilityWarcry, active: 1, next: tail}
	decoy := &activeAbilityDeactivateTestRecord4FC440{name: "decoy", unit: unit, ability: AbilityHarpoon, active: 4}
	w := activeAbilityDeactivateWarriorWorld4FC440(unit)
	w.abilityArg = AbilityHarpoon
	w.head = head
	w.after["exec-next:head=tail"] = func() {
		head.unit = 99
		head.ability = AbilityHarpoon
		head.next = decoy
	}

	activeAbilityDeactivate4FC440(w.hooks())
	wantSuffix := []string{
		"exec-unit:head=0000000000000006", "exec-next:head=tail",
		"exec-ability:head=3", "exec-active:head=00000000",
	}
	if got := w.events[len(w.events)-len(wantSuffix):]; !reflect.DeepEqual(got, wantSuffix) {
		t.Fatalf("event suffix = %q, want %q", got, wantSuffix)
	}
	if len(w.events) != 11 || head.active != 0 || tail.active != 3 || decoy.active != 4 {
		t.Fatalf("events/active = %q/%d/%d/%d; cached-key match should stop at head", w.events, head.active, tail.active, decoy.active)
	}
}

func TestActiveAbilityDeactivate4FC440CachedNextTraversalAndMiss(t *testing.T) {
	const unit = uint64(7)
	tail := &activeAbilityDeactivateTestRecord4FC440{name: "tail", unit: unit, ability: AbilityWarcry, active: math.MaxUint32}
	head := &activeAbilityDeactivateTestRecord4FC440{name: "head", unit: 8, ability: AbilityHarpoon, active: 1, next: tail}
	decoy := &activeAbilityDeactivateTestRecord4FC440{name: "decoy", unit: unit, ability: AbilityHarpoon, active: 2}
	w := activeAbilityDeactivateWarriorWorld4FC440(unit)
	w.abilityArg = AbilityHarpoon
	w.head = head
	w.after["exec-next:head=tail"] = func() {
		head.unit = unit
		head.next = decoy
	}

	activeAbilityDeactivate4FC440(w.hooks())
	wantSuffix := []string{
		"exec-unit:head=0000000000000008", "exec-next:head=tail",
		"exec-unit:tail=0000000000000007", "exec-next:tail=nil", "exec-ability:tail=2",
	}
	if got := w.events[len(w.events)-len(wantSuffix):]; !reflect.DeepEqual(got, wantSuffix) {
		t.Fatalf("event suffix = %q, want %q", got, wantSuffix)
	}
	if head.active != 1 || tail.active != math.MaxUint32 || decoy.active != 2 {
		t.Fatal("miss changed an Active value")
	}
}

func TestActiveAbilityDeactivate4FC440FaultPrefixes(t *testing.T) {
	const unit = uint64(9)
	all := []string{
		"unit-arg=0000000000000009", "unit-class:0000000000000009=04",
		"update:0000000000000009=update", "player:update=player", "player-class:player=0",
		"head=record", "ability-arg=1", "exec-unit:record=0000000000000009", "exec-next:record=nil",
		"exec-ability:record=1", "exec-active:record=00000000",
	}
	for faultAt := 1; faultAt <= len(all); faultAt++ {
		t.Run(fmt.Sprintf("fault-%d", faultAt), func(t *testing.T) {
			w := activeAbilityDeactivateWarriorWorld4FC440(unit)
			w.head = &activeAbilityDeactivateTestRecord4FC440{
				name: "record", unit: unit, ability: AbilityBerserk, active: math.MaxUint32,
			}
			w.faultAt = faultAt
			defer func() {
				if recover() == nil {
					t.Fatal("fault was not propagated")
				}
				if want := all[:faultAt]; !reflect.DeepEqual(w.events, want) {
					t.Fatalf("events = %q, want %q", w.events, want)
				}
			}()
			activeAbilityDeactivate4FC440(w.hooks())
		})
	}
}
