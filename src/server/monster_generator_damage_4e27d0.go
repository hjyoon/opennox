package server

import (
	"math"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
)

const (
	monsterGeneratorDamageRecentFrames4E27D0 = uint32(20)
	monsterGeneratorDamageFxPeriod4E27D0     = uint32(30)
	monsterGeneratorDamageFxOpcode4E27D0     = byte(0xf0)
	monsterGeneratorDamageFxSubtype4E27D0    = byte(26)
	monsterGeneratorDamageSound4E27D0        = 1001
	monsterGeneratorDamageStatusMid4E27D0    = uint32(0x100)
	monsterGeneratorDamageStatusLow4E27D0    = uint32(0x200)

	monsterGeneratorDamageEpsilonBits4E27D0  = uint32(0x3c23d70a) // 0.0099999998
	monsterGeneratorDamageThirdBits4E27D0    = uint32(0x3eaa7efa) // 0.333
	monsterGeneratorDamageTwoThirdBits4E27D0 = uint32(0x3f2a7efa) // 0.66600001
)

// Keep each original x87 instruction at a binary64 boundary. Nox uses
// 53-bit precision control, and each original FSTP is represented explicitly.
//
//go:noinline
func monsterGeneratorDamageAdd64_4E27D0(a, b float64) float64 { return a + b }

//go:noinline
func monsterGeneratorDamageSub64_4E27D0(a, b float64) float64 { return a - b }

//go:noinline
func monsterGeneratorDamageMul64_4E27D0(a, b float64) float64 { return a * b }

//go:noinline
func monsterGeneratorDamageSpill32_4E27D0(value float64) float32 { return float32(value) }

func monsterGeneratorDamageRoundThreshold4E27D0(max uint16, factorBits uint32) int32 {
	factor := math.Float32frombits(factorBits)
	scaled := monsterGeneratorDamageSpill32_4E27D0(
		monsterGeneratorDamageMul64_4E27D0(float64(uint32(max)), float64(factor)),
	)
	return int32(math.RoundToEven(float64(scaled)))
}

type monsterGeneratorDamageHooks4E27D0[O comparable, U, H any] struct {
	loadFlags       func(O) uint32
	loadUpdate      func(O) U
	frame           func() uint32
	loadEffectFrame func(O) uint32
	loadPosX        func(O) float32
	loadPosY        func(O) float32
	normalize       func(*types.Pointf)
	pointFX         func(byte, byte, types.Pointf)
	audio           func(int, O)
	getHP           func(O) uint16
	defaultDamage   func(O, O, O, int32, object.DamageType) bool
	scriptDamage    func(U, O, O, ScriptEventType)
	loadHealth      func(O) H
	loadHealthMax   func(H) uint16
	loadHealthCur   func(H) uint16
	loadXStatus     func(O) uint32
	setXStatus      func(O, uint32)
	unsetXStatus    func(O, uint32)
}

func monsterGeneratorDamageEffectPosition4E27D0[O comparable, U, H any](
	target, weapon O,
	hooks monsterGeneratorDamageHooks4E27D0[O, U, H],
) types.Pointf {
	var zero O
	if weapon == zero {
		return types.Pointf{
			X: hooks.loadPosX(target),
			Y: hooks.loadPosY(target),
		}
	}

	epsilon := math.Float32frombits(monsterGeneratorDamageEpsilonBits4E27D0)
	var point types.Pointf
	weaponX := hooks.loadPosX(weapon)
	targetX := hooks.loadPosX(target)
	point.X = monsterGeneratorDamageSpill32_4E27D0(
		monsterGeneratorDamageAdd64_4E27D0(
			monsterGeneratorDamageSub64_4E27D0(float64(weaponX), float64(targetX)),
			float64(epsilon),
		),
	)
	weaponY := hooks.loadPosY(weapon)
	targetY := hooks.loadPosY(target)
	point.Y = monsterGeneratorDamageSpill32_4E27D0(
		monsterGeneratorDamageAdd64_4E27D0(
			monsterGeneratorDamageSub64_4E27D0(float64(weaponY), float64(targetY)),
			float64(epsilon),
		),
	)
	hooks.normalize(&point)

	point.X = monsterGeneratorDamageSpill32_4E27D0(
		monsterGeneratorDamageAdd64_4E27D0(
			monsterGeneratorDamageMul64_4E27D0(float64(point.X), 22.0),
			float64(hooks.loadPosX(target)),
		),
	)
	point.Y = monsterGeneratorDamageSpill32_4E27D0(
		monsterGeneratorDamageAdd64_4E27D0(
			monsterGeneratorDamageMul64_4E27D0(float64(point.Y), 22.0),
			float64(hooks.loadPosY(target)),
		),
	)
	return point
}

// monsterGeneratorDamage4E27D0 preserves GAME.EXE 004E27D0, including the
// cached update pointer, unsigned frame subtraction, post-damage live field
// reloads, and the original one-third/two-thirds x87 rounding boundaries.
func monsterGeneratorDamage4E27D0[O comparable, U, H any](
	target, source, weapon O,
	damage int32,
	typ object.DamageType,
	hooks monsterGeneratorDamageHooks4E27D0[O, U, H],
) bool {
	flags := hooks.loadFlags(target)
	update := hooks.loadUpdate(target)
	if flags&uint32(object.FlagDead|object.FlagDestroyed) != 0 {
		return false
	}

	frame := hooks.frame()
	if frame-hooks.loadEffectFrame(target) > monsterGeneratorDamageRecentFrames4E27D0 ||
		frame%monsterGeneratorDamageFxPeriod4E27D0 == 0 {
		point := monsterGeneratorDamageEffectPosition4E27D0(target, weapon, hooks)
		hooks.pointFX(monsterGeneratorDamageFxOpcode4E27D0, monsterGeneratorDamageFxSubtype4E27D0, point)
		hooks.audio(monsterGeneratorDamageSound4E27D0, target)
	}

	oldHP := hooks.getHP(target)
	result := hooks.defaultDamage(target, source, weapon, damage, typ)
	if hooks.getHP(target) < oldHP {
		hooks.scriptDamage(update, source, target, NoxEventGeneratorDamage)
	}

	flags = hooks.loadFlags(target)
	if flags&uint32(object.FlagDead|object.FlagDestroyed) != 0 {
		return result
	}

	health := hooks.loadHealth(target)
	max := hooks.loadHealthMax(health)
	current := hooks.loadHealthCur(health)
	third := monsterGeneratorDamageRoundThreshold4E27D0(max, monsterGeneratorDamageThirdBits4E27D0)
	if int32(current) <= third {
		if hooks.loadXStatus(target)&monsterGeneratorDamageStatusMid4E27D0 != 0 {
			hooks.unsetXStatus(target, monsterGeneratorDamageStatusMid4E27D0)
		}
		if hooks.loadXStatus(target)&monsterGeneratorDamageStatusLow4E27D0 == 0 {
			hooks.setXStatus(target, monsterGeneratorDamageStatusLow4E27D0)
		}
		return result
	}

	health = hooks.loadHealth(target)
	max = hooks.loadHealthMax(health)
	twoThirds := monsterGeneratorDamageRoundThreshold4E27D0(max, monsterGeneratorDamageTwoThirdBits4E27D0)
	if int32(current) <= twoThirds && hooks.loadXStatus(target)&monsterGeneratorDamageStatusMid4E27D0 == 0 {
		hooks.setXStatus(target, monsterGeneratorDamageStatusMid4E27D0)
	}
	return result
}
