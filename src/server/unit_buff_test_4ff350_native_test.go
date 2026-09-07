package server

import (
	"math"
	"reflect"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

func TestUnitBuffTest4FF350NativeLayout(t *testing.T) {
	wantBuffs := uintptr(340)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantBuffs = 344
	}
	if got := unsafe.Offsetof(Object{}.Buffs); got != wantBuffs {
		t.Fatalf("Object.Buffs offset = %d, want %d", got, wantBuffs)
	}
	if got := unsafe.Sizeof(Object{}.Buffs); got != 4 {
		t.Fatalf("Object.Buffs width = %d, want exact dword", got)
	}
}

func TestUnitBuffTestNative4FF350AccessorOrder(t *testing.T) {
	unit := new(Object)
	var events []string
	deps := unitBuffTestNativeDeps4FF350{
		loadBuffArg: func() int32 {
			events = append(events, "buff")
			return -1
		},
		loadBuffs: func(got *Object) uint32 {
			events = append(events, "buffs")
			if got != unit {
				t.Fatalf("unit = %p, want full native identity %p", got, unit)
			}
			return uint32(1) << 31
		},
	}
	if got := unitBuffTestNative4FF350(unit, deps); got != 1 {
		t.Fatalf("result = %d, want canonical one", got)
	}
	if want := []string{"buff", "buffs"}; !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %q, want %q", events, want)
	}

	events = nil
	if got := unitBuffTestNative4FF350(nil, deps); got != 0 {
		t.Fatalf("nil result = %d, want canonical zero", got)
	}
	if len(events) != 0 {
		t.Fatalf("nil events = %q, want no dependency access", events)
	}
}

func TestUnitBuffTest4FF350ServerBindingPreservesNativePointer(t *testing.T) {
	unit, freeUnit := alloc.New(Object{})
	t.Cleanup(freeUnit)
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(unit)) <= math.MaxUint32 {
		t.Fatalf("unit pointer = %p, want native address above 4 GiB", unit)
	}

	unit.Buffs = uint32(1)<<0 | uint32(1)<<31
	tests := []struct {
		buff int32
		want int32
	}{
		{0, 1},
		{31, 1},
		{32, 1},
		{-1, 1},
		{1, 0},
		{33, 0},
		{math.MinInt32, 1},
		{math.MaxInt32, 1},
	}
	for _, tc := range tests {
		if got := unit.UnitBuffTest4FF350(tc.buff); got != tc.want {
			t.Errorf("buff %d result = %d, want %d", tc.buff, got, tc.want)
		}
	}

	if !unit.HasEnchant(EnchantID(32)) {
		t.Fatal("HasEnchant(32) rejected the original modulo-32 alias for buff 0")
	}
	if unit.HasEnchant(EnchantID(33)) {
		t.Fatal("HasEnchant(33) selected a clear bit instead of aliasing buff 1")
	}
	if got := (*Object)(nil).UnitBuffTest4FF350(-1); got != 0 {
		t.Fatalf("nil receiver result = %d, want canonical zero", got)
	}
	if (*Object)(nil).HasEnchant(ENCHANT_INVISIBLE) {
		t.Fatal("nil HasEnchant receiver returned true")
	}

	runtime.KeepAlive(unit)
}
