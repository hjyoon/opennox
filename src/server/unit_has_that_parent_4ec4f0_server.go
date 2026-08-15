package server

func unitHasThatParentNative4EC4F0(obj, owner *Object) bool {
	return unitHasThatParent4EC4F0(obj, owner, unitHasThatParentHooks4EC4F0[*Object]{
		loadOwner: func(current *Object) *Object {
			return current.ObjOwner
		},
	})
}
