package server

import "math"

const (
	playerCollideBerserkAbility4E8460      = uint32(1)
	playerCollideState4E8460               = uint32(13)
	playerCollideUnitClassMask4E8460       = uint32(0x00000006)
	playerCollidePlayerClass4E8460         = uint32(0x00000004)
	playerCollideDoorClass4E8460           = uint32(0x00000080)
	playerCollideMassBypassClass4E8460     = uint32(0x00400000)
	playerCollideMoveAfterDamage4E8460     = uint32(0x00020006)
	playerCollideDeadFlag4E8460            = uint32(0x00008000)
	playerCollideMassBypassFlagLow4E8460   = uint8(0x80)
	playerCollideRejectFlagLow4E8460       = uint8(0x09)
	playerCollideRejectClassLow4E8460      = uint8(0x01)
	playerCollideHeldEnchant4E8460         = uint32(5)
	playerCollideDeathEnchant4E8460        = uint32(16)
	playerCollideBerserkerCrashSound4E8460 = uint32(171)
	playerCollideMapDamage4E8460           = int32(100)
	playerCollideMapDamageType4E8460       = uint32(2)
	playerCollideMeleeDamageType4E8460     = uint32(2)
	playerCollideGridInverseBits4E8460     = uint32(0x3d321643)
)

type playerCollideHooks4E8460[O, H, W comparable, P any] struct {
	abilityActive  func(O, uint32) int32
	class          func(O) uint32
	flags          func(O) uint32
	flagsLow       func(O) uint8
	health         func(O) H
	healthCurrent  func(H) uint16
	healthMax      func(H) uint16
	mass           func(O) float32
	doorState      func(O) uint8
	setState       func(O, uint32)
	earthquake     func(O, int32)
	disableAbility func(O, uint32)
	balanceFloat   func(string) float64
	floatToInt     func(float32) int32
	bounce         func(O, O)
	findParent     func(O) O
	damage         func(O, O, O, int32, uint32)
	applyEnchant   func(O, uint32, uint32, uint32)
	collisionWall  func(O) W
	wallTile       func(W) uint8
	wallFlags      func(uint8) uint32
	audio          func(uint32, O, int32, uint32)
	newPosY        func(O) float32
	newPosX        func(O) float32
	damageMap      func(int32, int32, int32, uint32, O)
	damageClear    func(O, int32)
	move           func(O)
	hasEnchant     func(O, uint32) int32
	enchantTimer   func(O, uint32) uint32
	frameRate      func() uint32
	enchantPower   func(O, uint32) uint32
	disableEnchant func(O, uint32)
}

