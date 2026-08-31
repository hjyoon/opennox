package server

import (
	"fmt"
	"reflect"
	"testing"
)

type activeAbilityUnitMembershipTestRecord4FC2B0 struct {
	name     string
	unit     uint64
	ability  Ability
	frame    uint32
	active   uint32
	next     *activeAbilityUnitMembershipTestRecord4FC2B0
	previous *activeAbilityUnitMembershipTestRecord4FC2B0
}

type activeAbilityUnitMembershipTestWorld4FC2B0 struct {
	unitArg       uint64
	unitClasses   map[uint64]uint8
	updates       map[uint64]string
	players       map[string]string
	playerClasses map[string]uint8
	head          *activeAbilityUnitMembershipTestRecord4FC2B0
	events        []string
	faultAt       int
	after         map[string]func()
}

func activeAbilityUnitMembershipRecordName4FC2B0(record *activeAbilityUnitMembershipTestRecord4FC2B0) string {
	if record == nil {
		return "nil"
	}
	return record.name
}

func activeAbilityUnitMembershipOpaqueName4FC2B0(value string) string {
	if value == "" {
		return "nil"
	}
	return value
}

func (w *activeAbilityUnitMembershipTestWorld4FC2B0) record(event string) {
	w.events = append(w.events, event)
	if w.faultAt != 0 && len(w.events) == w.faultAt {
		panic(event)
	}
	if after := w.after[event]; after != nil {
		after()
	}
}

func (w *activeAbilityUnitMembershipTestWorld4FC2B0) hooks() activeAbilityUnitMembershipHooks4FC2B0[
	uint64,
	string,
	string,
	*activeAbilityUnitMembershipTestRecord4FC2B0,
] {
	return activeAbilityUnitMembershipHooks4FC2B0[
		uint64,
		string,
		string,
		*activeAbilityUnitMembershipTestRecord4FC2B0,
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
			w.record(fmt.Sprintf("update:%016x=%s", unit, activeAbilityUnitMembershipOpaqueName4FC2B0(update)))
			return update
		},
		loadPlayer: func(update string) string {
			player := w.players[update]
			w.record("player:" + update + "=" + activeAbilityUnitMembershipOpaqueName4FC2B0(player))
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
		loadExecHead: func() *activeAbilityUnitMembershipTestRecord4FC2B0 {
			head := w.head
			w.record("head=" + activeAbilityUnitMembershipRecordName4FC2B0(head))
			return head
		},
		loadExecUnit: func(record *activeAbilityUnitMembershipTestRecord4FC2B0) uint64 {
			unit := record.unit
			w.record(fmt.Sprintf("exec-unit:%s=%016x", record.name, unit))
			return unit
		},
		loadExecNext: func(record *activeAbilityUnitMembershipTestRecord4FC2B0) *activeAbilityUnitMembershipTestRecord4FC2B0 {
			next := record.next
			w.record("exec-next:" + record.name + "=" + activeAbilityUnitMembershipRecordName4FC2B0(next))
			return next
		},
	}
}

func activeAbilityUnitMembershipWarriorWorld4FC2B0(unit uint64) activeAbilityUnitMembershipTestWorld4FC2B0 {
	return activeAbilityUnitMembershipTestWorld4FC2B0{
		unitArg:       unit,
		unitClasses:   map[uint64]uint8{unit: activeAbilityUnitMembershipPlayerClass4FC2B0},
		updates:       map[uint64]string{unit: "update"},
		players:       map[string]string{"update": "player"},
		playerClasses: map[string]uint8{"player": activeAbilityUnitMembershipWarrior4FC2B0},
		after:         make(map[string]func()),
	}
}

