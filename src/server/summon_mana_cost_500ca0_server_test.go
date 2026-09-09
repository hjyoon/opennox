package server

import (
	"math"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
)

func TestSummonManaCostNative500CA0PreservesPointerAndSignedCost(t *testing.T) {
	unit := &Object{ObjClass: object.ClassPlayer | object.Class(0x80000000)}
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(unit)) <= math.MaxUint32 {
		t.Fatalf("unit pointer = %p, want native address above 4 GiB", unit)
	}

	loads := 0
	got := summonManaCostNative500CA0(75, unit, func(address uint32) int32 {
		loads++
		if address != 0x005bc370 {
			t.Fatalf("cost address = %#08x, want 0x005bc370", address)
		}
		return math.MinInt32 + 0x500ca0
	})
	if want := int32(math.MinInt32 + 0x500ca0); got != want || loads != 1 {
		t.Fatalf("cost/loads = %d/%d, want %d/1", got, loads, want)
	}
	runtime.KeepAlive(unit)
}

func TestSummonManaCostNative500CA0ExactGates(t *testing.T) {
	loadCost := func(uint32) int32 {
		t.Fatal("cost loaded across null or non-Player gate")
		return 0
	}
	if got := summonManaCostNative500CA0(75, nil, loadCost); got != 0 {
		t.Fatalf("null result = %d, want 0", got)
	}
	if got := summonManaCostNative500CA0(75, &Object{ObjClass: object.ClassMonster}, loadCost); got != 0 {
		t.Fatalf("Monster result = %d, want 0", got)
	}
	if got := summonManaCostNative500CA0(75, &Object{ObjClass: object.Class(0x00000400)}, loadCost); got != 0 {
		t.Fatalf("upper-byte Player bit result = %d, want 0", got)
	}
}

func TestSummonManaCostNative500CA0BlobTranslation(t *testing.T) {
	tests := []struct {
		address uint32
		want    uintptr
	}{
		{address: 0x00587000, want: 0},
		{address: 0x005bc244, want: 217668},
		{address: 0x005bc370, want: 217968},
		{address: 0x005bc40c, want: 218124},
	}
	for _, tc := range tests {
		if got := summonManaCostBlobOffset500CA0(tc.address); got != tc.want {
			t.Errorf("address %#08x offset = %d, want %d", tc.address, got, tc.want)
		}
	}
}

func TestSummonManaCostNative500CA0Layout(t *testing.T) {
	wantObjectSize := uintptr(780)
	wantClass := uintptr(8)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantObjectSize = 928
		wantClass = 12
	}
	checks := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"Object size", unsafe.Sizeof(Object{}), wantObjectSize},
		{"Object.ObjClass", unsafe.Offsetof(Object{}.ObjClass), wantClass},
		{"Object.ObjClass width", unsafe.Sizeof(Object{}.ObjClass), 4},
		{"cost result width", unsafe.Sizeof(int32(0)), 4},
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Errorf("%s on %s/%s = %d, want %d", check.name, runtime.GOOS, runtime.GOARCH, check.got, check.want)
		}
	}
}
