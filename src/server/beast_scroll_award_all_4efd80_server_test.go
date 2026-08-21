package server

import (
	"fmt"
	"reflect"
	"runtime"
	"testing"
	"unsafe"

	noxflags "github.com/opennox/opennox/v1/common/flags"
)

func TestBeastScrollAwardAll4EFD80NativeLayout(t *testing.T) {
	wantPlayerSize := uintptr(4828)
	wantScrollLevels := uintptr(4244)
	wantProtection := uintptr(4640)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantPlayerSize = 6160
		wantScrollLevels = 5540
		wantProtection = 5944
	}

	checks := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"Player size", unsafe.Sizeof(Player{}), wantPlayerSize},
		{"Player.BeastScrollLvl", unsafe.Offsetof(Player{}.BeastScrollLvl), wantScrollLevels},
		{"Player.Prot4640", unsafe.Offsetof(Player{}.Prot4640), wantProtection},
		{"beast-scroll level width", unsafe.Sizeof(Player{}.BeastScrollLvl[0]), 4},
		{"beast-scroll level count", uintptr(len(Player{}.BeastScrollLvl)), 41},
		{"protection width", unsafe.Sizeof(Player{}.Prot4640), 4},
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Errorf("%s on %s/%s = %d, want %d", check.name, runtime.GOOS, runtime.GOARCH, check.got, check.want)
		}
	}
}

func TestBeastScrollAwardAllNative4EFD80BindsLivePlayerFields(t *testing.T) {
	player := &Player{Prot4640: 0x100}
	for index := range player.BeastScrollLvl {
		player.BeastScrollLvl[index] = uint32(0x1000 + index)
	}

	var events []string
	awardIndex := int32(0)
	beastScrollAwardAllNative4EFD80(player, beastScrollAwardAllNativeDeps4EFD80{
		resetProtection: func(token uint32, value int32) {
			events = append(events, fmt.Sprintf("reset:%x:%d", token, value))
			if token != 0x100 || value != 0 {
				t.Fatalf("reset args = %#x/%d, want 0x100/0", token, value)
			}
			player.Prot4640 = 0x200
		},
		loadEngineFlags: func() uint8 {
			events = append(events, "flags")
			if player.Prot4640 != 0x200 {
				t.Fatalf("engine flags loaded before reset mutation: token=%#x", player.Prot4640)
			}
			return 0x10
		},
		awardProtection: func(token uint32, index, level int32) {
			awardIndex++
			if index != awardIndex || level != 1 {
				t.Fatalf("award index/level = %d/%d, want %d/1", index, level, awardIndex)
			}
			if token != uint32(0x1ff+index) {
				t.Fatalf("award %d token = %#x, want %#x", index, token, uint32(0x1ff+index))
			}
			if player.BeastScrollLvl[index] != 1 {
				t.Fatalf("award %d observed level %d before store", index, player.BeastScrollLvl[index])
			}
			player.Prot4640++
		},
		applyProtection: func(token uint32, levels *[41]uint32, count int32) {
			events = append(events, fmt.Sprintf("apply:%x:%d", token, count))
			if token != 0x228 || levels != &player.BeastScrollLvl || count != 41 {
				t.Fatalf("apply args = %#x/%p/%d, want 0x228/%p/41", token, levels, count, &player.BeastScrollLvl)
			}
		},
	})

	wantEvents := []string{"reset:100:0", "flags", "apply:228:41"}
	if !reflect.DeepEqual(events, wantEvents) {
		t.Fatalf("events = %v, want %v", events, wantEvents)
	}
	if player.BeastScrollLvl[0] != 0x1000 {
		t.Fatalf("beast-scroll level 0 = %#x, want unchanged", player.BeastScrollLvl[0])
	}
	for index := 1; index < len(player.BeastScrollLvl); index++ {
		if player.BeastScrollLvl[index] != 1 {
			t.Fatalf("beast-scroll level %d = %d, want 1", index, player.BeastScrollLvl[index])
		}
	}
}

func preserveBeastScrollAwardAllFlags4EFD80(t *testing.T) {
	t.Helper()
	engine := noxflags.GetEngine()
	t.Cleanup(func() {
		noxflags.ResetEngine()
		noxflags.SetEngine(engine)
	})
}

func TestBeastScrollAwardAll4EFD80ServerReadsFlagsAfterReset(t *testing.T) {
	preserveBeastScrollAwardAllFlags4EFD80(t)
	noxflags.ResetEngine()
	noxflags.SetEngine(noxflags.EngineAdmin)

	player := &Player{Prot4640: 0x12345678}
	awards := 0
	applies := 0
	new(Server).BeastScrollAwardAll4EFD80(player, BeastScrollAwardAllRuntime4EFD80{
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
		ApplyProtection: func(token uint32, levels *[41]uint32, count int32) {
			applies++
			if token != 0x12345678 || levels != &player.BeastScrollLvl || count != 41 {
				t.Fatalf("apply args = %#x/%p/%d", token, levels, count)
			}
		},
	})
	if awards != 40 || applies != 1 {
		t.Fatalf("award/apply calls = %d/%d, want 40/1", awards, applies)
	}
}

func TestBeastScrollAwardAllNative4EFD80NilPlayerFaultsBeforeServices(t *testing.T) {
	calls := 0
	deps := beastScrollAwardAllNativeDeps4EFD80{
		resetProtection: func(uint32, int32) { calls++ },
		loadEngineFlags: func() uint8 { calls++; return 0 },
		awardProtection: func(uint32, int32, int32) { calls++ },
		applyProtection: func(uint32, *[41]uint32, int32) { calls++ },
	}
	deferred := false
	func() {
		defer func() {
			if recover() == nil {
				t.Fatal("nil player did not fault")
			}
			deferred = true
		}()
		beastScrollAwardAllNative4EFD80(nil, deps)
	}()
	if !deferred || calls != 0 {
		t.Fatalf("deferred/calls = %t/%d, want true/0", deferred, calls)
	}
}
