package server

import (
	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/unit/ai"
)

const (
	monsterRetreatFoodRangeQuest5455E0   = float32(640)
	monsterRetreatFoodRangeDefault5455E0 = float32(250)
	monsterSearchEdibleMaxDistance544A00 = float32(10000000)
)

type monsterSearchEdibleHooks544A00 struct {
	eachInCircle func(types.Pointf, float32, func(*Object) bool)
	canInteract  func(*Object, *Object, int) bool
	online       bool
}

// monsterSearchEdible544A00 restores GAME.EXE 00544A00 and 00544A40. The
// original stored the winning object in a PE32 global; returning *Object keeps
// the same selection semantics without truncating a native pointer.
func monsterSearchEdible544A00(unit *Object, radius float32, hooks monsterSearchEdibleHooks544A00) *Object {
	if unit == nil || hooks.eachInCircle == nil || hooks.canInteract == nil {
		return nil
	}
	nearestDistance := monsterSearchEdibleMaxDistance544A00
	var nearest *Object
	hooks.eachInCircle(unit.PosVec, radius, func(candidate *Object) bool {
		if candidate == nil || !candidate.ObjClass.Has(object.ClassFood) {
			return true
		}
		food := candidate.ObjSubClass.AsFood()
		if food.Has(object.FoodJug) || food.Has(object.FoodMushroom) && unit.Poison540 == 0 {
			return true
		}
		if food.Has(object.FoodPotion) &&
			!(food.Has(object.FoodHealthPotion) && hooks.online && unit.ObjSubClass.AsMonster().Has(object.MonsterNPC)) {
			return true
		}
		if !hooks.canInteract(unit, candidate, 0) {
			return true
		}
		delta := candidate.PosVec.Sub(unit.PosVec)
		distance := delta.X*delta.X + delta.Y*delta.Y
		if distance < nearestDistance {
			nearestDistance = distance
			nearest = candidate
		}
		return true
	})
	return nearest
}

// MonsterSearchEdible544A00 binds the native-width edible search to the live
// spatial index and visibility service.
func (s *Server) MonsterSearchEdible544A00(unit *Object, radius float32) *Object {
	return monsterSearchEdible544A00(unit, radius, monsterSearchEdibleHooks544A00{
		eachInCircle: s.Map.EachObjInCircle,
		canInteract:  s.CanInteract,
		online:       noxflags.HasGame(noxflags.GameOnline),
	})
}

type monsterActionRetreatHooks545440 struct {
	frame       func() uint32
	tickRate    func() uint32
	random      func(int, int) int
	castRelated func(*Object) bool
	searchFood  func(*Object, float32) *Object
	quest       bool
	push        func(ai.ActionType, ...any) *AIStackItem
	pop         func() int
}

func monsterCanResumeAttack545520(unit *Object) bool {
	if unit == nil || unit.UpdateData == nil || unit.HealthData == nil {
		return false
	}
	health := float64(1)
	if unit.HealthData.Max != 0 {
		health = float64(unit.HealthData.Cur) / float64(unit.HealthData.Max)
	}
	return health >= float64(unit.UpdateDataMonster().ResumeLevel)
}

func monsterRetreatCheckEdibles5455E0(unit *Object, hooks monsterActionRetreatHooks545440) {
	radius := monsterRetreatFoodRangeDefault5455E0
	if hooks.quest {
		radius = monsterRetreatFoodRangeQuest5455E0
	}
	food := hooks.searchFood(unit, radius)
	if food != nil {
		hooks.push(ai.DEPENDENCY_NOT_HEALTHY)
		hooks.push(ai.DEPENDENCY_NO_VISIBLE_ENEMY)
		hooks.push(ai.DEPENDENCY_OBJECT_AT_VISIBLE_LOCATION, food.PosVec, food)
		hooks.push(ai.ACTION_PICKUP_OBJECT, food)
		hooks.push(ai.ACTION_MOVE_TO, food.PosVec, food)
		return
	}
	hooks.push(ai.DEPENDENCY_NOT_HEALTHY)
	hooks.push(ai.DEPENDENCY_NO_VISIBLE_ENEMY)
	hooks.push(ai.DEPENDENCY_NO_VISIBLE_FOOD)
	hooks.push(ai.ACTION_ROAM, uint32(0), uint32(0), int32(-128))
}

// monsterActionRetreat545440 restores GAME.EXE 00545440 through 005455E0.
// Engine calls are hooks so the PE32 branch and stack order can be tested
// independently of native pointer width.
func monsterActionRetreat545440(unit *Object, hooks monsterActionRetreatHooks545440) {
	if unit == nil || unit.UpdateData == nil || unit.HealthData == nil || hooks.push == nil || hooks.pop == nil {
		return
	}
	update := unit.UpdateDataMonster()
	canResume := monsterCanResumeAttack545520(unit)
	antiMagicCaster := update.StatusFlags.Has(object.MonStatusCanCastSpells) && unit.HasEnchant(ENCHANT_ANTI_MAGIC)
	if canResume && !antiMagicCaster {
		hooks.pop()
		return
	}
	if enemy := update.CurrentEnemy; enemy != nil {
		castRelated := false
		if !unit.HasEnchant(ENCHANT_ANTI_MAGIC) && hooks.castRelated != nil {
			castRelated = hooks.castRelated(unit)
		}
		if !castRelated {
			delay := hooks.random(4*int(hooks.tickRate()), 6*int(hooks.tickRate()))
			hooks.push(ai.DEPENDENCY_TIME, hooks.frame()+uint32(delay))
			hooks.push(ai.ACTION_FLEE, enemy.PosVec, uint32(0))
		}
		return
	}
	if !canResume {
		monsterRetreatCheckEdibles5455E0(unit, hooks)
	}
}

// MonsterActionRetreat545440 binds the native-width retreat action to the
// live server. The self-buff branch of 00541050 is conservatively treated as
// unavailable here; non-casters (including the War01A retreating creatures)
// therefore match the original path exactly.
func (s *Server) MonsterActionRetreat545440(unit *Object) {
	monsterActionRetreat545440(unit, monsterActionRetreatHooks545440{
		frame:       s.Frame,
		tickRate:    s.TickRate,
		random:      s.Rand.Logic.IntClamp,
		castRelated: func(*Object) bool { return false },
		searchFood:  s.MonsterSearchEdible544A00,
		quest:       noxflags.HasGame(noxflags.GameModeQuest),
		push:        unit.MonsterPushAction,
		pop:         unit.MonsterPopAction,
	})
}
