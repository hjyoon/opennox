package server

import (
	"reflect"
	"testing"
)

type unusedManaTransferTestPlayer4EAD50 struct {
	current uint16
	maximum uint16
}

type unusedManaTransferTestSource4EAD50 struct {
	mana int32
}

type unusedManaTransferTestTeam4EAD50 struct {
	id uint8
}

type unusedManaTransferTestObject4EAD50 struct {
	class        uint8
	playerUpdate *unusedManaTransferTestPlayer4EAD50
	sourceUpdate *unusedManaTransferTestSource4EAD50
	teamID       uint8
	hasTeam      int32
}

type unusedManaTransferTestState4EAD50 struct {
	t         *testing.T
	events    []string
	team      *unusedManaTransferTestTeam4EAD50
	contains  int32
	addResult uint16
	onAdd     func()
}

func (s *unusedManaTransferTestState4EAD50) hooks() unusedManaTransferHooks4EAD50[
	*unusedManaTransferTestObject4EAD50,
	*unusedManaTransferTestPlayer4EAD50,
	*unusedManaTransferTestSource4EAD50,
	*unusedManaTransferTestTeam4EAD50,
] {
	return unusedManaTransferHooks4EAD50[
		*unusedManaTransferTestObject4EAD50,
		*unusedManaTransferTestPlayer4EAD50,
		*unusedManaTransferTestSource4EAD50,
		*unusedManaTransferTestTeam4EAD50,
	]{
		loadClassByte: func(obj *unusedManaTransferTestObject4EAD50) uint8 {
			s.events = append(s.events, "class")
			return obj.class
		},
		loadPlayerUpdate: func(obj *unusedManaTransferTestObject4EAD50) *unusedManaTransferTestPlayer4EAD50 {
			s.events = append(s.events, "player-update")
			return obj.playerUpdate
		},
		loadManaCurrent: func(update *unusedManaTransferTestPlayer4EAD50) uint16 {
			s.events = append(s.events, "current")
			return update.current
		},
		loadSourceUpdate: func(obj *unusedManaTransferTestObject4EAD50) *unusedManaTransferTestSource4EAD50 {
			s.events = append(s.events, "source-update")
			return obj.sourceUpdate
		},
		loadManaMax: func(update *unusedManaTransferTestPlayer4EAD50) uint16 {
			s.events = append(s.events, "maximum")
			return update.maximum
		},
		loadSourceMana: func(update *unusedManaTransferTestSource4EAD50) int32 {
			s.events = append(s.events, "source-mana")
			return update.mana
		},
		hasTeam: func(obj *unusedManaTransferTestObject4EAD50) int32 {
			if obj == nil {
				s.t.Fatal("team query received nil object")
			}
			if obj.playerUpdate != nil {
				s.events = append(s.events, "target-has-team")
			} else {
				s.events = append(s.events, "source-has-team")
			}
			return obj.hasTeam
		},
		loadObjectTeamID: func(obj *unusedManaTransferTestObject4EAD50) uint8 {
			s.events = append(s.events, "object-team-id")
			return obj.teamID
		},
		findTeamByID: func(id uint8) *unusedManaTransferTestTeam4EAD50 {
			s.events = append(s.events, "find-team")
			if s.team != nil && id != s.team.id {
				s.t.Fatalf("team lookup ID = %#x, want %#x", id, s.team.id)
			}
			return s.team
		},
		loadTeamID: func(team *unusedManaTransferTestTeam4EAD50) uint8 {
			s.events = append(s.events, "team-id")
			return team.id
		},
		teamContains: func(obj *unusedManaTransferTestObject4EAD50, id uint8) int32 {
			s.events = append(s.events, "team-contains")
			if obj == nil || id != s.team.id {
				s.t.Fatalf("team containment args = %p/%#x", obj, id)
			}
			return s.contains
		},
		addPlayerMana: func(obj *unusedManaTransferTestObject4EAD50, amount int16) uint16 {
			s.events = append(s.events, "add-mana")
			if obj == nil || amount != 1 {
				s.t.Fatalf("mana add args = %p/%d", obj, amount)
			}
			if s.onAdd != nil {
				s.onAdd()
			}
			return s.addResult
		},
		storeSourceMana: func(update *unusedManaTransferTestSource4EAD50, mana int32) {
			s.events = append(s.events, "store-source-mana")
			update.mana = mana
		},
	}
}

