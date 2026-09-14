package server

import "github.com/opennox/libs/types"

// ProjectileTrailRuntime53AEC0 keeps random sparks on the native object path.
type ProjectileTrailRuntime53AEC0 struct {
	RandomFloat func(float32, float32) float32
	CreateSpark func(types.Pointf, int, int, types.Pointf, float32, *Object)
}

// ProjectileTrailUpdate53AEC0 restores GAME.EXE 0053AEC0 without passing the
// projectile as a 32-bit integer owner to the legacy Spark factory.
func ProjectileTrailUpdate53AEC0(projectile *Object, runtime ProjectileTrailRuntime53AEC0) {
	if projectile == nil {
		return
	}
	cosine, sine := SinCosDir(byte(projectile.Direction1))
	projectile.ForceVec.X += cosine * projectile.SpeedCur * 0.25
	projectile.ForceVec.Y += sine * projectile.SpeedCur * 0.25
	step := projectile.PosVec.Sub(projectile.PrevPos).Mul(0.25)
	for i := 0; i < 4; i++ {
		base := projectile.PrevPos.Add(step.Mul(float32(i)))
		base.X = float32(int32(base.X))
		base.Y = float32(int32(base.Y))
		for j := 0; j < 2; j++ {
			velocityY := runtime.RandomFloat(-2, 2)
			velocityX := runtime.RandomFloat(-2, 2)
			posY := base.Y + runtime.RandomFloat(-4, 4)
			posX := base.X + runtime.RandomFloat(-4, 4)
			runtime.CreateSpark(types.Ptf(posX, posY), 1, 6, types.Ptf(velocityX, velocityY), 0, projectile)
		}
	}
}
