package server

// dropEligibilityNative4EDCD0 binds the restored predicate to Object fields.
// The logical fields preserve the original 32-bit owner+8/item+16 reads while
// allowing native pointers to expand on 64-bit targets.
func dropEligibilityNative4EDCD0(owner, item *Object) int32 {
	return dropEligibility4EDCD0(dropEligibilityHooks4EDCD0[*Object, *Object]{
		loadItemArg: func() *Object {
			return item
		},
		loadItemFlags: func(item *Object) uint32 {
			return uint32(item.ObjFlags)
		},
		loadOwnerArg: func() *Object {
			return owner
		},
		loadOwnerClassLow: func(owner *Object) uint8 {
			return uint8(owner.ObjClass)
		},
	})
}
