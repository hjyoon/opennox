package server

import (
	"fmt"
	"math"
	"reflect"
	"testing"
)

type playerLevelSetTestPlayer4EF410 struct {
	name            string
	class           uint8
	level           uint8
	experienceToken uint32
	levelToken      uint32
	abilities       [6]uint32
}

type playerLevelSetTestUpdate4EF410 struct {
	name   string
	player *playerLevelSetTestPlayer4EF410
}

type playerLevelSetTestObject4EF410 struct {
	name       string
	experience float32
	update     *playerLevelSetTestUpdate4EF410
}

type playerLevelSetTestWorld4EF410 struct {
	unit       *playerLevelSetTestObject4EF410
	levelArg   uint8
	tables     []float64
	flags      []int32
	tableLoads int
	flagLoads  int
	events     []string
	after      map[string]func()
}

func playerLevelSetObjectName4EF410(unit *playerLevelSetTestObject4EF410) string {
	if unit == nil {
		return "nil"
	}
	return unit.name
}

func playerLevelSetUpdateName4EF410(update *playerLevelSetTestUpdate4EF410) string {
	if update == nil {
		return "nil"
	}
	return update.name
}

func playerLevelSetPlayerName4EF410(player *playerLevelSetTestPlayer4EF410) string {
	if player == nil {
		return "nil"
	}
	return player.name
}

func (w *playerLevelSetTestWorld4EF410) record(event string) {
	w.events = append(w.events, event)
	if after := w.after[event]; after != nil {
		delete(w.after, event)
		after()
	}
}

func (w *playerLevelSetTestWorld4EF410) hooks() playerLevelSetHooks4EF410[
	*playerLevelSetTestObject4EF410,
	*playerLevelSetTestUpdate4EF410,
	*playerLevelSetTestPlayer4EF410,
] {
	return playerLevelSetHooks4EF410[
		*playerLevelSetTestObject4EF410,
		*playerLevelSetTestUpdate4EF410,
		*playerLevelSetTestPlayer4EF410,
	]{
		loadLevelArg: func() uint8 {
			level := w.levelArg
			w.record(fmt.Sprintf("level-arg:%02x", level))
			return level
		},
		loadUnitArg: func() *playerLevelSetTestObject4EF410 {
			unit := w.unit
			w.record("arg:" + playerLevelSetObjectName4EF410(unit))
			return unit
		},
		loadUpdateData: func(unit *playerLevelSetTestObject4EF410) *playerLevelSetTestUpdate4EF410 {
			name := playerLevelSetObjectName4EF410(unit)
			if unit == nil {
				event := "update:" + name
				w.record(event)
				panic(event)
			}
			update := unit.update
			w.record("update:" + name + "=" + playerLevelSetUpdateName4EF410(update))
			return update
		},
		loadPlayer: func(update *playerLevelSetTestUpdate4EF410) *playerLevelSetTestPlayer4EF410 {
			name := playerLevelSetUpdateName4EF410(update)
			if update == nil {
				event := "player:" + name
				w.record(event)
				panic(event)
			}
			player := update.player
			w.record("player:" + name + "=" + playerLevelSetPlayerName4EF410(player))
			return player
		},
		loadXPTable: func(key string, index int32) float64 {
			w.tableLoads++
			value := w.tables[w.tableLoads-1]
			w.record(fmt.Sprintf("table:%d:%s:%d=%016x", w.tableLoads, key, index, math.Float64bits(value)))
			return value
		},
		storeExperience: func(unit *playerLevelSetTestObject4EF410, value float32) {
			event := fmt.Sprintf("store-experience:%s=%08x", playerLevelSetObjectName4EF410(unit), math.Float32bits(value))
			w.record(event)
			if unit == nil {
				panic(event)
			}
			unit.experience = value
		},
		loadExperienceToken: func(player *playerLevelSetTestPlayer4EF410) uint32 {
			name := playerLevelSetPlayerName4EF410(player)
			if player == nil {
				event := "experience-token:" + name
				w.record(event)
				panic(event)
			}
			token := player.experienceToken
			w.record(fmt.Sprintf("experience-token:%s=%08x", name, token))
			return token
		},
		protectExperience: func(token uint32, value float32) {
			w.record(fmt.Sprintf("protect-experience:%08x:%08x", token, math.Float32bits(value)))
		},
		reportExperience: func(unit *playerLevelSetTestObject4EF410) {
			w.record("report:" + playerLevelSetObjectName4EF410(unit))
		},
		loadLevelToken: func(player *playerLevelSetTestPlayer4EF410) uint32 {
			name := playerLevelSetPlayerName4EF410(player)
			if player == nil {
				event := "level-token:" + name
				w.record(event)
				panic(event)
			}
			token := player.levelToken
			w.record(fmt.Sprintf("level-token:%s=%08x", name, token))
			return token
		},
		storeLevel: func(player *playerLevelSetTestPlayer4EF410, level uint8) {
			event := fmt.Sprintf("store-level:%s=%02x", playerLevelSetPlayerName4EF410(player), level)
			w.record(event)
			if player == nil {
				panic(event)
			}
			player.level = level
		},
		protectLevel: func(token uint32, level uint8) {
			w.record(fmt.Sprintf("protect-level:%08x:%02x", token, level))
		},
		readValues: func(unit *playerLevelSetTestObject4EF410, value int32) {
			w.record(fmt.Sprintf("read-values:%s:%d", playerLevelSetObjectName4EF410(unit), value))
		},
		gameFlag: func(flag uint32) int32 {
			w.flagLoads++
			value := w.flags[w.flagLoads-1]
			w.record(fmt.Sprintf("flag:%d:%08x=%d", w.flagLoads, flag, value))
			return value
		},
		loadPlayerClass: func(player *playerLevelSetTestPlayer4EF410) uint8 {
			name := playerLevelSetPlayerName4EF410(player)
			if player == nil {
				event := "class:" + name
				w.record(event)
				panic(event)
			}
			class := player.class
			w.record(fmt.Sprintf("class:%s=%02x", name, class))
			return class
		},
		loadAbilityLevel: func(player *playerLevelSetTestPlayer4EF410, ability int32) uint32 {
			name := playerLevelSetPlayerName4EF410(player)
			if player == nil {
				event := fmt.Sprintf("ability:%s:%d", name, ability)
				w.record(event)
				panic(event)
			}
			value := player.abilities[ability]
			w.record(fmt.Sprintf("ability:%s:%d=%08x", name, ability, value))
			return value
		},
		bookAbility: func(kind, ability, index int32) {
			w.record(fmt.Sprintf("book:%d:%d:%d", kind, ability, index))
		},
		pauseFX: func(unit *playerLevelSetTestObject4EF410, value int32) {
			w.record(fmt.Sprintf("pause:%s:%d", playerLevelSetObjectName4EF410(unit), value))
		},
	}
}

