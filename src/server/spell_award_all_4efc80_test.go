package server

import (
	"fmt"
	"reflect"
	"testing"
)

type spellAwardAllTestObject4EFC80 struct {
	name string
}

type spellAwardAllTestPlayer4EFC80 struct {
	name       string
	protection uint32
	levels     [137]uint32
	class      uint8
	unit       *spellAwardAllTestObject4EFC80
}

type spellAwardAllTestWorld4EFC80 struct {
	player      *spellAwardAllTestPlayer4EFC80
	engineFlags uint8
	gameResult  int32
	events      []string
	faultAt     int
	after       map[string]func()
}

func spellAwardAllPlayerName4EFC80(player *spellAwardAllTestPlayer4EFC80) string {
	if player == nil {
		return "nil"
	}
	return player.name
}

func spellAwardAllObjectName4EFC80(unit *spellAwardAllTestObject4EFC80) string {
	if unit == nil {
		return "nil"
	}
	return unit.name
}

func (w *spellAwardAllTestWorld4EFC80) record(event string) {
	w.events = append(w.events, event)
	if w.faultAt != 0 && len(w.events) == w.faultAt {
		panic(event)
	}
}

func (w *spellAwardAllTestWorld4EFC80) finish(event string) {
	if after := w.after[event]; after != nil {
		after()
	}
}

func (w *spellAwardAllTestWorld4EFC80) hooks() spellAwardAllHooks4EFC80[
	*spellAwardAllTestPlayer4EFC80,
	*spellAwardAllTestObject4EFC80,
] {
	return spellAwardAllHooks4EFC80[
		*spellAwardAllTestPlayer4EFC80,
		*spellAwardAllTestObject4EFC80,
	]{
		loadPlayerArg: func() *spellAwardAllTestPlayer4EFC80 {
			player := w.player
			event := "arg:" + spellAwardAllPlayerName4EFC80(player)
			w.record(event)
			w.finish(event)
			return player
		},
		loadProtection: func(player *spellAwardAllTestPlayer4EFC80) uint32 {
			if player == nil {
				event := "token:nil"
				w.record(event)
				panic(event)
			}
			protection := player.protection
			event := fmt.Sprintf("token:%s=%08x", player.name, protection)
			w.record(event)
			w.finish(event)
			return protection
		},
		resetProtection: func(protection uint32, value int32) {
			event := fmt.Sprintf("reset:%08x:%d", protection, value)
			w.record(event)
			w.finish(event)
		},
		loadEngineFlags: func() uint8 {
			flags := w.engineFlags
			event := fmt.Sprintf("flags:%02x", flags)
			w.record(event)
			w.finish(event)
			return flags
		},
		storeSpellLevel: func(player *spellAwardAllTestPlayer4EFC80, index int32, value uint32) {
			event := fmt.Sprintf("store:%s:%d=%d", spellAwardAllPlayerName4EFC80(player), index, value)
			w.record(event)
			if player == nil {
				panic(event)
			}
			player.levels[index] = value
			w.finish(event)
		},
		awardProtection: func(protection uint32, index, level int32) {
			event := fmt.Sprintf("award:%08x:%d:%d", protection, index, level)
			w.record(event)
			w.finish(event)
		},
		gameFlagsCheck: func(mask uint32) int32 {
			result := w.gameResult
			event := fmt.Sprintf("game:%08x=%d", mask, result)
			w.record(event)
			w.finish(event)
			return result
		},
		loadPlayerClass: func(player *spellAwardAllTestPlayer4EFC80) uint8 {
			if player == nil {
				event := "class:nil"
				w.record(event)
				panic(event)
			}
			class := player.class
			event := fmt.Sprintf("class:%s=%d", player.name, class)
			w.record(event)
			w.finish(event)
			return class
		},
		loadPlayerUnit: func(player *spellAwardAllTestPlayer4EFC80) *spellAwardAllTestObject4EFC80 {
			if player == nil {
				event := "unit:nil-player"
				w.record(event)
				panic(event)
			}
			unit := player.unit
			event := "unit:" + player.name + "=" + spellAwardAllObjectName4EFC80(unit)
			w.record(event)
			w.finish(event)
			return unit
		},
		grantSpell: func(unit *spellAwardAllTestObject4EFC80, spellID, a3, a4, a5 int32) {
			event := fmt.Sprintf("grant:%s:%d:%d:%d:%d", spellAwardAllObjectName4EFC80(unit), spellID, a3, a4, a5)
			w.record(event)
			w.finish(event)
		},
		applyProtection: func(protection uint32, player *spellAwardAllTestPlayer4EFC80, count int32) {
			event := fmt.Sprintf("apply:%08x:%s:%d", protection, spellAwardAllPlayerName4EFC80(player), count)
			w.record(event)
			w.finish(event)
		},
	}
}

