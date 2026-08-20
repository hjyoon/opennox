package server

import (
	"fmt"
	"math"
	"reflect"
	"testing"
)

type experienceLevelTestPlayer4EF2E0 struct {
	name  string
	level uint8
	token uint32
}

type experienceLevelTestUpdate4EF2E0 struct {
	name   string
	player *experienceLevelTestPlayer4EF2E0
}

type experienceLevelTestObject4EF2E0 struct {
	name       string
	experience float32
	netCode    uint32
	update     *experienceLevelTestUpdate4EF2E0
}

type experienceLevelTestWorld4EF2E0 struct {
	unit          *experienceLevelTestObject4EF2E0
	gameGetResult int32
	gameSubResult bool
	threshold     float64
	gameFlag      int32
	message       string
	events        []string
	after         map[string]func()
}

func experienceLevelObjectName4EF2E0(unit *experienceLevelTestObject4EF2E0) string {
	if unit == nil {
		return "nil"
	}
	return unit.name
}

func experienceLevelUpdateName4EF2E0(update *experienceLevelTestUpdate4EF2E0) string {
	if update == nil {
		return "nil"
	}
	return update.name
}

func experienceLevelPlayerName4EF2E0(player *experienceLevelTestPlayer4EF2E0) string {
	if player == nil {
		return "nil"
	}
	return player.name
}

func (w *experienceLevelTestWorld4EF2E0) record(event string) {
	w.events = append(w.events, event)
	if after := w.after[event]; after != nil {
		delete(w.after, event)
		after()
	}
}

