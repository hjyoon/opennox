package server

import (
	"fmt"
	"reflect"
	"runtime"
	"testing"
	"unsafe"

	playerlib "github.com/opennox/libs/player"

	noxflags "github.com/opennox/opennox/v1/common/flags"
)

func TestSpellAwardAll4EFC80NativeLayout(t *testing.T) {
	wantPlayerSize := uintptr(4828)
	wantInfo := uintptr(2185)
	wantClass := uintptr(2251)
	wantSpellLevels := uintptr(3696)
	wantProtection := uintptr(4636)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantPlayerSize = 6160
		wantInfo = 2189
		wantClass = 2255
		wantSpellLevels = 4992
		wantProtection = 5940
	}

	checks := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"Player size", unsafe.Sizeof(Player{}), wantPlayerSize},
		{"Player.PlayerUnit", unsafe.Offsetof(Player{}.PlayerUnit), 2056},
		{"Player.info", unsafe.Offsetof(Player{}.info), wantInfo},
		{"Player.info.playerClass", unsafe.Offsetof(Player{}.info) + unsafe.Offsetof(PlayerInfo{}.playerClass), wantClass},
		{"Player.SpellLvl", unsafe.Offsetof(Player{}.SpellLvl), wantSpellLevels},
		{"Player.Prot4636", unsafe.Offsetof(Player{}.Prot4636), wantProtection},
		{"PlayerInfo.playerClass", unsafe.Offsetof(PlayerInfo{}.playerClass), 66},
		{"spell level width", unsafe.Sizeof(Player{}.SpellLvl[0]), 4},
		{"spell level count", uintptr(len(Player{}.SpellLvl)), 137},
		{"protection width", unsafe.Sizeof(Player{}.Prot4636), 4},
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Errorf("%s on %s/%s = %d, want %d", check.name, runtime.GOOS, runtime.GOARCH, check.got, check.want)
		}
	}
}

func TestSpellAwardAllNative4EFC80BindsLivePlayerFields(t *testing.T) {
	firstUnit := &Object{}
	secondUnit := &Object{}
	player := &Player{PlayerUnit: firstUnit, Prot4636: 0x100}
	player.Info().SetPlayerClass(playerlib.Wizard)
	for index := range player.SpellLvl {
		player.SpellLvl[index] = uint32(0x1000 + index)
	}

	var events []string
	awardIndex := int32(0)
	spellAwardAllNative4EFC80(player, spellAwardAllNativeDeps4EFC80{
		resetProtection: func(token uint32, value int32) {
			events = append(events, fmt.Sprintf("reset:%x:%d", token, value))
			if token != 0x100 || value != 0 {
				t.Fatalf("reset args = %#x/%d, want 0x100/0", token, value)
			}
			player.Prot4636 = 0x200
		},
		loadEngineFlags: func() uint8 {
			events = append(events, "flags")
			if player.Prot4636 != 0x200 {
				t.Fatalf("engine flags loaded before reset mutation: token=%#x", player.Prot4636)
			}
			return 0
		},
		awardProtection: func(token uint32, index, level int32) {
			awardIndex++
			if index != awardIndex || level != 0 {
				t.Fatalf("award index/level = %d/%d, want %d/0", index, level, awardIndex)
			}
			if token != uint32(0x1ff+index) {
				t.Fatalf("award %d token = %#x, want %#x", index, token, uint32(0x1ff+index))
			}
			if player.SpellLvl[index] != 0 {
				t.Fatalf("award %d observed level %d before store", index, player.SpellLvl[index])
			}
			player.Prot4636++
		},
		gameFlagsCheck: func(mask uint32) int32 {
			events = append(events, fmt.Sprintf("game:%x", mask))
			if mask != 0x1000 || awardIndex != 136 {
				t.Fatalf("game args/state = %#x/%d, want 0x1000/136", mask, awardIndex)
			}
			player.Info().SetPlayerClass(playerlib.Conjurer)
			return -1
		},
		grantSpell: func(unit *Object, spellID, a3, a4, a5 int32) {
			events = append(events, fmt.Sprintf("grant:%p:%d:%d:%d:%d", unit, spellID, a3, a4, a5))
			if a3 != 1 || a4 != 1 || a5 != 1 {
				t.Fatalf("grant trailing args = %d/%d/%d, want 1/1/1", a3, a4, a5)
			}
			switch spellID {
			case 9:
				if unit != firstUnit {
					t.Fatalf("first Conjurer unit = %p, want %p", unit, firstUnit)
				}
				player.PlayerUnit = secondUnit
			case 41:
				if unit != secondUnit {
					t.Fatalf("second Conjurer unit = %p, want live %p", unit, secondUnit)
				}
			default:
				t.Fatalf("unexpected spell ID %d", spellID)
			}
		},
		applyProtection: func(token uint32, levels *[137]uint32, count int32) {
			events = append(events, fmt.Sprintf("apply:%x:%d", token, count))
			if token != 0x288 || levels != &player.SpellLvl || count != 137 {
				t.Fatalf("apply args = %#x/%p/%d, want 0x288/%p/137", token, levels, count, &player.SpellLvl)
			}
		},
	})

	wantEvents := []string{
		"reset:100:0",
		"flags",
		"game:1000",
		fmt.Sprintf("grant:%p:9:1:1:1", firstUnit),
		fmt.Sprintf("grant:%p:41:1:1:1", secondUnit),
		"apply:288:137",
	}
	if !reflect.DeepEqual(events, wantEvents) {
		t.Fatalf("events = %v, want %v", events, wantEvents)
	}
	if player.SpellLvl[0] != 0x1000 {
		t.Fatalf("spell level 0 = %#x, want unchanged", player.SpellLvl[0])
	}
	for index := 1; index < len(player.SpellLvl); index++ {
		if player.SpellLvl[index] != 0 {
			t.Fatalf("spell level %d = %d, want 0", index, player.SpellLvl[index])
		}
	}
}

