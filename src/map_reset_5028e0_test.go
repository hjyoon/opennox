package opennox

import (
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

func TestMapReset5028E0ClearsOriginalSlotsAndPreservesLazyBuffers(t *testing.T) {
	const blob = 0x5D4594
	zeroSlots := []uintptr{
		1599484, 1599488, 1599492, 1599496,
		1599500, 1599504, 1599508, 1599512,
		1599516, 1599520, 1599524, 1599528,
		1599536, 1599544, 1599552, 1599560,
		1599564, 1599568,
	}
	oldSlots := make([]uint32, len(zeroSlots))
	for i, off := range zeroSlots {
		oldSlots[i] = *memmap.PtrUint32(blob, off)
	}
	oldLastID := *memmap.PtrUint32(blob, 1599572)
	oldAdjacent := *memmap.PtrUint32(blob, 1599576)
	oldIndex := legacy.Get_dword_5d4594_1599480()
	oldFlag := legacy.Get_dword_5d4594_1599476()
	oldListA := legacy.Get_dword_5d4594_1599540()
	oldListB := legacy.Get_dword_5d4594_1599532()
	oldListC := legacy.Get_dword_5d4594_1599556()
	oldListD := legacy.Get_dword_5d4594_1599548()
	oldBufferA := legacy.Get_dword_5d4594_1599588()
	oldBufferB := legacy.Get_dword_5d4594_1599592()

	list, freeList := alloc.Malloc(16)
	bufferA, freeBufferA := alloc.Malloc(2048)
	bufferB, freeBufferB := alloc.Malloc(2048)
	var newBufferA, newBufferB unsafe.Pointer
	t.Cleanup(func() {
		legacy.Set_dword_5d4594_1599480(oldIndex)
		legacy.Set_dword_5d4594_1599476(oldFlag)
		legacy.Set_dword_5d4594_1599540(oldListA)
		legacy.Set_dword_5d4594_1599532(oldListB)
		legacy.Set_dword_5d4594_1599556(oldListC)
		legacy.Set_dword_5d4594_1599548(oldListD)
		legacy.Set_dword_5d4594_1599588(oldBufferA)
		legacy.Set_dword_5d4594_1599592(oldBufferB)
		*memmap.PtrUint32(blob, 1599572) = oldLastID
		*memmap.PtrUint32(blob, 1599576) = oldAdjacent
		for i, off := range zeroSlots {
			*memmap.PtrUint32(blob, off) = oldSlots[i]
		}
		if newBufferA != nil {
			alloc.FreePtr(newBufferA)
		}
		if newBufferB != nil {
			alloc.FreePtr(newBufferB)
		}
		freeBufferB()
		freeBufferA()
		freeList()
	})

	legacy.Set_dword_5d4594_1599480(17)
	legacy.Set_dword_5d4594_1599476(1)
	legacy.Set_dword_5d4594_1599540(list)
	legacy.Set_dword_5d4594_1599532(list)
	legacy.Set_dword_5d4594_1599556(list)
	legacy.Set_dword_5d4594_1599548(list)
	legacy.Set_dword_5d4594_1599588(bufferA)
	legacy.Set_dword_5d4594_1599592(bufferB)
	*memmap.PtrUint32(blob, 1599572) = 17
	*memmap.PtrUint32(blob, 1599576) = 0x12345678
	for _, off := range zeroSlots {
		*memmap.PtrUint32(blob, off) = 0xA5A5A5A5
	}
	s := &Server{Server: &server.Server{}}
	s.MapGroups.Refs = &server.MapGroupRef{}
	s.Nox_xxx_mapReset5028E0()

	if got := legacy.Get_dword_5d4594_1599480(); got != math.MaxUint32 {
		t.Fatalf("map index = %#x, want -1", got)
	}
	if got := *memmap.PtrUint32(blob, 1599572); got != math.MaxUint32 {
		t.Fatalf("last map ID = %#x, want -1", got)
	}
	if got := legacy.Get_dword_5d4594_1599476(); got != 0 {
		t.Fatalf("map flag = %d, want 0", got)
	}
	if legacy.Get_dword_5d4594_1599540() != nil || legacy.Get_dword_5d4594_1599532() != nil ||
		legacy.Get_dword_5d4594_1599556() != nil || legacy.Get_dword_5d4594_1599548() != nil || s.MapGroups.Refs != nil {
		t.Fatal("map list head retained after reset")
	}
	for _, off := range zeroSlots {
		if got := *memmap.PtrUint32(blob, off); got != 0 {
			t.Fatalf("map slot +%d = %#x, want 0", off, got)
		}
	}
	if got := *memmap.PtrUint32(blob, 1599576); got != 0x12345678 {
		t.Fatalf("adjacent map slot = %#x, want unchanged value", got)
	}
	if legacy.Get_dword_5d4594_1599588() != bufferA || legacy.Get_dword_5d4594_1599592() != bufferB {
		t.Fatal("reset replaced existing map buffers")
	}

	legacy.Set_dword_5d4594_1599588(nil)
	legacy.Set_dword_5d4594_1599592(nil)
	s.Nox_xxx_mapReset5028E0()
	newBufferA = legacy.Get_dword_5d4594_1599588()
	newBufferB = legacy.Get_dword_5d4594_1599592()
	if newBufferA == nil || newBufferB == nil || newBufferA == newBufferB {
		t.Fatal("reset did not allocate two distinct map buffers")
	}
	for _, p := range []unsafe.Pointer{newBufferA, newBufferB} {
		for i, b := range unsafe.Slice((*byte)(p), 2048) {
			if b != 0 {
				t.Fatalf("new map buffer byte %d = %#x, want zero", i, b)
			}
		}
	}
}