func (w *experienceLevelTestWorld4EF2E0) hooks() experienceLevelUpdateHooks4EF2E0[
	*experienceLevelTestObject4EF2E0,
	*experienceLevelTestUpdate4EF2E0,
	*experienceLevelTestPlayer4EF2E0,
	string,
] {
	return experienceLevelUpdateHooks4EF2E0[
		*experienceLevelTestObject4EF2E0,
		*experienceLevelTestUpdate4EF2E0,
		*experienceLevelTestPlayer4EF2E0,
		string,
	]{
		loadUnitArg: func() *experienceLevelTestObject4EF2E0 {
			unit := w.unit
			w.record("arg:" + experienceLevelObjectName4EF2E0(unit))
			return unit
		},
		loadUpdateData: func(unit *experienceLevelTestObject4EF2E0) *experienceLevelTestUpdate4EF2E0 {
			name := experienceLevelObjectName4EF2E0(unit)
			if unit == nil {
				event := "update:" + name
				w.record(event)
				panic(event)
			}
			update := unit.update
			w.record("update:" + name + "=" + experienceLevelUpdateName4EF2E0(update))
			return update
		},
		loadPlayer: func(update *experienceLevelTestUpdate4EF2E0) *experienceLevelTestPlayer4EF2E0 {
			name := experienceLevelUpdateName4EF2E0(update)
			if update == nil {
				event := "player:" + name
				w.record(event)
				panic(event)
			}
			player := update.player
			w.record("player:" + name + "=" + experienceLevelPlayerName4EF2E0(player))
			return player
		},
		gameGet: func() int32 {
			result := w.gameGetResult
			w.record(fmt.Sprintf("game-get:%d", result))
			return result
		},
		gameSubActive: func() bool {
			result := w.gameSubResult
			w.record(fmt.Sprintf("game-sub:%t", result))
			return result
		},
		loadLevel: func(player *experienceLevelTestPlayer4EF2E0) uint8 {
			name := experienceLevelPlayerName4EF2E0(player)
			if player == nil {
				event := "level:" + name
				w.record(event)
				panic(event)
			}
			level := player.level
			w.record(fmt.Sprintf("level:%s=%02x", name, level))
			return level
		},
		loadXPTable: func(key string, index int32) float64 {
			value := w.threshold
			w.record(fmt.Sprintf("table:%s:%d=%016x", key, index, math.Float64bits(value)))
			return value
		},
		loadExperience: func(unit *experienceLevelTestObject4EF2E0) float32 {
			name := experienceLevelObjectName4EF2E0(unit)
			if unit == nil {
				event := "experience:" + name
				w.record(event)
				panic(event)
			}
			value := unit.experience
			w.record(fmt.Sprintf("experience:%s=%08x", name, math.Float32bits(value)))
			return value
		},
		storeLevel: func(player *experienceLevelTestPlayer4EF2E0, level uint8) {
			event := fmt.Sprintf("store-level:%s=%02x", experienceLevelPlayerName4EF2E0(player), level)
			w.record(event)
			if player == nil {
				panic(event)
			}
			player.level = level
		},
		loadLevelToken: func(player *experienceLevelTestPlayer4EF2E0) uint32 {
			name := experienceLevelPlayerName4EF2E0(player)
			if player == nil {
				event := "token:" + name
				w.record(event)
				panic(event)
			}
			token := player.token
			w.record(fmt.Sprintf("token:%s=%08x", name, token))
			return token
		},
		protectLevel: func(token uint32, level uint8) {
			w.record(fmt.Sprintf("protect:%08x:%02x", token, level))
		},
		readValues: func(unit *experienceLevelTestObject4EF2E0, reward int32) {
			w.record(fmt.Sprintf("read-values:%s:%d", experienceLevelObjectName4EF2E0(unit), reward))
		},
		gameFlag: func(flag uint32) int32 {
			result := w.gameFlag
			w.record(fmt.Sprintf("game-flag:%08x=%d", flag, result))
			return result
		},
		pauseFX: func(unit *experienceLevelTestObject4EF2E0, mode int32) {
			w.record(fmt.Sprintf("pause:%s:%d", experienceLevelObjectName4EF2E0(unit), mode))
		},
		loadNetCode: func(unit *experienceLevelTestObject4EF2E0) uint32 {
			name := experienceLevelObjectName4EF2E0(unit)
			if unit == nil {
				event := "net-code:" + name
				w.record(event)
				panic(event)
			}
			code := unit.netCode
			w.record(fmt.Sprintf("net-code:%s=%08x", name, code))
			return code
		},
		audio: func(id uint32, unit *experienceLevelTestObject4EF2E0, kind int32, code uint32) {
			w.record(fmt.Sprintf("audio:%d:%s:%d:%08x", id, experienceLevelObjectName4EF2E0(unit), kind, code))
		},
		loadString: func(key, path string, line int) string {
			w.record(fmt.Sprintf("string:%s:%s:%d", key, path, line))
			return w.message
		},
		sendLineMessage: func(unit *experienceLevelTestObject4EF2E0, message string) {
			w.record("line:" + experienceLevelObjectName4EF2E0(unit) + ":" + message)
		},
	}
}

func newExperienceLevelTestWorld4EF2E0() *experienceLevelTestWorld4EF2E0 {
	player := &experienceLevelTestPlayer4EF2E0{name: "player", level: 4, token: 0x12345678}
	update := &experienceLevelTestUpdate4EF2E0{name: "update", player: player}
	return &experienceLevelTestWorld4EF2E0{
		unit:      &experienceLevelTestObject4EF2E0{name: "unit", experience: 500, netCode: 0x10203040, update: update},
		threshold: 400,
		message:   "level-message",
		after:     make(map[string]func()),
	}
}

func experienceLevelMustPanic4EF2E0(t *testing.T, run func()) {
	t.Helper()
	defer func() {
		if recover() == nil {
			t.Fatal("call did not panic")
		}
	}()
	run()
}

