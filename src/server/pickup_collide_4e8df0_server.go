package server

import "unsafe"

type pickupCollideNativeDeps4E8DF0 struct {
	frame          func() uint32
	fps            func() uint32
	placeInventory func(*Object, *Object, int, int) bool
}

// PickupCollideRuntime4E8DF0 supplies the inventory transition still owned by
// the legacy object subsystem.
type PickupCollideRuntime4E8DF0 struct {
	PlaceInventory func(unit, item *Object, flag1, flag2 int) bool
}

func pickupCollideNative4E8DF0(
	item, unit *Object,
	collision unsafe.Pointer,
	deps pickupCollideNativeDeps4E8DF0,
) uintptr {
	result := pickupCollide4E8DF0(
		item,
		unit,
		collision,
		pickupCollideHooks4E8DF0[*Object, *PlayerUpdateData]{
			loadClassByte: func(obj *Object) uint8 {
				return uint8(obj.ObjClass)
			},
			frame: deps.frame,
			loadPickupFrame: func(obj *Object) uint32 {
				return obj.Field32
			},
			fps: deps.fps,
			loadUpdateData: func(obj *Object) *PlayerUpdateData {
				// The original uses the class byte cached at entry and performs
				// one direct update-data pointer load here; do not call
				// UpdateDataPlayer, which would reload the live class.
				return (*PlayerUpdateData)(obj.UpdateData)
			},
			loadMovementFlagsByte: func(update *PlayerUpdateData) uint8 {
				return uint8(update.MovementFlags)
			},
			placeInventory: func(unit, item *Object, flag1, flag2 int32) uint32 {
				if deps.placeInventory(unit, item, int(flag1), int(flag2)) {
					return 1
				}
				return 0
			},
		},
	)
	if result.inventoryAttempted {
		return uintptr(result.inventoryResult)
	}
	return uintptr(unit.CObj())
}

// PickupCollide4E8DF0 binds Pickup collision to native-width Object pointers.
// Its uintptr result preserves both possible IA-32 EAX meanings: the untouched
// unit pointer on a guard path, or the inventory routine's zero/one result.
func (s *Server) PickupCollide4E8DF0(
	item, unit *Object,
	collision unsafe.Pointer,
	runtime PickupCollideRuntime4E8DF0,
) uintptr {
	return pickupCollideNative4E8DF0(item, unit, collision, pickupCollideNativeDeps4E8DF0{
		frame:          s.Frame,
		fps:            s.TickRate,
		placeInventory: runtime.PlaceInventory,
	})
}
