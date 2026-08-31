package server

import "math"

const (
	warcryProximityWarrior4FC4C0 = uint8(0)
	warcryProximityEpsilon4FC4C0 = float32(0.1)
	warcryProximityLimit4FC4C0   = float32(300.0)
)

type warcryProximityScanHooks4FC4C0[P, O comparable] struct {
	firstPlayer     func() P
	loadTargetArg   func() O
	loadPlayerUnit  func(P) O
	loadPlayerClass func(P) uint8
	isAbilityActive func(O, Ability) int32
	loadPosX        func(O) float32
	loadPosY        func(O) float32
	mapCheck        func(O, O) int32
	nextPlayer      func(P) P
}

// Keep every x87 arithmetic instruction at an explicit binary64 boundary.
// GAME.EXE executes this expression with 53-bit precision control; noinline
// also prevents a target compiler from contracting or reassociating it.
//
//go:noinline
func warcryProximitySub64_4FC4C0(left, right float64) float64 {
	return left - right
}

//go:noinline
func warcryProximityMul64_4FC4C0(left, right float64) float64 {
	return left * right
}

//go:noinline
func warcryProximityAdd64_4FC4C0(left, right float64) float64 {
	return left + right
}

//go:noinline
func warcryProximitySqrt64_4FC4C0(value float64) float64 {
	return math.Sqrt(value)
}

func warcryProximityDistance4FC4C0(targetX, unitX, targetY, unitY float32) float64 {
	deltaX := warcryProximitySub64_4FC4C0(float64(targetX), float64(unitX))
	deltaY := warcryProximitySub64_4FC4C0(float64(targetY), float64(unitY))
	deltaYSquared := warcryProximityMul64_4FC4C0(deltaY, deltaY)
	deltaXSquared := warcryProximityMul64_4FC4C0(deltaX, deltaX)
	return warcryProximitySqrt64_4FC4C0(
		warcryProximityAdd64_4FC4C0(deltaYSquared, deltaXSquared),
	)
}

// warcryProximityScan4FC4C0 preserves GAME.EXE 004FC4C0. The player-list
// head is acquired before the target argument is observed. Each candidate's
// unit is checked for nil before its PlayerClass byte, and only class zero
// (Warrior) reaches the fixed AbilityWarcry membership test. That callback may
// mutate the Player, so the unit is deliberately reloaded before position and
// visibility reads. The target argument itself remains cached across the scan.
//
// Coordinates are observed in target-X, unit-X, target-Y, unit-Y order. The
// original x87 comparison tests C0 only: ordered values pass strictly below
// 300.0 after adding the binary32 0.1 constant, while unordered (NaN) values
// pass as well. A nonzero visibility result returns canonical one immediately;
// exhausted traversal returns canonical zero.
func warcryProximityScan4FC4C0[P, O comparable](hooks warcryProximityScanHooks4FC4C0[P, O]) int32 {
	player := hooks.firstPlayer()
	var zeroPlayer P
	if player == zeroPlayer {
		return 0
	}
	target := hooks.loadTargetArg()

	var zeroObject O
	for {
		unit := hooks.loadPlayerUnit(player)
		if unit != zeroObject &&
			hooks.loadPlayerClass(player) == warcryProximityWarrior4FC4C0 &&
			hooks.isAbilityActive(unit, AbilityWarcry) != 0 {
			unit = hooks.loadPlayerUnit(player)
			targetX := hooks.loadPosX(target)
			unitX := hooks.loadPosX(unit)
			targetY := hooks.loadPosY(target)
			unitY := hooks.loadPosY(unit)
			distance := warcryProximityDistance4FC4C0(targetX, unitX, targetY, unitY)
			withEpsilon := warcryProximityAdd64_4FC4C0(
				distance,
				float64(warcryProximityEpsilon4FC4C0),
			)
			if !(withEpsilon >= float64(warcryProximityLimit4FC4C0)) &&
				hooks.mapCheck(unit, target) != 0 {
				return 1
			}
		}

		player = hooks.nextPlayer(player)
		if player == zeroPlayer {
			return 0
		}
	}
}
