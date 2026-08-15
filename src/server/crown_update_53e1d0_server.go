package server

import "github.com/opennox/libs/types"

// CrownUpdateRuntime53E1D0 supplies the restored CrownPickup callback and the
// root object-position mutation that still updates broader map state.
type CrownUpdateRuntime53E1D0 struct {
	Pickup func(who, crown *Object, flag1, flag2 int32) uint32
	Move   func(crown *Object, destination types.Pointf)
}

type crownUpdateNativeDeps53E1D0 struct {
	pickup     func(*Object, *Object, int32, int32) uint32
	clearOwner func(*Object)
	trace      func(types.Pointf, types.Pointf, MapTraceFlags) bool
	move       func(*Object, types.Pointf)
}

func crownUpdateNative53E1D0(
	crown *Object,
	deps crownUpdateNativeDeps53E1D0,
) {
	crownUpdate53E1D0(
		crown,
		crownUpdateHooks53E1D0[*Object, *CrownUpdateData]{
			loadUpdate: func(obj *Object) *CrownUpdateData {
				return (*CrownUpdateData)(obj.UpdateData)
			},
			loadPickupTarget: func(update *CrownUpdateData) *Object {
				return update.PickupTarget
			},
			loadFlags: func(obj *Object) uint32 {
				return uint32(obj.ObjFlags)
			},
			pickup: deps.pickup,
			loadField0: func(update *CrownUpdateData) *Object {
				return update.Field0
			},
			loadFlagsLow: func(obj *Object) uint8 {
				return uint8(obj.ObjFlags)
			},
			clearField0: func(update *CrownUpdateData) {
				update.Field0 = nil
			},
			loadOwner: func(obj *Object) *Object {
				return obj.ObjOwner
			},
			clearOwner: deps.clearOwner,
			loadRadius: func(obj *Object) float32 {
				return obj.Shape.Circle.R
			},
			loadPosX: func(obj *Object) float32 {
				return obj.PosVec.X
			},
			loadPosY: func(obj *Object) float32 {
				return obj.PosVec.Y
			},
			loadDirection: func(obj *Object) int16 {
				return int16(obj.Direction1)
			},
			loadDirectionCos: func(direction int16) float32 {
				cosine, _ := SinCosDir(byte(direction))
				return cosine
			},
			loadDirectionSin: func(direction int16) float32 {
				_, sine := SinCosDir(byte(direction))
				return sine
			},
			trace: func(from, to types.Pointf, flags uint8) bool {
				return deps.trace(from, to, MapTraceFlags(flags))
			},
			move: deps.move,
		},
	)
}

func crownUpdateServerDeps53E1D0(
	s *Server,
	runtime CrownUpdateRuntime53E1D0,
) crownUpdateNativeDeps53E1D0 {
	return crownUpdateNativeDeps53E1D0{
		pickup:     runtime.Pickup,
		clearOwner: s.ObjClearOwner,
		trace:      s.MapTraceRay,
		move:       runtime.Move,
	}
}

// CrownUpdate53E1D0 binds the original three-word Crown update record to
// native object pointers on 32- and 64-bit targets.
func (s *Server) CrownUpdate53E1D0(
	crown *Object,
	runtime CrownUpdateRuntime53E1D0,
) {
	crownUpdateNative53E1D0(crown, crownUpdateServerDeps53E1D0(s, runtime))
}
