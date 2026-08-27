package server

import (
	"unsafe"

	"github.com/opennox/libs/object"
)

// InventoryServPlaceRuntime4F36F0 supplies operations that remain owned by
// the root/legacy runtime. Object and callback pointers stay native-width.
type InventoryServPlaceRuntime4F36F0 struct {
	DefaultPickup  PickupDefaultRuntime4F31E0
	RefreshCollide func(*Object)
	ScriptPickup   func(*ScriptCallback, *Object, *Object)
}

type inventoryServPlaceNativeDeps4F36F0 struct {
	itemTypeAllowed func(uint16) int32
	callPickup      func(PickupFuncPtr, *Object, *Object, int32, int32) int32
	defaultPickup   func(*Object, *Object, int32, int32) int32
	refreshCollide  func(*Object)
	scriptPickup    func(*ScriptCallback, *Object, *Object)
}

func inventoryServPlaceNative4F36F0(
	owner, item *Object,
	arg3, arg4 int32,
	deps inventoryServPlaceNativeDeps4F36F0,
) int32 {
	return inventoryServPlace4F36F0(
		owner,
		item,
		inventoryServPlaceHooks4F36F0[*Object, PickupFuncPtr, unsafe.Pointer]{
			loadOwnerCarryCapacity: func(owner *Object) uint16 {
				return owner.CarryCapacity
			},
			loadItemFlagsLow: func(item *Object) uint8 {
				return uint8(item.ObjFlags)
			},
			loadOwnerFlags: func(owner *Object) uint32 {
				return uint32(owner.ObjFlags)
			},
			loadItemType: func(item *Object) uint16 {
				return item.TypeInd
			},
			itemTypeAllowed: deps.itemTypeAllowed,
			loadOwnerClassLow: func(owner *Object) uint8 {
				return uint8(owner.ObjClass)
			},
			loadPickup: func(item *Object) PickupFuncPtr {
				return item.Pickup
			},
			loadArg4: func() int32 {
				return arg4
			},
			loadArg3: func() int32 {
				return arg3
			},
			callPickup: deps.callPickup,
			defaultPickup: func(owner, item *Object, arg3, arg4 int32) int32 {
				return deps.defaultPickup(owner, item, arg3, arg4)
			},
			loadItemFlags: func(item *Object) uint32 {
				return uint32(item.ObjFlags)
			},
			storeItemFlags: func(item *Object, value uint32) {
				item.ObjFlags = object.Flags(value)
			},
			loadCollide: func(item *Object) unsafe.Pointer {
				return item.Collide
			},
			refreshCollide: deps.refreshCollide,
			loadScriptPickupFunc: func(item *Object) int32 {
				return item.ScriptPickup.Func
			},
			callScriptPickup: func(owner, item *Object) {
				deps.scriptPickup(&item.ScriptPickup, owner, item)
			},
			storeScriptPickupFunc: func(item *Object, value int32) {
				item.ScriptPickup.Func = value
			},
		},
	)
}

func inventoryServPlaceServerDeps4F36F0(
	s *Server,
	runtime InventoryServPlaceRuntime4F36F0,
) inventoryServPlaceNativeDeps4F36F0 {
	return inventoryServPlaceNativeDeps4F36F0{
		itemTypeAllowed: func(typeInd uint16) int32 {
			typ := s.Types.ByInd(int(typeInd))
			if typ == nil || !typ.Allowed() {
				return 0
			}
			return 1
		},
		callPickup: func(pickup PickupFuncPtr, owner, item *Object, arg3, arg4 int32) int32 {
			return pickup.CallInt32(owner, item, arg3, arg4)
		},
		defaultPickup: func(owner, item *Object, arg3, arg4 int32) int32 {
			return s.PickupDefault4F31E0(owner, item, arg3, arg4, runtime.DefaultPickup)
		},
		refreshCollide: runtime.RefreshCollide,
		scriptPickup:   runtime.ScriptPickup,
	}
}

// InventoryServPlace4F36F0 runs GAME.EXE 004F36F0 against native-width
// Object and callback layouts and returns the pickup callback's exact int32.
func (s *Server) InventoryServPlace4F36F0(
	owner, item *Object,
	arg3, arg4 int32,
	runtime InventoryServPlaceRuntime4F36F0,
) int32 {
	return inventoryServPlaceNative4F36F0(
		owner,
		item,
		arg3,
		arg4,
		inventoryServPlaceServerDeps4F36F0(s, runtime),
	)
}
