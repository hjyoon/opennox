package server

import (
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
)

func TestPlayerGetMaxMana4EECB0NativeLayouts(t *testing.T) {
	wantObjectSize := uintptr(780)
	wantClass := uintptr(8)
	wantUpdate := uintptr(748)
	wantPlayerUpdateSize := uintptr(556)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantObjectSize = 928
		wantClass = 12
		wantUpdate = 872
		wantPlayerUpdateSize = 656
	}
	checks := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"Object size", unsafe.Sizeof(Object{}), wantObjectSize},
		{"Object.ObjClass", unsafe.Offsetof(Object{}.ObjClass), wantClass},
		{"Object.UpdateData", unsafe.Offsetof(Object{}.UpdateData), wantUpdate},
		{"PlayerUpdateData size", unsafe.Sizeof(PlayerUpdateData{}), wantPlayerUpdateSize},
		{"PlayerUpdateData.ManaMax", unsafe.Offsetof(PlayerUpdateData{}.ManaMax), 8},
		{"mana return width", unsafe.Sizeof(uint16(0)), 2},
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Errorf("%s on %s/%s = %d, want %d", check.name, runtime.GOOS, runtime.GOARCH, check.got, check.want)
		}
	}
}

func TestPlayerGetMaxMana4EECB0NativeClassGates(t *testing.T) {
	if got := PlayerGetMaxMana4EECB0(nil); got != 0 {
		t.Fatalf("nil object result = %#04x, want zero", got)
	}
	if got := PlayerGetMaxMana4EECB0(&Object{ObjClass: object.ClassMonster}); got != 0 {
		t.Fatalf("Monster result = %#04x, want zero", got)
	}
	if got := PlayerGetMaxMana4EECB0(&Object{ObjClass: object.Class(0x00000400)}); got != 0 {
		t.Fatalf("upper-byte Player bit result = %#04x, want zero", got)
	}

	update := &PlayerUpdateData{ManaCur: 0x1234, ManaMax: 0xabcd}
	player := &Object{ObjClass: object.ClassPlayer, UpdateData: unsafe.Pointer(update)}
	if got := PlayerGetMaxMana4EECB0(player); got != 0xabcd {
		t.Fatalf("Player result = %#04x, want 0xabcd", got)
	}
	player.ObjClass |= object.ClassMonster
	if got := PlayerGetMaxMana4EECB0(player); got != 0xabcd {
		t.Fatalf("Player|Monster result = %#04x, want Player maximum 0xabcd", got)
	}
}

func TestPlayerGetMaxMana4EECB0NativeHasNoUpdateNilGuard(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("nil Player UpdateData did not panic")
		}
	}()
	PlayerGetMaxMana4EECB0(&Object{ObjClass: object.ClassPlayer})
}
