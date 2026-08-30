package server

import (
	"math"
	"testing"
	"unsafe"
)

func TestInventoryLookupNative4F78E0PreservesPointersAndFullNetCode(t *testing.T) {
	holder := &Object{}
	first := &Object{NetCode: 0x0000beef}
	matched := &Object{NetCode: math.MaxUint32, InvHolder: holder}
	duplicate := &Object{NetCode: math.MaxUint32, InvHolder: holder}
	holder.InvFirstItem = first
	first.InvNextItem = matched
	matched.InvNextItem = duplicate

	if got := InventoryContains4F78E0(holder, matched); got != 1 {
		t.Fatalf("contains result = %d, want 1", got)
	}
	if got := EquippedItemByCode4F7920(holder, math.MaxUint32); got != matched {
		t.Fatalf("lookup result = %p, want first exact match %p", got, matched)
	}
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(matched)) <= math.MaxUint32 {
		t.Skip("Go heap address did not exercise the high-pointer case; the cgo export test does")
	}
}

func TestInventoryContainsNative4F78E0RequiresHolderRelationship(t *testing.T) {
	holder := &Object{}
	item := &Object{}
	holder.InvFirstItem = item

	if got := InventoryContains4F78E0(holder, item); got != 0 {
		t.Fatalf("contains result = %d, want 0 for missing holder relationship", got)
	}
	item.InvHolder = holder
	if got := InventoryContains4F78E0(holder, item); got != 1 {
		t.Fatalf("contains result = %d, want 1 after relationship is live", got)
	}
}
