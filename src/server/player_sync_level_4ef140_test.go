package server

import (
	"fmt"
	"math"
	"reflect"
	"testing"
)

type playerSyncLevelTestPlayer4EF140 struct {
	name  string
	level uint8
	token uint32
}

type playerSyncLevelTestUpdate4EF140 struct {
	name   string
	player *playerSyncLevelTestPlayer4EF140
}

type playerSyncLevelTestObject4EF140 struct {
	name       string
	experience float32
	update     *playerSyncLevelTestUpdate4EF140
}

type playerSyncLevelTestWorld4EF140 struct {
	unit           *playerSyncLevelTestObject4EF140
	gameFlagResult int32
	xpTable        [11]float64
	readResult     string
	events         []string
	faultAt        int
	after          map[string]func()
	protectedToken uint32
	protectedLevel uint8
	readUnit       *playerSyncLevelTestObject4EF140
	readReward     int32
}

func playerSyncLevelObjectName4EF140(unit *playerSyncLevelTestObject4EF140) string {
	if unit == nil {
		return "nil"
	}
	return unit.name
}

func playerSyncLevelUpdateName4EF140(update *playerSyncLevelTestUpdate4EF140) string {
	if update == nil {
		return "nil"
	}
	return update.name
}

func playerSyncLevelPlayerName4EF140(player *playerSyncLevelTestPlayer4EF140) string {
	if player == nil {
		return "nil"
	}
	return player.name
}

func (w *playerSyncLevelTestWorld4EF140) record(event string) {
	w.events = append(w.events, event)
	if w.faultAt != 0 && len(w.events) == w.faultAt {
		panic(event)
	}
}

func (w *playerSyncLevelTestWorld4EF140) finish(event string) {
	if after := w.after[event]; after != nil {
		after()
	}
}

func (w *playerSyncLevelTestWorld4EF140) hooks() playerSyncLevelHooks4EF140[
	*playerSyncLevelTestObject4EF140,
	*playerSyncLevelTestUpdate4EF140,
	*playerSyncLevelTestPlayer4EF140,
	string,
] {
	return playerSyncLevelHooks4EF140[
		*playerSyncLevelTestObject4EF140,
		*playerSyncLevelTestUpdate4EF140,
		*playerSyncLevelTestPlayer4EF140,
		string,
	]{
		loadUnitArg: func() *playerSyncLevelTestObject4EF140 {
			unit := w.unit
			event := "arg:" + playerSyncLevelObjectName4EF140(unit)
			w.record(event)
			w.finish(event)
			return unit
		},
		loadUpdateData: func(unit *playerSyncLevelTestObject4EF140) *playerSyncLevelTestUpdate4EF140 {
			name := playerSyncLevelObjectName4EF140(unit)
			if unit == nil {
				event := "update:" + name
				w.record(event)
				panic("update:nil")
			}
			update := unit.update
			event := "update:" + name + "=" + playerSyncLevelUpdateName4EF140(update)
			w.record(event)
			w.finish(event)
			return update
		},
		loadPlayer: func(update *playerSyncLevelTestUpdate4EF140) *playerSyncLevelTestPlayer4EF140 {
			name := playerSyncLevelUpdateName4EF140(update)
			if update == nil {
				event := "player:" + name
				w.record(event)
				panic("player:nil-update")
			}
			player := update.player
			event := "player:" + name + "=" + playerSyncLevelPlayerName4EF140(player)
			w.record(event)
			w.finish(event)
			return player
		},
		gameFlagsCheck: func(mask uint32) int32 {
			result := w.gameFlagResult
			event := fmt.Sprintf("flag:%08x=%d", mask, result)
			w.record(event)
			w.finish(event)
			return result
		},
		loadXPTable: func(index int32) float64 {
			value := w.xpTable[index]
			event := fmt.Sprintf("table:%d=%016x", index, math.Float64bits(value))
			w.record(event)
			w.finish(event)
			return value
		},
		loadExperience: func(unit *playerSyncLevelTestObject4EF140) float32 {
			name := playerSyncLevelObjectName4EF140(unit)
			if unit == nil {
				event := "experience:" + name
				w.record(event)
				panic("experience:nil")
			}
			value := unit.experience
			event := fmt.Sprintf("experience:%s=%08x", name, math.Float32bits(value))
			w.record(event)
			w.finish(event)
			return value
		},
		loadLevelProtection: func(player *playerSyncLevelTestPlayer4EF140) uint32 {
			name := playerSyncLevelPlayerName4EF140(player)
			if player == nil {
				event := "protection:" + name
				w.record(event)
				panic("protection:nil")
			}
			token := player.token
			event := fmt.Sprintf("protection:%s=%08x", name, token)
			w.record(event)
			w.finish(event)
			return token
		},
		storeLevel: func(player *playerSyncLevelTestPlayer4EF140, level uint8) {
			event := fmt.Sprintf("store-level:%s=%02x", playerSyncLevelPlayerName4EF140(player), level)
			w.record(event)
			if player == nil {
				panic("store-level:nil")
			}
			player.level = level
			w.finish(event)
		},
		protectLevel: func(token uint32, level uint8) {
			event := fmt.Sprintf("protect:%08x:%02x", token, level)
			w.record(event)
			w.protectedToken, w.protectedLevel = token, level
			w.finish(event)
		},
		readValues: func(unit *playerSyncLevelTestObject4EF140, reward int32) string {
			result := w.readResult
			event := fmt.Sprintf("read-values:%s:%d", playerSyncLevelObjectName4EF140(unit), reward)
			w.record(event)
			w.readUnit, w.readReward = unit, reward
			w.finish(event)
			return result
		},
	}
}

