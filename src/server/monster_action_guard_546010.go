package server

import (
	"math"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/common/unit/ai"
)

type monsterActionGuardHooks546010 struct {
	Nearest          bool
	ObserveFacing    func(types.Pointf)
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

func monsterGuardFacing546010(unit *Object, direction types.Pointf, nearest bool) bool {
	index := int32(int16(unit.Direction1)) * 2
	y := monsterGuardMul546010(float64(monsterConfusedDirectionFloat545140(index+1)), objectForceMemory32_52DF80(direction.Y), nearest)
	x := monsterGuardMul546010(float64(monsterConfusedDirectionFloat545140(index)), objectForceMemory32_52DF80(direction.X), nearest)
	return monsterGuardAdd546010(y, x, nearest) > float64(math.Float32frombits(0x3f666666))
}

func monsterGuardRecentSound546010(unit *Object, hooks monsterActionGuardHooks546010) bool {
	update := unit.UpdateDataMonster()
	if update.Field97 == 0 {
		return false
	}
	if hooks.frame()-update.Field101 >= 3*hooks.tickRate() {
		return false
	}
	hooks.push(ai.DEPENDENCY_NO_INTERESTING_SOUND)
	hooks.push(ai.DEPENDENCY_NO_VISIBLE_ENEMY)
	if item := hooks.push(ai.ACTION_WAIT); item != nil {
		fps := hooks.tickRate()
		delay := hooks.random(int(int32(fps)), int(int32(fps*2)))
		item.Args[0] = uintptr(hooks.frame() + uint32(delay))
	}
	if item := hooks.push(ai.ACTION_FACE_LOCATION); item != nil {
		item.Args[0] = uintptr(math.Float32bits(update.Field99X))
		item.Args[1] = uintptr(math.Float32bits(update.Field99Y))
	}
	update.Field97 = 0
	return true
}

// monsterActionGuard546010 retains the entry update/head addresses, then
// reloads live data after each accepted Push in the original store order.
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
	if !unit.Sub_534440() && hooks.noticeThreat() != 0 {
		return
	}
	fight := func() {
		if item := hooks.push(ai.ACTION_FIGHT); item != nil {
			enemy := update.CurrentEnemy
			item.Args[0] = uintptr(math.Float32bits(enemy.PosVec.X))
			item.Args[1] = uintptr(math.Float32bits(enemy.PosVec.Y))
			item.Args[2] = uintptr(hooks.frame())
		}
	}
	if update.CurrentEnemy != nil {
		if unit.Nox_xxx_monsterCanAttackAtWill_534390() {
			fight()
			return
		}
		if unit.Sub_5343C0() {
			enemy := update.CurrentEnemy
			x := monsterGuardSub546010(float64(head.ArgF32(0)), objectForceMemory32_52DF80(enemy.PosVec.X), hooks.Nearest)
			y := monsterGuardSub546010(float64(head.ArgF32(1)), objectForceMemory32_52DF80(enemy.PosVec.Y), hooks.Nearest)
			sight := float64(update.SightRange)
			if monsterGuardMul546010(sight, sight, hooks.Nearest) > monsterGuardSquare546010(x, y, hooks.Nearest) {
				if hooks.isPlant() {
					if item := hooks.push(ai.DEPENDENCY_ENEMY_CLOSER_THAN); item != nil {
						item.Args[0] = uintptr(math.Float32bits(monsterGuardSpill546010(monsterGuardMul546010(float64(update.SightRange), float64(math.Float32frombits(0x3f866666)), hooks.Nearest), hooks.Nearest)))
					}
				} else {
					if item := hooks.push(ai.DEPENDENCY_UNDER_ATTACK); item != nil {
						item.Args[0] = 0
					}
					if item := hooks.push(ai.DEPENDENCY_LOCATION_CLOSER_THAN); item != nil {
						item.Args[0] = uintptr(math.Float32bits(monsterGuardSpill546010(monsterGuardMul546010(float64(update.SightRange), 1.5, hooks.Nearest), hooks.Nearest)))
						item.Args[2] = uintptr(head.ArgU32(0))
						item.Args[3] = uintptr(head.ArgU32(1))
					}
					hooks.push(ai.DEPENDENCY_OR)
				}
				fight()
			}
		}
	}
	if !hooks.isMimic() && hooks.lookAtDamager() {
		return
	}
	if unit.Nox_xxx_monsterCanAttackAtWill_534390() {
		if hooks.interestingSound() != 0 {
			return
		}
	} else if !hooks.isMimic() && monsterGuardRecentSound546010(unit, hooks) {
		return
	}
	facing := func(direction types.Pointf) bool {
		if hooks.ObserveFacing != nil {
			hooks.ObserveFacing(direction)
		}
		return monsterGuardFacing546010(unit, direction, hooks.Nearest)
	}
	if (hooks.frame()+unit.NetCode)&15 == 0 {
		x := monsterGuardSub546010(float64(head.ArgF32(0)), objectForceMemory32_52DF80(unit.PosVec.X), hooks.Nearest)
		y := monsterGuardSub546010(float64(head.ArgF32(1)), objectForceMemory32_52DF80(unit.PosVec.Y), hooks.Nearest)
		if monsterGuardSquare546010(x, y, hooks.Nearest) > 64 {
			if unit.Nox_xxx_monsterCanAttackAtWill_534390() {
				hooks.push(ai.DEPENDENCY_NO_VISIBLE_ENEMY)
				hooks.push(ai.DEPENDENCY_NO_INTERESTING_SOUND)
			}
			hooks.push(ai.DEPENDENCY_NOT_UNDER_ATTACK)
			if item := hooks.push(ai.ACTION_MOVE_TO); item != nil {
				item.Args[0] = uintptr(head.ArgU32(0))
				y := head.ArgU32(1)
				item.Args[2] = 0
				item.Args[1] = uintptr(y)
			}
			return
		}
		if hooks.isMimic() {
			direction := head.ArgU32(2)
			if !facing(monsterGuardDirection546010(direction)) {
				if item := hooks.push(ai.ACTION_FACE_ANGLE); item != nil {
					item.Args[0] = uintptr(direction)
				}
			}
		} else if update.Field282_1 != 0 {
			target := update.CurrentEnemy
			if target == nil {
				// If no player qualifies, the original retains the first seen
				// object. ClassPlayer is the BYTE+8 bit, independent of flags.
				target = update.SeenEnemies[0]
				for i := 0; i < int(update.Field282_1) && i < len(update.SeenEnemies); i++ {
					if update.SeenEnemies[i].ObjClass.Has(object.ClassPlayer) {
						target = update.SeenEnemies[i]
						break
					}
				}
			}
			if !facing(monsterGuardTargetDirection546010(unit, target, hooks.Nearest)) {
				if item := hooks.push(ai.ACTION_FACE_OBJECT); item != nil {
					item.Args[0] = uintptr(target.CObj())
				}
			}
		} else {
			direction := head.ArgU32(2)
			if !facing(monsterGuardDirection546010(direction)) {
				hooks.push(ai.DEPENDENCY_NO_VISIBLE_ENEMY)
				if item := hooks.push(ai.ACTION_FACE_ANGLE); item != nil {
					item.Args[0] = uintptr(direction)
				}
			}
		}
	}
	if hooks.isMimic() || hooks.frame()-update.Field137 <= hooks.tickRate()/2 ||
		monsterGuardUnorderedEqual546010(unit.PosVec.X, unit.PrevPos.X) && monsterGuardUnorderedEqual546010(unit.PosVec.Y, unit.PrevPos.Y) {
		if !unit.HasEnchant(ENCHANT_ANTI_MAGIC) {
			hooks.healSomeone()
		}
	} else if item := hooks.push(ai.ACTION_FACE_LOCATION); item != nil {
		item.Args[0] = uintptr(math.Float32bits(unit.PrevPos.X))
		item.Args[1] = uintptr(math.Float32bits(unit.PrevPos.Y))
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
