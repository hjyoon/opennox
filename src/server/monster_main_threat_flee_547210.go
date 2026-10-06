package server

import (
	"math"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/opennox/v1/common/unit/ai"
)

// monsterMainThreatFlee547210 restores GAME.EXE 00547567..00547771.
// Aggression, movement, enchant and action predicates read the live unit;
// enemy, range, spell flags and cooldowns belong to MainAI's cached record.
// Blink runs before the whole-stack FLEE test and uses the immediate 541300
// service, not the animation scheduler. The outer distance comparison retains
// the unrounded result, whereas the half-range test reads its float32 spill.
func (s *Server) monsterMainThreatFlee547210(unit *Object, update *MonsterUpdateData, soundSet unsafe.Pointer, runtime MonsterMainRuntime547210) bool {
	if unit == nil || unit.UpdateData == nil || update == nil ||
		!(unit.UpdateDataMonster().Aggression >= monsterMainPassiveAggressionLimit547210) ||
		!(unit.SpeedBase >= float32(0.0099999998)) {
		return false
	}
	if head := unit.UpdateDataMonster().AIStackHead(); head != nil {
		switch head.Type() {
		case ai.ACTION_CAST_SPELL_ON_OBJECT, ai.ACTION_CAST_SPELL_ON_LOCATION, ai.ACTION_CAST_DURATION_SPELL:
			return false
		}
	}
	if unit.HasEnchant(ENCHANT_CONFUSED) || s.MonsterMoveAttemptRecent534810(unit) || update.CurrentEnemy == nil {
		return false
	}
	distanceService := runtime.Distance
	if distanceService == nil {
		distanceService = ObjectDistance4E6C00
	}
	distance := distanceService(unit, update.CurrentEnemy)
	// 005475D5 FSTS stores under ToZero before the retained outer FCOMP;
	// 00547635 compares this cached binary32 word, not nearest-even distance.
	spilledDistance := monsterMoveToRunSpill544440(distance)
	fleeRange := float64(update.FleeRange)
	// x87 C0 alone admits unordered values at the outer comparison.
	if !(distance < fleeRange) && !math.IsNaN(distance) && !math.IsNaN(fleeRange) {
		return false
	}
	if update.StatusFlags.Has(object.MonStatusCanCastSpells) && update.Field376 != 0 &&
		!unit.HasEnchant(ENCHANT_ANTI_MAGIC) && s.Frame() >= update.Field371 &&
		fleeRange*0.5 > spilledDistance {
		random := runtime.RandomInt
		if random == nil && s.Rand.Logic != nil {
			random = s.Rand.Logic.IntClamp
		}
		if runtime.CastSpell == nil || random == nil {
			return false // do not replace an unavailable immediate cast with FLEE
		}
		arg := SpellAcceptArg{Obj: unit, Pos: unit.PosVec}
		runtime.CastSpell(4, unit, &arg)
		// Bounds are unsigned WORDs reread after casting, and gameFrame is read
		// after the RNG callback. Both remain tied to the entry-cached record.
		delay := random(int(update.Field370_0), int(update.Field370_2))
		update.Field371 = s.Frame() + uint32(delay)
		return true
	}
	if unit.UpdateDataMonster().HasAction(ai.ACTION_FLEE) || update.FleeRange == 0 || math.IsNaN(float64(update.FleeRange)) {
		return false
	}
	if action := unit.MonsterPushAction(ai.ACTION_SET_ANGLE); action != nil {
		action.Args[0] = uintptr(uint32(int32(int16(unit.Direction1)) + 128))
	}
	unit.MonsterPushAction(ai.DEPENDENCY_NOT_CORNERED)
	if action := unit.MonsterPushAction(ai.DEPENDENCY_ENEMY_CLOSER_THAN); action != nil {
		// 005476F8..00547704 reads the cached record after the accepted push,
		// retains FADDS at precision 53/ToZero, then FSTPS to binary32 ToZero.
		rangeSum := monsterMoveToRunAddChop53_544434(float64(update.FleeRange), 30)
		action.Args[0] = uintptr(math.Float32bits(float32(monsterMoveToRunSpill544440(rangeSum))))
	}
	if action := unit.MonsterPushAction(ai.ACTION_FLEE); action != nil {
		enemy := update.CurrentEnemy // the push may cancel an action and replace it
		action.Args[0] = uintptr(math.Float32bits(enemy.PosVec.X))
		action.Args[1] = uintptr(math.Float32bits(enemy.PosVec.Y))
		action.Args[2] = 0
	}
	playSound := false
	if runtime.RandomInt != nil {
		playSound = runtime.RandomInt(0, 1) != 0
	} else if s.Rand.Logic != nil {
		playSound = s.Rand.Logic.IntClamp(0, 1) != 0
	}
	if playSound && soundSet != nil && runtime.AudioEvent != nil {
		runtime.AudioEvent(*(*uint32)(unsafe.Add(soundSet, 48)), unit)
	}
	return true
}