func newPlayerSyncLevelTestWorld4EF140() *playerSyncLevelTestWorld4EF140 {
	player := &playerSyncLevelTestPlayer4EF140{name: "player", level: 7, token: 0x12345678}
	update := &playerSyncLevelTestUpdate4EF140{name: "update", player: player}
	w := &playerSyncLevelTestWorld4EF140{
		unit:       &playerSyncLevelTestObject4EF140{name: "unit", experience: 250, update: update},
		readResult: "read-result",
		after:      make(map[string]func()),
	}
	for index := range w.xpTable {
		w.xpTable[index] = float64(index * 100)
	}
	return w
}

func TestPlayerSyncLevel4EF140CooperativePathOrderAndWholeFlag(t *testing.T) {
	for _, flag := range []int32{1, 2, -1, math.MinInt32} {
		t.Run(fmt.Sprintf("%d", flag), func(t *testing.T) {
			w := newPlayerSyncLevelTestWorld4EF140()
			w.gameFlagResult = flag
			player := w.unit.update.player
			if got := playerSyncLevel4EF140(w.hooks()); got != "read-result" {
				t.Fatalf("result = %q, want read-result", got)
			}
			if player.level != 10 || w.protectedToken != 0x12345678 || w.protectedLevel != 10 {
				t.Fatalf("level/protection = %d/%08x/%d", player.level, w.protectedToken, w.protectedLevel)
			}
			if w.readUnit != w.unit || w.readReward != 0 {
				t.Fatalf("read values args = %p/%d, want %p/0", w.readUnit, w.readReward, w.unit)
			}
			want := []string{
				"arg:unit",
				"update:unit=update",
				"player:update=player",
				fmt.Sprintf("flag:00002000=%d", flag),
				"protection:player=12345678",
				"store-level:player=0a",
				"protect:12345678:0a",
				"read-values:unit:0",
			}
			if !reflect.DeepEqual(w.events, want) {
				t.Fatalf("events = %q, want %q", w.events, want)
			}
		})
	}
}

