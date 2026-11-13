package legacy

import (
	"runtime"
	"runtime/debug"
	"sync/atomic"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/legacy/common/alloc/memguard"
)

// A literal invalid native address may overlap Go/386's heap. Keep malformed
// pointer roles in real foreign NOACCESS storage on 32-bit hosts so GC can
// safely retain them while the declared read still faults. On 64-bit hosts
// callers continue using their original reference getter and literal address.
func originalInvalidNativePointer32(t *testing.T) unsafe.Pointer {
	t.Helper()
	if unsafe.Sizeof(uintptr(0)) != 4 {
		return nil
	}
	guard, release := memguard.New(0x10000)
	t.Cleanup(release)
	return unsafe.Pointer(&guard[len(guard)/2])
}

func TestOriginalInvalidNativePointer32GCAndFault(t *testing.T) {
	ptr := originalInvalidNativePointer32(t)
	if unsafe.Sizeof(uintptr(0)) != 4 {
		if ptr != nil {
			t.Fatal("64-bit fixtures must retain their original getter addresses")
		}
		return
	}
	if ptr == nil {
		t.Fatal("32-bit malformed-pointer fixture did not own foreign storage")
	}
	roots := []*uint32{(*uint32)(ptr), (*uint32)(unsafe.Add(ptr, 128))}
	runtime.GC()
	old := debug.SetPanicOnFault(true)
	defer debug.SetPanicOnFault(old)
	faulted := func() (faulted bool) {
		defer func() { faulted = recover() != nil }()
		atomic.LoadUint32(roots[1])
		return false
	}()
	if !faulted {
		t.Fatal("malformed native pointer became readable")
	}
	runtime.KeepAlive(roots)
}
