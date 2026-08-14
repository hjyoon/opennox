package server

import "math"

const (
	pixieTargetClassMask4EA080   = uint32(0x00020006)
	pixiePlayerClass4EA080       = uint32(0x00000004)
	pixieRejectFlags4EA080       = uint32(0x00008020)
	pixieOwnerSuppressFlag4EA080 = uint32(0x00000002)
	pixieReflectEnchant4EA080    = uint32(27)
	pixieDamageAudio4EA080       = uint32(96)
	pixieReflectAudio4EA080      = uint32(122)
	pixieDamageType4EA080        = uint32(11)
	pixieGridInverseBits4EA080   = uint32(0x3d321643)
)

type pixieCollideHooks4EA080[O, C comparable, D any] struct {
	loadCollideData func(O) D
	isEnemy         func(O, O) int32
	loadClass       func(O) uint32
	loadFlags       func(O) uint32
	loadOwner       func(O) O
	checkInversion  func(O, O) int32
	changeOwner     func(O, O)
	hasEnchant      func(O, uint32) int32
	loadDirection   func(O) int16
	checkDirection  func(O, int16, O) int32
	loadDamage      func(D) int32
	findParent      func(O) O
	targetDamage    func(O, O, O, int32, uint32) int32
	audio           func(uint32, O)
	delayedDelete   func(O)
	wallReflect     func(C, O)
	vectorDirection func(O) int32
	loadVelocityX   func(O) float32
	loadVelocityY   func(O) float32
	loadNewPosX     func(O) float32
	loadNewPosY     func(O) float32
	storeDirection2 func(O, uint16)
	storeNewPosX    func(O, float32)
	storeNewPosY    func(O, float32)
	floatToInt      func(float32) int32
	damageMap       func(int32, int32, int32, uint32, O)
}

// pixieCollide4EA080 preserves GAME.EXE 004EA080. The eight-byte collide-data
// pointer is cached before either collision branch. Target handling caches the
// class word after the enemy callback, retains the original silent rejection
// paths, and reads damage before the live parent lookup and Damage callback.
// The wall path keeps the unspilled x87 Y sum through the grid multiplication,
// while X is rounded to its stored binary32 value and reloaded after the first
// coordinate conversion.
func pixieCollide4EA080[O, C comparable, D any](
	source, target O,
	collision C,
	hooks pixieCollideHooks4EA080[O, C, D],
) {
	data := hooks.loadCollideData(source)

	var zeroObject O
	if target != zeroObject {
		if hooks.isEnemy(source, target) == 0 {
			return
		}
		targetClass := hooks.loadClass(target)
		if targetClass&pixieTargetClassMask4EA080 == 0 {
			return
		}
		if hooks.loadFlags(target)&pixieRejectFlags4EA080 != 0 {
			return
		}

		owner := hooks.loadOwner(source)
		if owner != zeroObject &&
			hooks.loadClass(owner)&pixiePlayerClass4EA080 == pixiePlayerClass4EA080 &&
			hooks.loadFlags(owner)&pixieOwnerSuppressFlag4EA080 == pixieOwnerSuppressFlag4EA080 {
			return
		}

		if targetClass&pixiePlayerClass4EA080 != 0 {
			if hooks.checkInversion(target, source) != 0 {
				hooks.changeOwner(source, target)
				return
			}
			if hooks.hasEnchant(target, pixieReflectEnchant4EA080) != 0 {
				direction := hooks.loadDirection(target)
				if hooks.checkDirection(target, direction, source)&1 != 0 {
					hooks.changeOwner(source, target)
					hooks.audio(pixieReflectAudio4EA080, target)
					return
				}
			}
		}

		damage := hooks.loadDamage(data)
		parent := hooks.findParent(source)
		_ = hooks.targetDamage(target, parent, source, damage, pixieDamageType4EA080)
		hooks.audio(pixieDamageAudio4EA080, source)
		hooks.delayedDelete(source)
		return
	}

	var zeroCollision C
	if collision == zeroCollision {
		hooks.delayedDelete(source)
		return
	}

	hooks.wallReflect(collision, source)
	direction := hooks.vectorDirection(source)

	velocityX := hooks.loadVelocityX(source)
	oldNewX := hooks.loadNewPosX(source)
	newXExtended := float64(velocityX) + float64(oldNewX)
	newX := float32(newXExtended)
	hooks.storeDirection2(source, uint16(direction))
	hooks.storeNewPosX(source, newX)

	velocityY := hooks.loadVelocityY(source)
	oldNewY := hooks.loadNewPosY(source)
	newYExtended := float64(velocityY) + float64(oldNewY)
	hooks.storeNewPosY(source, float32(newYExtended))
	damage := hooks.loadDamage(data)

	gridInverse := float64(math.Float32frombits(pixieGridInverseBits4EA080))
	gridY := hooks.floatToInt(float32(newYExtended * gridInverse))
	liveNewX := hooks.loadNewPosX(source)
	gridX := hooks.floatToInt(float32(float64(liveNewX) * gridInverse))
	hooks.damageMap(gridX, gridY, damage, pixieDamageType4EA080, source)
}
