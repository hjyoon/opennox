package server

type unitSetOwnerNativeDeps4EC290 struct {
	clearOwner     func(*Object)
	resetMonster   func(*Object)
	markUnitUpdate func(*Object)
}

func unitSetOwnerNative4EC290(owner, obj *Object, deps unitSetOwnerNativeDeps4EC290) {
	unitSetOwner4EC290(owner, obj, unitSetOwnerHooks4EC290[*Object]{
		clearOwner: deps.clearOwner,
		loadFlags: func(obj *Object) uint32 {
			return uint32(obj.ObjFlags)
		},
		loadOwner: func(obj *Object) *Object {
			return obj.ObjOwner
		},
		loadFirstOwned: func(owner *Object) *Object {
			return owner.Field129
		},
		storeNextOwned: func(obj, next *Object) {
			obj.Field128 = next
		},
		storeFirstOwned: func(owner, first *Object) {
			owner.Field129 = first
		},
		storeOwner: func(obj, owner *Object) {
			obj.ObjOwner = owner
		},
		loadClass: func(obj *Object) uint32 {
			return uint32(obj.ObjClass)
		},
		resetMonster:   deps.resetMonster,
		markUnitUpdate: deps.markUnitUpdate,
	})
}

func (s *Server) unitSetOwner4EC290(owner, obj *Object) {
	unitSetOwnerNative4EC290(owner, obj, unitSetOwnerNativeDeps4EC290{
		clearOwner: s.ObjClearOwner,
		resetMonster: func(obj *Object) {
			obj.Nox_xxx_monsterResetEnemy_5346F0()
		},
		markUnitUpdate: func(obj *Object) {
			obj.Nox_xxx_monsterMarkUpdate_4E8020()
		},
	})
}
