package server

var objectCountInventoryTypeHooks4E7D30 = countInventoryTypeHooks4E7D30[*Object]{
	loadFirst: func(owner *Object) *Object {
		return owner.InvFirstItem
	},
	loadType: func(obj *Object) uint16 {
		return obj.TypeInd
	},
	loadFlagsLow: func(obj *Object) uint8 {
		return uint8(obj.ObjFlags)
	},
	loadNext: func(obj *Object) *Object {
		return obj.InvNextItem
	},
}

func (obj *Object) CountInventoryWithType(typeInd int32) int32 { // nox_xxx_inventoryCountObjects_4E7D30
	return countInventoryType4E7D30(obj, typeInd, objectCountInventoryTypeHooks4E7D30)
}
