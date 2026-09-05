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
	buffSelf       func(*Object) bool
	castOffensive  func(*Object, *Object) bool
	castRelated    func(*Object, *Object) bool
	distance       func(*Object, *Object) float64
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

func monsterFightPush531C10(hooks monsterActionFightHooks531EC0, action ai.ActionType, args ...any) *AIStackItem {
	item := hooks.push(action)
	if item != nil {
		item.SetArgs(args...)
	}
	return item
}

func monsterFightSchedulePursuit531C10(target *Object, hooks monsterActionFightHooks531EC0) {
	monsterFightPush531C10(hooks, ai.DEPENDENCY_NO_NEW_ENEMY, target)
	monsterFightPush531C10(hooks, ai.DEPENDENCY_ALIVE, target)
	item := hooks.push(ai.ACTION_MOVE_TO)
	if item != nil {
		item.SetArgs(target.PosVec, target)
	}
}

func monsterFightScheduleMelee531C60(unit, target *Object, hooks monsterActionFightHooks531EC0) {
	update := unit.UpdateDataMonster()
	monsterFightPush531C10(hooks, ai.DEPENDENCY_NO_NEW_ENEMY, target)
	monsterFightPush531C10(hooks, ai.DEPENDENCY_ALIVE, target)
	if monsterFightCanShoot534280(unit, update) {
		item := hooks.push(ai.DEPENDENCY_OBJECT_CLOSER_THAN)
		if item != nil && update.MonsterDef != nil {
			item.SetArgs(update.MonsterDef.MissileAttackRange212*0.60000002, uint32(0), target)
		}
	}
	monsterFightPush531C10(hooks, ai.DEPENDENCY_CAN_SEE, target)
	hooks.push(ai.ACTION_MELEE_ATTACK)
	monsterFightPush531C10(hooks, ai.ACTION_FACE_OBJECT, target)
	item := hooks.push(ai.DEPENDENCY_OBJECT_FARTHER_THAN)
	if item != nil && update.MonsterDef != nil {
		item.SetArgs(update.MonsterDef.MeleeAttackRange112, uint32(0), target)
	}
	if unit.SubClass().Has(object.SubClass(object.MonsterNPC)) {
		hooks.push(ai.DEPENDENCY_WAIT_FOR_STAMINA)
		hooks.push(ai.DEPENDENCY_OR)
	}
	item = hooks.push(ai.ACTION_MOVE_TO)
	if item != nil {
		item.SetArgs(target.PosVec, target)
	}
}

func monsterFightHoldPosition534710(unit *Object) bool {
	update := unit.UpdateDataMonster()
	return update.StatusFlags.Has(object.MonStatusHoldYourGround) || update.AIStackHead().Type() == ai.ACTION_GUARD
}

func monsterFightScheduleMissile531D50(unit, target *Object, hooks monsterActionFightHooks531EC0) {
	update := unit.UpdateDataMonster()
	monsterFightPush531C10(hooks, ai.DEPENDENCY_NO_NEW_ENEMY, target)
	monsterFightPush531C10(hooks, ai.DEPENDENCY_CAN_SEE, target)
	item := hooks.push(ai.ACTION_MISSILE_ATTACK)
	if item != nil {
		item.SetArgs(target.PosVec, target)
	}
	monsterFightPush531C10(hooks, ai.ACTION_FACE_OBJECT, target)
	if monsterFightHoldPosition534710(unit) {
		return
	}
	monsterFightPush531C10(hooks, ai.DEPENDENCY_BLOCKED_LINE_OF_FIRE, target)
	item = hooks.push(ai.DEPENDENCY_OBJECT_FARTHER_THAN)
	if item != nil && update.MonsterDef != nil {
		item.SetArgs(update.MonsterDef.MissileAttackRange212, uint32(0), target)
	}
	hooks.push(ai.DEPENDENCY_OR)
	item = hooks.push(ai.ACTION_MOVE_TO)
	if item != nil {
		item.SetArgs(target.PosVec, target)
	}
}

