package legacy

import (
	"math"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

func TestSpellDurationNextExport4FE940UsesLiveNativeLink(t *testing.T) {
	record, freeRecord := alloc.New(server.DurSpell{ID: 0x1234})
	first, freeFirst := alloc.New(server.DurSpell{ID: 0xabcd})
	second, freeSecond := alloc.New(server.DurSpell{ID: 0x5678})
	t.Cleanup(freeRecord)
	t.Cleanup(freeFirst)
	t.Cleanup(freeSecond)

	if unsafe.Sizeof(uintptr(0)) == 8 {
		for name, pointer := range map[string]unsafe.Pointer{
			"record": unsafe.Pointer(record),
			"first":  unsafe.Pointer(first),
			"second": unsafe.Pointer(second),
		} {
			if uintptr(pointer) <= math.MaxUint32 {
				t.Fatalf("%s pointer = %p, want native address above 4 GiB", name, pointer)
			}
		}
	}

	record.Next = first
	if got := spellDurationNextExportCall4FE940(unsafe.Pointer(record)); got != unsafe.Pointer(first) {
		t.Fatalf("first CGo result = %p, want %p", got, first)
	}
	record.Next = second
	if got := spellDurationNextExportCall4FE940(unsafe.Pointer(record)); got != unsafe.Pointer(second) {
		t.Fatalf("second CGo result = %p, want live replacement %p", got, second)
	}
	record.Next = nil
	if got := spellDurationNextExportCall4FE940(unsafe.Pointer(record)); got != nil {
		t.Fatalf("nil-link CGo result = %p, want nil", got)
	}
	runtime.KeepAlive(record)
	runtime.KeepAlive(first)
	runtime.KeepAlive(second)
}