func preserveSpellAwardAllFlags4EFC80(t *testing.T) {
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

func TestSpellAwardAll4EFC80ServerReadsFlagsAfterReset(t *testing.T) {
	preserveSpellAwardAllFlags4EFC80(t)
	noxflags.ResetGame()
	noxflags.SetGame(noxflags.GameModeQuest)
	noxflags.ResetEngine()
	noxflags.SetEngine(noxflags.EngineAdmin)

	unit := &Object{}
	player := &Player{PlayerUnit: unit, Prot4636: 0x12345678}
	player.Info().SetPlayerClass(playerlib.Wizard)
	awards := 0
	grants := 0
	applies := 0
	new(Server).SpellAwardAll4EFC80(player, SpellAwardAllRuntime4EFC80{
		ResetProtection: func(token uint32, value int32) {
			if token != 0x12345678 || value != 0 {
				t.Fatalf("reset args = %#x/%d", token, value)
			}
			noxflags.UnsetEngine(noxflags.EngineAdmin)
		},
		AwardProtection: func(token uint32, index, level int32) {
			awards++
			if token != 0x12345678 || index != int32(awards) || level != 0 {
				t.Fatalf("award %d args = %#x/%d/%d", awards, token, index, level)
			}
		},
		GrantSpell: func(got *Object, spellID, a3, a4, a5 int32) {
			grants++
			if got != unit || spellID != 27 || a3 != 1 || a4 != 1 || a5 != 1 {
				t.Fatalf("grant args = %p/%d/%d/%d/%d", got, spellID, a3, a4, a5)
			}
		},
		ApplyProtection: func(token uint32, levels *[137]uint32, count int32) {
			applies++
			if token != 0x12345678 || levels != &player.SpellLvl || count != 137 {
				t.Fatalf("apply args = %#x/%p/%d", token, levels, count)
			}
		},
	})
	if awards != 136 || grants != 1 || applies != 1 {
		t.Fatalf("award/grant/apply calls = %d/%d/%d, want 136/1/1", awards, grants, applies)
	}
}

func TestSpellAwardAllNative4EFC80NilPlayerFaultsBeforeServices(t *testing.T) {
	calls := 0
	deps := spellAwardAllNativeDeps4EFC80{
		resetProtection: func(uint32, int32) { calls++ },
		loadEngineFlags: func() uint8 { calls++; return 0 },
		awardProtection: func(uint32, int32, int32) { calls++ },
		gameFlagsCheck:  func(uint32) int32 { calls++; return 0 },
		grantSpell:      func(*Object, int32, int32, int32, int32) { calls++ },
		applyProtection: func(uint32, *[137]uint32, int32) { calls++ },
	}
	deferred := false
	func() {
		defer func() {
			if recover() == nil {
				t.Fatal("nil player did not fault")
			}
			deferred = true
		}()
		spellAwardAllNative4EFC80(nil, deps)
	}()
	if !deferred || calls != 0 {
		t.Fatalf("deferred/calls = %t/%d, want true/0", deferred, calls)
	}
}
