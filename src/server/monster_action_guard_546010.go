package server

import (
	"math"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/unit/ai"
)

type monsterActionGuardHooks546010 struct {
	frame            func() uint32
	tickRate         func() uint32
	noticeThreat     func() int
	lookAtDamager    func() bool
	interestingSound func() int
	random           func(int, int) int
	isMimic          func() bool
	isPlant          func() bool
	healSomeone      func() int
	push             func(ai.ActionType, ...any) *AIStackItem
}

func monsterGuardFacing546010(unit *Object, direction types.Pointf) bool {
	forward := unit.Direction1.Vec()
	return float64(forward.X)*float64(direction.X)+float64(forward.Y)*float64(direction.Y) > 0.89999998
}

func monsterGuardFacingObject546010(unit, target *Object) bool {
	delta := target.PosVec.Sub(unit.PosVec)
	distance := math.Sqrt(float64(delta.X*delta.X+delta.Y*delta.Y)) + 0.001
	return monsterGuardFacing546010(unit, types.Ptf(float32(float64(delta.X)/distance), float32(float64(delta.Y)/distance)))
}

func monsterGuardRecentSound546010(unit *Object, hooks monsterActionGuardHooks546010) bool {
	update := unit.UpdateDataMonster()
	frame := hooks.frame()
	if update.Field97 == 0 || frame-update.Field101 >= 3*hooks.tickRate() {
		return false
	}
	hooks.push(ai.DEPENDENCY_NO_INTERESTING_SOUND)
	hooks.push(ai.DEPENDENCY_NO_VISIBLE_ENEMY)
	hooks.push(ai.ACTION_WAIT, frame+uint32(hooks.random(int(hooks.tickRate()), 2*int(hooks.tickRate()))))
	hooks.push(ai.ACTION_FACE_LOCATION, types.Ptf(update.Field99X, update.Field99Y))
	update.Field97 = 0
	return true
}

// monsterActionGuard546010 restores GAME.EXE 00546010 without interpreting a
// native-width Object, MonsterUpdateData, or AI stack pointer through PE32 C
// offsets. The hook boundary records the engine calls made by the original
// routine while keeping its branch and push ordering independently testable.
func monsterActionGuard546010(unit *Object, hooks monsterActionGuardHooks546010) {
	if unit == nil || unit.UpdateData == nil || !unit.ObjClass.Has(object.ClassMonster) {
		return
	}
	update := unit.UpdateDataMonster()
	if update.AIStackInd < 0 || int(update.AIStackInd) >= len(update.AIStack) {
		return
	}
	head := update.AIStackHead()
	if head.Type() != ai.ACTION_GUARD {
		return
	}

	if unit.Sub_534440() || hooks.noticeThreat() == 0 {
		if enemy := update.CurrentEnemy; enemy != nil {
			if unit.Nox_xxx_monsterCanAttackAtWill_534390() {
				hooks.push(ai.ACTION_FIGHT, enemy.PosVec, hooks.frame())
				return
			}
			if unit.Sub_5343C0() {
				guardPos := head.ArgPos(0)
				delta := guardPos.Sub(enemy.PosVec)
				if float64(update.SightRange)*float64(update.SightRange) > float64(delta.X*delta.X+delta.Y*delta.Y) {
					if hooks.isPlant() {
						hooks.push(ai.DEPENDENCY_ENEMY_CLOSER_THAN, update.SightRange*1.05)
					} else {
						hooks.push(ai.DEPENDENCY_UNDER_ATTACK, uint32(0))
						hooks.push(ai.DEPENDENCY_LOCATION_CLOSER_THAN, update.SightRange*1.5, guardPos)
						hooks.push(ai.DEPENDENCY_OR)
					}
					hooks.push(ai.ACTION_FIGHT, enemy.PosVec, hooks.frame())
				}
			}
		}

		mimic := hooks.isMimic()
		if !mimic && hooks.lookAtDamager() {
			return
		}
		if unit.Nox_xxx_monsterCanAttackAtWill_534390() {
			if hooks.interestingSound() != 0 {
				return
			}
		} else if !mimic && monsterGuardRecentSound546010(unit, hooks) {
			return
		}

		if (byte(hooks.frame())+byte(unit.NetCode))&0xf == 0 {
			guardPos := head.ArgPos(0)
			delta := guardPos.Sub(unit.PosVec)
			if float64(delta.X*delta.X+delta.Y*delta.Y) > 64.0 {
				if unit.Nox_xxx_monsterCanAttackAtWill_534390() {
					hooks.push(ai.DEPENDENCY_NO_VISIBLE_ENEMY)
					hooks.push(ai.DEPENDENCY_NO_INTERESTING_SOUND)
				}
				hooks.push(ai.DEPENDENCY_NOT_UNDER_ATTACK)
				hooks.push(ai.ACTION_MOVE_TO, guardPos, 0)
				return
			}

			if mimic {
				direction := byte(head.ArgU32(2))
				cosine, sine := SinCosDir(direction)
				if !monsterGuardFacing546010(unit, types.Ptf(cosine, sine)) {
					hooks.push(ai.ACTION_FACE_ANGLE, uint32(direction))
				}
			} else if update.Field282_1 != 0 {
				target := update.CurrentEnemy
				if target == nil {
					count := int(update.Field282_1)
					if count > len(update.SeenEnemies) {
						count = len(update.SeenEnemies)
					}
					for i := 0; i < count; i++ {
						target = update.SeenEnemies[i]
						if target != nil && target.ObjFlags.Has(object.FlagEnabled) {
							break
						}
					}
				}
				if target != nil && !monsterGuardFacingObject546010(unit, target) {
					hooks.push(ai.ACTION_FACE_OBJECT, target)
				}
			} else {
				direction := byte(head.ArgU32(2))
				cosine, sine := SinCosDir(direction)
				if !monsterGuardFacing546010(unit, types.Ptf(cosine, sine)) {
					hooks.push(ai.DEPENDENCY_NO_VISIBLE_ENEMY)
					hooks.push(ai.ACTION_FACE_ANGLE, uint32(direction))
				}
			}
		}

		if mimic || hooks.frame()-update.Field137 <= hooks.tickRate()/2 || unit.PosVec == unit.PrevPos {
			if !unit.HasEnchant(ENCHANT_ANTI_MAGIC) {
				hooks.healSomeone()
			}
		} else {
			hooks.push(ai.ACTION_FACE_LOCATION, unit.PrevPos)
		}
	}
}

// MonsterActionGuard546010 binds the restored guard action to the live server.
// interestingSound is supplied by the top-level game package while its tile
// and pathfinding compatibility callbacks remain there.
func (s *Server) MonsterActionGuard546010(unit *Object, interestingSound func(*Object) int) {
	monsterActionGuard546010(unit, monsterActionGuardHooks546010{
		frame:            s.Frame,
		tickRate:         s.TickRate,
		noticeThreat:     func() int { return unit.Sub_545E60() },
		lookAtDamager:    unit.MonsterLookAtDamager,
		interestingSound: func() int { return interestingSound(unit) },
		random:           s.Rand.Logic.IntClamp,
		isMimic:          func() bool { return s.IsMimic(unit) },
		isPlant:          func() bool { return s.IsPlant(unit) },
		healSomeone:      func() int { return s.MonsterHealSomeone5411A0(unit) },
		push:             unit.MonsterPushAction,
	})
}
