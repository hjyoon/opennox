package server

import (
	"runtime"
	"testing"
	"unsafe"
)

func TestUnitSetMaxHP4EE7C0NativeLayouts(t *testing.T) {
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
		{"HealthData.Field2", unsafe.Offsetof(HealthData{}.Field2), 2},
		{"HealthData.Max", unsafe.Offsetof(HealthData{}.Max), 4},
		{"HP width", unsafe.Sizeof(uint16(0)), 2},
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Errorf("%s on %s/%s = %d, want %d", check.name, runtime.GOOS, runtime.GOARCH, check.got, check.want)
		}
	}
}

func TestUnitSetMaxHP4EE7C0NativeEntryGates(t *testing.T) {
	if got := UnitSetMaxHP4EE7C0(nil, 0xabcd); got != nil {
		t.Fatalf("nil object result = %p, want nil", got)
	}
	if got := UnitSetMaxHP4EE7C0(&Object{}, 0xabcd); got != nil {
		t.Fatalf("nil HealthData result = %p, want nil", got)
	}
}

func TestUnitSetMaxHP4EE7C0NativeStoresAndReturnsHealth(t *testing.T) {
	health := &HealthData{Cur: 0x1234, Field2: 0x5678, Max: 0x9abc, Field16: 0xdef01234}
	unit := &Object{HealthData: health}
	if got := UnitSetMaxHP4EE7C0(unit, 0xfedc); got != health {
		t.Fatalf("result = %p, want HealthData %p", got, health)
	}
	if health.Max != 0xfedc {
		t.Fatalf("maximum = %#04x, want 0xfedc", health.Max)
	}
	if health.Cur != 0x1234 || health.Field2 != 0x5678 || health.Field16 != 0xdef01234 {
		t.Fatalf("adjacent HealthData changed: %+v", *health)
	}
}
