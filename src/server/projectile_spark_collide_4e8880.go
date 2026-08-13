package server

import "math"

const (
	projectileSparkDamageType4E8880      = uint32(11)
	projectileSparkGridInverseBits4E8880 = uint32(0x3d321643)
)

type projectileSparkCollideHooks4E8880[O comparable, C any] struct {
	loadCollideData func(O) C
	loadDamage      func(C) int32
	findParent      func(O) O
	damage          func(O, O, O, int32, uint32) uint8
	loadNewPosY     func(O) float32
	loadNewPosX     func(O) float32
	floatToInt      func(float32) int32
	damageMap       func(int32, int32, int32, uint32, O)
	delayedDelete   func(O)
}

// projectileSparkCollide4E8880 preserves GAME.EXE 004E8880. The collision-
// data pointer is loaded before the target branch and retained across every
// later callback. The registered third collision argument is not read.
func projectileSparkCollide4E8880[O comparable, C, D any](
	projectile, other O,
	collision D,
	hooks projectileSparkCollideHooks4E8880[O, C],
) {
	_ = collision
	collideData := hooks.loadCollideData(projectile)

	var zeroObject O
	if other != zeroObject {
		damage := hooks.loadDamage(collideData)
		parent := hooks.findParent(projectile)
		if hooks.damage(other, parent, projectile, damage, projectileSparkDamageType4E8880) != 0 {
			hooks.delayedDelete(projectile)
		}
		return
	}

	y := hooks.loadNewPosY(projectile)
	damage := hooks.loadDamage(collideData)
	gridInverse := math.Float32frombits(projectileSparkGridInverseBits4E8880)
	gridY := hooks.floatToInt(y * gridInverse)
	x := hooks.loadNewPosX(projectile)
	gridX := hooks.floatToInt(x * gridInverse)
	hooks.damageMap(gridX, gridY, damage, projectileSparkDamageType4E8880, projectile)
	hooks.delayedDelete(projectile)
}
