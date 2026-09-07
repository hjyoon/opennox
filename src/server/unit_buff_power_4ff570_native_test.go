package server

import (
	"math"
	"reflect"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

func TestUnitBuffPower4FF570NativeLayout(t *testing.T) {
	wantOffset := uintptr(408)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantOffset = 412
	}
	checks := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"BuffsPower offset", unsafe.Offsetof(Object{}.BuffsPower), wantOffset},
		{"BuffsPower width", unsafe.Sizeof(Object{}.BuffsPower), 32},
		{"BuffsPower element width", unsafe.Sizeof(Object{}.BuffsPower[0]), 1},
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Errorf("Object %s = %d, want %d", check.name, check.got, check.want)
		}
	}
}

func TestUnitBuffPowerNative4FF570AccessorOrderAndWidths(t *testing.T) {
	unit, freeUnit := alloc.New(Object{})
	t.Cleanup(freeUnit)
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(unit)) <= math.MaxUint32 {
		t.Fatalf("unit pointer = %p, want native address above 4 GiB", unit)
	}

	var events []string
	deps := unitBuffPowerNativeDeps4FF570{
		loadBuffArg: func(buff int32) int32 {
			events = append(events, "buff")
			if buff != math.MinInt32 {
				t.Fatalf("buff = %d, want exact signed dword minimum", buff)
			}
			return buff
		},
		loadUnitArg: func(got *Object) *Object {
			events = append(events, "unit")
			if got != unit {
				t.Fatalf("unit = %p, want full native identity %p", got, unit)
			}
			return got
		},
		loadPower: func(got *Object, buff int32) uint8 {
			events = append(events, "power")
			if got != unit || buff != math.MinInt32 {
				t.Fatalf("power args = (%p, %d), want (%p, %d)", got, buff, unit, int32(math.MinInt32))
			}
			return math.MaxUint8
		},
	}
	if got := unitBuffPowerNative4FF570(unit, math.MinInt32, deps); got != math.MaxUint8 {
		t.Fatalf("result = %#02x, want exact byte 0xff", got)
	}
	if want := []string{"buff", "unit", "power"}; !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %q, want exact accessor order %q", events, want)
	}
	runtime.KeepAlive(unit)
}

func TestUnitBuffPower4FF570ServerBinding(t *testing.T) {
	unit, freeUnit := alloc.New(Object{})
	t.Cleanup(freeUnit)
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(unit)) <= math.MaxUint32 {
		t.Fatalf("unit pointer = %p, want native address above 4 GiB", unit)
	}

	unit.BuffsPower[0] = 0x80
	unit.BuffsPower[31] = math.MaxUint8
	if got := unit.UnitBuffPower4FF570(0); got != 0x80 {
		t.Fatalf("slot 0 result = %#02x, want 0x80", got)
	}
	if got := unit.UnitBuffPower4FF570(31); got != math.MaxUint8 {
		t.Fatalf("slot 31 result = %#02x, want 0xff", got)
	}
	if got := unit.EnchantPower(EnchantID(31)); got != math.MaxUint8 {
		t.Fatalf("EnchantPower(31) = %d, want 255", got)
	}
	runtime.KeepAlive(unit)
}

func TestUnitBuffPower4FF570NativeInvalidAccessFaults(t *testing.T) {
	unit := new(Object)
	tests := []struct {
		name string
		run  func()
	}{
		{"nil-unit", func() { (*Object)(nil).UnitBuffPower4FF570(0) }},
		{"negative-one", func() { unit.UnitBuffPower4FF570(-1) }},
		{"slot-32", func() { unit.UnitBuffPower4FF570(32) }},
		{"minimum-dword", func() { unit.UnitBuffPower4FF570(math.MinInt32) }},
		{"maximum-dword", func() { unit.UnitBuffPower4FF570(math.MaxInt32) }},
		{"EnchantPower-slot-32", func() { unit.EnchantPower(EnchantID(32)) }},
		{"nil-EnchantPower", func() { (*Object)(nil).EnchantPower(0) }},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Fatal("invalid native access returned instead of faulting")
				}
			}()
			tc.run()
		})
	}
}