func TestUnusedManaTransfer4EAD50TargetGuardsPrecedeSource(t *testing.T) {
	state := &unusedManaTransferTestState4EAD50{t: t}
	poisonSource := (*unusedManaTransferTestObject4EAD50)(nil)

	unusedManaTransfer4EAD50(
		poisonSource,
		(*unusedManaTransferTestObject4EAD50)(nil),
		state.hooks(),
	)
	if len(state.events) != 0 {
		t.Fatalf("nil-target events = %#v", state.events)
	}

	for _, class := range []uint8{0, 0x02, 0x80} {
		state.events = nil
		unusedManaTransfer4EAD50(
			poisonSource,
			&unusedManaTransferTestObject4EAD50{class: class},
			state.hooks(),
		)
		if want := []string{"class"}; !reflect.DeepEqual(state.events, want) {
			t.Fatalf("class %#x events = %#v, want %#v", class, state.events, want)
		}
	}
}

func TestUnusedManaTransfer4EAD50UntetheredOrderAndCachedLiveDecrement(t *testing.T) {
	playerUpdate := &unusedManaTransferTestPlayer4EAD50{current: 7, maximum: 9}
	original := &unusedManaTransferTestSource4EAD50{mana: 3}
	replacement := &unusedManaTransferTestSource4EAD50{mana: 99}
	source := &unusedManaTransferTestObject4EAD50{sourceUpdate: original}
	target := &unusedManaTransferTestObject4EAD50{class: 0x84, playerUpdate: playerUpdate}
	state := &unusedManaTransferTestState4EAD50{t: t, addResult: 0xbeef}
	hooks := state.hooks()
	originalLoadCurrent := hooks.loadManaCurrent
	hooks.loadManaCurrent = func(update *unusedManaTransferTestPlayer4EAD50) uint16 {
		current := originalLoadCurrent(update)
		target.playerUpdate = &unusedManaTransferTestPlayer4EAD50{current: 9, maximum: 0}
		return current
	}
	state.onAdd = func() {
		source.sourceUpdate = replacement
		original.mana = -0x80000000
	}

	unusedManaTransfer4EAD50(source, target, hooks)

	wantEvents := []string{
		"class", "player-update", "current", "source-update", "maximum",
		"source-mana", "source-has-team", "add-mana", "source-mana",
		"store-source-mana",
	}
	if !reflect.DeepEqual(state.events, wantEvents) {
		t.Fatalf("events = %#v, want %#v", state.events, wantEvents)
	}
	if original.mana != 0x7fffffff {
		t.Fatalf("cached source mana = %#x, want wrapping %#x", uint32(original.mana), uint32(0x7fffffff))
	}
	if replacement.mana != 99 {
		t.Fatalf("replacement source mana = %d, want 99", replacement.mana)
	}
}

func TestUnusedManaTransfer4EAD50TeamPathUsesLiveIDs(t *testing.T) {
	playerUpdate := &unusedManaTransferTestPlayer4EAD50{current: 1, maximum: 2}
	sourceUpdate := &unusedManaTransferTestSource4EAD50{mana: 2}
	source := &unusedManaTransferTestObject4EAD50{sourceUpdate: sourceUpdate, hasTeam: -1}
	target := &unusedManaTransferTestObject4EAD50{
		class: 0x04, playerUpdate: playerUpdate, teamID: 0x31, hasTeam: 2,
	}
	team := &unusedManaTransferTestTeam4EAD50{id: 0x31}
	state := &unusedManaTransferTestState4EAD50{t: t, team: team, contains: -7}
	hooks := state.hooks()
	originalHasTeam := hooks.hasTeam
	hooks.hasTeam = func(obj *unusedManaTransferTestObject4EAD50) int32 {
		result := originalHasTeam(obj)
		if obj == target {
			target.teamID = 0x42
			team.id = 0x42
		}
		return result
	}

	unusedManaTransfer4EAD50(source, target, hooks)

	wantEvents := []string{
		"class", "player-update", "current", "source-update", "maximum",
		"source-mana", "source-has-team", "target-has-team", "object-team-id",
		"find-team", "team-id", "team-contains", "add-mana", "source-mana",
		"store-source-mana",
	}
	if !reflect.DeepEqual(state.events, wantEvents) {
		t.Fatalf("events = %#v, want %#v", state.events, wantEvents)
	}
	if sourceUpdate.mana != 1 {
		t.Fatalf("source mana = %d, want 1", sourceUpdate.mana)
	}
}

