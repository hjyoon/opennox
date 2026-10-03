package legacy

import (
	"bytes"
	"math"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

func TestItemEnchantmentIcon413420CEntryAllBytePatternsAndHighReturn(t *testing.T) {
	image, free := alloc.Malloc(8)
	t.Cleanup(free)
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(image) <= math.MaxUint32 {
		t.Fatal("image must exceed 4 GiB")
	}
	old := itemEnchantmentIconCall413420
	t.Cleanup(func() { itemEnchantmentIconCall413420 = old })
	var expected byte
	var result unsafe.Pointer
	calls := 0
	itemEnchantmentIconCall413420 = func(flag byte) unsafe.Pointer {
		calls++
		if flag != expected {
			t.Fatalf("byte=%x want=%x", flag, expected)
		}
		return result
	}
	for i := 0; i < 256; i++ {
		expected = byte(i)
		for _, result = range []unsafe.Pointer{image, nil} {
			if got := itemEnchantmentIconCEntry413420(expected); got != result {
				t.Fatalf("byte=%x got=%p want=%p", i, got, result)
			}
		}
	}
	if calls != 512 {
		t.Fatalf("delegations=%d", calls)
	}
}

func TestItemEnchantmentIcon413420NativeSideSlotsPreservePackedNeighbors(t *testing.T) {
	flags := [6]byte{8, 16, 1, 4, 2, 32}
	names := [6]string{"BrillianceItemIcon", "SpeedItemIcon", "FireProtectItemIcon", "LightningProtectItemIcon", "PoisonProtectItemIcon", "RegenerationItemIcon"}
	var images [6]unsafe.Pointer
	var namePointers [6]*byte
	var oldNames, oldImages [6]unsafe.Pointer
	packed := memmap.Slice(0x587000, 27328)[:120]
	oldPacked := bytes.Clone(packed)
	oldLoaded, oldLoad := memmap.Uint32(0x5D4594, 251624), itemEnchantmentIconLoad413420
	for i, name := range names {
		buf, freeName := alloc.Make([]byte{}, len(name)+1)
		copy(buf, name)
		t.Cleanup(freeName)
		namePointers[i] = &buf[0]
		image, freeImage := alloc.Malloc(8)
		t.Cleanup(freeImage)
		images[i] = image
		if unsafe.Sizeof(uintptr(0)) == 8 && (uintptr(image) <= math.MaxUint32 || uintptr(unsafe.Pointer(&buf[0])) <= math.MaxUint32) {
			t.Fatal("native image/name must exceed 4 GiB")
		}
		nameSlot := memmap.PtrPtr(0x587000, 27336+20*uintptr(i))
		imageSlot := memmap.PtrPtr(0x587000, 27340+20*uintptr(i))
		oldNames[i], oldImages[i] = *nameSlot, *imageSlot
		*nameSlot, *imageSlot = unsafe.Pointer(&buf[0]), nil
		*memmap.PtrUint8(0x587000, 27332+20*uintptr(i)) = flags[i]
	}
	t.Cleanup(func() {
		itemEnchantmentIconLoad413420 = oldLoad
		*memmap.PtrUint32(0x5D4594, 251624) = oldLoaded
		copy(packed, oldPacked)
		for i := range oldNames {
			*memmap.PtrPtr(0x587000, 27336+20*uintptr(i)) = oldNames[i]
			*memmap.PtrPtr(0x587000, 27340+20*uintptr(i)) = oldImages[i]
		}
	})
	before := bytes.Clone(packed)
	var loadedNames []string
	itemEnchantmentIconLoad413420 = func(name *byte) unsafe.Pointer {
		i := len(loadedNames) % 6
		if name != namePointers[i] || alloc.GoString(name) != names[i] || memmap.Uint32(0x5D4594, 251624) != 0 {
			t.Fatalf("load index=%d name=%p text=%q", i, name, alloc.GoString(name))
		}
		loadedNames = append(loadedNames, alloc.GoString(name))
		return images[i]
	}
	for cycle := 0; cycle < 2; cycle++ {
		*memmap.PtrUint32(0x5D4594, 251624) = 0
		loadedNames = nil
		for query := 0; query < 256; query++ {
			var want unsafe.Pointer
			for i, flag := range flags {
				if byte(query) == flag {
					want = images[i]
				}
			}
			if got := itemEnchantmentIconCEntry413420(byte(query)); got != want {
				t.Fatalf("cycle=%d byte=%x got=%p want=%p", cycle, query, got, want)
			}
		}
		if !reflect.DeepEqual(loadedNames, names[:]) || memmap.Uint32(0x5D4594, 251624) != 1 || !bytes.Equal(packed, before) {
			t.Fatalf("names=%v ready=%d packed neighbors changed=%t", loadedNames, memmap.Uint32(0x5D4594, 251624), !bytes.Equal(packed, before))
		}
		for i, want := range images {
			if got := *memmap.PtrPtr(0x587000, 27340+20*uintptr(i)); got != want {
				t.Fatalf("cache %d got=%p want=%p", i, got, want)
			}
		}
	}
}
