package server

import (
	"runtime"
	"testing"
	"unsafe"
)

func TestUnitGetMaxHP4EE7A0NativeLayouts(t *testing.T) {
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
		{"HealthData.Max", unsafe.Offsetof(HealthData{}.Max), 4},
		{"HP width", unsafe.Sizeof(uint16(0)), 2},
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Errorf("%s on %s/%s = %d, want %d", check.name, runtime.GOOS, runtime.GOARCH, check.got, check.want)
		}
	}
}

func TestUnitGetMaxHP4EE7A0NativeEntryGates(t *testing.T) {
	if got := UnitGetMaxHP4EE7A0(nil); got != 0 {
		t.Fatalf("nil object result = %#04x, want zero", got)
	}
	if got := UnitGetMaxHP4EE7A0(&Object{}); got != 0 {
		t.Fatalf("nil HealthData result = %#04x, want zero", got)
	}
}

func TestUnitGetMaxHP4EE7A0NativeBindsMaximumWord(t *testing.T) {
	unit := &Object{HealthData: &HealthData{Cur: 0xfedc, Field2: 0x1234, Max: 0xabcd}}
	if got := UnitGetMaxHP4EE7A0(unit); got != 0xabcd {
		t.Fatalf("result = %#04x, want 0xabcd", got)
	}
}
