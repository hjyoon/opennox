package server

import "github.com/opennox/libs/types"

type respawnInventoryClassNativeDeps4EDD00 struct {
	detachInventory func(*Object, *Object)
	randomReachable func(float32, *types.Pointf, *types.Pointf) *types.Pointf
	createAt        func(*Object, *Object, types.Pointf)
}

// respawnInventoryClassNative4EDD00 binds the restored control flow to
// native-width Object pointers. Named fields replace the original +8, +56,
// +496, and +504 ABI32 accesses.
func respawnInventoryClassNative4EDD00(
	owner *Object,
	classMask uint32,
	deps respawnInventoryClassNativeDeps4EDD00,
) {
	respawnInventoryClass4EDD00(respawnInventoryClassHooks4EDD00[
		*Object,
		*Object,
		types.Pointf,
	]{
		loadOwnerArg: func() *Object {
			return owner
		},
		loadInventoryHead: func(owner *Object) *Object {
			return owner.InvFirstItem
		},
		loadClassMaskArg: func() uint32 {
			return classMask
		},
		loadInventoryNext: func(item *Object) *Object {
			return item.InvNextItem
		},
		loadItemClass: func(item *Object) uint32 {
			return uint32(item.ObjClass)
		},
		detachInventory: deps.detachInventory,
		ownerPosition: func(owner *Object) *types.Pointf {
			return &owner.PosVec
		},
		randomReachable: deps.randomReachable,
		loadPointY: func(point *types.Pointf) float32 {
			return point.Y
		},
		loadPointX: func(point *types.Pointf) float32 {
			return point.X
		},
		createAt: func(item, owner *Object, x, y float32) {
			deps.createAt(item, owner, types.Pointf{X: x, Y: y})
		},
	})
}

type dropPlayerInventoryClassNativeDeps4EDD70 struct {
	firstPlayer     func() *Object
	nextPlayer      func(*Object) *Object
	randomReachable func(float32, *types.Pointf, *types.Pointf) *types.Pointf
	drop            func(*Object, *Object, *types.Pointf) int32
}

// dropPlayerInventoryClassNative4EDD70 binds both loops to native-width
// Object pointers and named inventory fields. No pointer is converted through
// int, uintptr, or a fixed 32-bit slot.
func dropPlayerInventoryClassNative4EDD70(
	deps dropPlayerInventoryClassNativeDeps4EDD70,
) {
	dropPlayerInventoryClass4EDD70(dropPlayerInventoryClassHooks4EDD70[
		*Object,
		*Object,
		types.Pointf,
	]{
		firstPlayer: deps.firstPlayer,
		loadInventoryHead: func(player *Object) *Object {
			return player.InvFirstItem
		},
		loadItemClass: func(item *Object) uint32 {
			return uint32(item.ObjClass)
		},
		loadInventoryNext: func(item *Object) *Object {
			return item.InvNextItem
		},
		playerPosition: func(player *Object) *types.Pointf {
			return &player.PosVec
		},
		randomReachable: deps.randomReachable,
		drop:            deps.drop,
		nextPlayer:      deps.nextPlayer,
	})
}