func TestPlayerSyncLevel4EF140NormalThresholdAndUnorderedRules(t *testing.T) {
	tests := []struct {
		name       string
		experience float32
		mutate     func(*playerSyncLevelTestWorld4EF140)
		wantLevel  uint8
		wantCalls  int
	}{
		{name: "below first", experience: -1, wantLevel: 0xff, wantCalls: 1},
		{name: "equal first", experience: 0, wantLevel: 0, wantCalls: 2},
		{name: "between", experience: 250, wantLevel: 2, wantCalls: 4},
		{name: "equal maximum", experience: 1000, wantLevel: 10, wantCalls: 11},
		{name: "above maximum", experience: 2000, wantLevel: 10, wantCalls: 11},
		{name: "positive infinity experience", experience: float32(math.Inf(1)), wantLevel: 10, wantCalls: 11},
		{name: "nan experience", experience: math.Float32frombits(0x7fc12345), wantLevel: 10, wantCalls: 11},
		{
			name:       "nan threshold continues",
			experience: 250,
			mutate: func(w *playerSyncLevelTestWorld4EF140) {
				w.xpTable[3] = math.Float64frombits(0x7ff8000000001234)
			},
			wantLevel: 3,
			wantCalls: 5,
		},
		{
			name:       "positive infinity threshold breaks",
			experience: math.MaxFloat32,
			mutate: func(w *playerSyncLevelTestWorld4EF140) {
				w.xpTable[0] = math.Inf(1)
			},
			wantLevel: 0xff,
			wantCalls: 1,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			w := newPlayerSyncLevelTestWorld4EF140()
			w.unit.experience = test.experience
			if test.mutate != nil {
				test.mutate(w)
			}
			player := w.unit.update.player
			if got := playerSyncLevel4EF140(w.hooks()); got != "read-result" {
				t.Fatalf("result = %q, want read-result", got)
			}
			if player.level != test.wantLevel || w.protectedLevel != test.wantLevel {
				t.Fatalf("stored/protected level = %#02x/%#02x, want %#02x", player.level, w.protectedLevel, test.wantLevel)
			}
			calls := 0
			for _, event := range w.events {
				if len(event) >= len("table:") && event[:len("table:")] == "table:" {
					calls++
				}
			}
			if calls != test.wantCalls {
				t.Fatalf("table calls = %d, want %d; events=%q", calls, test.wantCalls, w.events)
			}
		})
	}
}

func TestPlayerSyncLevel4EF140CachesEntryPointersAndLateValues(t *testing.T) {
	w := newPlayerSyncLevelTestWorld4EF140()
	w.gameFlagResult = 1
	originalUnit := w.unit
	originalUpdate := originalUnit.update
	originalPlayer := originalUpdate.player
	replacementPlayer := &playerSyncLevelTestPlayer4EF140{name: "replacement-player", level: 1, token: 0x11111111}
	replacementUpdate := &playerSyncLevelTestUpdate4EF140{name: "replacement-update", player: replacementPlayer}
	replacementUnit := &playerSyncLevelTestObject4EF140{name: "replacement-unit", update: replacementUpdate}
	w.after["arg:unit"] = func() {
		w.unit = replacementUnit
	}
	w.after["update:unit=update"] = func() {
		originalUnit.update = replacementUpdate
	}
	w.after["player:update=player"] = func() {
		originalUpdate.player = replacementPlayer
	}
	w.after["flag:00002000=1"] = func() {
		originalPlayer.token = 0xabcdef01
	}
	w.after["protection:player=abcdef01"] = func() {
		originalPlayer.token = 0x22222222
	}

	if got := playerSyncLevel4EF140(w.hooks()); got != "read-result" {
		t.Fatalf("result = %q, want read-result", got)
	}
	if originalPlayer.level != 10 || replacementPlayer.level != 1 {
		t.Fatalf("original/replacement levels = %d/%d", originalPlayer.level, replacementPlayer.level)
	}
	if w.protectedToken != 0xabcdef01 || w.protectedLevel != 10 {
		t.Fatalf("protection = %08x/%d, want abcdef01/10", w.protectedToken, w.protectedLevel)
	}
	if w.readUnit != originalUnit || w.readReward != 0 {
		t.Fatalf("read values args = %p/%d, want original %p/0", w.readUnit, w.readReward, originalUnit)
	}
	want := []string{
		"arg:unit",
		"update:unit=update",
		"player:update=player",
		"flag:00002000=1",
		"protection:player=abcdef01",
		"store-level:player=0a",
		"protect:abcdef01:0a",
		"read-values:unit:0",
	}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %q, want %q", w.events, want)
	}
}

