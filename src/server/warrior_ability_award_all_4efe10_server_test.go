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

func TestWarriorAbilityAwardAll4EFE10NativeLayout(t *testing.T) {
	wantPlayerSize := uintptr(4828)
	wantInfo := uintptr(2185)
	wantClass := uintptr(2251)
	wantAbilities := uintptr(3696)
	wantProtection := uintptr(4636)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantPlayerSize = 6160
		wantInfo = 2189
		wantClass = 2255
		wantAbilities = 4992
		wantProtection = 5940
	}

	checks := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"Player size", unsafe.Sizeof(Player{}), wantPlayerSize},
		{"Player.info", unsafe.Offsetof(Player{}.info), wantInfo},
		{"Player.info.playerClass", unsafe.Offsetof(Player{}.info) + unsafe.Offsetof(PlayerInfo{}.playerClass), wantClass},
		{"Player.SpellLvl", unsafe.Offsetof(Player{}.SpellLvl), wantAbilities},
		{"Player.Prot4636", unsafe.Offsetof(Player{}.Prot4636), wantProtection},
		{"PlayerInfo.playerClass", unsafe.Offsetof(PlayerInfo{}.playerClass), 66},
		{"ability level width", unsafe.Sizeof(Player{}.SpellLvl[0]), 4},
		{"ability level count", uintptr(len(Player{}.SpellLvl)), 137},
		{"protection width", unsafe.Sizeof(Player{}.Prot4636), 4},
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Errorf("%s on %s/%s = %d, want %d", check.name, runtime.GOOS, runtime.GOARCH, check.got, check.want)
		}
	}
}

func TestWarriorAbilityAwardAllNative4EFE10BindsLivePlayerFields(t *testing.T) {
	player := &Player{Prot4636: 0x100}
	player.Info().SetPlayerClass(playerlib.Warrior)
	for index := range player.SpellLvl {
		player.SpellLvl[index] = uint32(0x1000 + index)
	}

	var events []string
	awardIndex := int32(0)
	warriorAbilityAwardAllNative4EFE10(player, warriorAbilityAwardAllNativeDeps4EFE10{
		loadEngineFlags: func() uint8 {
			events = append(events, "flags")
			player.Info().SetPlayerClass(playerlib.Conjurer)
			return 0x10
		},
		awardProtection: func(token uint32, index, level int32) {
			awardIndex++
			if index != awardIndex || level != 5 {
				t.Fatalf("award index/level = %d/%d, want %d/5", index, level, awardIndex)
			}
			if token != uint32(0xff+index) {
				t.Fatalf("award %d token = %#x, want %#x", index, token, uint32(0xff+index))
			}
			if player.SpellLvl[index] != 5 {
				t.Fatalf("award %d observed level %d before store", index, player.SpellLvl[index])
			}
			player.Prot4636++
			events = append(events, fmt.Sprintf("award:%d", index))
		},
	})

	if want := []string{"flags", "award:1", "award:2", "award:3", "award:4", "award:5"}; !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
	if player.SpellLvl[0] != 0x1000 {
		t.Fatalf("ability level 0 = %#x, want unchanged", player.SpellLvl[0])
	}
	for index := 1; index <= 5; index++ {
		if player.SpellLvl[index] != 5 {
			t.Fatalf("ability level %d = %d, want 5", index, player.SpellLvl[index])
		}
	}
	if player.SpellLvl[6] != 0x1006 {
		t.Fatalf("ability level 6 = %#x, want unchanged", player.SpellLvl[6])
	}
}

func TestWarriorAbilityAwardAllNative4EFE10NonWarriorSkipsDependencies(t *testing.T) {
	player := &Player{Prot4636: 0x12345678}
	player.Info().SetPlayerClass(playerlib.Wizard)
	player.SpellLvl[1] = 11
	calls := 0
	warriorAbilityAwardAllNative4EFE10(player, warriorAbilityAwardAllNativeDeps4EFE10{
		loadEngineFlags: func() uint8 {
			calls++
			return 0x10
		},
		awardProtection: func(uint32, int32, int32) {
			calls++
		},
	})
	if calls != 0 || player.SpellLvl[1] != 11 {
		t.Fatalf("calls/level = %d/%d, want 0/11", calls, player.SpellLvl[1])
	}
}

func preserveWarriorAbilityAwardAllFlags4EFE10(t *testing.T) {
	t.Helper()
	engine := noxflags.GetEngine()
	t.Cleanup(func() {
		noxflags.ResetEngine()
		noxflags.SetEngine(engine)
	})
}

func TestWarriorAbilityAwardAll4EFE10ServerReadsLiveFlags(t *testing.T) {
	preserveWarriorAbilityAwardAllFlags4EFE10(t)
	noxflags.ResetEngine()
	noxflags.SetEngine(noxflags.EngineAdmin)

	player := &Player{Prot4636: 0x12345678}
	player.Info().SetPlayerClass(playerlib.Warrior)
	awards := 0
	new(Server).WarriorAbilityAwardAll4EFE10(player, WarriorAbilityAwardAllRuntime4EFE10{
		AwardProtection: func(token uint32, index, level int32) {
			awards++
			if token != 0x12345678 || index != int32(awards) || level != 5 {
				t.Fatalf("award %d args = %#x/%d/%d", awards, token, index, level)
			}
		},
	})
	if awards != 5 {
		t.Fatalf("award calls = %d, want 5", awards)
	}
}

func TestWarriorAbilityAwardAllNative4EFE10NilPlayerFaultsBeforeDependencies(t *testing.T) {
	calls := 0
	deps := warriorAbilityAwardAllNativeDeps4EFE10{
		loadEngineFlags: func() uint8 {
			calls++
			return 0
		},
		awardProtection: func(uint32, int32, int32) {
			calls++
		},
	}
	deferred := false
	func() {
		defer func() {
			if recover() == nil {
				t.Fatal("nil player did not fault")
			}
			deferred = true
		}()
		warriorAbilityAwardAllNative4EFE10(nil, deps)
	}()
	if !deferred || calls != 0 {
		t.Fatalf("deferred/calls = %t/%d, want true/0", deferred, calls)
	}
}
