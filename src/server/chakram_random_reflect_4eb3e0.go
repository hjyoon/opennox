package server

type chakramRandomReflectNativeDeps4EB3E0 struct {
	randomInt  func(int32, int32) int32
	moveUpdate func(*Object)
}

// chakramRandomReflectNative4EB3E0 preserves GAME.EXE 004EB3E0 with native
// Object pointers and the original PC53/toward-zero arithmetic. The velocity
// length and both final components are spilled to binary32. Both positions
// are reset from the live PrevPos before the movement update callback runs.
func chakramRandomReflectNative4EB3E0(obj *Object, deps chakramRandomReflectNativeDeps4EB3E0) {
	velocityX := obj.VelVec.X
	velocityY := obj.VelVec.Y
	x := float64(velocityX)
	y := float64(velocityY)
	mul, add := objectForceMul53_52DF80, objectForceAdd53_52DF80
	spill := experienceSpill32TowardZero
	square := add(mul(x, x), mul(y, y))
	speed := float64(spill(monsterMoveForceSqrtChop53_50D50C(square)))
	direction := directionFromVector509ED0(velocityX, velocityY)
	randomized := uint8(deps.randomInt(-64, 64) + direction + 128)
	cosine, sine := SinCosDir(randomized)
	obj.VelVec.X = spill(mul(speed, objectForceMemory32_52DF80(cosine)))
	obj.VelVec.Y = spill(mul(speed, objectForceMemory32_52DF80(sine)))
	obj.NewPos = obj.PrevPos
	obj.PosVec = obj.PrevPos
	deps.moveUpdate(obj)
}

// ChakramRandomReflect4EB3E0 uses the live game RNG for a Chakram rebound.
func (s *Server) ChakramRandomReflect4EB3E0(obj *Object, moveUpdate func(*Object)) {
	chakramRandomReflectNative4EB3E0(obj, chakramRandomReflectNativeDeps4EB3E0{
		randomInt: func(minimum, maximum int32) int32 {
			return int32(s.Rand.Logic.IntClamp(int(minimum), int(maximum)))
		},
		moveUpdate: moveUpdate,
	})
}
