//go:build safe

package legacy

import (
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/legacy/internal/cfixture"
)

func TestSafeMemoryReturnsCallerDestination(t *testing.T) {
	check := func(t *testing.T, dst unsafe.Pointer) {
		t.Helper()
		buffer := unsafe.Slice((*byte)(dst), 16)
		buffer[0], buffer[15] = 0x5a, 0x5a
		if status := cfixture.SafeMemoryResults(dst); status != 0 {
			t.Fatalf("safe C result or contents for %p: %d, %x", dst, status, buffer)
		}
	}
	t.Run("native", func(t *testing.T) {
		dst, free := alloc.Malloc(16)
		defer free()
		check(t, dst)
	})
	t.Run("borrowed", func(t *testing.T) {
		var dst [16]byte
		check(t, unsafe.Pointer(&dst[0]))
	})
	t.Run("explicitly pinned", func(t *testing.T) {
		var dst [16]byte
		var pin runtime.Pinner
		pin.Pin(&dst[0])
		defer pin.Unpin()
		check(t, unsafe.Pointer(&dst[0]))
	})
}
