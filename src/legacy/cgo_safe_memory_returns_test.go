//go:build safe

package legacy

import (
	"bytes"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/legacy/internal/cfixture"
)

func TestSafeLegacyCBufferAllocations(t *testing.T) {
	for _, text := range []string{"", "left\x00right\xff"} {
		p := CString(text)
		got := unsafe.Slice((*byte)(unsafe.Pointer(p)), len(text)+1)
		if string(got[:len(text)]) != text || got[len(text)] != 0 {
			t.Errorf("CString(%q) = %x", text, got)
		}
		StrFree(p)
	}
	for _, input := range [][]byte{nil, {0, 0xff, 7}} {
		p := (*byte)(CBytes(input))
		if p == nil || !bytes.Equal(unsafe.Slice(p, len(input)), input) {
			t.Errorf("CBytes(%x) produced the wrong buffer", input)
		}
		StrFree(p)
	}
}

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
