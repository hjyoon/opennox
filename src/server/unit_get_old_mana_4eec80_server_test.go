package server

import (
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
)

func TestUnitGetOldMana4EEC80NativeLayouts(t *testing.T) {
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
		{"PlayerUpdateData.ManaCur", unsafe.Offsetof(PlayerUpdateData{}.ManaCur), 4},
		{"mana return width", unsafe.Sizeof(uint16(0)), 2},
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Errorf("%s on %s/%s = %d, want %d", check.name, runtime.GOOS, runtime.GOARCH, check.got, check.want)
		}
	}
}

func TestUnitGetOldMana4EEC80NativeClassGates(t *testing.T) {
	if got := UnitGetOldMana4EEC80(nil); got != 0 {
		t.Fatalf("nil object result = %#04x, want zero", got)
	}
	if got := UnitGetOldMana4EEC80(&Object{ObjClass: object.ClassImmobile}); got != 0 {
		t.Fatalf("non-unit result = %#04x, want zero", got)
	}
	if got := UnitGetOldMana4EEC80(&Object{ObjClass: object.ClassMonster}); got != 1000 {
		t.Fatalf("Monster result = %d, want 1000", got)
	}

	update := &PlayerUpdateData{ManaCur: 0xabcd}
	player := &Object{ObjClass: object.ClassPlayer, UpdateData: unsafe.Pointer(update)}
	if got := UnitGetOldMana4EEC80(player); got != 0xabcd {
		t.Fatalf("Player result = %#04x, want 0xabcd", got)
	}
	player.ObjClass |= object.ClassMonster
	if got := UnitGetOldMana4EEC80(player); got != 0xabcd {
		t.Fatalf("Player|Monster result = %#04x, want Player mana 0xabcd", got)
	}
}

func TestUnitGetOldMana4EEC80NativeHasNoUpdateNilGuard(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("nil Player UpdateData did not panic")
		}
	}()
	UnitGetOldMana4EEC80(&Object{ObjClass: object.ClassPlayer})
}
