package server

import (
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/unit/ai"
)

// MonsterActionFightStartRuntime531E20 contains the engine services called by
// GAME.EXE 00531E20. Keeping them at the boundary makes the original call
// order testable while all pointer-bearing state stays in Go.
type MonsterActionFightStartRuntime531E20 struct {
	AudioEvent       func(uint32, *Object)
	ScriptCallback   func(*ScriptCallback, *Object, *Object, ScriptEventType)
	CopyFrameCounter func()
	UpdateSight      func(*Object)
}

func monsterActionFightStart531E20(unit *Object, runtime MonsterActionFightStartRuntime531E20) {
	if unit == nil || unit.UpdateData == nil || !unit.Class().Has(object.ClassMonster) {
		return
	}
	update := unit.UpdateDataMonster()
	if runtime.AudioEvent != nil && update.SoundSet122 != nil {
		runtime.AudioEvent(*(*uint32)(unsafe.Add(update.SoundSet122, 5*4)), unit)
	}
	if runtime.ScriptCallback != nil {
		runtime.ScriptCallback(&update.ScriptChangeFocus, update.CurrentEnemy, unit, NoxEventMonsterFightStart)
	}
	update.StatusFlags |= object.MonStatusAlert
	if runtime.CopyFrameCounter != nil {
		runtime.CopyFrameCounter()
	}
	if runtime.UpdateSight != nil {
		runtime.UpdateSight(unit)
	}
	if !update.StatusFlags.Has(object.MonStatusNeverRun) {
		update.StatusFlags |= object.MonStatusRunning
	}
}

// MonsterActionFightStart531E20 restores GAME.EXE 00531E20 without loading
// Object.UpdateData through the original 32-bit offset-748 pointer slot.
func (s *Server) MonsterActionFightStart531E20(unit *Object, runtime MonsterActionFightStartRuntime531E20) {
	monsterActionFightStart531E20(unit, runtime)
}

func monsterActionFightEnd531E90(unit *Object) {
	if unit == nil || unit.UpdateData == nil || !unit.Class().Has(object.ClassMonster) {
		return
	}
	update := unit.UpdateDataMonster()
	update.StatusFlags &^= object.MonStatusAlert
	if !update.StatusFlags.Has(object.MonStatusAlwaysRun) {
		update.StatusFlags &^= object.MonStatusRunning
	}
}

// MonsterActionFightEnd531E90 restores GAME.EXE 00531E90.
func (s *Server) MonsterActionFightEnd531E90(unit *Object) {
	monsterActionFightEnd531E90(unit)
}

type monsterActionFightHooks531EC0 struct {
	frame          func() uint32
	tickRate       func() uint32
	findDeadTarget func(types.Pointf, uint32) bool
	push           func(ai.ActionType, ...any) *AIStackItem
	pop            func() int
}

func monsterFightCanShoot534280(unit *Object, update *MonsterUpdateData) bool {
	if unit.SubClass().Has(object.SubClass(object.MonsterNPC)) {
		return update.WeaponEquipFlags&0x47f00fe != 0
	}
	return update.MonsterDef != nil && update.MonsterDef.MissileName148[0] != 0
}

func monsterFightCanMelee534220(unit *Object, update *MonsterUpdateData) bool {
	if update.StatusFlags.Has(object.MonStatusCanCastSpells) {
		return false
	}
	if unit.SubClass().Has(object.SubClass(object.MonsterNPC)) {
		return !monsterFightCanShoot534280(unit, update)
	}
	return update.MonsterDef != nil && update.MonsterDef.MeleeAttackRange112 > 0
}

func monsterFightSchedulePursuit531C10(target *Object, hooks monsterActionFightHooks531EC0) {
	hooks.push(ai.DEPENDENCY_NO_NEW_ENEMY, target)
	hooks.push(ai.DEPENDENCY_ALIVE, target)
	hooks.push(ai.ACTION_MOVE_TO, target.PosVec, target)
}