func newSpellAwardAllTestWorld4EFC80() *spellAwardAllTestWorld4EFC80 {
	player := &spellAwardAllTestPlayer4EFC80{
		name:       "player",
		protection: 0x12345678,
		unit:       &spellAwardAllTestObject4EFC80{name: "unit"},
	}
	for index := range player.levels {
		player.levels[index] = uint32(0x1000 + index)
	}
	return &spellAwardAllTestWorld4EFC80{
		player: player,
		after:  make(map[string]func()),
	}
}

func spellAwardAllLoopEvents4EFC80(level int32, protection uint32) []string {
	events := make([]string, 0, 136*3)
	for index := int32(1); index < 137; index++ {
		events = append(events,
			fmt.Sprintf("store:player:%d=%d", index, level),
			fmt.Sprintf("token:player=%08x", protection),
			fmt.Sprintf("award:%08x:%d:%d", protection, index, level),
		)
	}
	return events
}

func spellAwardAllExpectedEvents4EFC80(engine uint8, game int32, class uint8) []string {
	level := int32(0)
	if engine&0x10 != 0 {
		level = 3
	}
	events := []string{
		"arg:player",
		"token:player=12345678",
		"reset:12345678:0",
		fmt.Sprintf("flags:%02x", engine),
	}
	events = append(events, spellAwardAllLoopEvents4EFC80(level, 0x12345678)...)
	if engine&0x10 == 0 {
		events = append(events, fmt.Sprintf("game:00001000=%d", game))
		if game != 0 {
			events = append(events, fmt.Sprintf("class:player=%d", class))
			switch class {
			case 1:
				events = append(events, "unit:player=unit", "grant:unit:27:1:1:1")
			case 2:
				events = append(events,
					"unit:player=unit", "grant:unit:9:1:1:1",
					"unit:player=unit", "grant:unit:41:1:1:1",
				)
			}
		}
	}
	return append(events, "token:player=12345678", "apply:12345678:player:137")
}

func TestSpellAwardAll4EFC80AdminAndDisabledPaths(t *testing.T) {
	for _, test := range []struct {
		name   string
		engine uint8
		level  uint32
	}{
		{name: "disabled zero", engine: 0x00, level: 0},
		{name: "disabled unrelated", engine: 0x20, level: 0},
		{name: "admin exact", engine: 0x10, level: 3},
		{name: "admin with unrelated", engine: 0x91, level: 3},
	} {
		t.Run(test.name, func(t *testing.T) {
			world := newSpellAwardAllTestWorld4EFC80()
			world.engineFlags = test.engine
			spellAwardAll4EFC80(world.hooks())

			if got, want := world.events, spellAwardAllExpectedEvents4EFC80(test.engine, 0, 0); !reflect.DeepEqual(got, want) {
				t.Fatalf("events differ: got %d, want %d; tail=%v", len(got), len(want), got[len(got)-4:])
			}
			if got := world.player.levels[0]; got != 0x1000 {
				t.Fatalf("level[0] = %#x, want unchanged", got)
			}
			for index := 1; index < len(world.player.levels); index++ {
				if got := world.player.levels[index]; got != test.level {
					t.Fatalf("level[%d] = %d, want %d", index, got, test.level)
				}
			}
		})
	}
}

