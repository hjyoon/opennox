package server

import (
	"fmt"
	"math"
	"reflect"
	"testing"
)

type playerUnitInitObject4EFE80 struct {
	name   string
	update *playerUnitInitUpdate4EFE80
}

type playerUnitInitUpdate4EFE80 struct {
	name       string
	player     *playerUnitInitPlayer4EFE80
	extraLives int32
}

type playerUnitInitPlayer4EFE80 struct {
	name string
}

type playerUnitInitWorld4EFE80 struct {
	unitArg       *playerUnitInitObject4EFE80
	gold          uint32
	questFlag     int32
	balance       float32
	converted     int32
	defaultResult uint8
	events        []string
	faultAt       int
	after         map[string]func()
}

func playerUnitInitObjectName4EFE80(value *playerUnitInitObject4EFE80) string {
	if value == nil {
		return "nil"
	}
	return value.name
}

func playerUnitInitUpdateName4EFE80(value *playerUnitInitUpdate4EFE80) string {
	if value == nil {
		return "nil"
	}
	return value.name
}

func playerUnitInitPlayerName4EFE80(value *playerUnitInitPlayer4EFE80) string {
	if value == nil {
		return "nil"
	}
	return value.name
}

func (w *playerUnitInitWorld4EFE80) record(event string) {
	w.events = append(w.events, event)
	if w.faultAt != 0 && len(w.events) == w.faultAt {
		panic(event)
	}
	if after := w.after[event]; after != nil {
		after()
	}
}

func (w *playerUnitInitWorld4EFE80) hooks() playerUnitInitHooks4EFE80[
	*playerUnitInitObject4EFE80,
	*playerUnitInitUpdate4EFE80,
	*playerUnitInitPlayer4EFE80,
] {
	return playerUnitInitHooks4EFE80[
		*playerUnitInitObject4EFE80,
		*playerUnitInitUpdate4EFE80,
		*playerUnitInitPlayer4EFE80,
	]{
		loadUnitArg: func() *playerUnitInitObject4EFE80 {
			unit := w.unitArg
			w.record("arg:" + playerUnitInitObjectName4EFE80(unit))
			return unit
		},
		loadUpdateData: func(unit *playerUnitInitObject4EFE80) *playerUnitInitUpdate4EFE80 {
			if unit == nil {
				w.record("update:nil")
				panic("update:nil")
			}
			update := unit.update
			w.record("update:" + unit.name + "=" + playerUnitInitUpdateName4EFE80(update))
			return update
		},
		getGold: func(unit *playerUnitInitObject4EFE80) uint32 {
			w.record(fmt.Sprintf("get-gold:%s=%08x", playerUnitInitObjectName4EFE80(unit), w.gold))
			return w.gold
		},
		subGold: func(unit *playerUnitInitObject4EFE80, value uint32) {
			w.record(fmt.Sprintf("sub-gold:%s:%08x", playerUnitInitObjectName4EFE80(unit), value))
		},
		syncLevel: func(unit *playerUnitInitObject4EFE80) {
			w.record("sync:" + playerUnitInitObjectName4EFE80(unit))
		},
		loadPlayer: func(update *playerUnitInitUpdate4EFE80) *playerUnitInitPlayer4EFE80 {
			if update == nil {
				w.record("player:nil")
				panic("player:nil")
			}
			player := update.player
			w.record("player:" + update.name + "=" + playerUnitInitPlayerName4EFE80(player))
			return player
		},
		awardBeastScrolls: func(player *playerUnitInitPlayer4EFE80) {
			event := "award-scroll:" + playerUnitInitPlayerName4EFE80(player)
			w.record(event)
			if player == nil {
				panic(event)
			}
		},
		awardSpells: func(player *playerUnitInitPlayer4EFE80) {
			event := "award-spell:" + playerUnitInitPlayerName4EFE80(player)
			w.record(event)
			if player == nil {
				panic(event)
			}
		},
		readValues: func(unit *playerUnitInitObject4EFE80, reward int32) {
			w.record(fmt.Sprintf("read-values:%s:%d", playerUnitInitObjectName4EFE80(unit), reward))
		},
		awardWarriorAbilities: func(player *playerUnitInitPlayer4EFE80) {
			event := "award-ability:" + playerUnitInitPlayerName4EFE80(player)
			w.record(event)
			if player == nil {
				panic(event)
			}
		},
		gameFlag: func(flag uint32) int32 {
			w.record(fmt.Sprintf("flag:%08x=%d", flag, w.questFlag))
			return w.questFlag
		},
		balanceFloat: func(key string) float32 {
			w.record(fmt.Sprintf("balance:%s=%08x", key, math.Float32bits(w.balance)))
			return w.balance
		},
		floatToInt: func(value float32) int32 {
			w.record(fmt.Sprintf("float-to-int:%08x=%d", math.Float32bits(value), w.converted))
			return w.converted
		},
		storeExtraLives: func(update *playerUnitInitUpdate4EFE80, value int32) {
			event := fmt.Sprintf("extra-lives:%s=%08x", playerUnitInitUpdateName4EFE80(update), uint32(value))
			w.record(event)
			if update == nil {
				panic(event)
			}
			update.extraLives = value
		},
		makeDefaultItems: func(unit *playerUnitInitObject4EFE80, restoreStats, keepItems int32) uint8 {
			w.record(fmt.Sprintf("default-items:%s:%d:%d=%02x", playerUnitInitObjectName4EFE80(unit), restoreStats, keepItems, w.defaultResult))
			return w.defaultResult
		},
	}
}

