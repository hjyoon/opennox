package legacy

import (
	"math"
	"testing"
	"unsafe"
)

func TestMapgenRecordLookup5029A0PreservesNativePointer(t *testing.T) {
	address, ok := mapgenRecordLookupContract5029A0()
	if !ok {
		t.Fatal("mapgen record lookup or entry access broke its original index contract")
	}
	if unsafe.Sizeof(uintptr(0)) > 4 && address <= math.MaxUint32 {
		t.Fatalf("record address = %#x, want a pointer above the PE32 range", address)
	}
}
