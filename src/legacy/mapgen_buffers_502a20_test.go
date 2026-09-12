package legacy

import (
	"math"
	"testing"
	"unsafe"
)

func TestMapgenBuffers502A20PreserveOriginalContractAndNativePointers(t *testing.T) {
	first, second, ok := mapgenBufferContract502A20()
	if !ok {
		t.Fatal("mapgen count, save-by-name, or buffer accessor contract failed")
	}
	if first == second {
		t.Fatal("the two mapgen name buffers alias")
	}
	if unsafe.Sizeof(uintptr(0)) > 4 && (first <= math.MaxUint32 || second <= math.MaxUint32) {
		t.Fatalf("mapgen buffer pointers %#x, %#x were not above the PE32 range", first, second)
	}
}
