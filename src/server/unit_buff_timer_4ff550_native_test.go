package server

import (
	"math"
	"reflect"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

func TestUnitBuffTimer4FF550NativeLayout(t *testing.T) {
	wantOffset := uintptr(344)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantOffset = 348
	}
	checks := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"BuffsDur offset", unsafe.Offsetof(Object{}.BuffsDur), wantOffset},
		{"BuffsDur width", unsafe.Sizeof(Object{}.BuffsDur), 64},
		{"BuffsDur element width", unsafe.Sizeof(Object{}.BuffsDur[0]), 2},
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Errorf("Object %s = %d, want %d", check.name, check.got, check.want)
		}
	}
}

func TestUnitBuffTimerNative4FF550AccessorOrderAndWidths(t *testing.T) {
	unit, freeUnit := alloc.New(Object{})
	t.Cleanup(freeUnit)
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(unit)) <= math.MaxUint32 {
		t.Fatalf("unit pointer = %p, want native address above 4 GiB", unit)
	}

	var events []string
	deps := unitBuffTimerNativeDeps4FF550{
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
		loadDuration: func(got *Object, buff int32) uint16 {
			events = append(events, "duration")
			if got != unit || buff != math.MinInt32 {
				t.Fatalf("duration args = (%p, %d), want (%p, %d)", got, buff, unit, int32(math.MinInt32))
			}
			return math.MaxUint16
		},
	}
	if got := unitBuffTimerNative4FF550(unit, math.MinInt32, deps); got != math.MaxUint16 {
		t.Fatalf("result = %#08x, want zero-extended 0x0000ffff", got)
	}
	if want := []string{"buff", "unit", "duration"}; !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %q, want exact accessor order %q", events, want)
	}
	runtime.KeepAlive(unit)
}

func TestUnitBuffTimer4FF550ServerBinding(t *testing.T) {
	unit, freeUnit := alloc.New(Object{})
	t.Cleanup(freeUnit)
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(unit)) <= math.MaxUint32 {
		t.Fatalf("unit pointer = %p, want native address above 4 GiB", unit)
	}

	unit.BuffsDur[0] = 0x8000
	unit.BuffsDur[31] = math.MaxUint16
	if got := unit.UnitBuffTimer4FF550(0); got != 0x8000 {
		t.Fatalf("slot 0 result = %#08x, want 0x00008000", got)
	}
	if got := unit.UnitBuffTimer4FF550(31); got != math.MaxUint16 {
		t.Fatalf("slot 31 result = %#08x, want 0x0000ffff", got)
	}
	if got := unit.EnchantDur(EnchantID(31)); got != math.MaxUint16 {
		t.Fatalf("EnchantDur(31) = %d, want 65535", got)
	}
	runtime.KeepAlive(unit)
}

func TestUnitBuffTimer4FF550NativeInvalidAccessFaults(t *testing.T) {
	unit := new(Object)
	tests := []struct {
		name string
		run  func()
	}{
		{"nil-unit", func() { (*Object)(nil).UnitBuffTimer4FF550(0) }},
		{"negative-one", func() { unit.UnitBuffTimer4FF550(-1) }},
		{"slot-32", func() { unit.UnitBuffTimer4FF550(32) }},
		{"minimum-dword", func() { unit.UnitBuffTimer4FF550(math.MinInt32) }},
		{"maximum-dword", func() { unit.UnitBuffTimer4FF550(math.MaxInt32) }},
		{"EnchantDur-slot-32", func() { unit.EnchantDur(EnchantID(32)) }},
		{"nil-EnchantDur", func() { (*Object)(nil).EnchantDur(0) }},
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
