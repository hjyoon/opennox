package server

import (
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
)

func TestPlayerCanCarryItem513B00KeepsNativePointersAndCheapestTie(t *testing.T) {
	s := &Server{}
	s.itemDropRules.initialized = true
	s.itemDropRules.typeInd[0] = 6 // An item listed by 0053EBF0 is not a drop candidate.
	item := &Object{TypeInd: 42}
	filteredClass := &Object{TypeInd: 1, ObjClass: object.ClassFood}
	filteredFlags := &Object{TypeInd: 2, ObjFlags: object.Flags(0x100)}
	glyph := &Object{TypeInd: 3}
	listed := &Object{TypeInd: 6}
	first := &Object{TypeInd: 4}
	second := &Object{TypeInd: 5}
	filteredClass.InvNextItem = filteredFlags
	filteredFlags.InvNextItem = glyph
	glyph.InvNextItem = listed
	listed.InvNextItem = first
	first.InvNextItem = second
	owner := &Object{InvFirstItem: filteredClass, PosVec: types.Pointf{X: 12, Y: 34}}
	if unsafe.Sizeof(uintptr(0)) > 4 && (uintptr(unsafe.Pointer(item)) <= uintptr(^uint32(0)) || uintptr(unsafe.Pointer(first)) <= uintptr(^uint32(0))) {
		t.Skip("allocator did not place test objects above the PE32 address range")
	}
	var capacityCalls, costCalls, pointCalls, dropCalls, warnings int
	var warned bool
	rt := PlayerCanCarryItemRuntime513B00{
		InventoryCapacity: func(typeID uint16, amount int32) int32 {
			capacityCalls++
			if typeID != 42 || amount != 1 {
				t.Fatalf("capacity args = %d/%d", typeID, amount)
			}
			return 2
		},
		PickupCount: func() int32 { return 2 },
		ItemCost: func(candidate *Object) int32 {
			costCalls++
			if candidate != first && candidate != second {
				t.Fatalf("ineligible cost candidate %p", candidate)
			}
			return 7
		},
		RandomPoint: func(radius float32, center, output *types.Pointf) {
			pointCalls++
			if radius != 50 || center != &owner.PosVec {
				t.Fatalf("point args = %v/%p", radius, center)
			}
			*output = types.Pointf{X: 20, Y: 40}
		},
		Drop: func(gotOwner, chosen *Object, point *types.Pointf) int32 {
			dropCalls++
			if gotOwner != owner || chosen != first || *point != (types.Pointf{X: 20, Y: 40}) {
				t.Fatalf("drop = %p/%p/%v", gotOwner, chosen, *point)
			}
			return 0 // Even a failed drop is followed by the original one-shot warning.
		},
		AlreadyWarned: func() bool { return warned },
		Warn: func(gotOwner *Object) {
			warnings++
			if gotOwner != owner {
				t.Fatalf("warn owner = %p", gotOwner)
			}
			warned = true
		},
	}
	s.PlayerCanCarryItem513B00(owner, item, 3, rt)
	s.PlayerCanCarryItem513B00(owner, item, 3, rt)
	if capacityCalls != 2 || costCalls != 4 || pointCalls != 2 || dropCalls != 2 || warnings != 1 {
		t.Fatalf("effects = capacity %d cost %d point %d drop %d warning %d", capacityCalls, costCalls, pointCalls, dropCalls, warnings)
	}
}

func TestPlayerCanCarryItem513B00CapacityAndPriceThreshold(t *testing.T) {
	s := &Server{}
	s.itemDropRules.initialized = true
	owner := &Object{InvFirstItem: &Object{TypeInd: 7}}
	item := &Object{TypeInd: 8}
	var count, costCalls int
	rt := PlayerCanCarryItemRuntime513B00{
		InventoryCapacity: func(uint16, int32) int32 { return 1 },
		PickupCount:       func() int32 { count++; return 0 },
		ItemCost:          func(*Object) int32 { costCalls++; return 999999 },
	}
	s.PlayerCanCarryItem513B00(owner, item, 0, rt)
	if count != 1 || costCalls != 0 {
		t.Fatalf("available slot effects = count %d cost %d", count, costCalls)
	}
	rt.PickupCount = func() int32 { return 1 }
	s.PlayerCanCarryItem513B00(owner, item, 0, rt)
	if costCalls != 1 {
		t.Fatalf("price threshold cost calls = %d", costCalls)
	}
}
