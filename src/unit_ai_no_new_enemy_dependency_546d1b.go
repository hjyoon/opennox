package opennox

import "github.com/opennox/opennox/v1/server"

// GAME.EXE 00546D34 keeps each difference, Y-square, X-square and sum in
// x87 registers. Ordinary gameplay uses 53-bit precision with chop; the
// finite binary32 coordinate domain fits the binary64 exponent range.
// Reuse the local directed-rounding operations without changing thread FPU
// state, spilling to binary32, taking a square root or subtracting shapes.
func aiDependencyCenterDistanceSquared546D34(unit, target *server.Object) float64 {
	dx := aiDependencyAddChop53_546B63(float64(target.PosVec.X), -float64(unit.PosVec.X))
	dy := aiDependencyAddChop53_546B63(float64(target.PosVec.Y), -float64(unit.PosVec.Y))
	ySquare := aiDependencySquareChop53_546B63(dy)
	xSquare := aiDependencySquareChop53_546B63(dx)
	return aiDependencyAddChop53_546B63(ySquare, xSquare)
}

// Original condition 60 (table 005470E4 -> 00546D1B) reads Args0 before
// the entry-cached update's CurrentEnemy. Nil old fails, then nil enemy
// passes without accessing positions. Ordered new >= old passes; C0 also
// marks unordered, which fails. No class, flag, health or identity gate.
func aiDependencyNoNewEnemy546D1B(unit *server.Object, cached *server.MonsterUpdateData, slot *server.AIStackItem) bool {
	old := slot.ArgObj(0)
	if old == nil {
		return false
	}
	enemy := cached.CurrentEnemy
	if enemy == nil {
		return true
	}
	oldSquare := aiDependencyCenterDistanceSquared546D34(unit, old)
	newSquare := aiDependencyCenterDistanceSquared546D34(unit, enemy)
	return newSquare >= oldSquare
}
