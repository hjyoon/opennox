package server

import (
	"runtime"
	"testing"
	"unsafe"
)

func TestUnitGetHP4EE780NativeLayouts(t *testing.T) {
	wantObjectSize := uintptr(780)
	wantHealth := uintptr(556)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantObjectSize = 928
		wantHealth = 616
	}
	checks := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"Object size", unsafe.Sizeof(Object{}), wantObjectSize},
		{"Object.HealthData", unsafe.Offsetof(Object{}.HealthData), wantHealth},
		{"HealthData size", unsafe.Sizeof(HealthData{}), 20},
		{"HealthData.Cur", unsafe.Offsetof(HealthData{}.Cur), 0},
		{"HP width", unsafe.Sizeof(uint16(0)), 2},
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Errorf("%s on %s/%s = %d, want %d", check.name, runtime.GOOS, runtime.GOARCH, check.got, check.want)
		}
	}
}

func TestUnitGetHP4EE780NativeEntryGates(t *testing.T) {
	if got := UnitGetHP4EE780(nil); got != 0 {
		t.Fatalf("nil object result = %#04x, want zero", got)
	}
	if got := UnitGetHP4EE780(&Object{}); got != 0 {
		t.Fatalf("nil HealthData result = %#04x, want zero", got)
	}
}

func TestUnitGetHP4EE780NativeBindsCurrentWord(t *testing.T) {
	unit := &Object{HealthData: &HealthData{Cur: 0xfedc, Field2: 0x1234, Max: 0xabcd}}
	if got := UnitGetHP4EE780(unit); got != 0xfedc {
		t.Fatalf("result = %#04x, want 0xfedc", got)
	}
}
