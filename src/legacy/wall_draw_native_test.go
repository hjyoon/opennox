package legacy

import (
	"math"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

func TestWallDrawUsesNativeImagePointers(t *testing.T) {
	if !clientWallDrawImageArgumentNative473C10() {
		t.Fatal("wall edge draw image argument is not a native pointer")
	}

	image, freeImage := alloc.Malloc(1)
	t.Cleanup(freeImage)
	addr := uintptr(image)
	if unsafe.Sizeof(addr) == 8 && addr <= math.MaxUint32 {
		t.Fatalf("wall image pointer = %p, want native address above 4 GiB", image)
	}
	if got := clientWallImageAddressRoundTrip473C10(addr); got != addr {
		t.Fatalf("wall image pointer round trip = %#x, want %#x", got, addr)
	}
	runtime.KeepAlive(image)
}