func newPlayerLevelSetTestWorld4EF410() *playerLevelSetTestWorld4EF410 {
	player := &playerLevelSetTestPlayer4EF410{
		name:            "player",
		level:           2,
		experienceToken: 0x11223344,
		levelToken:      0x55667788,
	}
	player.abilities = [6]uint32{99, 11, 0, 33, 0, 55}
	update := &playerLevelSetTestUpdate4EF410{name: "update", player: player}
	return &playerLevelSetTestWorld4EF410{
		unit:     &playerLevelSetTestObject4EF410{name: "unit", experience: 1, update: update},
		levelArg: 4,
		tables:   []float64{100.25, 200.5},
		flags:    []int32{1, 1},
		after:    make(map[string]func()),
	}
}

func playerLevelSetMustPanic4EF410(t *testing.T, run func()) {
	t.Helper()
	defer func() {
		if recover() == nil {
			t.Fatal("call did not panic")
		}
	}()
	run()
}

func TestPlayerLevelSet4EF410OrderCachingAndLiveReads(t *testing.T) {
	w := newPlayerLevelSetTestWorld4EF410()
	entryUnit := w.unit
	entryPlayer := entryUnit.update.player
	replacementPlayer := &playerLevelSetTestPlayer4EF410{name: "replacement", level: 90}
	replacementUpdate := &playerLevelSetTestUpdate4EF410{name: "replacement-update", player: replacementPlayer}
	replacementUnit := &playerLevelSetTestObject4EF410{name: "replacement-unit", update: replacementUpdate}

	w.after["level-arg:04"] = func() { w.levelArg = 99 }
	w.after["arg:unit"] = func() { w.unit = replacementUnit }
	w.after["update:unit=update"] = func() { entryUnit.update = replacementUpdate }
	w.after["player:update=player"] = func() { entryUnit.update.player = replacementPlayer }
	w.after["table:1:XPTable:4=4059100000000000"] = func() {
		w.tables[1] = 300.75
		entryPlayer.experienceToken = 0xaabbccdd
	}
	w.after["table:2:XPTable:4=4072cc0000000000"] = func() {
		entryUnit.experience = 999
		entryPlayer.experienceToken = 0x10203040
	}
	w.after["protect-experience:10203040:43966000"] = func() { entryPlayer.levelToken = 0xdeadbeef }
	w.after["report:unit"] = func() { entryPlayer.levelToken = 0xcafebabe }
	w.after["level-token:player=cafebabe"] = func() { entryPlayer.levelToken = 7 }
	w.after["protect-level:cafebabe:04"] = func() { entryPlayer.level = 77 }
	w.after["read-values:unit:0"] = func() {
		entryPlayer.class = 0
		entryPlayer.abilities = [6]uint32{0, 1, 2, 0, 4, 5}
	}
	w.after["flag:1:00000800=1"] = func() { entryPlayer.class = 0 }
	w.after["ability:player:1=00000001"] = func() { entryPlayer.abilities[2] = 22 }
	w.after["book:3:2:1"] = func() { w.flags[1] = -1 }

	playerLevelSet4EF410(w.hooks())
	want := []string{
		"level-arg:04",
		"arg:unit",
		"update:unit=update",
		"player:update=player",
		"table:1:XPTable:4=4059100000000000",
		"store-experience:unit=42c88000",
		"table:2:XPTable:4=4072cc0000000000",
		"experience-token:player=10203040",
		"protect-experience:10203040:43966000",
		"report:unit",
		"level-token:player=cafebabe",
		"store-level:player=04",
		"protect-level:cafebabe:04",
		"read-values:unit:0",
		"flag:1:00000800=1",
		"class:player=00",
		"ability:player:1=00000001",
		"book:3:1:0",
		"ability:player:2=00000016",
		"book:3:2:1",
		"ability:player:3=00000000",
		"ability:player:4=00000004",
		"book:3:4:3",
		"ability:player:5=00000005",
		"book:3:5:4",
		"flag:2:00000800=-1",
		"pause:unit:0",
	}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %q, want %q", w.events, want)
	}
	if w.unit != replacementUnit || entryUnit.update != replacementUpdate || replacementPlayer.level != 90 {
		t.Fatalf("entry caches were not retained")
	}
	if entryUnit.experience != 999 || entryPlayer.level != 77 {
		t.Fatalf("callback mutations lost: experience=%v level=%d", entryUnit.experience, entryPlayer.level)
	}
}

