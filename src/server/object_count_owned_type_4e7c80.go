package server

var objectCountOwnedTypeHooks4E7C80 = countOwnedTypeHooks4E7C80[*Object]{
	loadFirst: func(owner *Object) *Object {
		return owner.Field129
	},
	loadType: func(obj *Object) uint16 {
		return obj.TypeInd
	},
	loadFlagsLow: func(obj *Object) uint8 {
		return uint8(obj.ObjFlags)
	},
	loadNext: func(obj *Object) *Object {
		return obj.Field128
	},
}

func (obj *Object) CountSubOfType(typeInd int32) int32 { // nox_xxx_unitIsUnitTT_4E7C80
	return countOwnedType4E7C80(obj, typeInd, objectCountOwnedTypeHooks4E7C80)
}