func TestSpellAwardAll4EFC80QuestDefaults(t *testing.T) {
	for _, test := range []struct {
		name  string
		class uint8
	}{
		{name: "warrior", class: 0},
		{name: "wizard", class: 1},
		{name: "conjurer", class: 2},
		{name: "unknown", class: 0xff},
	} {
		t.Run(test.name, func(t *testing.T) {
			world := newSpellAwardAllTestWorld4EFC80()
			world.engineFlags = 0x04
			world.gameResult = -7
			world.player.class = test.class
			spellAwardAll4EFC80(world.hooks())
			want := spellAwardAllExpectedEvents4EFC80(0x04, -7, test.class)
			if !reflect.DeepEqual(world.events, want) {
				t.Fatalf("events differ: got tail=%v, want tail=%v", world.events[len(world.events)-8:], want[len(want)-8:])
			}
		})
	}
}

func TestSpellAwardAll4EFC80LiveReloadsAndCachedDecision(t *testing.T) {
	world := newSpellAwardAllTestWorld4EFC80()
	world.gameResult = 1
	world.after["reset:12345678:0"] = func() {
		world.engineFlags = 0x10
	}
	world.after["award:12345678:1:3"] = func() {
		world.player.protection = 0x87654321
		world.engineFlags = 0
	}
	spellAwardAll4EFC80(world.hooks())
	if world.events[3] != "flags:10" {
		t.Fatalf("flags event = %q, want post-reset admin", world.events[3])
	}
	if world.events[8] != "token:player=87654321" || world.events[9] != "award:87654321:2:3" {
		t.Fatalf("second live token events = %v", world.events[8:10])
	}
	for _, event := range world.events {
		if len(event) >= 5 && event[:5] == "game:" {
			t.Fatalf("cached Admin path queried game flags: %v", event)
		}
	}
	if got := world.events[len(world.events)-2:]; !reflect.DeepEqual(got, []string{
		"token:player=87654321", "apply:87654321:player:137",
	}) {
		t.Fatalf("final events = %v", got)
	}
}

func TestSpellAwardAll4EFC80QuestClassAndUnitAreLive(t *testing.T) {
	world := newSpellAwardAllTestWorld4EFC80()
	world.gameResult = 1
	world.player.class = 1
	replacement := &spellAwardAllTestObject4EFC80{name: "replacement"}
	world.after["game:00001000=1"] = func() {
		world.player.class = 2
	}
	world.after["grant:unit:9:1:1:1"] = func() {
		world.player.unit = replacement
		world.player.protection = 0xaabbccdd
	}
	spellAwardAll4EFC80(world.hooks())
	wantTail := []string{
		"game:00001000=1",
		"class:player=2",
		"unit:player=unit",
		"grant:unit:9:1:1:1",
		"unit:player=replacement",
		"grant:replacement:41:1:1:1",
		"token:player=aabbccdd",
		"apply:aabbccdd:player:137",
	}
	if got := world.events[len(world.events)-len(wantTail):]; !reflect.DeepEqual(got, wantTail) {
		t.Fatalf("tail = %v, want %v", got, wantTail)
	}
}

func TestSpellAwardAll4EFC80HasNoNilPlayerGuard(t *testing.T) {
	world := newSpellAwardAllTestWorld4EFC80()
	world.player = nil
	defer func() {
		if got := recover(); got != "token:nil" {
			t.Fatalf("panic = %v, want token:nil", got)
		}
		want := []string{"arg:nil", "token:nil"}
		if !reflect.DeepEqual(world.events, want) {
			t.Fatalf("events = %v, want %v", world.events, want)
		}
	}()
	spellAwardAll4EFC80(world.hooks())
}

func TestSpellAwardAll4EFC80EveryObservableFaultPrefix(t *testing.T) {
	base := func() *spellAwardAllTestWorld4EFC80 {
		world := newSpellAwardAllTestWorld4EFC80()
		world.gameResult = 1
		world.player.class = 2
		return world
	}
	complete := base()
	spellAwardAll4EFC80(complete.hooks())
	want := append([]string(nil), complete.events...)

	for fault := 1; fault <= len(want); fault++ {
		t.Run(fmt.Sprintf("event-%03d", fault), func(t *testing.T) {
			world := base()
			world.faultAt = fault
			defer func() {
				if got := recover(); got != want[fault-1] {
					t.Fatalf("panic = %v, want %q", got, want[fault-1])
				}
				if prefix := want[:fault]; !reflect.DeepEqual(world.events, prefix) {
					t.Fatalf("events = %v, want prefix %v", world.events, prefix)
				}
			}()
			spellAwardAll4EFC80(world.hooks())
		})
	}
}
