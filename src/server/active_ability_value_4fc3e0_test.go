package server

import (
	"fmt"
	"math"
	"reflect"
	"testing"
)

type activeAbilityValueTestRecord4FC3E0 struct {
	name     string
	unit     uint64
	ability  Ability
	active   uint32
	deadline uint32
	next     *activeAbilityValueTestRecord4FC3E0
}

type activeAbilityValueTestWorld4FC3E0 struct {
	unitArg       uint64
	abilityArg    Ability
	unitClasses   map[uint64]uint8
	updates       map[uint64]string
	players       map[string]string
	playerClasses map[string]uint8
	head          *activeAbilityValueTestRecord4FC3E0
	events        []string
	faultAt       int
	after         map[string]func()
}

func activeAbilityValueRecordName4FC3E0(record *activeAbilityValueTestRecord4FC3E0) string {
	if record == nil {
		return "nil"
	}
	return record.name
}

func activeAbilityValueOpaqueName4FC3E0(value string) string {
	if value == "" {
		return "nil"
	}
	return value
}

func (w *activeAbilityValueTestWorld4FC3E0) record(event string) {
	w.events = append(w.events, event)
	if w.faultAt != 0 && len(w.events) == w.faultAt {
		panic(event)
	}
	if after := w.after[event]; after != nil {
		after()
	}
}

func (w *activeAbilityValueTestWorld4FC3E0) hooks() activeAbilityValueHooks4FC3E0[
	uint64,
	string,
	string,
	*activeAbilityValueTestRecord4FC3E0,
] {
	return activeAbilityValueHooks4FC3E0[
		uint64,
		string,
		string,
		*activeAbilityValueTestRecord4FC3E0,
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
			w.record(fmt.Sprintf("update:%016x=%s", unit, activeAbilityValueOpaqueName4FC3E0(update)))
			return update
		},
		loadPlayer: func(update string) string {
			player := w.players[update]
			w.record("player:" + update + "=" + activeAbilityValueOpaqueName4FC3E0(player))
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
		loadExecHead: func() *activeAbilityValueTestRecord4FC3E0 {
			head := w.head
			w.record("head=" + activeAbilityValueRecordName4FC3E0(head))
			return head
		},
		loadAbilityArg: func() Ability {
			ability := w.abilityArg
			w.record(fmt.Sprintf("ability-arg=%d", ability))
			return ability
		},
		loadExecUnit: func(record *activeAbilityValueTestRecord4FC3E0) uint64 {
			unit := record.unit
			w.record(fmt.Sprintf("exec-unit:%s=%016x", record.name, unit))
			return unit
		},
		loadExecNext: func(record *activeAbilityValueTestRecord4FC3E0) *activeAbilityValueTestRecord4FC3E0 {
			next := record.next
			w.record("exec-next:" + record.name + "=" + activeAbilityValueRecordName4FC3E0(next))
			return next
		},
		loadExecAbility: func(record *activeAbilityValueTestRecord4FC3E0) Ability {
			ability := record.ability
			w.record(fmt.Sprintf("exec-ability:%s=%d", record.name, ability))
			return ability
		},
		loadExecActive: func(record *activeAbilityValueTestRecord4FC3E0) uint32 {
			active := record.active
			w.record(fmt.Sprintf("exec-active:%s=%08x", record.name, active))
			return active
		},
	}
}

func activeAbilityValueWarriorWorld4FC3E0(unit uint64) activeAbilityValueTestWorld4FC3E0 {
	return activeAbilityValueTestWorld4FC3E0{
		unitArg:       unit,
		abilityArg:    AbilityBerserk,
		unitClasses:   map[uint64]uint8{unit: activeAbilityValuePlayerClass4FC3E0},
		updates:       map[uint64]string{unit: "update"},
		players:       map[string]string{"update": "player"},
		playerClasses: map[string]uint8{"player": activeAbilityValueWarrior4FC3E0},
		after:         make(map[string]func()),
	}
}