func TestActiveAbilityUnitMembership4FC2B0NativeIdentityOrderAndIgnoredFields(t *testing.T) {
	const (
		unit     = uint64(0x1234567889abcdef)
		lowAlias = uint64(0x0000000089abcdef)
	)
	match := &activeAbilityUnitMembershipTestRecord4FC2B0{
		name: "match", unit: unit, ability: Ability(-2147483648), frame: 0xfeedface, active: 0,
	}
	low := &activeAbilityUnitMembershipTestRecord4FC2B0{
		name: "low-alias", unit: lowAlias, ability: AbilityWarcry, frame: 7, active: 1, next: match,
	}
	match.previous = low
	w := activeAbilityUnitMembershipWarriorWorld4FC2B0(unit)
	w.head = low

	if got := activeAbilityUnitMembership4FC2B0(w.hooks()); got != 1 {
		t.Fatalf("membership = %d, want canonical 1", got)
	}
	want := []string{
		"unit-arg=1234567889abcdef", "unit-class:1234567889abcdef=04",
		"update:1234567889abcdef=update", "player:update=player", "player-class:player=0",
		"head=low-alias", "exec-unit:low-alias=0000000089abcdef", "exec-next:low-alias=match",
		"exec-unit:match=1234567889abcdef",
	}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %q, want %q", w.events, want)
	}
	if low.next != match || match.previous != low || match.ability != Ability(-2147483648) ||
		match.frame != 0xfeedface || match.active != 0 {
		t.Fatal("unit membership inspected or changed ignored record fields or topology")
	}
}