func TestPlayerLevelSet4EF410SignedClampAndTwoSpills(t *testing.T) {
	for _, test := range []struct {
		name      string
		input     uint8
		wantLevel uint8
		wantIndex int32
	}{
		{name: "zero", input: 0, wantLevel: 0, wantIndex: 0},
		{name: "ten", input: 10, wantLevel: 10, wantIndex: 10},
		{name: "eleven clamps", input: 11, wantLevel: 10, wantIndex: 10},
		{name: "signed maximum clamps", input: 0x7f, wantLevel: 10, wantIndex: 10},
		{name: "signed minimum stays", input: 0x80, wantLevel: 0x80, wantIndex: -128},
		{name: "minus one stays", input: 0xff, wantLevel: 0xff, wantIndex: -1},
	} {
		t.Run(test.name, func(t *testing.T) {
			w := newPlayerLevelSetTestWorld4EF410()
			w.levelArg = test.input
			w.tables = []float64{
				math.Float64frombits(0x4059100000000001),
				math.Float64frombits(0x4072cc0000000001),
			}
			w.flags = []int32{0, 0}
			playerLevelSet4EF410(w.hooks())
			if got := w.unit.update.player.level; got != test.wantLevel {
				t.Fatalf("level = %02x, want %02x", got, test.wantLevel)
			}
			if got := math.Float32bits(w.unit.experience); got != math.Float32bits(float32(w.tables[0])) {
				t.Fatalf("experience bits = %08x, want %08x", got, math.Float32bits(float32(w.tables[0])))
			}
			for i := 1; i <= 2; i++ {
				prefix := fmt.Sprintf("table:%d:XPTable:%d=", i, test.wantIndex)
				found := false
				for _, event := range w.events {
					found = found || len(event) >= len(prefix) && event[:len(prefix)] == prefix
				}
				if !found {
					t.Fatalf("table %d signed index absent from %q", i, w.events)
				}
			}
			wantProtect := fmt.Sprintf("protect-experience:11223344:%08x", math.Float32bits(float32(w.tables[1])))
			if !containsPlayerLevelSetEvent4EF410(w.events, wantProtect) {
				t.Fatalf("rounded protection %q absent from %q", wantProtect, w.events)
			}
		})
	}
}

