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

func TestItemEnchantmentTooltip413480CEntryAllBytePatternsAndHighReturn(t *testing.T) {
	text, free := alloc.Make([]uint16{}, 3)
	t.Cleanup(free)
	copy(text, []uint16{0xac00, 0xb098, 0})
	pointer := unsafe.Pointer(&text[0])
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(pointer) <= math.MaxUint32 {
		t.Fatal("UTF-16 text must exceed 4 GiB")
	}
	old := itemEnchantmentTooltipCall413480
	t.Cleanup(func() { itemEnchantmentTooltipCall413480 = old })
	var expected byte
	var result unsafe.Pointer
	calls := 0
	itemEnchantmentTooltipCall413480 = func(flag byte) unsafe.Pointer {
		calls++
		if flag != expected {
			t.Fatalf("byte=%x want=%x", flag, expected)
		}
		return result
	}
	for i := 0; i < 256; i++ {
		expected = byte(i)
		for _, result = range []unsafe.Pointer{pointer, nil} {
			if got := itemEnchantmentTooltipCEntry413480(expected); got != result {
				t.Fatalf("byte=%x got=%p want=%p", i, got, result)
			}
		}
	}
	if calls != 512 || alloc.GoString16(&text[0]) != "가나" {
		t.Fatalf("delegations=%d text=%q", calls, alloc.GoString16(&text[0]))
	}
}

func TestItemEnchantmentTooltip413480NativeSideSlotsAndPackedData(t *testing.T) {
	flags := [6]byte{8, 16, 1, 4, 2, 32}
	keys := [6]string{
		"modifier.db:BrillianceItemEnchantDesc", "modifier.db:SpeedItemEnchantDesc",
		"modifier.db:FireProtectItemEnchantDesc", "modifier.db:LightningProtItemEnchantDesc",
		"modifier.db:PoisonProtectItemEnchantDesc", "thing.db:LesserHeal",
	}
	var keyPointers [6]*byte
	var results, oldKeys [6]unsafe.Pointer
	packed := memmap.Slice(0x587000, 27328)[:120]
	oldPacked := bytes.Clone(packed)
	oldLoaded, oldLoad := memmap.Uint32(0x5D4594, 251624), itemEnchantmentTooltipLoad413480
	for i := range oldKeys {
		oldKeys[i] = *memmap.PtrPtr(0x587000, 27344+20*uintptr(i))
	}
	// Register each allocation before its restoration callback, so LIFO
	// cleanup restores side slots before freeing their temporary native data.
	for i, key := range keys {
		buf, freeKey := alloc.Make([]byte{}, len(key)+1)
		copy(buf, key)
		t.Cleanup(freeKey)
		keyPointers[i] = &buf[0]
		text, freeText := alloc.Make([]uint16{}, 2)
		t.Cleanup(freeText)
		text[0] = uint16(0xac00 + i)
		results[i] = unsafe.Pointer(&text[0])
		if unsafe.Sizeof(uintptr(0)) == 8 && (uintptr(results[i]) <= math.MaxUint32 || uintptr(unsafe.Pointer(&buf[0])) <= math.MaxUint32) {
			t.Fatal("native key/text must exceed 4 GiB")
		}
	}
	t.Cleanup(func() {
		itemEnchantmentTooltipLoad413480 = oldLoad
		*memmap.PtrUint32(0x5D4594, 251624) = oldLoaded
		copy(packed, oldPacked)
		for i, key := range oldKeys {
			*memmap.PtrPtr(0x587000, 27344+20*uintptr(i)) = key
		}
	})
	for i := range keys {
		*memmap.PtrPtr(0x587000, 27344+20*uintptr(i)) = unsafe.Pointer(keyPointers[i])
		*memmap.PtrUint8(0x587000, 27332+20*uintptr(i)) = flags[i]
	}
	*memmap.PtrUint32(0x5D4594, 251624) = 0xa5a55a5a
	before := bytes.Clone(packed)
	var loads []int
	itemEnchantmentTooltipLoad413480 = func(key *byte, source string, line int32) unsafe.Pointer {
		if source != `C:\NoxPost\src\common\Object\Modifier.c` || line != 2087 {
			t.Fatalf("source=%q line=%d", source, line)
		}
		for i, want := range keyPointers {
			if key == want && alloc.GoString(key) == keys[i] {
				loads = append(loads, i)
				return results[i]
			}
		}
		t.Fatalf("unexpected key=%p text=%q", key, alloc.GoString(key))
		return nil
	}
	for cycle := 0; cycle < 2; cycle++ {
		loads = nil
		var wantLoads []int
		for query := 0; query < 256; query++ {
			var want unsafe.Pointer
			for i, flag := range flags {
				if byte(query) == flag {
					want, wantLoads = results[i], append(wantLoads, i)
				}
			}
			if got := itemEnchantmentTooltipCEntry413480(byte(query)); got != want {
				t.Fatalf("cycle=%d byte=%x got=%p want=%p", cycle, query, got, want)
			}
		}
		if !reflect.DeepEqual(loads, wantLoads) || !bytes.Equal(packed, before) || memmap.Uint32(0x5D4594, 251624) != 0xa5a55a5a {
			t.Fatalf("loads=%v want=%v packed unchanged=%t loaded=%x", loads, wantLoads, bytes.Equal(packed, before), memmap.Uint32(0x5D4594, 251624))
		}
	}
}
