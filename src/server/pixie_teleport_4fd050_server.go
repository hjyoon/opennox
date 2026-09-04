package server

import "math"

type pixieTeleportNativeDeps4FD050 struct {
	moveUpdate func(*Object)
}

func pixieTeleportNative4FD050(
	pixie, owner *Object,
	deps pixieTeleportNativeDeps4FD050,
) {
	pixieTeleport4FD050(pixieTeleportHooks4FD050[*Object]{
		loadOwnerArg: func() *Object {
			return owner
		},
		loadPixieArg: func() *Object {
			return pixie
		},
		loadOwnerXBits: func(owner *Object) uint32 {
			return math.Float32bits(owner.PosVec.X)
		},
		loadOwnerYBits: func(owner *Object) uint32 {
			return math.Float32bits(owner.PosVec.Y)
		},
		storeNewPosXBits: func(pixie *Object, value uint32) {
			pixie.NewPos.X = math.Float32frombits(value)
		},
		storeNewPosYBits: func(pixie *Object, value uint32) {
			pixie.NewPos.Y = math.Float32frombits(value)
		},
		storePosXBits: func(pixie *Object, value uint32) {
			pixie.PosVec.X = math.Float32frombits(value)
		},
		storePosYBits: func(pixie *Object, value uint32) {
			pixie.PosVec.Y = math.Float32frombits(value)
		},
		storePrevPosXBits: func(pixie *Object, value uint32) {
			pixie.PrevPos.X = math.Float32frombits(value)
		},
		storePrevPosYBits: func(pixie *Object, value uint32) {
			pixie.PrevPos.Y = math.Float32frombits(value)
		},
		moveUpdate: deps.moveUpdate,
	})
}

// PixieTeleport4FD050 binds GAME.EXE 004FD050 to native-width Object
// pointers. Coordinates are copied as raw float32 bits in the original store
// order before the already-restored 00517970 move-update service is invoked.
//
//go:noinline
func (*Server) PixieTeleport4FD050(
	pixie, owner *Object,
	moveUpdate func(*Object),
) {
	pixieTeleportNative4FD050(pixie, owner, pixieTeleportNativeDeps4FD050{
		moveUpdate: moveUpdate,
	})
}
