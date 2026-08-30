package server

import (
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
)

func TestPlayerManaAddNative4EEB80BindsLiveProtectionState(t *testing.T) {
	player1 := &Player{ProtUnitManaCur: 0x12345678}
	player2 := &Player{ProtUnitManaCur: 0x89abcdef}
	update := &PlayerUpdateData{
		ManaCur:  100,
		ManaPrev: 7,
		ManaMax:  120,
		Player:   player1,
	}
	unit := &Object{ObjClass: object.ClassPlayer, UpdateData: unsafe.Pointer(update)}
	protectCalls, repairCalls := 0, 0
	result := playerManaAddNative4EEB80(unit, 30, playerManaAddNativeDeps4EEB80{
		protectMana: func(token uint32, delta int16) {
			protectCalls++
			if token != player1.ProtUnitManaCur || delta != 30 {
				t.Fatalf("protect args = (%#x,%d)", token, delta)
			}
			if update.ManaPrev != 100 || update.ManaCur != 120 {
				t.Fatalf("state before protect = previous/current %d/%d, want 100/120", update.ManaPrev, update.ManaCur)
			}
			update.ManaCur = 91
			update.ManaMax = 90
			update.Player = player2
		},
		protectPlayerHPMana: func(token uint32, maximum uint16) uint16 {
			repairCalls++
			if token != player2.ProtUnitManaCur || maximum != 90 {
				t.Fatalf("repair args = (%#x,%d)", token, maximum)
			}
			return 0xbeef
		},
	})
	if result != 0xbeef || protectCalls != 1 || repairCalls != 1 {
		t.Fatalf("result/protect/repair = %#x/%d/%d, want 0xbeef/1/1", result, protectCalls, repairCalls)
	}
}

func TestPlayerManaAddNative4EEB80EntryGatesPreserveLowPointerWord(t *testing.T) {
	forbidden := playerManaAddNativeDeps4EEB80{
		protectMana: func(uint32, int16) {
			t.Fatal("protection update across entry gate")
		},
		protectPlayerHPMana: func(uint32, uint16) uint16 {
			t.Fatal("repair across entry gate")
			return 0
		},
	}
	if got := playerManaAddNative4EEB80(nil, 99, forbidden); got != 0 {
		t.Fatalf("nil result = %#x, want 0", got)
	}
	unit := &Object{ObjClass: object.ClassMonster}
	want := uint16(uintptr(unsafe.Pointer(unit)))
	if got := playerManaAddNative4EEB80(unit, 99, forbidden); got != want {
		t.Fatalf("non-Player result = %#x, want pointer low word %#x", got, want)
	}
}

func TestPlayerManaAddNative4EEB80WholeAmountSlotAndNoRepair(t *testing.T) {
	player := &Player{ProtUnitManaCur: 0xfedcba98}
	update := &PlayerUpdateData{ManaCur: 7, ManaPrev: 3, ManaMax: 100, Player: player}
	unit := &Object{ObjClass: object.ClassPlayer, UpdateData: unsafe.Pointer(update)}
	var gotToken uint32
	var gotDelta int16
	result := playerManaAddNative4EEB80(unit, 0x10001, playerManaAddNativeDeps4EEB80{
		protectMana: func(token uint32, delta int16) {
			gotToken, gotDelta = token, delta
		},
		protectPlayerHPMana: func(uint32, uint16) uint16 {
			t.Fatal("unexpected repair")
			return 0
		},
	})
	if result != 100 || update.ManaPrev != 7 || update.ManaCur != 8 {
		t.Fatalf("result/previous/current = %d/%d/%d, want 100/7/8", result, update.ManaPrev, update.ManaCur)
	}
	if gotToken != player.ProtUnitManaCur || gotDelta != 1 {
		t.Fatalf("protect args = (%#x,%d), want (%#x,1)", gotToken, gotDelta, player.ProtUnitManaCur)
	}
}