func containsPlayerLevelSetEvent4EF410(events []string, want string) bool {
	for _, event := range events {
		if event == want {
			return true
		}
	}
	return false
}

func TestPlayerLevelSet4EF410IndependentCoopChecks(t *testing.T) {
	for _, test := range []struct {
		name          string
		flags         []int32
		class         uint8
		wantClassRead bool
		wantAbilities bool
		wantPause     bool
	}{
		{name: "both zero", flags: []int32{0, 0}},
		{name: "late enable", flags: []int32{0, 2}, wantPause: true},
		{name: "early only warrior", flags: []int32{-1, 0}, wantClassRead: true, wantAbilities: true},
		{name: "early nonwarrior", flags: []int32{1, 0}, class: 1, wantClassRead: true},
		{name: "both warrior", flags: []int32{1, 1}, wantClassRead: true, wantAbilities: true, wantPause: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			w := newPlayerLevelSetTestWorld4EF410()
			w.flags = test.flags
			w.unit.update.player.class = test.class
			playerLevelSet4EF410(w.hooks())
			hasClass, hasAbility, hasPause := false, false, false
			for _, event := range w.events {
				hasClass = hasClass || len(event) >= 6 && event[:6] == "class:"
				hasAbility = hasAbility || len(event) >= 8 && event[:8] == "ability:"
				hasPause = hasPause || len(event) >= 6 && event[:6] == "pause:"
			}
			if hasClass != test.wantClassRead || hasAbility != test.wantAbilities || hasPause != test.wantPause {
				t.Fatalf("class/abilities/pause = %t/%t/%t, want %t/%t/%t; events=%q", hasClass, hasAbility, hasPause, test.wantClassRead, test.wantAbilities, test.wantPause, w.events)
			}
		})
	}
}

func TestPlayerLevelSet4EF410NilFaultBoundaries(t *testing.T) {
	t.Run("nil unit faults after both argument reads", func(t *testing.T) {
		w := newPlayerLevelSetTestWorld4EF410()
		w.unit = nil
		playerLevelSetMustPanic4EF410(t, func() { playerLevelSet4EF410(w.hooks()) })
		want := []string{"level-arg:04", "arg:nil", "update:nil"}
		if !reflect.DeepEqual(w.events, want) {
			t.Fatalf("events = %q, want %q", w.events, want)
		}
	})

	t.Run("nil update faults before clamp and tables", func(t *testing.T) {
		w := newPlayerLevelSetTestWorld4EF410()
		w.unit.update = nil
		w.levelArg = 11
		playerLevelSetMustPanic4EF410(t, func() { playerLevelSet4EF410(w.hooks()) })
		want := []string{"level-arg:0b", "arg:unit", "update:unit=nil", "player:nil"}
		if !reflect.DeepEqual(w.events, want) {
			t.Fatalf("events = %q, want %q", w.events, want)
		}
	})

	t.Run("nil player faults only after both table callbacks and experience store", func(t *testing.T) {
		w := newPlayerLevelSetTestWorld4EF410()
		w.unit.update.player = nil
		playerLevelSetMustPanic4EF410(t, func() { playerLevelSet4EF410(w.hooks()) })
		if w.tableLoads != 2 || math.Float32bits(w.unit.experience) != math.Float32bits(100.25) || w.events[len(w.events)-1] != "experience-token:nil" {
			t.Fatalf("tables/experience/events = %d/%08x/%q", w.tableLoads, math.Float32bits(w.unit.experience), w.events)
		}
	})
}

func TestPlayerLevelSet4EF410EveryObservationFaultPrefix(t *testing.T) {
	baseline := newPlayerLevelSetTestWorld4EF410()
	baseline.unit.update.player.abilities = [6]uint32{0, 1, 2, 3, 4, 5}
	playerLevelSet4EF410(baseline.hooks())

	for stop, event := range baseline.events {
		t.Run(fmt.Sprintf("%02d_%s", stop, event), func(t *testing.T) {
			w := newPlayerLevelSetTestWorld4EF410()
			w.unit.update.player.abilities = [6]uint32{0, 1, 2, 3, 4, 5}
			w.after[event] = func() { panic("fault") }
			playerLevelSetMustPanic4EF410(t, func() { playerLevelSet4EF410(w.hooks()) })
			want := baseline.events[:stop+1]
			if !reflect.DeepEqual(w.events, want) {
				t.Fatalf("events = %q, want prefix %q", w.events, want)
			}
		})
	}
}
