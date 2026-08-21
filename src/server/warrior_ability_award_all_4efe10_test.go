package server

import (
	"fmt"
	"reflect"
	"testing"
)

type warriorAbilityAwardAllTestPlayer4EFE10 struct {
	name       string
	class      uint8
	protection uint32
	levels     [6]uint32
}

type warriorAbilityAwardAllTestWorld4EFE10 struct {
	player      *warriorAbilityAwardAllTestPlayer4EFE10
	engineFlags uint8
	events      []string
	faultAt     int
	after       map[string]func()
}

func warriorAbilityAwardAllPlayerName4EFE10(player *warriorAbilityAwardAllTestPlayer4EFE10) string {
	if player == nil {
		return "nil"
	}
	return player.name
}

func (w *warriorAbilityAwardAllTestWorld4EFE10) record(event string) {
	w.events = append(w.events, event)
	if w.faultAt != 0 && len(w.events) == w.faultAt {
		panic(event)
	}
}

func (w *warriorAbilityAwardAllTestWorld4EFE10) finish(event string) {
	if after := w.after[event]; after != nil {
		after()
	}
}

func (w *warriorAbilityAwardAllTestWorld4EFE10) hooks() warriorAbilityAwardAllHooks4EFE10[*warriorAbilityAwardAllTestPlayer4EFE10] {
	return warriorAbilityAwardAllHooks4EFE10[*warriorAbilityAwardAllTestPlayer4EFE10]{
		loadPlayerArg: func() *warriorAbilityAwardAllTestPlayer4EFE10 {
			player := w.player
			event := "arg:" + warriorAbilityAwardAllPlayerName4EFE10(player)
			w.record(event)
			w.finish(event)
			return player
		},
		loadPlayerClass: func(player *warriorAbilityAwardAllTestPlayer4EFE10) uint8 {
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
		loadEngineFlags: func() uint8 {
			flags := w.engineFlags
			event := fmt.Sprintf("flags:%02x", flags)
			w.record(event)
			w.finish(event)
			return flags
		},
		storeAbilityLevel: func(player *warriorAbilityAwardAllTestPlayer4EFE10, index int32, value uint32) {
			event := fmt.Sprintf("store:%s:%d=%d", warriorAbilityAwardAllPlayerName4EFE10(player), index, value)
			w.record(event)
			if player == nil {
				panic(event)
			}
			player.levels[index] = value
			w.finish(event)
		},
		loadProtection: func(player *warriorAbilityAwardAllTestPlayer4EFE10) uint32 {
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
		awardProtection: func(protection uint32, index, level int32) {
			event := fmt.Sprintf("award:%08x:%d:%d", protection, index, level)
			w.record(event)
			w.finish(event)
		},
	}
}

func newWarriorAbilityAwardAllTestWorld4EFE10() *warriorAbilityAwardAllTestWorld4EFE10 {
	player := &warriorAbilityAwardAllTestPlayer4EFE10{
		name:       "player",
		protection: 0x12345678,
	}
	for index := range player.levels {
		player.levels[index] = uint32(0x1000 + index)
	}
	return &warriorAbilityAwardAllTestWorld4EFE10{
		player: player,
		after:  make(map[string]func()),
	}
}

func warriorAbilityAwardAllExpectedEvents4EFE10(engine uint8, protection uint32) []string {
	level := int32(0)
	if engine&warriorAbilityAwardAllAdminMask4EFE10 != 0 {
		level = warriorAbilityAwardAllAdminLevel4EFE10
	}
	events := []string{
		"arg:player",
		"class:player=0",
		fmt.Sprintf("flags:%02x", engine),
	}
	for index := warriorAbilityAwardAllFirstIndex4EFE10; index < warriorAbilityAwardAllEndIndex4EFE10; index++ {
		events = append(events,
			fmt.Sprintf("store:player:%d=%d", index, level),
			fmt.Sprintf("token:player=%08x", protection),
			fmt.Sprintf("award:%08x:%d:%d", protection, index, level),
		)
	}
	return events
}

func TestWarriorAbilityAwardAll4EFE10AdminAndDisabledPaths(t *testing.T) {
	for _, test := range []struct {
		name   string
		engine uint8
		level  uint32
	}{
		{name: "disabled zero", engine: 0x00, level: 0},
		{name: "disabled unrelated", engine: 0x20, level: 0},
		{name: "admin exact", engine: 0x10, level: 5},
		{name: "admin with unrelated", engine: 0x91, level: 5},
	} {
		t.Run(test.name, func(t *testing.T) {
			world := newWarriorAbilityAwardAllTestWorld4EFE10()
			world.engineFlags = test.engine
			warriorAbilityAwardAll4EFE10(world.hooks())

			want := warriorAbilityAwardAllExpectedEvents4EFE10(test.engine, 0x12345678)
			if !reflect.DeepEqual(world.events, want) {
				t.Fatalf("events = %v, want %v", world.events, want)
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

func TestWarriorAbilityAwardAll4EFE10NonWarriorReturnsBeforeFlags(t *testing.T) {
	for _, class := range []uint8{1, 2, 0xff} {
		t.Run(fmt.Sprintf("class-%d", class), func(t *testing.T) {
			world := newWarriorAbilityAwardAllTestWorld4EFE10()
			world.player.class = class
			before := world.player.levels
			warriorAbilityAwardAll4EFE10(world.hooks())

			want := []string{"arg:player", fmt.Sprintf("class:player=%d", class)}
			if !reflect.DeepEqual(world.events, want) {
				t.Fatalf("events = %v, want %v", world.events, want)
			}
			if world.player.levels != before {
				t.Fatalf("levels changed for class %d", class)
			}
		})
	}
}

func TestWarriorAbilityAwardAll4EFE10CachedPlayerDecisionAndLiveToken(t *testing.T) {
	world := newWarriorAbilityAwardAllTestWorld4EFE10()
	original := world.player
	replacement := &warriorAbilityAwardAllTestPlayer4EFE10{name: "replacement", protection: 0xaaaaaaaa}
	world.after["class:player=0"] = func() {
		world.engineFlags = 0x10
		original.class = 2
	}
	world.after["award:12345678:1:5"] = func() {
		original.protection = 0x87654321
		world.player = replacement
		world.engineFlags = 0
	}
	warriorAbilityAwardAll4EFE10(world.hooks())

	if world.events[2] != "flags:10" {
		t.Fatalf("flags event = %q, want post-class Admin", world.events[2])
	}
	if got := world.events[6:9]; !reflect.DeepEqual(got, []string{
		"store:player:2=5", "token:player=87654321", "award:87654321:2:5",
	}) {
		t.Fatalf("second iteration events = %v", got)
	}
	for index := 1; index < len(original.levels); index++ {
		if original.levels[index] != 5 {
			t.Fatalf("cached Player/decision level[%d] = %d, want 5", index, original.levels[index])
		}
	}
	if replacement.levels != [6]uint32{} {
		t.Fatalf("replacement levels changed: %v", replacement.levels)
	}
}

func TestWarriorAbilityAwardAll4EFE10HasNoNilPlayerGuard(t *testing.T) {
	world := newWarriorAbilityAwardAllTestWorld4EFE10()
	world.player = nil
	defer func() {
		if got := recover(); got != "class:nil" {
			t.Fatalf("panic = %v, want class:nil", got)
		}
		want := []string{"arg:nil", "class:nil"}
		if !reflect.DeepEqual(world.events, want) {
			t.Fatalf("events = %v, want %v", world.events, want)
		}
	}()
	warriorAbilityAwardAll4EFE10(world.hooks())
}

func TestWarriorAbilityAwardAll4EFE10EveryObservableFaultPrefix(t *testing.T) {
	base := func() *warriorAbilityAwardAllTestWorld4EFE10 {
		world := newWarriorAbilityAwardAllTestWorld4EFE10()
		world.engineFlags = 0x10
		return world
	}
	complete := base()
	warriorAbilityAwardAll4EFE10(complete.hooks())
	want := append([]string(nil), complete.events...)
	if len(want) != 18 {
		t.Fatalf("observable events = %d, want 18", len(want))
	}

	for fault := 1; fault <= len(want); fault++ {
		t.Run(fmt.Sprintf("event-%02d", fault), func(t *testing.T) {
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
			warriorAbilityAwardAll4EFE10(world.hooks())
		})
	}
}
