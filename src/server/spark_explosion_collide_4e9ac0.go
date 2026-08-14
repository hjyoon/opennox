package server

import "math"

const (
	sparkExplosionQuestFlag4E9AC0       = uint32(0x1000)
	sparkExplosionCoopFlag4E9AC0        = uint32(0x800)
	sparkExplosionReflectEnchant4E9AC0  = uint32(27)
	sparkExplosionPlayerClass4E9AC0     = uint8(0x4)
	sparkExplosionDamageType4E9AC0      = uint32(1)
	sparkExplosionReflectAudio4E9AC0    = uint32(122)
	sparkExplosionDetonateAudio4E9AC0   = uint32(42)
	sparkExplosionScorchType4E9AC0      = int32(2)
	sparkExplosionTwoThirdsBits4E9AC0   = uint32(0x3f2aaaab)
	sparkExplosionOneThirdBits4E9AC0    = uint32(0x3eaaaaab)
	sparkExplosionPushForceBits4E9AC0   = uint32(0x42480000)
	sparkExplosionInnerRadiusBits4E9AC0 = uint32(0x41700000)
)

type sparkExplosionCollideHooks4E9AC0[O comparable, D any] struct {
	loadCollideData func(O) D
	loadPower       func(D) uint8
	hasEnchant      func(O, uint32) int32
	loadDirection   func(O) int16
	checkDirection  func(O, int16, O) int32
	reflect         func(O, O)
	clearOwner      func(O)
	setOwner        func(O, O)
	audio           func(uint32, O, int32, uint32)
	gameFlagsCheck  func(uint32) int32
	findParent      func(O) O
	classLow        func(O) uint8
	isEnemy         func(O, O) int32
	mapPushUnits    func(O, float32, float32, float32, O, int32, int32)
	targetDamage    func(O, O, O, int32, uint32) int32
	mapDamageUnits  func(O, float32, float32, int32, uint32, O, O)
	sparkFX         func(O, uint8)
	scorch          func(O, int32)
	delayedDelete   func(O)
}

// sparkExplosionCollide4E9AC0 preserves GAME.EXE 004E9AC0. The one-byte
// collide-data pointer is cached before target inspection, but its power byte
// is reloaded at each original instruction site. Reflection still flows
// through the Quest-friendly gate before the detonation flag is tested. The
// registered third collision argument is not read.
func sparkExplosionCollide4E9AC0[O comparable, D any](
	source, target O,
	collision any,
	hooks sparkExplosionCollideHooks4E9AC0[O, D],
) {
	_ = collision
	data := hooks.loadCollideData(source)
	detonate := true

	var zero O
	if target != zero && hooks.hasEnchant(target, sparkExplosionReflectEnchant4E9AC0) != 0 {
		direction := hooks.loadDirection(target)
		if hooks.checkDirection(target, direction, source)&1 != 0 {
			hooks.reflect(source, target)
			hooks.clearOwner(source)
			hooks.setOwner(target, source)
			detonate = false
			hooks.audio(sparkExplosionReflectAudio4E9AC0, target, 0, 0)
		}
	}

	if hooks.gameFlagsCheck(sparkExplosionQuestFlag4E9AC0) != 0 {
		parent := hooks.findParent(source)
		if parent != zero &&
			target != zero &&
			hooks.classLow(parent)&sparkExplosionPlayerClass4E9AC0 != 0 &&
			hooks.classLow(target)&sparkExplosionPlayerClass4E9AC0 != 0 &&
			hooks.isEnemy(parent, target) == 0 {
			return
		}
	}
	if !detonate {
		return
	}

	power := hooks.loadPower(data)
	hooks.mapPushUnits(
		source,
		float32(power)*math.Float32frombits(sparkExplosionTwoThirdsBits4E9AC0),
		0,
		math.Float32frombits(sparkExplosionPushForceBits4E9AC0),
		source,
		0,
		0,
	)

	if target != zero {
		power = hooks.loadPower(data)
		damage := int32(power >> 1)
		parent := hooks.findParent(source)
		_ = hooks.targetDamage(
			target,
			parent,
			source,
			damage,
			sparkExplosionDamageType4E9AC0,
		)
	}

	coop := hooks.gameFlagsCheck(sparkExplosionCoopFlag4E9AC0)
	power = hooks.loadPower(data)
	inner := math.Float32frombits(sparkExplosionInnerRadiusBits4E9AC0)
	if coop != 0 {
		inner = 0
	}
	hooks.mapDamageUnits(
		source,
		float32(power)*math.Float32frombits(sparkExplosionOneThirdBits4E9AC0),
		inner,
		int32(power>>1),
		sparkExplosionDamageType4E9AC0,
		source,
		target,
	)

	power = hooks.loadPower(data)
	hooks.sparkFX(source, power)
	hooks.audio(sparkExplosionDetonateAudio4E9AC0, source, 0, 0)
	hooks.scorch(source, sparkExplosionScorchType4E9AC0)
	hooks.delayedDelete(source)
}
