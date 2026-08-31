package server

import (
	"fmt"
	"math"
	"reflect"
	"testing"
)

type activeAbilityMembershipTestRecord4FC250 struct {
	name     string
	unit     uint64
	ability  Ability
	active   uint32
	deadline uint32
	next     *activeAbilityMembershipTestRecord4FC250
}

type activeAbilityMembershipTestWorld4FC250 struct {
	unitArg       uint64
	abilityArg    Ability
	unitClasses   map[uint64]uint8
	updates       map[uint64]string
	players       map[string]string
	playerClasses map[string]uint8
	head          *activeAbilityMembershipTestRecord4FC250
	events        []string
	faultAt       int
	after         map[string]func()
}

func activeAbilityMembershipRecordName4FC250(record *activeAbilityMembershipTestRecord4FC250) string {
	if record == nil {
		return "nil"
	}
	return record.name
}

func activeAbilityMembershipOpaqueName4FC250(value string) string {
	if value == "" {
		return "nil"
	}
	return value
}

func (w *activeAbilityMembershipTestWorld4FC250) record(event string) {
	w.events = append(w.events, event)
	if w.faultAt != 0 && len(w.events) == w.faultAt {
		panic(event)
	}
	if after := w.after[event]; after != nil {
		after()
	}
}

func (w *activeAbilityMembershipTestWorld4FC250) hooks() activeAbilityMembershipHooks4FC250[
	uint64,
	string,
	string,
	*activeAbilityMembershipTestRecord4FC250,
] {
	return activeAbilityMembershipHooks4FC250[
		uint64,
		string,
		string,
		*activeAbilityMembershipTestRecord4FC250,
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
			w.record(fmt.Sprintf("update:%016x=%s", unit, activeAbilityMembershipOpaqueName4FC250(update)))
			return update
		},
		loadPlayer: func(update string) string {
			player := w.players[update]
			w.record("player:" + update + "=" + activeAbilityMembershipOpaqueName4FC250(player))
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
		loadExecHead: func() *activeAbilityMembershipTestRecord4FC250 {
			head := w.head
			w.record("head=" + activeAbilityMembershipRecordName4FC250(head))
			return head
		},
		loadAbilityArg: func() Ability {
			ability := w.abilityArg
			w.record(fmt.Sprintf("ability-arg=%d", ability))
			return ability
		},
		loadExecUnit: func(record *activeAbilityMembershipTestRecord4FC250) uint64 {
			unit := record.unit
			w.record(fmt.Sprintf("exec-unit:%s=%016x", record.name, unit))
			return unit
		},
		loadExecNext: func(record *activeAbilityMembershipTestRecord4FC250) *activeAbilityMembershipTestRecord4FC250 {
			next := record.next
			w.record("exec-next:" + record.name + "=" + activeAbilityMembershipRecordName4FC250(next))
			return next
		},
		loadExecAbility: func(record *activeAbilityMembershipTestRecord4FC250) Ability {
			ability := record.ability
			w.record(fmt.Sprintf("exec-ability:%s=%d", record.name, ability))
			return ability
		},
	}
}

func activeAbilityMembershipWarriorWorld4FC250(unit uint64) activeAbilityMembershipTestWorld4FC250 {
	return activeAbilityMembershipTestWorld4FC250{
		unitArg:       unit,
		abilityArg:    AbilityBerserk,
		unitClasses:   map[uint64]uint8{unit: activeAbilityMembershipPlayerClass4FC250},
		updates:       map[uint64]string{unit: "update"},
		players:       map[string]string{"update": "player"},
		playerClasses: map[string]uint8{"player": activeAbilityMembershipWarrior4FC250},
		after:         make(map[string]func()),
	}
}