func TestPlayerSyncLevel4EF140ReloadsExperienceAfterEachTableCallback(t *testing.T) {
	w := newPlayerSyncLevelTestWorld4EF140()
	w.unit.experience = 50
	w.after[fmt.Sprintf("table:1=%016x", math.Float64bits(100))] = func() {
		w.unit.experience = 150
	}
	w.after[fmt.Sprintf("experience:unit=%08x", math.Float32bits(150))] = func() {
		w.unit.experience = 50
	}
	playerSyncLevel4EF140(w.hooks())
	if w.unit.update.player.level != 1 {
		t.Fatalf("level = %d, want 1", w.unit.update.player.level)
	}
	wantMiddle := []string{
		fmt.Sprintf("table:0=%016x", math.Float64bits(0)),
		fmt.Sprintf("experience:unit=%08x", math.Float32bits(50)),
		fmt.Sprintf("table:1=%016x", math.Float64bits(100)),
		fmt.Sprintf("experience:unit=%08x", math.Float32bits(150)),
		fmt.Sprintf("table:2=%016x", math.Float64bits(200)),
		fmt.Sprintf("experience:unit=%08x", math.Float32bits(50)),
	}
	if !reflect.DeepEqual(w.events[4:10], wantMiddle) {
		t.Fatalf("loop events = %q, want %q", w.events[4:10], wantMiddle)
	}
}

func TestPlayerSyncLevel4EF140NilFaultBoundaries(t *testing.T) {
	t.Run("nil unit faults before game flag", func(t *testing.T) {
		w := newPlayerSyncLevelTestWorld4EF140()
		w.unit = nil
		playerSyncLevelMustPanic4EF140(t, w)
		want := []string{"arg:nil", "update:nil"}
		if !reflect.DeepEqual(w.events, want) {
			t.Fatalf("events = %q, want %q", w.events, want)
		}
	})
	t.Run("nil update faults before game flag", func(t *testing.T) {
		w := newPlayerSyncLevelTestWorld4EF140()
		w.unit.update = nil
		playerSyncLevelMustPanic4EF140(t, w)
		want := []string{"arg:unit", "update:unit=nil", "player:nil"}
		if !reflect.DeepEqual(w.events, want) {
			t.Fatalf("events = %q, want %q", w.events, want)
		}
	})
	t.Run("nil player survives load until direct protection", func(t *testing.T) {
		w := newPlayerSyncLevelTestWorld4EF140()
		w.unit.update.player = nil
		w.gameFlagResult = 1
		playerSyncLevelMustPanic4EF140(t, w)
		want := []string{
			"arg:unit", "update:unit=update", "player:update=nil",
			"flag:00002000=1", "protection:nil",
		}
		if !reflect.DeepEqual(w.events, want) {
			t.Fatalf("events = %q, want %q", w.events, want)
		}
	})
}

func TestPlayerSyncLevel4EF140EveryObservationFaultPrefix(t *testing.T) {
	paths := []struct {
		name string
		flag int32
	}{
		{name: "cooperative", flag: 1},
		{name: "normal", flag: 0},
	}
	for _, path := range paths {
		t.Run(path.name, func(t *testing.T) {
			baseline := newPlayerSyncLevelTestWorld4EF140()
			baseline.gameFlagResult = path.flag
			playerSyncLevel4EF140(baseline.hooks())
			full := append([]string(nil), baseline.events...)
			for faultAt := 1; faultAt <= len(full); faultAt++ {
				t.Run(fmt.Sprintf("fault-%02d", faultAt), func(t *testing.T) {
					w := newPlayerSyncLevelTestWorld4EF140()
					w.gameFlagResult = path.flag
					w.faultAt = faultAt
					playerSyncLevelMustPanic4EF140(t, w)
					if !reflect.DeepEqual(w.events, full[:faultAt]) {
						t.Fatalf("events = %q, want prefix %q", w.events, full[:faultAt])
					}
				})
			}
		})
	}
}

func playerSyncLevelMustPanic4EF140(t *testing.T, w *playerSyncLevelTestWorld4EF140) {
	t.Helper()
	defer func() {
		if recover() == nil {
			t.Fatal("playerSyncLevel4EF140 did not panic")
		}
	}()
	playerSyncLevel4EF140(w.hooks())
}