func TestPlayerManaAddNative4EEB80HasNoUpdateOrPlayerNilGuards(t *testing.T) {
	t.Run("nil update", func(t *testing.T) {
		unit := &Object{ObjClass: object.ClassPlayer}
		defer func() {
			if recover() == nil {
				t.Fatal("nil update did not fault")
			}
		}()
		playerManaAddNative4EEB80(unit, 1, playerManaAddNativeDeps4EEB80{
			protectMana: func(uint32, int16) { t.Fatal("protect after nil update") },
			protectPlayerHPMana: func(uint32, uint16) uint16 {
				t.Fatal("repair after nil update")
				return 0
			},
		})
	})

	t.Run("nil player", func(t *testing.T) {
		update := &PlayerUpdateData{ManaCur: 1, ManaMax: 10}
		unit := &Object{ObjClass: object.ClassPlayer, UpdateData: unsafe.Pointer(update)}
		defer func() {
			if recover() == nil {
				t.Fatal("nil player did not fault")
			}
		}()
		playerManaAddNative4EEB80(unit, 1, playerManaAddNativeDeps4EEB80{
			protectMana: func(uint32, int16) { t.Fatal("protect after nil player") },
			protectPlayerHPMana: func(uint32, uint16) uint16 {
				t.Fatal("repair after nil player")
				return 0
			},
		})
	})
}

func TestPlayerManaAdd4EEB80ServerMethodUsesNativeFields(t *testing.T) {
	player := &Player{ProtUnitManaCur: 77}
	update := &PlayerUpdateData{ManaCur: 10, ManaMax: 20, Player: player}
	unit := &Object{ObjClass: object.ClassPlayer, UpdateData: unsafe.Pointer(update)}
	var protected bool
	result := new(Server).PlayerManaAdd4EEB80(
		unit,
		3,
		func(token uint32, delta int16) {
			protected = token == 77 && delta == 3
		},
		func(uint32, uint16) uint16 {
			t.Fatal("unexpected repair")
			return 0
		},
	)
	if result != 20 || update.ManaPrev != 10 || update.ManaCur != 13 || !protected {
		t.Fatalf("result/previous/current/protected = %d/%d/%d/%v", result, update.ManaPrev, update.ManaCur, protected)
	}
}

func TestPlayerManaAdd4EEB80NativeLayouts(t *testing.T) {
	wantObjectSize := uintptr(780)
	wantClass := uintptr(8)
	wantUpdate := uintptr(748)
	wantUpdateSize := uintptr(556)
	wantPlayer := uintptr(276)
	wantPlayerSize := uintptr(4828)
	wantProtection := uintptr(4596)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantObjectSize = 928
		wantClass = 12
		wantUpdate = 872
		wantUpdateSize = 656
		wantPlayer = 336
		wantPlayerSize = 6160
		wantProtection = 5900
	}
	checks := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"Object size", unsafe.Sizeof(Object{}), wantObjectSize},
		{"Object.ObjClass", unsafe.Offsetof(Object{}.ObjClass), wantClass},
		{"Object.UpdateData", unsafe.Offsetof(Object{}.UpdateData), wantUpdate},
		{"PlayerUpdateData size", unsafe.Sizeof(PlayerUpdateData{}), wantUpdateSize},
		{"PlayerUpdateData.ManaCur", unsafe.Offsetof(PlayerUpdateData{}.ManaCur), 4},
		{"PlayerUpdateData.ManaPrev", unsafe.Offsetof(PlayerUpdateData{}.ManaPrev), 6},
		{"PlayerUpdateData.ManaMax", unsafe.Offsetof(PlayerUpdateData{}.ManaMax), 8},
		{"PlayerUpdateData.Player", unsafe.Offsetof(PlayerUpdateData{}.Player), wantPlayer},
		{"Player size", unsafe.Sizeof(Player{}), wantPlayerSize},
		{"Player.ProtUnitManaCur", unsafe.Offsetof(Player{}.ProtUnitManaCur), wantProtection},
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Errorf("%s on %s/%s = %d, want %d", check.name, runtime.GOOS, runtime.GOARCH, check.got, check.want)
		}
	}
}