func monsterFightSelectAttack531B40(unit, target *Object, hooks monsterActionFightHooks531EC0) {
	update := unit.UpdateDataMonster()
	if !unit.HasEnchant(ENCHANT_ANTI_MAGIC) && hooks.castRelated(unit, target) {
		return
	}
	if monsterFightCanShoot534280(unit, update) {
		if monsterFightCanMelee534220(unit, update) &&
			hooks.distance(unit, target) < float64(update.MonsterDef.MissileAttackRange212*0.5) {
			monsterFightScheduleMelee531C60(unit, target, hooks)
			return
		}
		monsterFightScheduleMissile531D50(unit, target, hooks)
		return
	}
	if monsterFightCanMelee534220(unit, update) {
		monsterFightScheduleMelee531C60(unit, target, hooks)
		return
	}
	if !update.StatusFlags.Has(object.MonStatusCanCastSpells) {
		monsterFightSchedulePursuit531C10(target, hooks)
	}
}

func monsterFightKillable528190(target *Object) bool {
	if target == nil || target.HealthData == nil {
		return false
	}
	return target.HealthData.Cur != 0 || target.HealthData.Max == 0
}

// monsterActionFight531EC0 restores the complete pointer-bearing FIGHT loop.
// It returns false before mutation when a required native service is absent;
// a fully bound call never needs the original PE32 fallback.
func monsterActionFight531EC0(unit *Object, hooks monsterActionFightHooks531EC0) bool {
	if unit == nil || unit.UpdateData == nil || !unit.Class().Has(object.ClassMonster) ||
		hooks.frame == nil || hooks.tickRate == nil || hooks.buffSelf == nil ||
		hooks.castOffensive == nil || hooks.castRelated == nil || hooks.distance == nil ||
		hooks.push == nil || hooks.pop == nil {
		return false
	}
	update := unit.UpdateDataMonster()
	head := update.AIStackHead()
	if head == nil || head.Type() != ai.ACTION_FIGHT {
		return false
	}

	if hooks.frame()-head.ArgU32(2) > 10*hooks.tickRate() {
		hooks.pop()
		return true
	}
	if update.CurrentEnemy != nil {
		head.Args[2] = uintptr(hooks.frame())
		target := update.CurrentEnemy
		if !monsterFightKillable528190(target) {
			hooks.pop()
			return true
		}
		if !unit.HasEnchant(ENCHANT_ANTI_MAGIC) {
			if hooks.buffSelf(unit) {
				return true
			}
			if hooks.castOffensive(unit, update.CurrentEnemy) {
				return true
			}
		}
		monsterFightSelectAttack531B40(unit, update.CurrentEnemy, hooks)
		return true
	}

	if monsterFightHoldPosition534710(unit) {
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
	item := hooks.push(ai.ACTION_MOVE_TO)
	if item != nil {
		item.SetArgs(head.ArgPos(0), uint32(0))
	}
	return true
}

// MonsterActionFightRuntime531EC0 supplies root-level services used by the
// original fight selector but owned outside the server package.
type MonsterActionFightRuntime531EC0 struct {
	Distance  func(*Object, *Object) float64
	CanSummon func(*Object, int) bool
}

// MonsterActionFight531EC0 binds the complete restored FIGHT branch to the
// live spell registry, duration list, logic RNG, object index, and action
// stack while preserving native-width pointers throughout.
func (s *Server) MonsterActionFight531EC0(unit *Object, runtime MonsterActionFightRuntime531EC0) bool {
	spellHooks := s.monsterFightSpellHooks540B90(runtime.CanSummon)
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
		buffSelf: func(unit *Object) bool {
			return monsterFightBuffSelf540B90(unit, spellHooks)
		},
		castOffensive: func(unit, target *Object) bool {
			return monsterFightCastOffensive540F20(unit, target, spellHooks)
		},
		castRelated: func(unit, target *Object) bool {
			return monsterFightCastRelated540D90(unit, target, spellHooks)
		},
		distance: runtime.Distance,
		push:     unit.MonsterPushAction,
		pop:      unit.MonsterPopAction,
	})
}
