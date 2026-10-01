package legacy

import (
	"bytes"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

func TestClientStatusAnimations473930NativeCache(t *testing.T) {
	birds, freeBirds := alloc.New(ImageRef{RefKind: 2})
	shield, freeShield := alloc.New(ImageRef{RefKind: 2})
	t.Cleanup(freeBirds)
	t.Cleanup(freeShield)
	if unsafe.Sizeof(uintptr(0)) > 4 && (uintptr(birds.C()) <= 0xffffffff || uintptr(shield.C()) <= 0xffffffff) {
		t.Fatalf("expected native animation references above 4 GiB: %p/%p", birds, shield)
	}
	for _, refs := range [][2]*ImageRef{{birds, shield}, {shield, birds}, {nil, shield}, {birds, nil}, {nil, nil}} {
		oldLoad := Nox_xxx_gLoadAnim
		oldBirds := *memmap.PtrPtr(0x5D4594, 1096456)
		oldShield := *memmap.PtrPtr(0x5D4594, 1096460)
		packed := memmap.Slice(0x5D4594, 1096452)[:16]
		before := append([]byte(nil), packed...)
		var calls []string
		var publishedBeforeSecondLoad unsafe.Pointer
		Nox_xxx_gLoadAnim = func(name string) *ImageRef {
			calls = append(calls, name)
			if name == "ConfusedBirdies" {
				return refs[0]
			}
			publishedBeforeSecondLoad = *memmap.PtrPtr(0x5D4594, 1096456)
			return refs[1]
		}
		Sub_473930() // Exercise the actual C -> Go startup ABI, not a mock caller.
		gotBirds := *memmap.PtrPtr(0x5D4594, 1096456)
		gotShield := *memmap.PtrPtr(0x5D4594, 1096460)
		unchanged := bytes.Equal(packed, before)
		Nox_xxx_gLoadAnim = oldLoad
		*memmap.PtrPtr(0x5D4594, 1096456) = oldBirds
		*memmap.PtrPtr(0x5D4594, 1096460) = oldShield
		copy(packed, before)
		if !reflect.DeepEqual(calls, []string{"ConfusedBirdies", "SphericalShieldAnim"}) {
			t.Fatalf("startup load order = %v", calls)
		}
		if gotBirds != unsafe.Pointer(refs[0]) || gotShield != unsafe.Pointer(refs[1]) || publishedBeforeSecondLoad != unsafe.Pointer(refs[0]) {
			t.Fatalf("native caches = %p/%p, first published=%p; want %p/%p", gotBirds, gotShield, publishedBeforeSecondLoad, refs[0], refs[1])
		}
		if unsafe.Sizeof(uintptr(0)) > 4 && !unchanged {
			t.Fatal("native pointer cache write changed packed PE32 bytes/neighbor fields")
		}
	}
}
