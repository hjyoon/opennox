package server

import (
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
)

func TestChestInit4F0400NativeLayout(t *testing.T) {
	wantObjectSize := uintptr(780)
	wantField5 := uintptr(20)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantObjectSize = 928
		wantField5 = 24
	}
	checks := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"Object size", unsafe.Sizeof(Object{}), wantObjectSize},
		{"Object.Field5", unsafe.Offsetof(Object{}.Field5), wantField5},
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Fatalf("%s = %d, want %d", check.name, check.got, check.want)
		}
	}
}

func TestChestInit4F0400NativeUsesLowByteAndExactObject(t *testing.T) {
	unit := &Object{Field5: 0x0e000001}
	calls := 0
	chestInitNative4F0400(unit, chestInitNativeDeps4F0400{
		setXStatus: func(got *Object, bit uint32) {
			calls++
			if got != unit || bit != 2 {
				t.Fatalf("set args = %p/%#x, want %p/2", got, bit, unit)
			}
			got.Field5 |= bit
		},
	})
	if calls != 1 || unit.Field5 != 0x0e000003 {
		t.Fatalf("calls/status = %d/%#x, want 1/0x0e000003", calls, unit.Field5)
	}
}

func TestChestInit4F0400NativeMaskedBitsSkipSet(t *testing.T) {
	for _, status := range []uint32{2, 4, 8, 0xffffff0e} {
		unit := &Object{Field5: status}
		calls := 0
		chestInitNative4F0400(unit, chestInitNativeDeps4F0400{
			setXStatus: func(*Object, uint32) {
				calls++
			},
		})
		if calls != 0 || unit.Field5 != status {
			t.Fatalf("status %#x: calls/result = %d/%#x", status, calls, unit.Field5)
		}
	}
}

func TestChestInit4F0400NativeNilUnitFaultsBeforeSet(t *testing.T) {
	calls := 0
	defer func() {
		if recover() == nil {
			t.Fatal("nil Object did not preserve the original Field5-load fault")
		}
		if calls != 0 {
			t.Fatalf("set calls = %d, want 0", calls)
		}
	}()
	chestInitNative4F0400(nil, chestInitNativeDeps4F0400{
		setXStatus: func(*Object, uint32) {
			calls++
		},
	})
}

func TestChestInit4F0400NativeBindsSetXStatus(t *testing.T) {
	s := new(Server)
	unit := &Object{ObjClass: object.ClassImmobile, Field5: 1, Field38: 0x12345678}
	for index := range unit.Field140 {
		unit.Field140[index] = uint32(index)
	}
	s.ChestInit4F0400(unit)
	if unit.Field5 != 3 || unit.Field38 != math.MaxUint32 {
		t.Fatalf("status/sync = %#x/%#x, want 3/max", unit.Field5, unit.Field38)
	}
	for index, value := range unit.Field140 {
		if value != 0x80000 {
			t.Fatalf("Field140[%d] = %#x, want 0x80000", index, value)
		}
	}

	masked := &Object{ObjClass: object.ClassImmobile, Field5: 2, Field38: 0x89abcdef}
	masked.Field140[0] = 0x12345678
	s.ChestInit4F0400(masked)
	if masked.Field5 != 2 || masked.Field38 != 0x89abcdef || masked.Field140[0] != 0x12345678 {
		t.Fatalf("masked object changed: status/sync/data = %#x/%#x/%#x", masked.Field5, masked.Field38, masked.Field140[0])
	}
}
