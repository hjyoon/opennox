package server

import "github.com/opennox/libs/types"

// DropOwnedCrownsRuntime4ED050 supplies the shared legacy Crown type cache and
// the CrownDrop callback. The cache remains shared with the still-raw
// GAME.EXE 004ED810 path, while all object and update-data values use native
// pointers.
type DropOwnedCrownsRuntime4ED050 struct {
	LoadCrownTypeCache  func() uint32
	LookupCrownType     func() uint32
	StoreCrownTypeCache func(uint32)
	DropCrown           func(owner, crown *Object, position *types.Pointf) uint32
}

type dropOwnedCrownsNativeDeps4ED050 struct {
	loadCrownTypeCache  func() uint32
	lookupCrownType     func() uint32
	storeCrownTypeCache func(uint32)
	dropCrown           func(*Object, *Object, *types.Pointf) uint32
}

func dropOwnedCrownsNative4ED050(
	owner, target *Object,
	deps dropOwnedCrownsNativeDeps4ED050,
) {
	dropOwnedCrowns4ED050(dropOwnedCrownsHooks4ED050[
		*Object,
		*CrownUpdateData,
		*types.Pointf,
	]{
		loadCrownTypeCache:  deps.loadCrownTypeCache,
		lookupCrownType:     deps.lookupCrownType,
		storeCrownTypeCache: deps.storeCrownTypeCache,
		loadOwnerArg: func() *Object {
			return owner
		},
		firstOwned: func(obj *Object) *Object {
			return obj.Field129
		},
		loadTargetArg: func() *Object {
			return target
		},
		loadTypeIndex: func(obj *Object) uint16 {
			return obj.TypeInd
		},
		loadUpdate: func(obj *Object) *CrownUpdateData {
			return (*CrownUpdateData)(obj.UpdateData)
		},
		ownerPosition: func(obj *Object) *types.Pointf {
			return &obj.PosVec
		},
		dropCrown: deps.dropCrown,
		storePickupTarget: func(update *CrownUpdateData, target *Object) {
			update.PickupTarget = target
		},
		nextOwned: func(obj *Object) *Object {
			return obj.Field128
		},
	})
}

// DropOwnedCrowns4ED050 drops every owned object whose zero-extended TypeInd
// matches the live shared Crown cache and records target in each Crown's
// pre-drop cached update record.
func (s *Server) DropOwnedCrowns4ED050(
	owner, target *Object,
	runtime DropOwnedCrownsRuntime4ED050,
) {
	dropOwnedCrownsNative4ED050(owner, target, dropOwnedCrownsNativeDeps4ED050{
		loadCrownTypeCache:  runtime.LoadCrownTypeCache,
		lookupCrownType:     runtime.LookupCrownType,
		storeCrownTypeCache: runtime.StoreCrownTypeCache,
		dropCrown:           runtime.DropCrown,
	})
}
