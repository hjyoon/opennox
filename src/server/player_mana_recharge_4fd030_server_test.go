package server

import (
	"math"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
)

func TestPlayerManaRechargeNative4FD030PreservesNativePointerAndWidths(t *testing.T) {
	unit := &Object{ObjClass: object.ClassPlayer | 0x100}
	replacement := &Object{ObjClass: object.ClassMonster}
	if unsafe.Sizeof(uintptr(0)) == 8 {
		for name, pointer := range map[string]uintptr{
			"unit":        uintptr(unsafe.Pointer(unit)),
			"replacement": uintptr(unsafe.Pointer(replacement)),
		} {
			if pointer <= math.MaxUint32 {
				t.Fatalf("%s pointer = %#x, want native address above 4 GiB", name, pointer)
			}
		}
	}

	original := unit
	addCalls := 0
	got := playerManaRechargeNative4FD030(unit, math.MinInt16, playerManaRechargeNativeDeps4FD030{
		addMana: func(gotUnit *Object, gotAmount int16) uint16 {
			addCalls++
			if gotUnit != original || gotAmount != math.MinInt16 {
				t.Fatalf("add args = %p/%d, want %p/INT16_MIN", gotUnit, gotAmount, original)
			}
			unit = replacement
			return math.MaxUint16
		},
	})
	if got != math.MaxUint16 || addCalls != 1 {
		t.Fatalf("result/add calls = %#x/%d, want UINT16_MAX/1", got, addCalls)
	}
	runtime.KeepAlive(original)
	runtime.KeepAlive(replacement)
}

func TestPlayerManaRechargeNative4FD030NonPlayerReturnsPointerLowWord(t *testing.T) {
	unit := &Object{ObjClass: object.ClassMonster | 0x400}
	want := uint16(uintptr(unsafe.Pointer(unit)))
	got := playerManaRechargeNative4FD030(unit, math.MaxInt16, playerManaRechargeNativeDeps4FD030{
		addMana: func(*Object, int16) uint16 {
			t.Fatal("non-Player invoked mana addition")
			return 0
		},
	})
	if got != want {
		t.Fatalf("result = %#x, want pointer low word %#x", got, want)
	}
	runtime.KeepAlive(unit)
}

func TestPlayerManaRechargeNative4FD030NilFaultsAtClassLoad(t *testing.T) {
	addCalls := 0
	defer func() {
		if recover() == nil {
			t.Fatal("nil unit did not fault")
		}
		if addCalls != 0 {
			t.Fatalf("nil unit invoked mana addition %d times", addCalls)
		}
	}()
	_ = playerManaRechargeNative4FD030(nil, 1, playerManaRechargeNativeDeps4FD030{
		addMana: func(*Object, int16) uint16 {
			addCalls++
			return 0
		},
	})
}

func TestPlayerManaRecharge4FD030ServerMethodUsesNativeAdapter(t *testing.T) {
	unit := &Object{ObjClass: object.ClassPlayer}
	result := new(Server).PlayerManaRecharge4FD030(unit, math.MaxInt16, func(gotUnit *Object, amount int16) uint16 {
		if gotUnit != unit || amount != math.MaxInt16 {
			t.Fatalf("add args = %p/%d, want %p/INT16_MAX", gotUnit, amount, unit)
		}
		return 0x8000
	})
	if result != 0x8000 {
		t.Fatalf("result = %#x, want 0x8000", result)
	}
}

func TestPlayerManaRecharge4FD030NativeLayout(t *testing.T) {
	wantSize := uintptr(780)
	wantClass := uintptr(8)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantSize = 928
		wantClass = 12
	}
	if got := unsafe.Sizeof(Object{}); got != wantSize {
		t.Fatalf("Object size on %s/%s = %d, want %d", runtime.GOOS, runtime.GOARCH, got, wantSize)
	}
	if got := unsafe.Offsetof(Object{}.ObjClass); got != wantClass {
		t.Fatalf("Object.ObjClass on %s/%s = %d, want %d", runtime.GOOS, runtime.GOARCH, got, wantClass)
	}
}
