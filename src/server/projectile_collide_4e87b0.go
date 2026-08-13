package server

const (
	projectileCollideThrowingStoneType4E87B0 = "ThrowingStone"
	projectileCollideImpShotType4E87B0       = "ImpShot"
	projectileCollideUrchinDamage4E87B0      = "UrchinStoneDamage"
	projectileCollideImpShotDamage4E87B0     = "ImpShotDamage"
	projectileCollideDamageType4E87B0        = uint32(11)
)

type projectileCollideHooks4E87B0[O, D, P comparable, C any] struct {
	loadCollideData       func(O) C
	loadThrowingStoneType func() uint32
	lookupType            func(string) uint32
	storeThrowingStone    func(uint32)
	storeImpShot          func(uint32)
	loadType              func(O) uint16
	loadImpShotType       func() uint32
	gameDataFloat         func(string) float64
	floatToInt            func(float32) int32
	loadDamage            func(C) int32
	findParentPlayer      func(O) O
	damage                func(O, O, O, int32, uint32) uint8
	traceHitPoint         func() P
	loadPointY            func(P) int32
	loadPointX            func(P) int32
	damageMap             func(int32, int32, int32, uint32, O)
	delayedDelete         func(O)
}

// projectileCollide4E87B0 preserves GAME.EXE 004E87B0. The initial
// ThrowingStone cache word is sampled before the collision-data pointer, and
// that pointer is cached before the sampled cache is tested or either name is
// looked up. Its damage word is read only for projectile types without a
// balance override. The registered third collision argument is deliberately
// not read.
func projectileCollide4E87B0[O, D, P comparable, C any](
	projectile, other O,
	collision D,
	hooks projectileCollideHooks4E87B0[O, D, P, C],
) {
	_ = collision
	throwingStoneType := hooks.loadThrowingStoneType()
	collideData := hooks.loadCollideData(projectile)
	if throwingStoneType == 0 {
		throwingStoneType = hooks.lookupType(projectileCollideThrowingStoneType4E87B0)
		hooks.storeThrowingStone(throwingStoneType)
		impShotType := hooks.lookupType(projectileCollideImpShotType4E87B0)
		hooks.storeImpShot(impShotType)
	}

	throwingStoneType = hooks.loadThrowingStoneType()
	typeIndex := uint32(hooks.loadType(projectile))
	var damage int32
	if typeIndex == throwingStoneType {
		damage = hooks.floatToInt(float32(hooks.gameDataFloat(projectileCollideUrchinDamage4E87B0)))
	} else if typeIndex == hooks.loadImpShotType() {
		damage = hooks.floatToInt(float32(hooks.gameDataFloat(projectileCollideImpShotDamage4E87B0)))
	} else {
		damage = hooks.loadDamage(collideData)
	}

	var zeroObject O
	if other != zeroObject {
		parent := hooks.findParentPlayer(projectile)
		if hooks.damage(other, parent, projectile, damage, projectileCollideDamageType4E87B0) == 0 {
			return
		}
		hooks.delayedDelete(projectile)
		return
	}

	point := hooks.traceHitPoint()
	var zeroPoint P
	if point != zeroPoint {
		y := hooks.loadPointY(point)
		x := hooks.loadPointX(point)
		hooks.damageMap(x, y, damage, projectileCollideDamageType4E87B0, projectile)
	}
	hooks.delayedDelete(projectile)
}
