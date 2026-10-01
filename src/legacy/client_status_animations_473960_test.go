package legacy

import (
	"bytes"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

func TestClientStatusAnimations473960ResetNativeCache(t *testing.T) {
	birds, freeBirds := alloc.New(ImageRef{RefKind: 2})
	shield, freeShield := alloc.New(ImageRef{RefKind: 2})
	t.Cleanup(freeBirds)
	t.Cleanup(freeShield)
	if unsafe.Sizeof(uintptr(0)) > 4 && (uintptr(birds.C()) <= 0xffffffff || uintptr(shield.C()) <= 0xffffffff) {
		t.Fatalf("expected native references above 4 GiB: %p/%p", birds, shield)
	}
	oldBirds := *memmap.PtrPtr(0x5D4594, 1096456)
	oldShield := *memmap.PtrPtr(0x5D4594, 1096460)
	packed := memmap.Slice(0x5D4594, 1096452)[:16]
	before := append([]byte(nil), packed...)
	t.Cleanup(func() {
		*memmap.PtrPtr(0x5D4594, 1096456) = oldBirds
		*memmap.PtrPtr(0x5D4594, 1096460) = oldShield
		copy(packed, before)
	})
	for _, refs := range [][2]*ImageRef{{birds, shield}, {nil, shield}, {birds, nil}, {nil, nil}} {
		*memmap.PtrPtr(0x5D4594, 1096456) = unsafe.Pointer(refs[0])
		*memmap.PtrPtr(0x5D4594, 1096460) = unsafe.Pointer(refs[1])
		for attempt := 0; attempt < 2; attempt++ {
			Sub_473960() // Actual session cleanup ABI; repeated reset is harmless.
			if got := *memmap.PtrPtr(0x5D4594, 1096456); got != nil {
				t.Fatalf("first animation cache survives reset: %p", got)
			}
			if got := *memmap.PtrPtr(0x5D4594, 1096460); got != nil {
				t.Fatalf("second animation cache survives reset: %p", got)
			}
			if unsafe.Sizeof(uintptr(0)) > 4 && !bytes.Equal(packed, before) {
				t.Fatal("native reset mutated packed PE32 bytes")
			}
			if !bytes.Equal(packed[:4], before[:4]) || !bytes.Equal(packed[12:], before[12:]) {
				t.Fatal("animation reset mutated neighboring packed fields")
			}
		}
	}
}
