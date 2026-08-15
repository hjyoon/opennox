package server

import "github.com/opennox/libs/types"

// CrownCollideRuntime4EBB50 supplies CrownPickup, whose original 004F3400
// body and four-argument pickup ABI remain a separate restoration unit.
type CrownCollideRuntime4EBB50 struct {
	Pickup func(who, crown *Object, flag1, flag2 int32) uint32
}

type crownCollideNativeDeps4EBB50 struct {
	pickup func(*Object, *Object, int32, int32) uint32
}

func crownCollideNative4EBB50(
	crown, target *Object,
	collision *types.Pointf,
	deps crownCollideNativeDeps4EBB50,
) uintptr {
	result := crownCollide4EBB50(
		crown,
		target,
		collision,
		crownCollideHooks4EBB50[*Object]{
			loadFlags: func(obj *Object) uint32 {
				return uint32(obj.ObjFlags)
			},
			loadClassLow: func(obj *Object) uint8 {
				return uint8(obj.ObjClass)
			},
			pickup: deps.pickup,
		},
	)
	if result.pickupAttempted {
		return uintptr(result.pickupResult)
	}
	return uintptr(target.CObj())
}

// CrownCollide4EBB50 binds Crown collision to native-width Object pointers.
// Its uintptr result preserves both IA-32 EAX meanings: the untouched target
// pointer on a guard path, or CrownPickup's 32-bit result after an attempt.
func (s *Server) CrownCollide4EBB50(
	crown, target *Object,
	collision *types.Pointf,
	runtime CrownCollideRuntime4EBB50,
) uintptr {
	return crownCollideNative4EBB50(crown, target, collision, crownCollideNativeDeps4EBB50{
		pickup: runtime.Pickup,
	})
}