func newPlayerUnitInitWorld4EFE80() *playerUnitInitWorld4EFE80 {
	player := &playerUnitInitPlayer4EFE80{name: "player-a"}
	update := &playerUnitInitUpdate4EFE80{name: "update-a", player: player, extraLives: 9}
	return &playerUnitInitWorld4EFE80{
		unitArg:       &playerUnitInitObject4EFE80{name: "unit-a", update: update},
		gold:          0xf1234567,
		questFlag:     1,
		balance:       math.Float32frombits(0xbfc00000),
		converted:     -2,
		defaultResult: 0xfe,
		after:         make(map[string]func()),
	}
}

func playerUnitInitExpectedQuestEvents4EFE80() []string {
	return []string{
		"arg:unit-a",
		"update:unit-a=update-a",
		"get-gold:unit-a=f1234567",
		"sub-gold:unit-a:f1234567",
		"sync:unit-a",
		"player:update-a=player-a",
		"award-scroll:player-a",
		"player:update-a=player-a",
		"award-spell:player-a",
		"read-values:unit-a:0",
		"player:update-a=player-a",
		"award-ability:player-a",
		"flag:00001000=1",
		"balance:QuestGameStartingExtraLives=bfc00000",
		"float-to-int:bfc00000=-2",
		"extra-lives:update-a=fffffffe",
		"default-items:unit-a:1:0=fe",
	}
}

func TestPlayerUnitInit4EFE80QuestOrderStateAndReturn(t *testing.T) {
	w := newPlayerUnitInitWorld4EFE80()
	if got := playerUnitInit4EFE80(w.hooks()); got != 0xfe {
		t.Fatalf("result = %#x, want 0xfe", got)
	}
	if want := playerUnitInitExpectedQuestEvents4EFE80(); !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %v, want %v", w.events, want)
	}
	if got := w.unitArg.update.extraLives; got != -2 {
		t.Fatalf("ExtraLives = %d, want -2", got)
	}
}

func TestPlayerUnitInit4EFE80QuestFlagIsWholeNonzero(t *testing.T) {
	for _, flag := range []int32{1, 2, -1, math.MinInt32} {
		t.Run(fmt.Sprintf("%d", flag), func(t *testing.T) {
			w := newPlayerUnitInitWorld4EFE80()
			w.questFlag = flag
			playerUnitInit4EFE80(w.hooks())
			if got := w.unitArg.update.extraLives; got != -2 {
				t.Fatalf("ExtraLives = %d, want -2", got)
			}
		})
	}
}

func TestPlayerUnitInit4EFE80NonQuestSkipsBalanceAndStore(t *testing.T) {
	w := newPlayerUnitInitWorld4EFE80()
	w.questFlag = 0
	if got := playerUnitInit4EFE80(w.hooks()); got != 0xfe {
		t.Fatalf("result = %#x, want 0xfe", got)
	}
	if got := w.unitArg.update.extraLives; got != 9 {
		t.Fatalf("ExtraLives = %d, want unchanged 9", got)
	}
	want := playerUnitInitExpectedQuestEvents4EFE80()[:13]
	want[len(want)-1] = "flag:00001000=0"
	want = append(want, "default-items:unit-a:1:0=fe")
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %v, want %v", w.events, want)
	}
}

