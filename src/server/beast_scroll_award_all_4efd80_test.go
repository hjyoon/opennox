package server

import (
	"fmt"
	"reflect"
	"testing"
)

type beastScrollAwardAllTestPlayer4EFD80 struct {
	name       string
	protection uint32
	levels     [41]uint32
}

type beastScrollAwardAllTestWorld4EFD80 struct {
	player      *beastScrollAwardAllTestPlayer4EFD80
	engineFlags uint8
	events      []string
	faultAt     int
	after       map[string]func()
}

func beastScrollAwardAllPlayerName4EFD80(player *beastScrollAwardAllTestPlayer4EFD80) string {
	if player == nil {
		return "nil"
	}
	return player.name
}

func (w *beastScrollAwardAllTestWorld4EFD80) record(event string) {
	w.events = append(w.events, event)
	if w.faultAt != 0 && len(w.events) == w.faultAt {
		panic(event)
	}
}

func (w *beastScrollAwardAllTestWorld4EFD80) finish(event string) {
	if after := w.after[event]; after != nil {
		after()
	}
}

func (w *beastScrollAwardAllTestWorld4EFD80) hooks() beastScrollAwardAllHooks4EFD80[*beastScrollAwardAllTestPlayer4EFD80] {
	return beastScrollAwardAllHooks4EFD80[*beastScrollAwardAllTestPlayer4EFD80]{
		loadPlayerArg: func() *beastScrollAwardAllTestPlayer4EFD80 {
			player := w.player
			event := "arg:" + beastScrollAwardAllPlayerName4EFD80(player)
			w.record(event)
			w.finish(event)
			return player
		},
		loadProtection: func(player *beastScrollAwardAllTestPlayer4EFD80) uint32 {
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
		storeScrollLevel: func(player *beastScrollAwardAllTestPlayer4EFD80, index int32, value uint32) {
			event := fmt.Sprintf("store:%s:%d=%d", beastScrollAwardAllPlayerName4EFD80(player), index, value)
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
		applyProtection: func(protection uint32, player *beastScrollAwardAllTestPlayer4EFD80, count int32) {
			event := fmt.Sprintf("apply:%08x:%s:%d", protection, beastScrollAwardAllPlayerName4EFD80(player), count)
			w.record(event)
			w.finish(event)
		},
	}
}

func newBeastScrollAwardAllTestWorld4EFD80() *beastScrollAwardAllTestWorld4EFD80 {
	player := &beastScrollAwardAllTestPlayer4EFD80{
		name:       "player",
		protection: 0x12345678,
	}
	for index := range player.levels {
		player.levels[index] = uint32(0x1000 + index)
	}
	return &beastScrollAwardAllTestWorld4EFD80{
		player: player,
		after:  make(map[string]func()),
	}
}

func beastScrollAwardAllExpectedEvents4EFD80(engine uint8, protection uint32) []string {
	level := int32(0)
	if engine&beastScrollAwardAllAdminMask4EFD80 != 0 {
		level = beastScrollAwardAllAdminLevel4EFD80
	}
	events := []string{
		"arg:player",
		"token:player=12345678",
		"reset:12345678:0",
		fmt.Sprintf("flags:%02x", engine),
	}
	for index := int32(1); index < beastScrollAwardAllLevelCount4EFD80; index++ {
		events = append(events,
			fmt.Sprintf("store:player:%d=%d", index, level),
			fmt.Sprintf("token:player=%08x", protection),
			fmt.Sprintf("award:%08x:%d:%d", protection, index, level),
		)
	}
	return append(events,
		fmt.Sprintf("token:player=%08x", protection),
		fmt.Sprintf("apply:%08x:player:41", protection),
	)
}

func TestBeastScrollAwardAll4EFD80AdminAndDisabledPaths(t *testing.T) {
	for _, test := range []struct {
		name   string
		engine uint8
		level  uint32
	}{
		{name: "disabled zero", engine: 0x00, level: 0},
		{name: "disabled unrelated", engine: 0x20, level: 0},
		{name: "admin exact", engine: 0x10, level: 1},
		{name: "admin with unrelated", engine: 0x91, level: 1},
	} {
		t.Run(test.name, func(t *testing.T) {
			world := newBeastScrollAwardAllTestWorld4EFD80()
			world.engineFlags = test.engine
			beastScrollAwardAll4EFD80(world.hooks())

			want := beastScrollAwardAllExpectedEvents4EFD80(test.engine, 0x12345678)
			if !reflect.DeepEqual(world.events, want) {
				t.Fatalf("events differ: got %d, want %d; tail=%v", len(world.events), len(want), world.events[len(world.events)-4:])
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

func TestBeastScrollAwardAll4EFD80LiveReloadsAndCachedDecision(t *testing.T) {
	world := newBeastScrollAwardAllTestWorld4EFD80()
	world.after["reset:12345678:0"] = func() {
		world.engineFlags = 0x10
	}
	world.after["award:12345678:1:1"] = func() {
		world.player.protection = 0x87654321
		world.engineFlags = 0
	}
	beastScrollAwardAll4EFD80(world.hooks())

	if world.events[3] != "flags:10" {
		t.Fatalf("flags event = %q, want post-reset admin", world.events[3])
	}
	if got := world.events[8:10]; !reflect.DeepEqual(got, []string{
		"token:player=87654321", "award:87654321:2:1",
	}) {
		t.Fatalf("second live token events = %v", got)
	}
	if got := world.events[len(world.events)-2:]; !reflect.DeepEqual(got, []string{
		"token:player=87654321", "apply:87654321:player:41",
	}) {
		t.Fatalf("final events = %v", got)
	}
	for index := 1; index < len(world.player.levels); index++ {
		if world.player.levels[index] != 1 {
			t.Fatalf("cached Admin decision changed level[%d] to %d", index, world.player.levels[index])
		}
	}
}

func TestBeastScrollAwardAll4EFD80HasNoNilPlayerGuard(t *testing.T) {
	world := newBeastScrollAwardAllTestWorld4EFD80()
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
	beastScrollAwardAll4EFD80(world.hooks())
}

func TestBeastScrollAwardAll4EFD80EveryObservableFaultPrefix(t *testing.T) {
	base := func() *beastScrollAwardAllTestWorld4EFD80 {
		world := newBeastScrollAwardAllTestWorld4EFD80()
		world.engineFlags = 0x10
		return world
	}
	complete := base()
	beastScrollAwardAll4EFD80(complete.hooks())
	want := append([]string(nil), complete.events...)
	if len(want) != 126 {
		t.Fatalf("observable events = %d, want 126", len(want))
	}

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
			beastScrollAwardAll4EFD80(world.hooks())
		})
	}
}