func TestActiveAbilityUnitMembership4FC2B0GatesAndRequiredPointerFaults(t *testing.T) {
	t.Run("nil unit faults on class read", func(t *testing.T) {
		w := activeAbilityUnitMembershipWarriorWorld4FC2B0(0)
		defer func() {
			if recover() == nil {
				t.Fatal("nil unit did not fault")
			}
			want := []string{"unit-arg=0000000000000000", "unit-class:nil"}
			if !reflect.DeepEqual(w.events, want) {
				t.Fatalf("events = %q, want %q", w.events, want)
			}
		}()
		activeAbilityUnitMembership4FC2B0(w.hooks())
	})

	t.Run("non-player avoids late reads", func(t *testing.T) {
		w := activeAbilityUnitMembershipWarriorWorld4FC2B0(1)
		w.unitClasses[1] = 0x82
		if got := activeAbilityUnitMembership4FC2B0(w.hooks()); got != 0 {
			t.Fatalf("membership = %d, want 0", got)
		}
		want := []string{"unit-arg=0000000000000001", "unit-class:0000000000000001=82"}
		if !reflect.DeepEqual(w.events, want) {
			t.Fatalf("events = %q, want %q", w.events, want)
		}
	})

	t.Run("nil UpdateData skips Player class", func(t *testing.T) {
		w := activeAbilityUnitMembershipWarriorWorld4FC2B0(2)
		delete(w.updates, 2)
		w.head = &activeAbilityUnitMembershipTestRecord4FC2B0{name: "match", unit: 2}
		if got := activeAbilityUnitMembership4FC2B0(w.hooks()); got != 1 {
			t.Fatalf("membership = %d, want 1", got)
		}
		want := []string{
			"unit-arg=0000000000000002", "unit-class:0000000000000002=04",
			"update:0000000000000002=nil", "head=match", "exec-unit:match=0000000000000002",
		}
		if !reflect.DeepEqual(w.events, want) {
			t.Fatalf("events = %q, want %q", w.events, want)
		}
	})

	t.Run("nil Player faults on class read", func(t *testing.T) {
		w := activeAbilityUnitMembershipWarriorWorld4FC2B0(3)
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
		activeAbilityUnitMembership4FC2B0(w.hooks())
	})

	t.Run("non-Warrior avoids head", func(t *testing.T) {
		w := activeAbilityUnitMembershipWarriorWorld4FC2B0(4)
		w.playerClasses["player"] = 2
		if got := activeAbilityUnitMembership4FC2B0(w.hooks()); got != 0 {
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

	t.Run("nil head misses", func(t *testing.T) {
		w := activeAbilityUnitMembershipWarriorWorld4FC2B0(5)
		if got := activeAbilityUnitMembership4FC2B0(w.hooks()); got != 0 {
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

func TestActiveAbilityUnitMembership4FC2B0ReadsLiveNextOnlyAfterMismatch(t *testing.T) {
	const unit = uint64(6)
	match := &activeAbilityUnitMembershipTestRecord4FC2B0{name: "match", unit: unit}
	stale := &activeAbilityUnitMembershipTestRecord4FC2B0{name: "stale", unit: 7}
	head := &activeAbilityUnitMembershipTestRecord4FC2B0{name: "head", unit: 8, next: stale}
	w := activeAbilityUnitMembershipWarriorWorld4FC2B0(unit)
	w.head = head
	w.after["exec-unit:head=0000000000000008"] = func() { head.next = match }

	if got := activeAbilityUnitMembership4FC2B0(w.hooks()); got != 1 {
		t.Fatalf("membership = %d, want live-Next match", got)
	}
	wantSuffix := []string{
		"exec-unit:head=0000000000000008", "exec-next:head=match", "exec-unit:match=0000000000000006",
	}
	if got := w.events[len(w.events)-len(wantSuffix):]; !reflect.DeepEqual(got, wantSuffix) {
		t.Fatalf("event suffix = %q, want %q", got, wantSuffix)
	}
}

func TestActiveAbilityUnitMembership4FC2B0MatchAvoidsNextAndMissesCanonically(t *testing.T) {
	const unit = uint64(7)

	t.Run("match avoids Next", func(t *testing.T) {
		w := activeAbilityUnitMembershipWarriorWorld4FC2B0(unit)
		w.head = &activeAbilityUnitMembershipTestRecord4FC2B0{name: "match", unit: unit}
		w.after["exec-unit:match=0000000000000007"] = func() { w.faultAt = len(w.events) + 1 }
		if got := activeAbilityUnitMembership4FC2B0(w.hooks()); got != 1 {
			t.Fatalf("membership = %d, want 1", got)
		}
		if got := w.events[len(w.events)-1]; got != "exec-unit:match=0000000000000007" {
			t.Fatalf("last event = %q, want matching Unit read", got)
		}
	})

	t.Run("miss returns canonical zero", func(t *testing.T) {
		w := activeAbilityUnitMembershipWarriorWorld4FC2B0(unit)
		w.head = &activeAbilityUnitMembershipTestRecord4FC2B0{name: "miss", unit: 8}
		if got := activeAbilityUnitMembership4FC2B0(w.hooks()); got != 0 {
			t.Fatalf("membership = %d, want 0", got)
		}
		wantSuffix := []string{"exec-unit:miss=0000000000000008", "exec-next:miss=nil"}
		if got := w.events[len(w.events)-len(wantSuffix):]; !reflect.DeepEqual(got, wantSuffix) {
			t.Fatalf("event suffix = %q, want %q", got, wantSuffix)
		}
	})
}

func TestActiveAbilityUnitMembership4FC2B0FaultPrefixes(t *testing.T) {
	const unit = uint64(9)
	all := []string{
		"unit-arg=0000000000000009", "unit-class:0000000000000009=04",
		"update:0000000000000009=update", "player:update=player", "player-class:player=0",
		"head=record", "exec-unit:record=0000000000000008", "exec-next:record=nil",
	}
	for faultAt := 1; faultAt <= len(all); faultAt++ {
		t.Run(fmt.Sprintf("fault-%d", faultAt), func(t *testing.T) {
			w := activeAbilityUnitMembershipWarriorWorld4FC2B0(unit)
			w.head = &activeAbilityUnitMembershipTestRecord4FC2B0{name: "record", unit: 8}
			w.faultAt = faultAt
			defer func() {
				if recover() == nil {
					t.Fatal("fault was not propagated")
				}
				if want := all[:faultAt]; !reflect.DeepEqual(w.events, want) {
					t.Fatalf("events = %q, want %q", w.events, want)
				}
			}()
			activeAbilityUnitMembership4FC2B0(w.hooks())
		})
	}
}