func TestExperienceLevelUpdate4EF2E0GameStateGateIsExactAndLazy(t *testing.T) {
	for _, test := range []struct {
		name        string
		gameGet     int32
		gameSub     bool
		wantSub     bool
		wantLevel   uint8
		wantStopped bool
	}{
		{name: "zero skips sub", gameGet: 0, gameSub: true, wantLevel: 5},
		{name: "two skips sub", gameGet: 2, gameSub: true, wantLevel: 5},
		{name: "minus one skips sub", gameGet: -1, gameSub: true, wantLevel: 5},
		{name: "exact one inactive", gameGet: 1, gameSub: false, wantSub: true, wantLevel: 5},
		{name: "exact one active", gameGet: 1, gameSub: true, wantSub: true, wantLevel: 4, wantStopped: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			w := newExperienceLevelTestWorld4EF2E0()
			w.gameGetResult, w.gameSubResult = test.gameGet, test.gameSub
			experienceLevelUpdate4EF2E0(w.hooks())
			if got := w.unit.update.player.level; got != test.wantLevel {
				t.Fatalf("level = %d, want %d", got, test.wantLevel)
			}
			hasSub, hasTable := false, false
			for _, event := range w.events {
				hasSub = hasSub || event == fmt.Sprintf("game-sub:%t", test.gameSub)
				hasTable = hasTable || len(event) >= 6 && event[:6] == "table:"
			}
			if hasSub != test.wantSub || hasTable == test.wantStopped {
				t.Fatalf("sub/table/events = %t/%t/%q", hasSub, hasTable, w.events)
			}
		})
	}
}

func TestExperienceLevelUpdate4EF2E0SignedIndexAndOrderedComparison(t *testing.T) {
	for _, test := range []struct {
		name       string
		level      uint8
		threshold  float64
		experience float32
		wantIndex  int32
		wantLevel  uint8
	}{
		{name: "positive", level: 4, threshold: 501, experience: 500, wantIndex: 5, wantLevel: 4},
		{name: "equal proceeds", level: 4, threshold: 500, experience: 500, wantIndex: 5, wantLevel: 5},
		{name: "lower proceeds", level: 4, threshold: 499, experience: 500, wantIndex: 5, wantLevel: 5},
		{name: "ff maps zero and wraps", level: 0xff, threshold: 0, experience: 0, wantIndex: 0, wantLevel: 0},
		{name: "signed maximum", level: 0x7f, threshold: 0, experience: 0, wantIndex: 128, wantLevel: 0x80},
		{name: "signed minimum", level: 0x80, threshold: 0, experience: 0, wantIndex: -127, wantLevel: 0x81},
		{name: "nan threshold proceeds", level: 9, threshold: math.NaN(), experience: 0, wantIndex: 10, wantLevel: 10},
		{name: "nan experience proceeds", level: 9, threshold: 1, experience: math.Float32frombits(0x7fc12345), wantIndex: 10, wantLevel: 10},
	} {
		t.Run(test.name, func(t *testing.T) {
			w := newExperienceLevelTestWorld4EF2E0()
			w.unit.update.player.level = test.level
			w.threshold, w.unit.experience = test.threshold, test.experience
			experienceLevelUpdate4EF2E0(w.hooks())
			if got := w.unit.update.player.level; got != test.wantLevel {
				t.Fatalf("level = %#02x, want %#02x; events=%q", got, test.wantLevel, w.events)
			}
			wantPrefix := fmt.Sprintf("table:XPTable:%d=", test.wantIndex)
			found := false
			for _, event := range w.events {
				found = found || len(event) >= len(wantPrefix) && event[:len(wantPrefix)] == wantPrefix
			}
			if !found {
				t.Fatalf("signed index %d absent from events %q", test.wantIndex, w.events)
			}
		})
	}
}