func TestUnusedManaTransfer4EAD50Gates(t *testing.T) {
	tests := []struct {
		name       string
		current    uint16
		maximum    uint16
		sourceMana int32
		sourceTeam int32
		targetTeam int32
		team       *unusedManaTransferTestTeam4EAD50
		contains   int32
	}{
		{name: "equal mana", current: 5, maximum: 5, sourceMana: 1},
		{name: "above mana", current: 0xffff, maximum: 0, sourceMana: 1},
		{name: "zero source", current: 1, maximum: 2, sourceMana: 0},
		{name: "negative source", current: 1, maximum: 2, sourceMana: -1},
		{name: "target has no team", current: 1, maximum: 2, sourceMana: 1, sourceTeam: 1},
		{name: "team lookup misses", current: 1, maximum: 2, sourceMana: 1, sourceTeam: 1, targetTeam: 1},
		{name: "team excludes source", current: 1, maximum: 2, sourceMana: 1, sourceTeam: 1, targetTeam: 1, team: &unusedManaTransferTestTeam4EAD50{id: 7}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			playerUpdate := &unusedManaTransferTestPlayer4EAD50{current: tc.current, maximum: tc.maximum}
			sourceUpdate := &unusedManaTransferTestSource4EAD50{mana: tc.sourceMana}
			source := &unusedManaTransferTestObject4EAD50{sourceUpdate: sourceUpdate, hasTeam: tc.sourceTeam}
			target := &unusedManaTransferTestObject4EAD50{
				class: 0x04, playerUpdate: playerUpdate, teamID: 7, hasTeam: tc.targetTeam,
			}
			state := &unusedManaTransferTestState4EAD50{t: t, team: tc.team, contains: tc.contains}
			unusedManaTransfer4EAD50(source, target, state.hooks())
			for _, event := range state.events {
				if event == "add-mana" || event == "store-source-mana" {
					t.Fatalf("gate called mutation; events = %#v", state.events)
				}
			}
			if sourceUpdate.mana != tc.sourceMana {
				t.Fatalf("source mana = %d, want %d", sourceUpdate.mana, tc.sourceMana)
			}
		})
	}
}

func TestUnusedManaTransfer4EAD50FaultOrder(t *testing.T) {
	t.Run("nil player update faults before source", func(t *testing.T) {
		state := &unusedManaTransferTestState4EAD50{t: t}
		var recovered any
		func() {
			defer func() { recovered = recover() }()
			unusedManaTransfer4EAD50(
				(*unusedManaTransferTestObject4EAD50)(nil),
				&unusedManaTransferTestObject4EAD50{class: 4},
				state.hooks(),
			)
		}()
		if recovered == nil {
			t.Fatal("nil player update did not fault")
		}
		if want := []string{"class", "player-update", "current"}; !reflect.DeepEqual(state.events, want) {
			t.Fatalf("events = %#v, want %#v", state.events, want)
		}
	})

	t.Run("nil source faults after target current", func(t *testing.T) {
		state := &unusedManaTransferTestState4EAD50{t: t}
		target := &unusedManaTransferTestObject4EAD50{
			class: 4, playerUpdate: &unusedManaTransferTestPlayer4EAD50{current: 1, maximum: 2},
		}
		var recovered any
		func() {
			defer func() { recovered = recover() }()
			unusedManaTransfer4EAD50(
				(*unusedManaTransferTestObject4EAD50)(nil), target, state.hooks(),
			)
		}()
		if recovered == nil {
			t.Fatal("nil source did not fault")
		}
		if want := []string{"class", "player-update", "current", "source-update"}; !reflect.DeepEqual(state.events, want) {
			t.Fatalf("events = %#v, want %#v", state.events, want)
		}
	})

	t.Run("full target suppresses nil source update dereference", func(t *testing.T) {
		state := &unusedManaTransferTestState4EAD50{t: t}
		source := &unusedManaTransferTestObject4EAD50{}
		target := &unusedManaTransferTestObject4EAD50{
			class: 4, playerUpdate: &unusedManaTransferTestPlayer4EAD50{current: 2, maximum: 2},
		}
		unusedManaTransfer4EAD50(source, target, state.hooks())
		want := []string{"class", "player-update", "current", "source-update", "maximum"}
		if !reflect.DeepEqual(state.events, want) {
			t.Fatalf("events = %#v, want %#v", state.events, want)
		}
	})

	t.Run("non-full target faults on nil source update", func(t *testing.T) {
		state := &unusedManaTransferTestState4EAD50{t: t}
		source := &unusedManaTransferTestObject4EAD50{}
		target := &unusedManaTransferTestObject4EAD50{
			class: 4, playerUpdate: &unusedManaTransferTestPlayer4EAD50{current: 1, maximum: 2},
		}
		var recovered any
		func() {
			defer func() { recovered = recover() }()
			unusedManaTransfer4EAD50(source, target, state.hooks())
		}()
		if recovered == nil {
			t.Fatal("nil source update did not fault")
		}
		want := []string{"class", "player-update", "current", "source-update", "maximum", "source-mana"}
		if !reflect.DeepEqual(state.events, want) {
			t.Fatalf("events = %#v, want %#v", state.events, want)
		}
	})
}