// playerCollide4E8460 preserves GAME.EXE 004E8460. collision is part of the
// registered three-argument collide ABI but the original body never reads it.
// All object and callback reads remain ordered so callback-side mutation is
// visible at the same later reload points as in the original x86 routine.
func playerCollide4E8460[O, H, W comparable, P any](
	player, other O,
	collision P,
	hooks playerCollideHooks4E8460[O, H, W, P],
) {
	_ = collision
	var zeroObject O
	var zeroWall W

	if hooks.abilityActive(player, playerCollideBerserkAbility4E8460) != 0 {
		hit := other == zeroObject
		if !hit {
			class := hooks.class(other)
			checkMass := class&playerCollideUnitClassMask4E8460 == 0
			if !checkMass {
				health := hooks.health(other)
				if hooks.healthCurrent(health) == 0 && hooks.healthMax(health) != 0 {
					checkMass = true
				}
			}
			if checkMass && class&playerCollideMassBypassClass4E8460 == 0 &&
				hooks.flagsLow(other)&playerCollideMassBypassFlagLow4E8460 == 0 {
				otherMass := hooks.mass(other)
				playerMass := hooks.mass(player)
				// x87 tests C0|C3 after comparing other against player. Only
				// an ordered, strictly greater value proceeds; NaN is rejected.
				if !(otherMass > playerMass) {
					goto transferDeathEnchant
				}
			}

			if class&playerCollideDoorClass4E8460 != 0 {
				if hooks.doorState(other) == 0 {
					goto transferDeathEnchant
				}
				hit = true
			} else {
				if hooks.flagsLow(other)&playerCollideRejectFlagLow4E8460 != 0 ||
					uint8(class)&playerCollideRejectClassLow4E8460 != 0 {
					goto transferDeathEnchant
				}
				hit = true
			}
		}

		if hit {
			hooks.setState(player, playerCollideState4E8460)
			hooks.earthquake(player, 10)
			hooks.disableAbility(player, playerCollideBerserkAbility4E8460)

			if other != zeroObject {
				damage := hooks.floatToInt(float32(hooks.balanceFloat("BerserkerDamage")))
				if hooks.class(other)&playerCollideMassBypassClass4E8460 == 0 {
					hooks.bounce(player, other)
				}
				parent := hooks.findParent(player)
				hooks.damage(other, parent, player, damage, playerCollideMeleeDamageType4E8460)
				if hooks.class(other)&playerCollideMoveAfterDamage4E8460 != 0 {
					hooks.move(player)
					goto transferDeathEnchant
				}
				duration := hooks.floatToInt(float32(hooks.balanceFloat("BerserkerStunDuration")))
				hooks.applyEnchant(player, playerCollideHeldEnchant4E8460, uint32(uint16(duration)), 5)
			} else {
				wall := hooks.collisionWall(player)
				if wall != zeroWall {
					tile := hooks.wallTile(wall)
					if hooks.wallFlags(tile)&5 == 0 {
						hooks.move(player)
						goto transferDeathEnchant
					}
				}
				hooks.audio(playerCollideBerserkerCrashSound4E8460, player, 0, 0)
				duration := hooks.floatToInt(float32(hooks.balanceFloat("BerserkerStunDuration")))
				hooks.applyEnchant(player, playerCollideHeldEnchant4E8460, uint32(uint16(duration)), 5)

				gridInverse := math.Float32frombits(playerCollideGridInverseBits4E8460)
				y := hooks.floatToInt(hooks.newPosY(player) * gridInverse)
				x := hooks.floatToInt(hooks.newPosX(player) * gridInverse)
				hooks.damageMap(x, y, playerCollideMapDamage4E8460, playerCollideMapDamageType4E8460, player)
			}

			ratio := hooks.balanceFloat("BerserkerPainRatio")
			health := hooks.health(player)
			pain := hooks.floatToInt(float32(ratio * float64(hooks.healthCurrent(health))))
			if pain < 1 {
				pain = 1
			}
			hooks.damageClear(player, pain)
			hooks.move(player)
		}
	}

transferDeathEnchant:
	if other == zeroObject || uint8(hooks.class(other))&uint8(playerCollidePlayerClass4E8460) == 0 {
		return
	}
	if hooks.flags(other)&playerCollideDeadFlag4E8460 != 0 {
		return
	}
	if hooks.hasEnchant(player, playerCollideDeathEnchant4E8460) == 0 {
		return
	}
	timer := hooks.enchantTimer(player, playerCollideDeathEnchant4E8460)
	firstRate := hooks.frameRate()
	if timer >= firstRate*14 {
		return
	}
	power := hooks.enchantPower(player, playerCollideDeathEnchant4E8460) & 0xff
	secondRate := hooks.frameRate()
	hooks.applyEnchant(other, playerCollideDeathEnchant4E8460, uint32(uint16(secondRate*15)), power)
	hooks.disableEnchant(player, playerCollideDeathEnchant4E8460)
}

// playerCollideRound4E8460 models the default x87 FISTP rounding mode used by
// nox_float2int. Invalid and out-of-range inputs produce the integer-indefinite
// value 0x80000000.
func playerCollideRound4E8460(value float32) int32 {
	if math.IsNaN(float64(value)) || value >= 2147483648 || value < -2147483648 {
		return math.MinInt32
	}
	return int32(math.RoundToEven(float64(value)))
}
