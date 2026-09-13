package server

import "github.com/opennox/libs/types"

const playerCanCarryItemInitialCost513B00 = int32(999999)

// PlayerCanCarryItemRuntime513B00 supplies the external services used by the
// original Coop pickup helper. Object and point arguments stay native-width.
type PlayerCanCarryItemRuntime513B00 struct {
	InventoryCapacity func(uint16, int32) int32
	PickupCount       func() int32
	ItemCost          func(*Object) int32
	RandomPoint       func(float32, *types.Pointf, *types.Pointf)
	Drop              func(*Object, *Object, *types.Pointf) int32
	AlreadyWarned     func() bool
	Warn              func(*Object)
}

// PlayerCanCarryItem513B00 preserves GAME.EXE 00513B00. The client inventory
// capacity is compared with the per-frame pickup count as a signed dword. If
// exhausted, the first eligible lowest-cost inventory item is dropped before
// the one-shot primary message. Equal-cost items retain the earlier item.
func (s *Server) PlayerCanCarryItem513B00(owner, incoming *Object, glyphType uint32, rt PlayerCanCarryItemRuntime513B00) {
	if rt.InventoryCapacity(incoming.TypeInd, 1)-rt.PickupCount() > 0 {
		return
	}
	var chosen *Object
	lowest := playerCanCarryItemInitialCost513B00
	for it := owner.InvFirstItem; it != nil; it = it.InvNextItem {
		if uint32(it.ObjClass)&0x10 != 0 || uint32(it.ObjFlags)&0x100 != 0 || uint32(it.TypeInd) == glyphType || s.ItemIsDroppable53EBF0(it) != 0 {
			continue
		}
		cost := rt.ItemCost(it)
		if cost < lowest {
			lowest, chosen = cost, it
		}
	}
	if chosen == nil {
		return
	}
	var point types.Pointf
	rt.RandomPoint(50, &owner.PosVec, &point)
	rt.Drop(owner, chosen, &point)
	if !rt.AlreadyWarned() {
		rt.Warn(owner)
	}
}
