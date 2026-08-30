package server

import (
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
)

func TestPlayerManaRefreshNative4EECF0BindsFieldsAndProtectionResult(t *testing.T) {
	player := &Player{ProtUnitManaCur: 0x89abcdef}
	update := &PlayerUpdateData{
		ManaCur:  0x1234,
		ManaPrev: 0x1111,
		ManaMax:  0x8765,
		Player:   player,
	}
	unit := &Object{ObjClass: object.ClassPlayer | object.ClassMonster, UpdateData: unsafe.Pointer(update)}
	protectCalls := 0
	result := playerManaRefreshNative4EECF0(unit, playerManaRefreshNativeDeps4EECF0{
		protectMana: func(token uint32, maximum int16) uintptr {
			protectCalls++
			if token != 0x89abcdef || maximum != int16(-30875) {
				t.Fatalf("protection args = (%#x,%d), want (0x89abcdef,-30875)", token, maximum)
			}
			if update.ManaPrev != 0x1234 || update.ManaCur != 0x8765 || update.ManaMax != 0x8765 {
				t.Fatalf("mana at protection = previous %#04x current %#04x maximum %#04x", update.ManaPrev, update.ManaCur, update.ManaMax)
			}
			return 0xdeadbeef
		},
	})
	if result != 0xdeadbeef || protectCalls != 1 {
		t.Fatalf("result/protect calls = %#x/%d, want 0xdeadbeef/1", result, protectCalls)
	}
	runtime.KeepAlive(unit)
	runtime.KeepAlive(update)
	runtime.KeepAlive(player)
}

func TestPlayerManaRefreshNative4EECF0EntryGatesReturnNativeAddress(t *testing.T) {
	forbidden := playerManaRefreshNativeDeps4EECF0{
		protectMana: func(uint32, int16) uintptr {
			t.Fatal("protection update across entry gate")
			return 0
		},
	}
	if got := playerManaRefreshNative4EECF0(nil, forbidden); got != 0 {
		t.Fatalf("nil result = %#x, want 0", got)
	}

	for _, class := range []object.Class{object.ClassMonster, object.Class(0x00000400), object.Class(0x80000000)} {
		unit := &Object{ObjClass: class}
		want := uintptr(unsafe.Pointer(unit))
		if got := playerManaRefreshNative4EECF0(unit, forbidden); got != want {
			t.Fatalf("class %#x result = %#x, want unit address %#x", class, got, want)
		}
		runtime.KeepAlive(unit)
	}
}

func TestPlayerManaRefreshNative4EECF0PassesMaximumAsSignedWord(t *testing.T) {
	for _, maximum := range []uint16{0, 1, 0x7fff, 0x8000, 0xffff} {
		player := &Player{ProtUnitManaCur: 77}
		update := &PlayerUpdateData{ManaCur: 9, ManaPrev: 8, ManaMax: maximum, Player: player}
		unit := &Object{ObjClass: object.ClassPlayer, UpdateData: unsafe.Pointer(update)}
		var gotMaximum int16
		if got := playerManaRefreshNative4EECF0(unit, playerManaRefreshNativeDeps4EECF0{
			protectMana: func(token uint32, value int16) uintptr {
				if token != 77 {
					t.Fatalf("token = %d, want 77", token)
				}
				gotMaximum = value
				return 19
			},
		}); got != 19 {
			t.Fatalf("maximum %#04x result = %d, want 19", maximum, got)
		}
		if update.ManaPrev != 9 || update.ManaCur != maximum || gotMaximum != int16(maximum) {
			t.Fatalf("maximum %#04x produced previous/current/delta %#04x/%#04x/%d", maximum, update.ManaPrev, update.ManaCur, gotMaximum)
		}
	}
}

func TestPlayerManaRefreshNative4EECF0HasNoUpdateOrPlayerNilGuards(t *testing.T) {
	t.Run("nil update", func(t *testing.T) {
		unit := &Object{ObjClass: object.ClassPlayer}
		defer func() {
			if recover() == nil {
				t.Fatal("nil update did not fault")
			}
		}()
		playerManaRefreshNative4EECF0(unit, playerManaRefreshNativeDeps4EECF0{
			protectMana: func(uint32, int16) uintptr {
				t.Fatal("protection after nil update")
				return 0
			},
		})
	})

	t.Run("nil player after both stores", func(t *testing.T) {
		update := &PlayerUpdateData{ManaCur: 15, ManaPrev: 99, ManaMax: 23}
		unit := &Object{ObjClass: object.ClassPlayer, UpdateData: unsafe.Pointer(update)}
		defer func() {
			if recover() == nil {
				t.Fatal("nil player did not fault")
			}
			if update.ManaPrev != 15 || update.ManaCur != 23 {
				t.Fatalf("mana before fault = previous/current %d/%d, want 15/23", update.ManaPrev, update.ManaCur)
			}
		}()
		playerManaRefreshNative4EECF0(unit, playerManaRefreshNativeDeps4EECF0{
			protectMana: func(uint32, int16) uintptr {
				t.Fatal("protection after nil player")
				return 0
			},
		})
	})
}

func TestPlayerManaRefresh4EECF0ServerMethodBindsProtection(t *testing.T) {
	player := &Player{ProtUnitManaCur: 123}
	update := &PlayerUpdateData{ManaCur: 10, ManaPrev: 2, ManaMax: 20, Player: player}
	unit := &Object{ObjClass: object.ClassPlayer, UpdateData: unsafe.Pointer(update)}
	if got := new(Server).PlayerManaRefresh4EECF0(unit, func(token uint32, maximum int16) uintptr {
		if token != 123 || maximum != 20 {
			t.Fatalf("protection args = %d/%d, want 123/20", token, maximum)
		}
		return 0x1234
	}); got != 0x1234 {
		t.Fatalf("result = %#x, want 0x1234", got)
	}
	if update.ManaPrev != 10 || update.ManaCur != 20 {
		t.Fatalf("mana = previous/current %d/%d, want 10/20", update.ManaPrev, update.ManaCur)
	}
}

func TestPlayerManaRefresh4EECF0NativeLayouts(t *testing.T) {
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
		{"native result", unsafe.Sizeof(playerManaRefreshNative4EECF0(nil, playerManaRefreshNativeDeps4EECF0{})), unsafe.Sizeof(uintptr(0))},
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Errorf("%s on %s/%s = %d, want %d", check.name, runtime.GOOS, runtime.GOARCH, check.got, check.want)
		}
	}
}
