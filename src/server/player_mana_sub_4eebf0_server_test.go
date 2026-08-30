package server

import (
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	noxflags "github.com/opennox/opennox/v1/common/flags"
)

func TestPlayerManaSubNative4EEBF0BindsCachedUpdateAndProtectionResult(t *testing.T) {
	player1 := &Player{ProtUnitManaCur: 0x11111111}
	player2 := &Player{ProtUnitManaCur: 0x89abcdef}
	update1 := &PlayerUpdateData{ManaCur: 7, ManaPrev: 3, Player: player1}
	update2 := &PlayerUpdateData{ManaCur: 100, ManaPrev: 9, Player: player2}
	unit := &Object{ObjClass: object.ClassPlayer, UpdateData: unsafe.Pointer(update1)}
	protectCalls := 0
	result := playerManaSubNative4EEBF0(unit, 10, playerManaSubNativeDeps4EEBF0{
		loadEngineGodMode: func() bool {
			unit.UpdateData = unsafe.Pointer(update2)
			return false
		},
		protectMana: func(token uint32, delta int16) uintptr {
			protectCalls++
			if token != player2.ProtUnitManaCur || delta != -10 {
				t.Fatalf("protection args = (%#x,%d), want (%#x,-10)", token, delta, player2.ProtUnitManaCur)
			}
			if update2.ManaPrev != 100 || update2.ManaCur != 90 {
				t.Fatalf("cached update previous/current = %d/%d, want 100/90", update2.ManaPrev, update2.ManaCur)
			}
			return 0xdeadbeef
		},
	})
	if result != 0xdeadbeef || protectCalls != 1 {
		t.Fatalf("result/protect calls = %#x/%d, want 0xdeadbeef/1", result, protectCalls)
	}
	if update1.ManaCur != 7 || update1.ManaPrev != 3 {
		t.Fatalf("stale update changed: previous/current = %d/%d", update1.ManaPrev, update1.ManaCur)
	}
	runtime.KeepAlive(unit)
	runtime.KeepAlive(update1)
	runtime.KeepAlive(update2)
}

func TestPlayerManaSubNative4EEBF0EntryAndGodModeReturnNativeAddresses(t *testing.T) {
	forbidden := playerManaSubNativeDeps4EEBF0{
		loadEngineGodMode: func() bool {
			t.Fatal("GodMode read across entry gate")
			return false
		},
		protectMana: func(uint32, int16) uintptr {
			t.Fatal("protection update across entry gate")
			return 0
		},
	}
	if got := playerManaSubNative4EEBF0(nil, 99, forbidden); got != 0 {
		t.Fatalf("nil result = %#x, want 0", got)
	}

	unit := &Object{ObjClass: object.ClassMonster}
	wantUnit := uintptr(unsafe.Pointer(unit))
	if got := playerManaSubNative4EEBF0(unit, 99, forbidden); got != wantUnit {
		t.Fatalf("non-Player result = %#x, want unit address %#x", got, wantUnit)
	}

	update := &PlayerUpdateData{ManaCur: 77, ManaPrev: 3}
	unit.ObjClass = object.ClassPlayer
	unit.UpdateData = unsafe.Pointer(update)
	got := playerManaSubNative4EEBF0(unit, 99, playerManaSubNativeDeps4EEBF0{
		loadEngineGodMode: func() bool { return true },
		protectMana: func(uint32, int16) uintptr {
			t.Fatal("protection update in GodMode")
			return 0
		},
	})
	wantUpdate := uintptr(unsafe.Pointer(update))
	if got != wantUpdate {
		t.Fatalf("GodMode result = %#x, want update address %#x", got, wantUpdate)
	}
	if update.ManaCur != 77 || update.ManaPrev != 3 {
		t.Fatalf("GodMode changed mana: previous/current = %d/%d", update.ManaPrev, update.ManaCur)
	}
	runtime.KeepAlive(unit)
	runtime.KeepAlive(update)
}

