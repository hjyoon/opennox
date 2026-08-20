package server

import (
	"fmt"
	"reflect"
	"testing"

	noxflags "github.com/opennox/opennox/v1/common/flags"
)

func preserveGodModeFlags4EF500(t *testing.T) {
	t.Helper()
	game := noxflags.GetGame()
	engine := noxflags.GetEngine()
	t.Cleanup(func() {
		noxflags.ResetGame()
		noxflags.SetGame(game)
		noxflags.ResetEngine()
		noxflags.SetEngine(engine)
	})
}

func TestGodModeController4EF500NativeCoopGate(t *testing.T) {
	preserveGodModeFlags4EF500(t)
	noxflags.ResetGame()
	noxflags.SetGame(noxflags.GameModeQuest)
	noxflags.ResetEngine()
	noxflags.SetEngine(noxflags.EngineFlag(0xa5000004))

	s := &Server{}
	s.Players.list = make([]Player, 2)
	s.Players.list[0].Active = 1
	s.Players.list[0].PlayerInd = 0
	fail := func(*Player) { t.Fatal("player callback crossed Coop gate") }
	s.GodModeController4EF500(1, GodModeControllerRuntime4EF500{
		AwardScrolls:   fail,
		AwardSpells:    fail,
		AwardAbilities: fail,
	})
	if got, want := uint32(noxflags.GetEngine()), uint32(0xa5000004); got != want {
		t.Fatalf("engine flags = %#08x, want %#08x", got, want)
	}
}

func TestGodModeController4EF500NativeFlagsAndLivePlayers(t *testing.T) {
	for _, test := range []struct {
		name        string
		value       uint32
		initial     uint32
		want        uint32
		adminAtCall bool
	}{
		{"exact-one-enables", 1, 0xa5000004, 0xa5000034, true},
		{"other-clears", 2, 0xa5000034, 0xa5000004, false},
	} {
		t.Run(test.name, func(t *testing.T) {
			preserveGodModeFlags4EF500(t)
			noxflags.ResetGame()
			noxflags.SetGame(noxflags.GameModeCoop | noxflags.GameModeQuest)
			noxflags.ResetEngine()
			noxflags.SetEngine(noxflags.EngineFlag(test.initial))

			s := &Server{}
			s.Players.list = make([]Player, 4)
			for i := range s.Players.list {
				s.Players.list[i].PlayerInd = uint8(i)
			}
			s.Players.list[0].Active = 1
			s.Players.list[1].Active = 1
			var events []string
			record := func(kind string, player *Player) {
				gotAdmin := noxflags.HasEngine(noxflags.EngineAdmin)
				events = append(events, fmt.Sprintf("%s:%d:%t", kind, player.Index(), gotAdmin))
				if gotAdmin != test.adminAtCall {
					t.Fatalf("%s player %d observed Admin=%t, want %t", kind, player.Index(), gotAdmin, test.adminAtCall)
				}
			}
			s.GodModeController4EF500(test.value, GodModeControllerRuntime4EF500{
				AwardScrolls: func(player *Player) {
					record("scrolls", player)
				},
				AwardSpells: func(player *Player) {
					record("spells", player)
				},
				AwardAbilities: func(player *Player) {
					record("abilities", player)
					if player.Index() == 0 {
						s.Players.list[1].Active = 0
						s.Players.list[2].Active = 1
					}
				},
			})

			if got := uint32(noxflags.GetEngine()); got != test.want {
				t.Fatalf("engine flags = %#08x, want %#08x", got, test.want)
			}
			wantEvents := []string{
				fmt.Sprintf("scrolls:0:%t", test.adminAtCall),
				fmt.Sprintf("spells:0:%t", test.adminAtCall),
				fmt.Sprintf("abilities:0:%t", test.adminAtCall),
				fmt.Sprintf("scrolls:2:%t", test.adminAtCall),
				fmt.Sprintf("spells:2:%t", test.adminAtCall),
				fmt.Sprintf("abilities:2:%t", test.adminAtCall),
			}
			if !reflect.DeepEqual(events, wantEvents) {
				t.Fatalf("events = %v, want %v", events, wantEvents)
			}
		})
	}
}

func TestStoreEngineFlags4EF500PreservesHostExtensionBits(t *testing.T) {
	if ^uint(0) == uint(^uint32(0)) {
		t.Skip("host uint has no extension bits")
	}
	preserveGodModeFlags4EF500(t)
	noxflags.ResetEngine()
	high := noxflags.EngineFlag(uint(1) << 40)
	noxflags.SetEngine(high | noxflags.EngineAdmin)
	storeEngineFlags4EF500(uint32(noxflags.EngineGodMode | noxflags.EngineFlag3))
	if got, want := noxflags.GetEngine(), high|noxflags.EngineGodMode|noxflags.EngineFlag3; got != want {
		t.Fatalf("engine flags = %#x, want %#x", got, want)
	}
}
