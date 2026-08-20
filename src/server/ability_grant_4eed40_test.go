package server

import (
	"fmt"
	"reflect"
	"testing"
)

type abilityGivePlayerAllTestPlayer4EED40 struct {
	name   string
	levels [137]uint32
}

type abilityGivePlayerAllTestUpdate4EED40 struct {
	name   string
	player *abilityGivePlayerAllTestPlayer4EED40
}

type abilityGivePlayerAllTestObject4EED40 struct {
	name   string
	update *abilityGivePlayerAllTestUpdate4EED40
}

type abilityGivePlayerAllTestReward4EED40 struct {
	unit      *abilityGivePlayerAllTestObject4EED40
	ability   int32
	rewardArg int32
}

type abilityGivePlayerAllTestWorld4EED40 struct {
	unit       *abilityGivePlayerAllTestObject4EED40
	count      int8
	rewardArg  int32
	table      map[int32]uint32
	game       int32
	quest      int32
	questState int32
	events     []string
	rewards    []abilityGivePlayerAllTestReward4EED40
	faultAt    int
	after      map[string]func()
}

func abilityGivePlayerAllObjectName4EED40(unit *abilityGivePlayerAllTestObject4EED40) string {
	if unit == nil {
		return "nil"
	}
	return unit.name
}

func abilityGivePlayerAllUpdateName4EED40(update *abilityGivePlayerAllTestUpdate4EED40) string {
	if update == nil {
		return "nil"
	}
	return update.name
}

func abilityGivePlayerAllPlayerName4EED40(player *abilityGivePlayerAllTestPlayer4EED40) string {
	if player == nil {
		return "nil"
	}
	return player.name
}

func (w *abilityGivePlayerAllTestWorld4EED40) record(event string) {
	w.events = append(w.events, event)
	if w.faultAt != 0 && len(w.events) == w.faultAt {
		panic(event)
	}
}

func (w *abilityGivePlayerAllTestWorld4EED40) finish(event string) {
	if after := w.after[event]; after != nil {
		after()
	}
}

func (w *abilityGivePlayerAllTestWorld4EED40) hooks() abilityGivePlayerAllHooks4EED40[
	*abilityGivePlayerAllTestObject4EED40,
	*abilityGivePlayerAllTestUpdate4EED40,
	*abilityGivePlayerAllTestPlayer4EED40,
] {
	return abilityGivePlayerAllHooks4EED40[
		*abilityGivePlayerAllTestObject4EED40,
		*abilityGivePlayerAllTestUpdate4EED40,
		*abilityGivePlayerAllTestPlayer4EED40,
	]{
		loadUnitArg: func() *abilityGivePlayerAllTestObject4EED40 {
			unit := w.unit
			event := "arg:" + abilityGivePlayerAllObjectName4EED40(unit)
			w.record(event)
			w.finish(event)
			return unit
		},
		loadUpdateData: func(unit *abilityGivePlayerAllTestObject4EED40) *abilityGivePlayerAllTestUpdate4EED40 {
			update := unit.update
			event := "update:" + abilityGivePlayerAllObjectName4EED40(unit) + "=" + abilityGivePlayerAllUpdateName4EED40(update)
			w.record(event)
			w.finish(event)
			return update
		},
		loadCountLow: func() int8 {
			count := w.count
			event := fmt.Sprintf("count:%d", count)
			w.record(event)
			w.finish(event)
			return count
		},
		loadPlayer: func(update *abilityGivePlayerAllTestUpdate4EED40) *abilityGivePlayerAllTestPlayer4EED40 {
			if update == nil {
				event := "player:nil-update"
				w.record(event)
				panic(event)
			}
			player := update.player
			event := "player:" + abilityGivePlayerAllUpdateName4EED40(update) + "=" + abilityGivePlayerAllPlayerName4EED40(player)
			w.record(event)
			w.finish(event)
			return player
		},
		loadAbilityID: func(index int32) uint32 {
			ability := w.table[index]
			event := fmt.Sprintf("ability:%d=%#x", index, ability)
			w.record(event)
			w.finish(event)
			return ability
		},
		gameFlagsCheck: func(mask uint32) int32 {
			result := w.game
			event := fmt.Sprintf("game:%#x=%d", mask, result)
			w.record(event)
			w.finish(event)
			return result
		},
		isQuest: func() int32 {
			result := w.quest
			event := fmt.Sprintf("quest:%d", result)
			w.record(event)
			w.finish(event)
			return result
		},
		questMode: func() int32 {
			result := w.questState
			event := fmt.Sprintf("quest-mode:%d", result)
			w.record(event)
			w.finish(event)
			return result
		},
		loadRewardArg: func() int32 {
			rewardArg := w.rewardArg
			event := fmt.Sprintf("reward-arg:%d", rewardArg)
			w.record(event)
			w.finish(event)
			return rewardArg
		},
		rewardAbility: func(unit *abilityGivePlayerAllTestObject4EED40, ability, rewardArg int32) {
			event := fmt.Sprintf("reward:%s:%d:%d", abilityGivePlayerAllObjectName4EED40(unit), ability, rewardArg)
			w.record(event)
			w.rewards = append(w.rewards, abilityGivePlayerAllTestReward4EED40{
				unit: unit, ability: ability, rewardArg: rewardArg,
			})
			w.finish(event)
		},
		storeAbilityLevel: func(player *abilityGivePlayerAllTestPlayer4EED40, index int32, value uint32) {
			event := fmt.Sprintf("store:%s:%d=%d", abilityGivePlayerAllPlayerName4EED40(player), index, value)
			w.record(event)
			if player == nil {
				panic("store:nil-player")
			}
			player.levels[index] = value
			w.finish(event)
		},
	}
}

