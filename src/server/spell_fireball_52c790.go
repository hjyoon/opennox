package server

import "github.com/opennox/libs/types"

// SpellFireballRuntime52C790 supplies the original type-table/creation, trace,
// placement, balance and audio boundaries. Object fields remain native-width.
type SpellFireballRuntime52C790 struct {
	NewObject func(level int32) *Object
	Direction func(direction int16) types.Pointf
	Trace     func(*types.Rectf) bool
	Place     func(projectile, owner *Object, pos types.Pointf)
	Balance   func(index int32) float64
	Sound     func(spell int32) int32
	Audio     func(sound int32, owner *Object)
}

// CastFireball52C790 preserves the original PC53/ToZero intermediates and
// distinct binary32 stores, including values retained after nonpopping FST.
func CastFireball52C790(spellID int32, firing *Object, level int32, rt SpellFireballRuntime52C790) int32 {
	projectile := rt.NewObject(level)
	if projectile == nil {
		return 1
	}
	radius := float64(firing.Shape.Circle.R)
	twiceRadius := objectForceAdd53_52DF80(radius, radius)
	direction := rt.Direction(int16(firing.Direction1))
	ray := types.Rectf{Min: firing.PosVec}
	x := experienceSpill32TowardZero(objectForceAdd53_52DF80(
		objectForceMul53_52DF80(twiceRadius, objectForceMemory32_52DF80(direction.X)), objectForceMemory32_52DF80(firing.PosVec.X)))
	ray.Max.X = experienceSpill32TowardZero(objectForceAdd53_52DF80(float64(x), objectForceMemory32_52DF80(firing.VelVec.X)))
	y := objectForceAdd53_52DF80(
		objectForceMul53_52DF80(twiceRadius, objectForceMemory32_52DF80(direction.Y)), objectForceMemory32_52DF80(firing.PosVec.Y))
	ray.Max.Y = experienceSpill32TowardZero(objectForceAdd53_52DF80(y, objectForceMemory32_52DF80(firing.VelVec.Y)))
	if !rt.Trace(&ray) {
		ray.Max = ray.Min
	}
	rt.Place(projectile, firing, ray.Max)
	coefficient := rt.Balance(level - 1)
	speed := objectForceMul53_52DF80(coefficient, objectForceMemory32_52DF80(projectile.SpeedCur))
	// FST stores speed without popping the retained PC53 product.
	projectile.SpeedCur = experienceSpill32TowardZero(speed)
	xSpeed := experienceSpill32TowardZero(objectForceMul53_52DF80(speed, objectForceMemory32_52DF80(direction.X)))
	projectile.VelVec.X = xSpeed
	ySpeed := objectForceMul53_52DF80(speed, objectForceMemory32_52DF80(direction.Y))
	projectile.VelVec.Y = experienceSpill32TowardZero(ySpeed)
	projectile.VelVec.X = experienceSpill32TowardZero(objectForceAdd53_52DF80(float64(xSpeed), objectForceMemory32_52DF80(firing.VelVec.X)))
	projectile.VelVec.Y = experienceSpill32TowardZero(objectForceAdd53_52DF80(ySpeed, objectForceMemory32_52DF80(firing.VelVec.Y)))
	// Direction is reloaded after every world service and stored to both words.
	finalDirection := firing.Direction1
	projectile.Direction1, projectile.Direction2 = finalDirection, finalDirection
	sound := rt.Sound(spellID)
	rt.Audio(sound, firing)
	return 1
}
