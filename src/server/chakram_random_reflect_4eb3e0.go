package server

import "math"

type chakramRandomReflectNativeDeps4EB3E0 struct {
	randomInt  func(int32, int32) int32
	moveUpdate func(*Object)
}

// chakramRandomReflectNative4EB3E0 preserves GAME.EXE 004EB3E0 with native
// Object pointers. Velocity length is spilled to binary32 before multiplying
// by the original 256-entry direction table. The randomized direction wraps
// modulo 256, and both current positions are reset from PrevPos before the
// movement update callback runs.
func chakramRandomReflectNative4EB3E0(obj *Object, deps chakramRandomReflectNativeDeps4EB3E0) {
	velocityX := obj.VelVec.X
	velocityY := obj.VelVec.Y
	speed := float32(math.Sqrt(
		float64(velocityX)*float64(velocityX) + float64(velocityY)*float64(velocityY),
	))
	direction := directionFromVector509ED0(velocityX, velocityY)
	randomized := uint8(deps.randomInt(-64, 64) + direction + 128)
	cosine, sine := SinCosDir(randomized)
	obj.VelVec.X = float32(float64(speed) * float64(cosine))
	obj.VelVec.Y = float32(float64(speed) * float64(sine))
	obj.NewPos = obj.PrevPos
	obj.PosVec = obj.PrevPos
	deps.moveUpdate(obj)
}