func newAbilityGivePlayerAllTestWorld4EED40() *abilityGivePlayerAllTestWorld4EED40 {
	player := &abilityGivePlayerAllTestPlayer4EED40{name: "player"}
	update := &abilityGivePlayerAllTestUpdate4EED40{name: "update", player: player}
	return &abilityGivePlayerAllTestWorld4EED40{
		unit:      &abilityGivePlayerAllTestObject4EED40{name: "unit", update: update},
		count:     1,
		rewardArg: 7,
		table:     map[int32]uint32{0: 1},
		after:     make(map[string]func()),
	}
}

func abilityGivePlayerAllRewardEvents4EED40() []string {
	return []string{
		"arg:unit",
		"update:unit=update",
		"count:1",
		"player:update=player",
		"ability:0=0x1",
		"game:0x1000=0",
		"quest:0",
		"quest-mode:0",
		"reward-arg:7",
		"ability:0=0x1",
		"reward:unit:1:7",
	}
}

func TestAbilityGivePlayerAll4EED40EntryOrderAndSignedCount(t *testing.T) {
	t.Run("nil unit", func(t *testing.T) {
		world := newAbilityGivePlayerAllTestWorld4EED40()
		world.unit = nil
		abilityGivePlayerAll4EED40(world.hooks())
		if want := []string{"arg:nil"}; !reflect.DeepEqual(world.events, want) {
			t.Fatalf("events = %#v, want %#v", world.events, want)
		}
	})

	for _, count := range []int8{0, -1, -128} {
		t.Run(fmt.Sprintf("count %d", count), func(t *testing.T) {
			world := newAbilityGivePlayerAllTestWorld4EED40()
			world.count = count
			abilityGivePlayerAll4EED40(world.hooks())
			want := []string{"arg:unit", "update:unit=update", fmt.Sprintf("count:%d", count), "player:update=player"}
			if !reflect.DeepEqual(world.events, want) {
				t.Fatalf("events = %#v, want %#v", world.events, want)
			}
		})
	}

	t.Run("positive int8 maximum", func(t *testing.T) {
		world := newAbilityGivePlayerAllTestWorld4EED40()
		world.count = 127
		world.table = make(map[int32]uint32)
		abilityGivePlayerAll4EED40(world.hooks())
		if len(world.events) != 4+127 || world.events[len(world.events)-1] != "ability:126=0x0" {
			t.Fatalf("events count/tail = %d/%q, want 131/ability:126=0x0", len(world.events), world.events[len(world.events)-1])
		}
	})
}

func TestAbilityGivePlayerAll4EED40LoadsPlayerBeforeCountGateAndHasNoNilGuards(t *testing.T) {
	t.Run("nil update faults after count", func(t *testing.T) {
		world := newAbilityGivePlayerAllTestWorld4EED40()
		world.unit.update = nil
		world.count = -1
		defer func() {
			if recover() == nil {
				t.Fatal("nil update did not fault")
			}
			want := []string{"arg:unit", "update:unit=nil", "count:-1", "player:nil-update"}
			if !reflect.DeepEqual(world.events, want) {
				t.Fatalf("events = %#v, want %#v", world.events, want)
			}
		}()
		abilityGivePlayerAll4EED40(world.hooks())
	})

	t.Run("nil player survives nonpositive count", func(t *testing.T) {
		world := newAbilityGivePlayerAllTestWorld4EED40()
		world.unit.update.player = nil
		world.count = 0
		abilityGivePlayerAll4EED40(world.hooks())
		want := []string{"arg:unit", "update:unit=update", "count:0", "player:update=nil"}
		if !reflect.DeepEqual(world.events, want) {
			t.Fatalf("events = %#v, want %#v", world.events, want)
		}
	})
}

