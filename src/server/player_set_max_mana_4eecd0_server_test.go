package server

import (
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
)

func TestPlayerSetMaxMana4EECD0NativeLayouts(t *testing.T) {
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
		{"PlayerUpdateData.ManaPrev", unsafe.Offsetof(PlayerUpdateData{}.ManaPrev), 6},
		{"PlayerUpdateData.ManaMax", unsafe.Offsetof(PlayerUpdateData{}.ManaMax), 8},
		{"maximum word width", unsafe.Sizeof(uint16(0)), 2},
		{"return register width", unsafe.Sizeof(uintptr(0)), unsafe.Sizeof(unsafe.Pointer(nil))},
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Errorf("%s on %s/%s = %d, want %d", check.name, runtime.GOOS, runtime.GOARCH, check.got, check.want)
		}
	}
}

func TestPlayerSetMaxMana4EECD0NativeEntryGatesReturnUnit(t *testing.T) {
	if got := PlayerSetMaxMana4EECD0(nil, 0xabcd); got != 0 {
		t.Fatalf("nil object result = %#x, want zero", got)
	}

	tests := []struct {
		name  string
		class object.Class
	}{
		{name: "other", class: object.Class(0x80000000)},
		{name: "upper byte player bit", class: object.Class(0x00000400)},
		{name: "monster", class: object.ClassMonster},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			unit := &Object{ObjClass: test.class}
			want := uintptr(unsafe.Pointer(unit))
			if got := PlayerSetMaxMana4EECD0(unit, 0xabcd); got != want {
				t.Fatalf("result = %#x, want unit identity %#x", got, want)
			}
		})
	}
}

func TestPlayerSetMaxMana4EECD0NativeStoresAndReturnsUpdate(t *testing.T) {
	player := &Player{}
	update := &PlayerUpdateData{
		ManaCur:  0x1234,
		ManaPrev: 0x5678,
		ManaMax:  0x9abc,
		Player:   player,
	}
	unit := &Object{
		ObjClass:   object.Class(0xa5a50000) | object.ClassPlayer | object.ClassMonster,
		UpdateData: unsafe.Pointer(update),
	}
	want := uintptr(unsafe.Pointer(update))
	if got := PlayerSetMaxMana4EECD0(unit, 0xfedc); got != want {
		t.Fatalf("result = %#x, want UpdateData identity %#x", got, want)
	}
	if update.ManaMax != 0xfedc {
		t.Fatalf("maximum = %#04x, want 0xfedc", update.ManaMax)
	}
	if update.ManaCur != 0x1234 || update.ManaPrev != 0x5678 || update.Player != player {
		t.Fatalf("adjacent PlayerUpdateData changed: %+v", *update)
	}
}

func TestPlayerSetMaxMana4EECD0NativeHasNoUpdateNilGuard(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("nil Player UpdateData did not panic")
		}
	}()
	PlayerSetMaxMana4EECD0(&Object{ObjClass: object.ClassPlayer}, 0xabcd)
}