func TestPlayerManaSubNative4EEBF0WholeSignedAmount(t *testing.T) {
	player := &Player{ProtUnitManaCur: 0xfedcba98}
	update := &PlayerUpdateData{ManaCur: 100, ManaPrev: 7, Player: player}
	unit := &Object{ObjClass: object.ClassPlayer, UpdateData: unsafe.Pointer(update)}
	var gotDelta int16
	result := playerManaSubNative4EEBF0(unit, -1, playerManaSubNativeDeps4EEBF0{
		loadEngineGodMode: func() bool { return false },
		protectMana: func(token uint32, delta int16) uintptr {
			if token != player.ProtUnitManaCur {
				t.Fatalf("token = %#x, want %#x", token, player.ProtUnitManaCur)
			}
			gotDelta = delta
			return 17
		},
	})
	if result != 17 || update.ManaPrev != 100 || update.ManaCur != 101 || gotDelta != 1 {
		t.Fatalf("result/previous/current/delta = %d/%d/%d/%d, want 17/100/101/1", result, update.ManaPrev, update.ManaCur, gotDelta)
	}
}

func TestPlayerManaSubNative4EEBF0HasNoUpdateOrPlayerNilGuards(t *testing.T) {
	t.Run("nil update", func(t *testing.T) {
		unit := &Object{ObjClass: object.ClassPlayer}
		defer func() {
			if recover() == nil {
				t.Fatal("nil update did not fault")
			}
		}()
		playerManaSubNative4EEBF0(unit, 1, playerManaSubNativeDeps4EEBF0{
			loadEngineGodMode: func() bool { return false },
			protectMana: func(uint32, int16) uintptr {
				t.Fatal("protection after nil update")
				return 0
			},
		})
	})

	t.Run("nil player after stores", func(t *testing.T) {
		update := &PlayerUpdateData{ManaCur: 15, ManaPrev: 99}
		unit := &Object{ObjClass: object.ClassPlayer, UpdateData: unsafe.Pointer(update)}
		defer func() {
			if recover() == nil {
				t.Fatal("nil player did not fault")
			}
			if update.ManaPrev != 15 || update.ManaCur != 5 {
				t.Fatalf("mana before fault = %d/%d, want 15/5", update.ManaPrev, update.ManaCur)
			}
		}()
		playerManaSubNative4EEBF0(unit, 10, playerManaSubNativeDeps4EEBF0{
			loadEngineGodMode: func() bool { return false },
			protectMana: func(uint32, int16) uintptr {
				t.Fatal("protection after nil player")
				return 0
			},
		})
	})
}

func TestPlayerManaSub4EEBF0ServerMethodUsesEngineGodMode(t *testing.T) {
	oldEngine := noxflags.GetEngine()
	noxflags.ResetEngine()
	defer func() {
		noxflags.ResetEngine()
		noxflags.SetEngine(oldEngine)
	}()

	player := &Player{ProtUnitManaCur: 77}
	update := &PlayerUpdateData{ManaCur: 10, ManaPrev: 2, Player: player}
	unit := &Object{ObjClass: object.ClassPlayer, UpdateData: unsafe.Pointer(update)}
	var gotToken uint32
	var gotDelta int16
	if got := new(Server).PlayerManaSub4EEBF0(unit, 3, func(token uint32, delta int16) uintptr {
		gotToken, gotDelta = token, delta
		return 0x1234
	}); got != 0x1234 {
		t.Fatalf("active result = %#x, want 0x1234", got)
	}
	if update.ManaPrev != 10 || update.ManaCur != 7 || gotToken != 77 || gotDelta != -3 {
		t.Fatalf("active previous/current/token/delta = %d/%d/%#x/%d", update.ManaPrev, update.ManaCur, gotToken, gotDelta)
	}

	update.ManaCur, update.ManaPrev = 20, 4
	noxflags.SetEngine(noxflags.EngineGodMode)
	want := uintptr(unsafe.Pointer(update))
	if got := new(Server).PlayerManaSub4EEBF0(unit, 9, func(uint32, int16) uintptr {
		t.Fatal("protection in GodMode")
		return 0
	}); got != want {
		t.Fatalf("GodMode result = %#x, want update address %#x", got, want)
	}
	if update.ManaCur != 20 || update.ManaPrev != 4 {
		t.Fatalf("GodMode changed mana: previous/current = %d/%d", update.ManaPrev, update.ManaCur)
	}
	runtime.KeepAlive(update)
}

func TestPlayerManaSub4EEBF0NativeLayouts(t *testing.T) {
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
