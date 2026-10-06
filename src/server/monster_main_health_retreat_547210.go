package server

import (
	"unsafe"

	"github.com/opennox/libs/object"
	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/unit/ai"
)

// monsterMainHealthRetreat547210 restores GAME.EXE 00547772..005478AE.
// Unlike the earlier partial classifier, this branch is independent of
// aggression, spell animations and most enchantments. Health ratio is spilled
// to binary32 before the threshold comparison; x87 unordered also retreats.
// Predicates use the live unit, while spell status, threshold and script block
// belong to the entry-cached update. Sound retains the entry-cached pointer.
func (s *Server) monsterMainHealthRetreat547210(unit *Object, update *MonsterUpdateData, soundSet unsafe.Pointer, runtime MonsterMainRuntime547210) bool {
	if unit == nil || unit.UpdateData == nil || update == nil || unit.HealthData == nil ||
		unit.HealthData.Max == 0 || !(unit.SpeedBase >= float32(0.0099999998)) ||
		s.MonsterMoveAttemptRecent534810(unit) {
		return false
	}
	live := unit.UpdateDataMonster()
	if live.HasAction(ai.ACTION_FLEE) || live.HasAction(ai.ACTION_RETREAT) || live.HasAction(ai.ACTION_RETREAT_TO_MASTER) {
		return false
	}
	health := unit.HealthData
	// FIDIV 00547803 uses precision 53, then FSTP 00547807 spills to
	// binary32 under gameplay's ToZero rounding, not host nearest-even.
	ratio := monsterMoveToRunSpill544440(monsterMoveForceDivChop53_50D581(float64(health.Cur), float64(health.Max)))
	antiMagicCaster := update.StatusFlags.Has(object.MonStatusCanCastSpells) && unit.HasEnchant(ENCHANT_ANTI_MAGIC)
	if float64(ratio) > float64(update.RetreatLevel) && !antiMagicCaster {
		return false
	}

	s.monsterMainPopAttackActions5471B0(unit)
	unit.MonsterPushAction(ai.DEPENDENCY_NOT_CORNERED)
	// Cancellation callbacks during pop/push may change these predicates.
	retreat := ai.ACTION_RETREAT
	if (update.StatusFlags.Has(object.MonStatusSummoned) || unit.ObjSubClass.AsMonster().Has(object.MonsterMonitor)) &&
		noxflags.HasGame(noxflags.GameModeCoop) {
		retreat = ai.ACTION_RETREAT_TO_MASTER
	}
	unit.MonsterPushAction(retreat)
	if soundSet != nil && runtime.AudioEvent != nil {
		runtime.AudioEvent(*(*uint32)(unsafe.Add(soundSet, 52)), unit)
	}
	if runtime.ScriptCallback != nil {
		runtime.ScriptCallback(&update.ScriptRetreat, nil, unit, NoxEventMonsterMoveXXX)
	}
	return true
}
