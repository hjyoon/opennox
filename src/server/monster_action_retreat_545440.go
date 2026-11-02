package server

import (
	"math"

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
	state := MonsterEdibleSearchState544A00{BestDistance: monsterSearchEdibleMaxDistance544A00}
	hooks.eachInCircle(unit.PosVec, radius, func(candidate *Object) bool {
		if candidate != nil {
			monsterSearchEdibleCandidate544A40(candidate, unit, &state, hooks.canInteract, func() bool { return hooks.online })
		}
		return true
	})
	return state.BestObject
}

// The raw callback reads live fields and shared scratch after world services.
// Its direct field reads retain the original null-input fault prefix.
func monsterSearchEdibleCandidate544A40(candidate, unit *Object, state *MonsterEdibleSearchState544A00, interact func(*Object, *Object, int) bool, online func() bool) {
	if !candidate.ObjClass.Has(object.ClassFood) {
		return
	}
	food := candidate.ObjSubClass.AsFood()
	if food.Has(object.FoodJug) || food.Has(object.FoodMushroom) && unit.Poison540 == 0 {
		return
	}
	food = candidate.ObjSubClass.AsFood()
	if food.Has(object.FoodPotion) && !(food.Has(object.FoodHealthPotion) && online() && unit.ObjSubClass.AsMonster().Has(object.MonsterNPC)) {
		return
	}
	if !interact(unit, candidate, 0) {
		return
	}
	// Original PC53 ToZero retains intermediates through FCOM and spills only
	// the winning distance. C0 alone also replaces on an unordered comparison.
	dx := monsterEdibleSubtract544A40(float64(candidate.PosVec.X), float64(unit.PosVec.X))
	dy := monsterEdibleSubtract544A40(float64(candidate.PosVec.Y), float64(unit.PosVec.Y))
	ySquared := monsterMoveToRunSquareChop53_544434(dy)
	xSquared := monsterMoveToRunSquareChop53_544434(dx)
	distance := monsterMoveToRunAddChop53_544434(ySquared, xSquared)
	if !(distance >= float64(state.BestDistance)) {
		state.BestDistance = float32(monsterMoveToRunSpill544440(distance))
		state.BestObject = candidate
	}
}

// x87 produces the negative indefinite NaN for infinity minus itself.
// Preserve operand NaN payloads through the normal arithmetic path.
func monsterEdibleSubtract544A40(a, b float64) float64 {
	if math.IsInf(a, 0) && a == b {
		return math.Float64frombits(0xfff8000000000000)
	}
	return monsterMoveToRunAddChop53_544434(a, -b)
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
	return monsterCanResumeAttackData545520(unit.HealthData, unit.UpdateDataMonster())
}

func monsterRetreatCheckEdibles5455E0(unit *Object, hooks monsterActionRetreatHooks545440) *AIStackItem {
	radius := monsterRetreatFoodRangeDefault5455E0
	if hooks.quest {
		radius = monsterRetreatFoodRangeQuest5455E0
	}
	food := hooks.searchFood(unit, radius)
	if food != nil {
		hooks.push(ai.DEPENDENCY_NOT_HEALTHY)
		hooks.push(ai.DEPENDENCY_NO_VISIBLE_ENEMY)
		if visible := hooks.push(ai.DEPENDENCY_OBJECT_AT_VISIBLE_LOCATION); visible != nil {
			// Push may invoke Cancel. Read each live coordinate only after
			// acceptance, then store the cached native food identity.
			visible.Args[0] = uintptr(math.Float32bits(food.PosVec.X))
			visible.Args[1] = uintptr(math.Float32bits(food.PosVec.Y))
			visible.Args[2] = uintptr(food.CObj())
		}
		if pickup := hooks.push(ai.ACTION_PICKUP_OBJECT); pickup != nil {
			pickup.Args[0] = uintptr(food.CObj())
		}
		move := hooks.push(ai.ACTION_MOVE_TO)
		if move != nil {
			move.Args[0] = uintptr(math.Float32bits(food.PosVec.X))
			y := uintptr(math.Float32bits(food.PosVec.Y))
			move.Args[2] = uintptr(food.CObj())
			move.Args[1] = y // 0054566A stores the pointer before the saved Y.
		}
		return move
	}
	hooks.push(ai.DEPENDENCY_NOT_HEALTHY)
	hooks.push(ai.DEPENDENCY_NO_VISIBLE_ENEMY)
	hooks.push(ai.DEPENDENCY_NO_VISIBLE_FOOD)
	roam := hooks.push(ai.ACTION_ROAM)
	if roam != nil {
		roam.Args[0] = 0
		// 0054569E writes a BYTE, not a sign-extended direction DWORD.
		roam.Args[2] = roam.Args[2]&^uintptr(0xff) | 0x80
	}
	return roam
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
	antiMagicCaster := monsterCanCast2Data5455B0(unit, update)
	if canResume && !antiMagicCaster {
		hooks.pop()
		return
	}
	if update.CurrentEnemy != nil {
		castRelated := false
		if !unit.HasEnchant(ENCHANT_ANTI_MAGIC) && hooks.castRelated != nil {
			castRelated = hooks.castRelated(unit)
		}
		if !castRelated {
			if time := hooks.push(ai.DEPENDENCY_TIME); time != nil {
				// The original reads FPS once after the successful push and
				// passes wrapped signed DWORD bounds to RNG, then reads frame.
				fps := hooks.tickRate()
				delay := hooks.random(int(int32(4*fps)), int(int32(6*fps)))
				time.SetArgs(hooks.frame() + uint32(delay))
			}
			if flee := hooks.push(ai.ACTION_FLEE); flee != nil {
				// Push may invoke Cancel. Reload from the entry-cached update,
				// only after acceptance, preserving the raw position DWORDs.
				flee.SetArgs(update.CurrentEnemy.PosVec, uint32(0))
			}
		}
		return
	}
	if !canResume {
		monsterRetreatCheckEdibles5455E0(unit, hooks)
	}
}

// MonsterActionRetreat545440 binds the native-width retreat action and the
// original 00541050 related-spell selector to the live server.
func (s *Server) MonsterActionRetreat545440(unit *Object) {
	monsterActionRetreat545440(unit, monsterActionRetreatHooks545440{
		frame:    s.Frame,
		tickRate: s.TickRate,
		random:   s.Rand.Logic.IntClamp,
		castRelated: func(unit *Object) bool {
			return monsterFleeCastRelated541050(unit, s.monsterFightSpellHooks540B90(nil))
		},
		searchFood: s.MonsterSearchEdible544A00,
		quest:      noxflags.HasGame(noxflags.GameModeQuest),
		push:       unit.MonsterPushAction,
		pop:        unit.MonsterPopAction,
	})
}