func TestActiveAbilityValue4FC3E0TraversalOrderNativeIdentityAndRawValue(t *testing.T) {
	const (
		unit     = uint64(0x1234567889abcdef)
		lowAlias = uint64(0x0000000089abcdef)
		active   = uint32(0x89abcdef)
	)
	match := &activeAbilityValueTestRecord4FC3E0{
		name: "match", unit: unit, ability: Ability(math.MinInt32), active: active, deadline: 0xfeedface,
	}
	wrongAbility := &activeAbilityValueTestRecord4FC3E0{
		name: "wrong-ability", unit: unit, ability: AbilityWarcry, active: 1, next: match,
	}
	low := &activeAbilityValueTestRecord4FC3E0{
		name: "low-alias", unit: lowAlias, ability: Ability(math.MinInt32), active: math.MaxUint32, next: wrongAbility,
	}
	w := activeAbilityValueWarriorWorld4FC3E0(unit)
	w.abilityArg = Ability(math.MinInt32)
	w.head = low

	if got := activeAbilityValue4FC3E0(w.hooks()); uint32(got) != active {
		t.Fatalf("active bits = %08x, want %08x", uint32(got), active)
	}
	want := []string{
		"unit-arg=1234567889abcdef", "unit-class:1234567889abcdef=04",
		"update:1234567889abcdef=update", "player:update=player", "player-class:player=0",
		"head=low-alias", "ability-arg=-2147483648",
		"exec-unit:low-alias=0000000089abcdef", "exec-next:low-alias=wrong-ability",
		"exec-unit:wrong-ability=1234567889abcdef", "exec-next:wrong-ability=match", "exec-ability:wrong-ability=2",
		"exec-unit:match=1234567889abcdef", "exec-next:match=nil", "exec-ability:match=-2147483648", "exec-active:match=89abcdef",
	}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %q, want %q", w.events, want)
	}
	if match.active != active || match.deadline != 0xfeedface {
		t.Fatal("value lookup mutated Active or deadline state")
	}
}

func TestActiveAbilityValue4FC3E0GatesNilUpdateAndRequiredPointerFaults(t *testing.T) {
	t.Run("nil unit faults on class read", func(t *testing.T) {
		w := activeAbilityValueWarriorWorld4FC3E0(0)
		defer func() {
			if recover() == nil {
				t.Fatal("nil unit did not fault")
			}
			want := []string{"unit-arg=0000000000000000", "unit-class:nil"}
			if !reflect.DeepEqual(w.events, want) {
				t.Fatalf("events = %q, want %q", w.events, want)
			}
		}()
		activeAbilityValue4FC3E0(w.hooks())
	})

	t.Run("non-player avoids late reads", func(t *testing.T) {
		w := activeAbilityValueWarriorWorld4FC3E0(1)
		w.unitClasses[1] = 0x82
		if got := activeAbilityValue4FC3E0(w.hooks()); got != 0 {
			t.Fatalf("active = %d, want 0", got)
		}
		want := []string{"unit-arg=0000000000000001", "unit-class:0000000000000001=82"}
		if !reflect.DeepEqual(w.events, want) {
			t.Fatalf("events = %q, want %q", w.events, want)
		}
	})

	t.Run("nil UpdateData skips class gate", func(t *testing.T) {
		w := activeAbilityValueWarriorWorld4FC3E0(2)
		delete(w.updates, 2)
		w.head = &activeAbilityValueTestRecord4FC3E0{name: "match", unit: 2, ability: AbilityBerserk, active: 0x7fffffff}
		if got := activeAbilityValue4FC3E0(w.hooks()); got != math.MaxInt32 {
			t.Fatalf("active = %d, want %d", got, math.MaxInt32)
		}
		want := []string{
			"unit-arg=0000000000000002", "unit-class:0000000000000002=04", "update:0000000000000002=nil",
			"head=match", "ability-arg=1", "exec-unit:match=0000000000000002", "exec-next:match=nil",
			"exec-ability:match=1", "exec-active:match=7fffffff",
		}
		if !reflect.DeepEqual(w.events, want) {
			t.Fatalf("events = %q, want %q", w.events, want)
		}
	})

	t.Run("nil Player faults on class read", func(t *testing.T) {
		w := activeAbilityValueWarriorWorld4FC3E0(3)
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
		activeAbilityValue4FC3E0(w.hooks())
	})

	t.Run("non-Warrior avoids head and ability", func(t *testing.T) {
		w := activeAbilityValueWarriorWorld4FC3E0(4)
		w.playerClasses["player"] = 2
		if got := activeAbilityValue4FC3E0(w.hooks()); got != 0 {
			t.Fatalf("active = %d, want 0", got)
		}
		want := []string{
			"unit-arg=0000000000000004", "unit-class:0000000000000004=04",
			"update:0000000000000004=update", "player:update=player", "player-class:player=2",
		}
		if !reflect.DeepEqual(w.events, want) {
			t.Fatalf("events = %q, want %q", w.events, want)
		}
	})

	t.Run("empty head avoids ability", func(t *testing.T) {
		w := activeAbilityValueWarriorWorld4FC3E0(5)
		w.abilityArg = Ability(math.MinInt32)
		if got := activeAbilityValue4FC3E0(w.hooks()); got != 0 {
			t.Fatalf("active = %d, want 0", got)
		}
		want := []string{
			"unit-arg=0000000000000005", "unit-class:0000000000000005=04",
			"update:0000000000000005=update", "player:update=player", "player-class:player=0", "head=nil",
		}
		if !reflect.DeepEqual(w.events, want) {
			t.Fatalf("events = %q, want %q", w.events, want)
		}
	})
}

