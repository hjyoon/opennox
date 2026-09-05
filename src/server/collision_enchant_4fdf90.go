package server

const (
	collisionEnchantInvisible4FDF90          = uint32(0)
	collisionEnchantShock4FDF90              = uint32(22)
	collisionEnchantShockRejectFlags4FDF90   = uint32(0x00008008)
	collisionEnchantInactiveFlags4FDF90      = uint32(0x00008020)
	collisionEnchantUnitClassLow4FDF90       = uint8(0x06)
	collisionEnchantPlayerClassLow4FDF90     = uint8(0x04)
	collisionEnchantUnitOrWallClass4FDF90    = uint32(0x00020006)
	collisionEnchantWallClass4FDF90          = uint32(0x00020000)
	collisionEnchantShockAudio4FDF90         = int32(135)
	collisionEnchantShockDamageType4FDF90    = uint32(9)
	collisionEnchantShockDamageBalance4FDF90 = "ShockDamage"
)

//go:noinline
func collisionEnchantSpillFloat32_4FDF90(value float64) float32 {
	return float32(value)
}

type collisionEnchantHooks4FDF90[O any] struct {
	loadSourceArg func() O
	hasEnchant    func(O, uint32) int32
	loadTargetArg func() O

	loadTargetFlags    func(O) uint32
	loadTargetClassLow func(O) uint8
	isEnemy            func(O, O) int32
	loadShockPower     func(O) uint8
	audio              func(int32, O, int32, uint32)
	disableEnchant     func(O, uint32)
	balanceFloatTable  func(string, int32) float64
	floatToInt         func(float32) int32
	callTargetDamage   func(O, O, O, int32, uint32) int32

	loadTargetClass    func(O) uint32
	unitsOnSameTeam    func(O, O) int32
	loadSourceClassLow func(O) uint8
}

// collisionEnchant4FDF90 preserves GAME.EXE 004FDF90's three consecutive
// collision-enchantment phases. The source argument is cached before the
// Shock predicate, while the target argument is deliberately captured after
// that callback even when the predicate returns zero.
//
// The Shock power byte is zero-extended before subtracting one. Shock audio
// and removal precede the indexed balance lookup, whose binary64 result is
// explicitly spilled to binary32 before the x87-compatible integer
// conversion. The target damage callback is always dispatched; the original
// indirect call has no nil-function guard.
//
// Target class and flags are loaded live in the later hostile-unit and
// Player-versus-wall phases. Consequently invisibility may be removed twice,
// and callback-side mutation remains visible at every later load.
func collisionEnchant4FDF90[O any](hooks collisionEnchantHooks4FDF90[O]) {
	source := hooks.loadSourceArg()
	hasShock := hooks.hasEnchant(source, collisionEnchantShock4FDF90)
	target := hooks.loadTargetArg()

	if hasShock != 0 &&
		hooks.loadTargetFlags(target)&collisionEnchantShockRejectFlags4FDF90 == 0 &&
		hooks.loadTargetClassLow(target)&collisionEnchantUnitClassLow4FDF90 != 0 &&
		hooks.isEnemy(source, target) != 0 {
		power := int32(hooks.loadShockPower(source)) - 1
		hooks.audio(collisionEnchantShockAudio4FDF90, source, 0, 0)
		hooks.disableEnchant(source, collisionEnchantShock4FDF90)
		damageFloat := collisionEnchantSpillFloat32_4FDF90(
			hooks.balanceFloatTable(collisionEnchantShockDamageBalance4FDF90, power),
		)
		damage := hooks.floatToInt(damageFloat)
		_ = hooks.callTargetDamage(
			target,
			source,
			source,
			damage,
			collisionEnchantShockDamageType4FDF90,
		)
	}

	targetClass := hooks.loadTargetClass(target)
	if targetClass&collisionEnchantUnitOrWallClass4FDF90 != 0 &&
		hooks.loadTargetFlags(target)&collisionEnchantInactiveFlags4FDF90 == 0 &&
		hooks.unitsOnSameTeam(target, source) == 0 {
		hooks.disableEnchant(source, collisionEnchantInvisible4FDF90)
	}

	if hooks.loadSourceClassLow(source)&collisionEnchantPlayerClassLow4FDF90 != 0 &&
		hooks.loadTargetClass(target)&collisionEnchantWallClass4FDF90 != 0 &&
		hooks.loadTargetFlags(target)&collisionEnchantInactiveFlags4FDF90 == 0 {
		hooks.disableEnchant(source, collisionEnchantInvisible4FDF90)
	}
}