func TestExperienceLevelUpdate4EF2E0SoloOrderAndLiveReloads(t *testing.T) {
	w := newExperienceLevelTestWorld4EF2E0()
	entryPlayer := w.unit.update.player
	replacementPlayer := &experienceLevelTestPlayer4EF2E0{name: "replacement", level: 50, token: 9}
	w.after["update:unit=update"] = func() {
		w.unit.update = &experienceLevelTestUpdate4EF2E0{name: "new-update", player: replacementPlayer}
	}
	w.after["table:XPTable:5=4079000000000000"] = func() {
		entryPlayer.level = 9
		w.unit.experience = 600
	}
	w.after["store-level:player=0a"] = func() { entryPlayer.token = 0xaabbccdd }
	w.after["game-flag:00000800=0"] = func() { w.unit.netCode = 0x55667788 }

	experienceLevelUpdate4EF2E0(w.hooks())
	want := []string{
		"arg:unit",
		"update:unit=update",
		"player:update=player",
		"game-get:0",
		"level:player=04",
		"table:XPTable:5=4079000000000000",
		"experience:unit=44160000",
		"level:player=09",
		"store-level:player=0a",
		"token:player=aabbccdd",
		"protect:aabbccdd:01",
		"read-values:unit:1",
		"game-flag:00000800=0",
		"net-code:unit=55667788",
		"audio:902:unit:2:55667788",
		`string:LevelUP:C:\NoxPost\src\Server\GameMech\explevel.c:253`,
		"line:unit:level-message",
	}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %q, want %q", w.events, want)
	}
	if entryPlayer.level != 10 || replacementPlayer.level != 50 {
		t.Fatalf("cached/replacement levels = %d/%d, want 10/50", entryPlayer.level, replacementPlayer.level)
	}
}

func TestExperienceLevelUpdate4EF2E0CoopBranchDefersSoloReads(t *testing.T) {
	for _, result := range []int32{1, 2, -1, math.MinInt32} {
		w := newExperienceLevelTestWorld4EF2E0()
		w.gameFlag = result
		experienceLevelUpdate4EF2E0(w.hooks())
		wantTail := []string{
			fmt.Sprintf("game-flag:00000800=%d", result),
			"pause:unit:0",
		}
		if len(w.events) < 2 || !reflect.DeepEqual(w.events[len(w.events)-2:], wantTail) {
			t.Fatalf("result %d tail = %q, want %q", result, w.events, wantTail)
		}
		for _, event := range w.events {
			if len(event) >= 9 && (event[:9] == "net-code:" || event[:6] == "audio:" || event[:7] == "string:" || event[:5] == "line:") {
				t.Fatalf("result %d performed deferred solo read: %q", result, w.events)
			}
		}
	}
}

func TestExperienceLevelUpdate4EF2E0FaultPrefixes(t *testing.T) {
	t.Run("nil unit", func(t *testing.T) {
		w := newExperienceLevelTestWorld4EF2E0()
		w.unit = nil
		experienceLevelMustPanic4EF2E0(t, func() { experienceLevelUpdate4EF2E0(w.hooks()) })
		if want := []string{"arg:nil", "update:nil"}; !reflect.DeepEqual(w.events, want) {
			t.Fatalf("events = %q, want %q", w.events, want)
		}
	})

	t.Run("nil update", func(t *testing.T) {
		w := newExperienceLevelTestWorld4EF2E0()
		w.unit.update = nil
		experienceLevelMustPanic4EF2E0(t, func() { experienceLevelUpdate4EF2E0(w.hooks()) })
		if want := []string{"arg:unit", "update:unit=nil", "player:nil"}; !reflect.DeepEqual(w.events, want) {
			t.Fatalf("events = %q, want %q", w.events, want)
		}
	})

	t.Run("nil player may take game-state early return", func(t *testing.T) {
		w := newExperienceLevelTestWorld4EF2E0()
		w.unit.update.player = nil
		w.gameGetResult, w.gameSubResult = 1, true
		experienceLevelUpdate4EF2E0(w.hooks())
		want := []string{"arg:unit", "update:unit=update", "player:update=nil", "game-get:1", "game-sub:true"}
		if !reflect.DeepEqual(w.events, want) {
			t.Fatalf("events = %q, want %q", w.events, want)
		}
	})

	t.Run("nil player faults at first level read", func(t *testing.T) {
		w := newExperienceLevelTestWorld4EF2E0()
		w.unit.update.player = nil
		experienceLevelMustPanic4EF2E0(t, func() { experienceLevelUpdate4EF2E0(w.hooks()) })
		if got := w.events[len(w.events)-1]; got != "level:nil" {
			t.Fatalf("last event = %q; all=%q", got, w.events)
		}
	})
}