func TestAbilityGivePlayerAll4EED40ZeroEntriesAndModeShortCircuit(t *testing.T) {
	tests := []struct {
		name       string
		ability    uint32
		game       int32
		quest      int32
		questState int32
		wantTail   []string
		wantLevel  uint32
		wantReward bool
	}{
		{name: "zero ability", ability: 0, wantTail: []string{"ability:0=0x0"}, wantLevel: 9},
		{name: "game flag", ability: 1, game: -1, wantTail: []string{"ability:0=0x1", "game:0x1000=-1", "store:player:0=0"}},
		{name: "quest", ability: 2, quest: 2, wantTail: []string{"ability:0=0x2", "game:0x1000=0", "quest:2", "store:player:0=0"}},
		{name: "quest state", ability: 4, questState: -3, wantTail: []string{"ability:0=0x4", "game:0x1000=0", "quest:0", "quest-mode:-3", "store:player:0=0"}},
		{name: "reward", ability: 5, wantTail: []string{"ability:0=0x5", "game:0x1000=0", "quest:0", "quest-mode:0", "reward-arg:7", "ability:0=0x5", "reward:unit:5:7"}, wantLevel: 9, wantReward: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			world := newAbilityGivePlayerAllTestWorld4EED40()
			world.table[0] = test.ability
			world.game, world.quest, world.questState = test.game, test.quest, test.questState
			world.unit.update.player.levels[0] = 9
			abilityGivePlayerAll4EED40(world.hooks())
			want := append([]string{"arg:unit", "update:unit=update", "count:1", "player:update=player"}, test.wantTail...)
			if !reflect.DeepEqual(world.events, want) {
				t.Fatalf("events = %#v, want %#v", world.events, want)
			}
			if got := world.unit.update.player.levels[0]; got != test.wantLevel {
				t.Fatalf("level = %d, want %d", got, test.wantLevel)
			}
			if got := len(world.rewards) != 0; got != test.wantReward {
				t.Fatalf("reward present = %t, want %t", got, test.wantReward)
			}
		})
	}
}

func TestAbilityGivePlayerAll4EED40CachesUnitPlayerAndReloadsRewardInputs(t *testing.T) {
	world := newAbilityGivePlayerAllTestWorld4EED40()
	unit1 := world.unit
	player1 := unit1.update.player
	unit2 := &abilityGivePlayerAllTestObject4EED40{name: "unit-2"}
	player2 := &abilityGivePlayerAllTestPlayer4EED40{name: "player-2"}
	world.after["player:update=player"] = func() {
		unit1.update.player = player2
		world.unit = unit2
		world.count = 2
	}
	world.after["quest-mode:0"] = func() {
		world.table[0] = 5
		world.rewardArg = -9
	}
	abilityGivePlayerAll4EED40(world.hooks())
	if len(world.rewards) != 1 {
		t.Fatalf("rewards = %#v, want one", world.rewards)
	}
	got := world.rewards[0]
	if got.unit != unit1 || got.ability != 5 || got.rewardArg != -9 {
		t.Fatalf("reward = %#v, want cached unit/reloaded 5/-9", got)
	}
	if player1 == player2 {
		t.Fatal("test players unexpectedly alias")
	}
	if events := world.events; events[len(events)-1] != "reward:unit:5:-9" {
		t.Fatalf("tail event = %q, want reward:unit:5:-9", events[len(events)-1])
	}
}

func TestAbilityGivePlayerAll4EED40RestrictedStoreUsesCachedPlayer(t *testing.T) {
	world := newAbilityGivePlayerAllTestWorld4EED40()
	player1 := world.unit.update.player
	player2 := &abilityGivePlayerAllTestPlayer4EED40{name: "player-2"}
	player1.levels[0], player2.levels[0] = 11, 22
	world.game = 1
	world.after["game:0x1000=1"] = func() { world.unit.update.player = player2 }
	abilityGivePlayerAll4EED40(world.hooks())
	if player1.levels[0] != 0 || player2.levels[0] != 22 {
		t.Fatalf("cached/live player levels = %d/%d, want 0/22", player1.levels[0], player2.levels[0])
	}
}

func TestAbilityGivePlayerAll4EED40RewardPathFaultPrefixes(t *testing.T) {
	want := abilityGivePlayerAllRewardEvents4EED40()
	for faultAt := 1; faultAt <= len(want); faultAt++ {
		t.Run(fmt.Sprintf("fault %d", faultAt), func(t *testing.T) {
			world := newAbilityGivePlayerAllTestWorld4EED40()
			world.faultAt = faultAt
			func() {
				defer func() {
					if recover() == nil {
						t.Fatalf("fault %d did not stop execution", faultAt)
					}
				}()
				abilityGivePlayerAll4EED40(world.hooks())
			}()
			if prefix := want[:faultAt]; !reflect.DeepEqual(world.events, prefix) {
				t.Fatalf("events = %#v, want prefix %#v", world.events, prefix)
			}
		})
	}
}

func TestAbilityGivePlayerAll4EED40RestrictedPathFaultPrefixes(t *testing.T) {
	want := []string{
		"arg:unit",
		"update:unit=update",
		"count:1",
		"player:update=player",
		"ability:0=0x1",
		"game:0x1000=1",
		"store:player:0=0",
	}
	for faultAt := 1; faultAt <= len(want); faultAt++ {
		t.Run(fmt.Sprintf("fault %d", faultAt), func(t *testing.T) {
			world := newAbilityGivePlayerAllTestWorld4EED40()
			world.game = 1
			world.faultAt = faultAt
			func() {
				defer func() {
					if recover() == nil {
						t.Fatalf("fault %d did not stop execution", faultAt)
					}
				}()
				abilityGivePlayerAll4EED40(world.hooks())
			}()
			if prefix := want[:faultAt]; !reflect.DeepEqual(world.events, prefix) {
				t.Fatalf("events = %#v, want prefix %#v", world.events, prefix)
			}
		})
	}
}