func monsterFightScheduleMelee531C60(unit, target *Object, hooks monsterActionFightHooks531EC0) {
	update := unit.UpdateDataMonster()
	def := update.MonsterDef
	hooks.push(ai.DEPENDENCY_NO_NEW_ENEMY, target)
	hooks.push(ai.DEPENDENCY_ALIVE, target)
	if monsterFightCanShoot534280(unit, update) {
		hooks.push(ai.DEPENDENCY_OBJECT_CLOSER_THAN, def.MissileAttackRange212*0.60000002, uint32(0), target)
	}
	hooks.push(ai.DEPENDENCY_CAN_SEE, target)
	hooks.push(ai.ACTION_MELEE_ATTACK)
	hooks.push(ai.ACTION_FACE_OBJECT, target)
	hooks.push(ai.DEPENDENCY_OBJECT_FARTHER_THAN, def.MeleeAttackRange112, uint32(0), target)
	if unit.SubClass().Has(object.SubClass(object.MonsterNPC)) {
		hooks.push(ai.DEPENDENCY_WAIT_FOR_STAMINA)
		hooks.push(ai.DEPENDENCY_OR)
	}
	hooks.push(ai.ACTION_MOVE_TO, target.PosVec, target)
}

func monsterFightKillable528190(target *Object) bool {
	if target == nil || target.HealthData == nil {
		return false
	}
	return target.HealthData.Cur != 0 || target.HealthData.Max == 0
}

// monsterActionFight531EC0 restores the pointer-bearing FIGHT loop for
// ordinary non-casting melee monsters. It returns false before mutation when
// the original branch would enter a spell or missile scheduler that has not
// yet crossed the native-width boundary.
func monsterActionFight531EC0(unit *Object, hooks monsterActionFightHooks531EC0) bool {
	if unit == nil || unit.UpdateData == nil || !unit.Class().Has(object.ClassMonster) ||
		hooks.frame == nil || hooks.tickRate == nil || hooks.push == nil || hooks.pop == nil {
		return false
	}
	update := unit.UpdateDataMonster()
	head := update.AIStackHead()
	if head == nil || head.Type() != ai.ACTION_FIGHT {
		return false
	}

	frame := hooks.frame()
	if frame-head.ArgU32(2) > 10*hooks.tickRate() {
		hooks.pop()
		return true
	}
	if target := update.CurrentEnemy; target != nil {
		if !monsterFightKillable528190(target) {
			hooks.pop()
			return true
		}
		// Non-anti-magic casters call three spell-selection routines before
		// choosing a physical attack. Leave that distinct branch unopened.
		if update.StatusFlags.Has(object.MonStatusCanCastSpells) && !unit.HasEnchant(ENCHANT_ANTI_MAGIC) {
			return false
		}
		// The missile scheduler leads into a separate pointer-bearing action.
		// Admit the melee-only and pursuit paths here; mixed/ranged monsters
		// remain explicit instead of silently receiving different behaviour.
		if monsterFightCanShoot534280(unit, update) {
			return false
		}
		head.Args[2] = uintptr(frame)
		if monsterFightCanMelee534220(unit, update) {
			monsterFightScheduleMelee531C60(unit, target, hooks)
		} else if !update.StatusFlags.Has(object.MonStatusCanCastSpells) {
			monsterFightSchedulePursuit531C10(target, hooks)
		}
		return true
	}

	if update.StatusFlags.Has(object.MonStatusHoldYourGround) || head.Type() == ai.ACTION_GUARD {
		hooks.pop()
		return true
	}
	remembered := head.ArgPos(0)
	if hooks.findDeadTarget != nil && hooks.findDeadTarget(remembered, update.Field300) {
		hooks.pop()
		if update.Field98 == update.Field300 {
			update.Field97 = 0
		}
		return true
	}
	delta := remembered.Sub(unit.PosVec)
	if float64(delta.X*delta.X+delta.Y*delta.Y) < 64.0 {
		hooks.pop()
		return true
	}
	hooks.push(ai.DEPENDENCY_NO_VISIBLE_ENEMY)
	hooks.push(ai.ACTION_MOVE_TO, remembered, uint32(0))
	return true
}

// MonsterActionFight531EC0 binds the restored melee FIGHT branch to the live
// object index and native-width action stack.
func (s *Server) MonsterActionFight531EC0(unit *Object) bool {
	return monsterActionFight531EC0(unit, monsterActionFightHooks531EC0{
		frame:    s.Frame,
		tickRate: s.TickRate,
		findDeadTarget: func(pos types.Pointf, netCode uint32) bool {
			found := false
			s.Map.EachObjInCircle(pos, 30, func(candidate *Object) bool {
				if candidate.NetCode == netCode && candidate.ObjFlags.Has(object.FlagDead) {
					found = true
					return false
				}
				return true
			})
			return found
		},
		push: unit.MonsterPushAction,
		pop:  unit.MonsterPopAction,
	})
}