func TestActiveAbilityMembership4FC250TraversalOrderNativeIdentityAndInactiveMatch(t *testing.T) {
	const (
		unit     = uint64(0x1234567889abcdef)
		lowAlias = uint64(0x0000000089abcdef)
	)
	match := &activeAbilityMembershipTestRecord4FC250{
		name: "match", unit: unit, ability: Ability(math.MinInt32), active: 0, deadline: 0xfeedface,
	}
	wrongAbility := &activeAbilityMembershipTestRecord4FC250{
		name: "wrong-ability", unit: unit, ability: AbilityWarcry, active: 1, next: match,
	}
	low := &activeAbilityMembershipTestRecord4FC250{
		name: "low-alias", unit: lowAlias, ability: Ability(math.MinInt32), next: wrongAbility,
	}
	w := activeAbilityMembershipWarriorWorld4FC250(unit)
	w.abilityArg = Ability(math.MinInt32)
	w.head = low

	if got := activeAbilityMembership4FC250(w.hooks()); got != 1 {
		t.Fatalf("membership = %d, want canonical 1", got)
	}
	want := []string{
		"unit-arg=1234567889abcdef", "unit-class:1234567889abcdef=04",
		"update:1234567889abcdef=update", "player:update=player", "player-class:player=0",
		"head=low-alias", "ability-arg=-2147483648",
		"exec-unit:low-alias=0000000089abcdef", "exec-next:low-alias=wrong-ability",
		"exec-unit:wrong-ability=1234567889abcdef", "exec-next:wrong-ability=match", "exec-ability:wrong-ability=2",
		"exec-unit:match=1234567889abcdef", "exec-next:match=nil", "exec-ability:match=-2147483648",
	}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %q, want %q", w.events, want)
	}
	if match.active != 0 || match.deadline != 0xfeedface {
		t.Fatal("membership lookup inspected or changed Active/deadline state")
	}
}

