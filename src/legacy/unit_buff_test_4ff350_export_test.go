package legacy

import (
	"math"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

func TestUnitBuffTestExport4FF350PreservesNativePointerAndSignedDword(t *testing.T) {
	unit, freeUnit := alloc.New(server.Object{})
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
		{33, 0},
		{math.MinInt32, 1},
		{math.MaxInt32, 1},
	}
	for _, tc := range tests {
		if got := unitBuffTestExportCall4FF350(unit, tc.buff); got != tc.want {
			t.Errorf("CGo buff %d result = %d, want %d", tc.buff, got, tc.want)
		}
	}
	if got := unitBuffTestExportCall4FF350(nil, math.MinInt32); got != 0 {
		t.Fatalf("CGo nil result = %d, want canonical zero", got)
	}

	runtime.KeepAlive(unit)
}