func TestActiveAbilityValue4FC3E0CachedKeysNextAndLiveValue(t *testing.T) {
	const unit = uint64(7)
	tail := &activeAbilityValueTestRecord4FC3E0{name: "tail", unit: unit, ability: AbilityHarpoon, active: 3}
	head := &activeAbilityValueTestRecord4FC3E0{name: "head", unit: unit, ability: AbilityWarcry, active: 1, next: tail}
	decoy := &activeAbilityValueTestRecord4FC3E0{name: "decoy", unit: unit, ability: AbilityHarpoon, active: 4}
	w := activeAbilityValueWarriorWorld4FC3E0(unit)
	w.abilityArg = AbilityHarpoon
	w.head = head
	w.after["exec-next:head=tail"] = func() {
		head.unit = 99
		head.ability = AbilityHarpoon
		head.next = decoy
	}
	w.after["exec-ability:head=3"] = func() { head.active = 0x87654321 }

	if got := activeAbilityValue4FC3E0(w.hooks()); uint32(got) != 0x87654321 {
		t.Fatalf("active bits = %08x, want live 87654321", uint32(got))
	}
	wantSuffix := []string{
		"exec-unit:head=0000000000000007", "exec-next:head=tail",
		"exec-ability:head=3", "exec-active:head=87654321",
	}
	if got := w.events[len(w.events)-len(wantSuffix):]; !reflect.DeepEqual(got, wantSuffix) {
		t.Fatalf("event suffix = %q, want %q", got, wantSuffix)
	}
	if len(w.events) != 11 {
		t.Fatalf("events = %q; match should not traverse the cached tail", w.events)
	}
}

func TestActiveAbilityValue4FC3E0CachedNextTraversalZeroMatchAndMiss(t *testing.T) {
	const unit = uint64(8)
	tail := &activeAbilityValueTestRecord4FC3E0{name: "tail", unit: unit, ability: AbilityWarcry, active: math.MaxUint32}
	head := &activeAbilityValueTestRecord4FC3E0{name: "head", unit: 9, ability: AbilityHarpoon, active: 1, next: tail}
	decoy := &activeAbilityValueTestRecord4FC3E0{name: "decoy", unit: unit, ability: AbilityHarpoon, active: 2}
	w := activeAbilityValueWarriorWorld4FC3E0(unit)
	w.abilityArg = AbilityHarpoon
	w.head = head
	w.after["exec-next:head=tail"] = func() { head.next = decoy }

	if got := activeAbilityValue4FC3E0(w.hooks()); got != 0 {
		t.Fatalf("active = %d, want miss 0", got)
	}
	wantSuffix := []string{
		"exec-unit:head=0000000000000009", "exec-next:head=tail",
		"exec-unit:tail=0000000000000008", "exec-next:tail=nil", "exec-ability:tail=2",
	}
	if got := w.events[len(w.events)-len(wantSuffix):]; !reflect.DeepEqual(got, wantSuffix) {
		t.Fatalf("event suffix = %q, want %q", got, wantSuffix)
	}

	zero := activeAbilityValueWarriorWorld4FC3E0(unit)
	zero.head = &activeAbilityValueTestRecord4FC3E0{name: "zero", unit: unit, ability: AbilityBerserk, active: 0}
	if got := activeAbilityValue4FC3E0(zero.hooks()); got != 0 {
		t.Fatalf("inactive match = %d, want 0", got)
	}
	wantZeroSuffix := []string{
		"exec-unit:zero=0000000000000008", "exec-next:zero=nil",
		"exec-ability:zero=1", "exec-active:zero=00000000",
	}
	if got := zero.events[len(zero.events)-len(wantZeroSuffix):]; !reflect.DeepEqual(got, wantZeroSuffix) {
		t.Fatalf("inactive-match suffix = %q, want %q", got, wantZeroSuffix)
	}
}

func TestActiveAbilityValue4FC3E0FaultPrefixes(t *testing.T) {
	const unit = uint64(9)
	all := []string{
		"unit-arg=0000000000000009", "unit-class:0000000000000009=04",
		"update:0000000000000009=update", "player:update=player", "player-class:player=0",
		"head=record", "ability-arg=1", "exec-unit:record=0000000000000009", "exec-next:record=nil",
		"exec-ability:record=1", "exec-active:record=80000000",
	}
	for faultAt := 1; faultAt <= len(all); faultAt++ {
		t.Run(fmt.Sprintf("fault-%d", faultAt), func(t *testing.T) {
			w := activeAbilityValueWarriorWorld4FC3E0(unit)
			w.head = &activeAbilityValueTestRecord4FC3E0{
				name: "record", unit: unit, ability: AbilityBerserk, active: 0x80000000,
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
			activeAbilityValue4FC3E0(w.hooks())
		})
	}
}
