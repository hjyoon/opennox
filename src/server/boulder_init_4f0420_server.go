package server

import "math"

// BoulderInit4F0420 binds GAME.EXE 004F0420 to the native-width Object while
// retaining each coordinate as an exact 32-bit payload. There is deliberately
// no nil guard.
//
//go:noinline
func BoulderInit4F0420(unit *Object) *Object {
	return boulderInit4F0420(unit, boulderInitHooks4F0420[*Object, uint32]{
		loadSourceX: func(unit *Object) uint32 {
			return math.Float32bits(unit.PosVec.X)
		},
		loadSourceY: func(unit *Object) uint32 {
			return math.Float32bits(unit.PosVec.Y)
		},
		storeTargetX: func(unit *Object, value uint32) {
			unit.Pos39.X = math.Float32frombits(value)
		},
		storeTargetY: func(unit *Object, value uint32) {
			unit.Pos39.Y = math.Float32frombits(value)
		},
	})
}
