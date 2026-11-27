package legacy

import (
	"bytes"
	"encoding/binary"
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

func TestNativePointerTransport(t *testing.T) {
	const width = int(unsafe.Sizeof(uintptr(0)))
	memory, freeMemory := alloc.MakeHigh([]byte{}, 3*width)
	defer freeMemory()
	target, freeTarget := alloc.MallocHigh(16)
	defer freeTarget()
	word := uintptr(target)
	if width == 8 && uint64(word) <= math.MaxUint32 {
		t.Fatal("native pointer fixture must be above 4 GiB")
	}
	if got := nativePointerFromWord(word); got != target || nativePointerFromWord(0) != nil {
		t.Fatalf("native word %#x decoded to %p, want %p", word, got, target)
	}
	for offset := 1; offset <= width; offset++ {
		want := bytes.Repeat([]byte{0xa5}, len(memory))
		copy(memory, want)
		slot := unsafe.Pointer(&memory[offset])
		field := want[offset : offset+width]
		if width == 8 {
			binary.NativeEndian.PutUint64(field, uint64(word))
		} else {
			binary.NativeEndian.PutUint32(field, uint32(word))
		}
		storeForeignPointer(slot, target)
		if got := loadUnalignedNativePointer(slot); got != target || !bytes.Equal(memory, want) {
			t.Fatalf("slot at offset %d holds %p and bytes %x, want %p and %x", offset, got, memory, target, want)
		}
		clear(field)
		storeForeignPointer(slot, nil)
		if got := loadUnalignedNativePointer(slot); got != nil || !bytes.Equal(memory, want) {
			t.Fatalf("cleared slot at offset %d holds %p and bytes %x, want nil and %x", offset, got, memory, want)
		}
	}
}