func TestActiveAbilityMembership4FC250GatesNilUpdateAndRequiredPointerFaults(t *testing.T) {
	t.Run("nil unit faults on class read", func(t *testing.T) {
		w := activeAbilityMembershipWarriorWorld4FC250(0)
		defer func() {
			if recover() == nil {
				t.Fatal("nil unit did not fault")
			}
			want := []string{"unit-arg=0000000000000000", "unit-class:nil"}
			if !reflect.DeepEqual(w.events, want) {
				t.Fatalf("events = %q, want %q", w.events, want)
			}
		}()
		activeAbilityMembership4FC250(w.hooks())
	})

	t.Run("non-player avoids late reads", func(t *testing.T) {
		w := activeAbilityMembershipWarriorWorld4FC250(1)
		w.unitClasses[1] = 0x82
		if got := activeAbilityMembership4FC250(w.hooks()); got != 0 {
			t.Fatalf("membership = %d, want 0", got)
		}
		want := []string{"unit-arg=0000000000000001", "unit-class:0000000000000001=82"}
		if !reflect.DeepEqual(w.events, want) {
			t.Fatalf("events = %q, want %q", w.events, want)
		}
	})

	t.Run("nil UpdateData skips class gate", func(t *testing.T) {
		w := activeAbilityMembershipWarriorWorld4FC250(2)
		delete(w.updates, 2)
		w.head = &activeAbilityMembershipTestRecord4FC250{name: "match", unit: 2, ability: AbilityBerserk}
		if got := activeAbilityMembership4FC250(w.hooks()); got != 1 {
			t.Fatalf("membership = %d, want 1", got)
		}
		want := []string{
			"unit-arg=0000000000000002", "unit-class:0000000000000002=04", "update:0000000000000002=nil",
			"head=match", "ability-arg=1", "exec-unit:match=0000000000000002", "exec-next:match=nil", "exec-ability:match=1",
		}
		if !reflect.DeepEqual(w.events, want) {
			t.Fatalf("events = %q, want %q", w.events, want)
		}
	})

	t.Run("nil Player faults on class read", func(t *testing.T) {
		w := activeAbilityMembershipWarriorWorld4FC250(3)
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
		activeAbilityMembership4FC250(w.hooks())
	})

	t.Run("non-Warrior avoids head and ability", func(t *testing.T) {
		w := activeAbilityMembershipWarriorWorld4FC250(4)
		w.playerClasses["player"] = 2
		if got := activeAbilityMembership4FC250(w.hooks()); got != 0 {
			t.Fatalf("membership = %d, want 0", got)
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
		w := activeAbilityMembershipWarriorWorld4FC250(5)
		w.abilityArg = Ability(math.MinInt32)
		if got := activeAbilityMembership4FC250(w.hooks()); got != 0 {
			t.Fatalf("membership = %d, want 0", got)
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

func TestActiveAbilityMembership4FC250CachesUnitAndNextBeforeLiveAbility(t *testing.T) {
	const unit = uint64(7)
	tail := &activeAbilityMembershipTestRecord4FC250{name: "tail", unit: unit, ability: AbilityHarpoon}
	head := &activeAbilityMembershipTestRecord4FC250{name: "head", unit: unit, ability: AbilityWarcry, next: tail}
	decoy := &activeAbilityMembershipTestRecord4FC250{name: "decoy", unit: unit, ability: AbilityHarpoon}
	w := activeAbilityMembershipWarriorWorld4FC250(unit)
	w.abilityArg = AbilityHarpoon
	w.head = head
	w.after["exec-next:head=tail"] = func() {
		head.unit = 99
		head.ability = AbilityHarpoon
		head.next = decoy
	}

	if got := activeAbilityMembership4FC250(w.hooks()); got != 1 {
		t.Fatalf("membership = %d, want cached-unit/live-ability match", got)
	}
	wantSuffix := []string{
		"exec-unit:head=0000000000000007", "exec-next:head=tail", "exec-ability:head=3",
	}
	if got := w.events[len(w.events)-len(wantSuffix):]; !reflect.DeepEqual(got, wantSuffix) {
		t.Fatalf("event suffix = %q, want %q", got, wantSuffix)
	}
	if len(w.events) != 10 {
		t.Fatalf("events = %q; match should not traverse the cached tail", w.events)
	}
}

func TestActiveAbilityMembership4FC250CachedNextTraversalAndMiss(t *testing.T) {
	const unit = uint64(8)
	tail := &activeAbilityMembershipTestRecord4FC250{name: "tail", unit: unit, ability: AbilityWarcry}
	head := &activeAbilityMembershipTestRecord4FC250{name: "head", unit: 9, ability: AbilityHarpoon, next: tail}
	decoy := &activeAbilityMembershipTestRecord4FC250{name: "decoy", unit: unit, ability: AbilityBerserk}
	w := activeAbilityMembershipWarriorWorld4FC250(unit)
	w.abilityArg = AbilityHarpoon
	w.head = head
	w.after["exec-next:head=tail"] = func() { head.next = decoy }

	if got := activeAbilityMembership4FC250(w.hooks()); got != 0 {
		t.Fatalf("membership = %d, want canonical 0", got)
	}
	wantSuffix := []string{
		"exec-unit:head=0000000000000009", "exec-next:head=tail",
		"exec-unit:tail=0000000000000008", "exec-next:tail=nil", "exec-ability:tail=2",
	}
	if got := w.events[len(w.events)-len(wantSuffix):]; !reflect.DeepEqual(got, wantSuffix) {
		t.Fatalf("event suffix = %q, want %q", got, wantSuffix)
	}
}

func TestActiveAbilityMembership4FC250FaultPrefixes(t *testing.T) {
	const unit = uint64(9)
	all := []string{
		"unit-arg=0000000000000009", "unit-class:0000000000000009=04",
		"update:0000000000000009=update", "player:update=player", "player-class:player=0",
		"head=record", "ability-arg=1", "exec-unit:record=0000000000000009", "exec-next:record=nil", "exec-ability:record=1",
	}
	for faultAt := 1; faultAt <= len(all); faultAt++ {
		t.Run(fmt.Sprintf("fault-%d", faultAt), func(t *testing.T) {
			w := activeAbilityMembershipWarriorWorld4FC250(unit)
			w.head = &activeAbilityMembershipTestRecord4FC250{name: "record", unit: unit, ability: AbilityBerserk}
			w.faultAt = faultAt
			defer func() {
				if recover() == nil {
					t.Fatal("fault was not propagated")
				}
				if want := all[:faultAt]; !reflect.DeepEqual(w.events, want) {
					t.Fatalf("events = %q, want %q", w.events, want)
				}
			}()
			activeAbilityMembership4FC250(w.hooks())
		})
	}
}
