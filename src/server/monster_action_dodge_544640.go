package server

import (
	"math"

	"github.com/opennox/libs/object"

	"github.com/opennox/opennox/v1/common/unit/ai"
)

const (
	monsterActionDodgeDistanceLimitBits544640   = uint32(0x41000000)
	monsterActionDodgeDistanceEpsilonBits544640 = uint32(0x38d1b717)
)

type monsterActionDodgeHooks544640 struct {
	hasEnchant func(EnchantID) bool
	pop        func() int
}

// monsterActionDodge544640 restores GAME.EXE 00544640 without interpreting
// native-width Object, MonsterUpdateData, MonsterDef, or AI-stack pointers
// through their PE32 offsets. The original first gates movement with
// SpeedBase, but scales and overwrites SpeedCur before writing VelVec.
func monsterActionDodge544640(unit *Object, hooks monsterActionDodgeHooks544640) bool {
	if unit == nil || unit.UpdateData == nil || !unit.ObjClass.Has(object.ClassMonster) || hooks.pop == nil {
		return false
	}
	update := unit.UpdateDataMonster()
	if update.AIStackInd < 0 || int(update.AIStackInd) >= len(update.AIStack) {
		return false
	}
	head := update.AIStackHead()
	if head == nil || head.Type() != ai.ACTION_DODGE {
		return false
	}

	// nox_xxx_monsterIsMoveing_534320 compares SpeedBase against the
	// binary32 representation of 0.01. The inverted ordered comparison also
	// makes NaN take the original pop path.
	if !(unit.SpeedBase >= float32(0.0099999998)) {
		hooks.pop()
		return true
	}
	if hooks.hasEnchant != nil {
		for _, enchant := range [...]EnchantID{ENCHANT_CONFUSED, ENCHANT_HELD, ENCHANT_CHARMING} {
			if hooks.hasEnchant(enchant) {
				return true
			}
		}
	}

	target := head.ArgPos(0)
	dx := float64(target.X) - float64(unit.PosVec.X)
	dy := float64(target.Y) - float64(unit.PosVec.Y)
	epsilon := float64(math.Float32frombits(monsterActionDodgeDistanceEpsilonBits544640))
	distance := math.Sqrt(dx*dx+dy*dy) + epsilon
	denominator := float32(distance)
	limit := float64(math.Float32frombits(monsterActionDodgeDistanceLimitBits544640))
	if !(distance >= limit) {
		hooks.pop()
		return true
	}
	if update.MonsterDef == nil {
		return true
	}

	// The x87 routine stores the product to SpeedCur without popping it.
	// VelVec.X therefore consumes the unrounded product, whereas VelVec.Y
	// reloads the rounded binary32 SpeedCur value.
	speed := float64(update.MonsterDef.RunMultiplier96) * float64(unit.SpeedCur)
	unit.SpeedCur = float32(speed)
	unit.VelVec.X = float32(speed * dx / float64(denominator))
	unit.VelVec.Y = float32(dy * float64(unit.SpeedCur) / float64(denominator))
	return true
}

// MonsterActionDodge544640 binds the restored action to the live native
// object, enchant storage, and AI stack.
func (s *Server) MonsterActionDodge544640(unit *Object) bool {
	if unit == nil {
		return false
	}
	return monsterActionDodge544640(unit, monsterActionDodgeHooks544640{
		hasEnchant: unit.HasEnchant,
		pop:        unit.MonsterPopAction,
	})
}