func TestPlayerUnitInit4EFE80CachesUnitAndUpdateButReloadsPlayers(t *testing.T) {
	w := newPlayerUnitInitWorld4EFE80()
	originalUnit := w.unitArg
	cached := originalUnit.update
	replacementUpdate := &playerUnitInitUpdate4EFE80{name: "update-b", extraLives: 77}
	replacementUnit := &playerUnitInitObject4EFE80{name: "unit-b", update: replacementUpdate}
	players := []*playerUnitInitPlayer4EFE80{
		cached.player,
		{name: "player-b"},
		{name: "player-c"},
		{name: "player-d"},
	}
	w.after["update:unit-a=update-a"] = func() {
		originalUnit.update = replacementUpdate
		w.unitArg = replacementUnit
	}
	w.after["award-scroll:player-a"] = func() { cached.player = players[1] }
	w.after["award-spell:player-b"] = func() { cached.player = players[2] }
	w.after["read-values:unit-a:0"] = func() { cached.player = players[3] }
	w.after["balance:QuestGameStartingExtraLives=bfc00000"] = func() {
		originalUnit.update = replacementUpdate
	}

	playerUnitInit4EFE80(w.hooks())
	wantSubsequence := []string{
		"award-scroll:player-a",
		"award-spell:player-b",
		"read-values:unit-a:0",
		"award-ability:player-d",
		"extra-lives:update-a=fffffffe",
		"default-items:unit-a:1:0=fe",
	}
	position := 0
	for _, event := range w.events {
		if position < len(wantSubsequence) && event == wantSubsequence[position] {
			position++
		}
	}
	if position != len(wantSubsequence) {
		t.Fatalf("events = %v, missing subsequence %v", w.events, wantSubsequence[position:])
	}
	if cached.extraLives != -2 || replacementUpdate.extraLives != 77 {
		t.Fatalf("cached/replacement ExtraLives = %d/%d", cached.extraLives, replacementUpdate.extraLives)
	}
}

func TestPlayerUnitInit4EFE80HasNoNilGuards(t *testing.T) {
	t.Run("unit", func(t *testing.T) {
		w := newPlayerUnitInitWorld4EFE80()
		w.unitArg = nil
		defer func() {
			if got := recover(); got != "update:nil" {
				t.Fatalf("panic = %v, want update:nil", got)
			}
			if want := []string{"arg:nil", "update:nil"}; !reflect.DeepEqual(w.events, want) {
				t.Fatalf("events = %v, want %v", w.events, want)
			}
		}()
		playerUnitInit4EFE80(w.hooks())
	})

	t.Run("update", func(t *testing.T) {
		w := newPlayerUnitInitWorld4EFE80()
		w.unitArg.update = nil
		defer func() {
			if got := recover(); got != "player:nil" {
				t.Fatalf("panic = %v, want player:nil", got)
			}
			want := []string{
				"arg:unit-a", "update:unit-a=nil", "get-gold:unit-a=f1234567",
				"sub-gold:unit-a:f1234567", "sync:unit-a", "player:nil",
			}
			if !reflect.DeepEqual(w.events, want) {
				t.Fatalf("events = %v, want %v", w.events, want)
			}
		}()
		playerUnitInit4EFE80(w.hooks())
	})

	t.Run("player", func(t *testing.T) {
		w := newPlayerUnitInitWorld4EFE80()
		w.unitArg.update.player = nil
		defer func() {
			if got := recover(); got != "award-scroll:nil" {
				t.Fatalf("panic = %v, want award-scroll:nil", got)
			}
			want := []string{
				"arg:unit-a", "update:unit-a=update-a", "get-gold:unit-a=f1234567",
				"sub-gold:unit-a:f1234567", "sync:unit-a", "player:update-a=nil",
				"award-scroll:nil",
			}
			if !reflect.DeepEqual(w.events, want) {
				t.Fatalf("events = %v, want %v", w.events, want)
			}
		}()
		playerUnitInit4EFE80(w.hooks())
	})
}

func testPlayerUnitInitFaultPrefixes4EFE80(t *testing.T, questFlag int32) {
	t.Helper()
	base := newPlayerUnitInitWorld4EFE80()
	base.questFlag = questFlag
	playerUnitInit4EFE80(base.hooks())
	want := append([]string(nil), base.events...)
	wantCount := 14
	if questFlag != 0 {
		wantCount = 17
	}
	if len(want) != wantCount {
		t.Fatalf("observable events = %d, want %d", len(want), wantCount)
	}

	for faultAt := 1; faultAt <= len(want); faultAt++ {
		t.Run(fmt.Sprintf("event-%02d", faultAt), func(t *testing.T) {
			w := newPlayerUnitInitWorld4EFE80()
			w.questFlag = questFlag
			w.faultAt = faultAt
			defer func() {
				if got := recover(); got != want[faultAt-1] {
					t.Fatalf("panic = %v, want %q", got, want[faultAt-1])
				}
				if prefix := want[:faultAt]; !reflect.DeepEqual(w.events, prefix) {
					t.Fatalf("events = %v, want prefix %v", w.events, prefix)
				}
			}()
			playerUnitInit4EFE80(w.hooks())
		})
	}
}

func TestPlayerUnitInit4EFE80EveryObservableFaultPrefix(t *testing.T) {
	t.Run("quest", func(t *testing.T) { testPlayerUnitInitFaultPrefixes4EFE80(t, 1) })
	t.Run("non-quest", func(t *testing.T) { testPlayerUnitInitFaultPrefixes4EFE80(t, 0) })
}
