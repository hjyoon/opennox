package server

import (
	"math"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/strman"
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/common/unit/ai"
)

// These identities describe GAME.EXE's strike callbacks, not monster names.
// Several stock types share a callback (e.g. Ogre and OgreBrute).
type MonsterStrikeKind549220 uint8

const (
	MonsterStrikeDefault549220 MonsterStrikeKind549220 = iota
	MonsterStrikeSpider549220
	MonsterStrikeOgre549220
	MonsterStrikeScorpion549220
	MonsterStrikeZombie549220
	MonsterStrikeStoneGolem549220
	MonsterStrikeMechGolem549220
	MonsterStrikeWasp549220
	MonsterStrikeGhost549220
	MonsterStrikeBomber549220
)

type MonsterStrikeSpecialRuntime549220 struct {
	MonsterStrikeDefaultRuntime549380
	Direction       func(Dir16) types.Pointf
	Facing          func(types.Pointf, int16, types.Pointf) int
	Distance        func(*Object, *Object) float32
	ActivatePoison  func(*Object, int32, int32) int32
	PriorityMessage func(*Object, strman.ID, byte)
	BuffApply       func(*Object, EnchantID, int, int)
	SetAreaResult   func(golem bool, value uint32)
	AreaResult      func(golem bool) uint32
	SetGolemMode    func(uint32)
}

type monsterStrikeSpecialHooks549220 struct {
	MonsterStrikeSpecialRuntime549220
	pick     func(*Object) *Object
	trace    func(types.Pointf, types.Pointf, MapTraceFlags) bool
	each     func(types.Pointf, float32, func(*Object) bool)
	quake    func(types.Pointf, int)
	random   func(int, int) int
	tickRate func() uint32
	frame    func() uint32
}

func monsterStrikePoison549690(unit, target *Object, h monsterStrikeSpecialHooks549220, message strman.ID) {
	// The definition pointer is cached BEFORE the random callback; its chance,
	// strength and maximum are loaded live afterward (00549690).
	def := unit.UpdateDataMonster().MonsterDef
	if def.MeleeAttackPoisonChange136 == 0 {
		return
	}
	roll := h.random(1, 100)
	if int32(roll) <= int32(def.MeleeAttackPoisonChange136) &&
		h.ActivatePoison(target, int32(def.MeleeAttackPoisonStrength140), int32(def.MeleeAttackPoisonMax144)) != 0 {
		h.PriorityMessage(target, message, 0)
	}
}

func monsterStrikeSpecialDamage549220(unit, target *Object, update *MonsterUpdateData, h monsterStrikeSpecialHooks549220) {
	def := update.MonsterDef
	typ, damage := def.MeleeAttackDamageType124, def.MeleeAttackDamage116
	h.Damage(target, unit, unit, int(int32(damage)), object.DamageType(typ))
}

func monsterStrikeSpecialForce549220(unit, target *Object, update *MonsterUpdateData, h monsterStrikeSpecialHooks549220, unconditional bool) {
	impact := update.MonsterDef.MeleeAttackImpact120
	if unconditional || impact > 0 {
		h.ApplyForce(target, unit.PosVec, float64(impact))
	}
}

func monsterStrikeSpecial549220(unit *Object, kind MonsterStrikeKind549220, h monsterStrikeSpecialHooks549220) int {
	if kind == MonsterStrikeBomber549220 {
		return 1 // Bomber's damage belongs to its separate suicide action.
	}
	if kind == MonsterStrikeOgre549220 || kind == MonsterStrikeStoneGolem549220 || kind == MonsterStrikeMechGolem549220 {
		golem := kind != MonsterStrikeOgre549220
		if golem {
			mode := uint32(0)
			if kind == MonsterStrikeMechGolem549220 {
				mode = 1
			}
			h.SetGolemMode(mode)
		}
		update := unit.UpdateDataMonster()
		radius := float32(float64(update.MonsterDef.MeleeAttackRange112) + float64(unit.Shape.Circle.R) + 30)
		h.SetAreaResult(golem, 0)
		h.each(unit.PosVec, radius, func(target *Object) bool {
			cachedUpdate := unit.UpdateDataMonster()
			if target == unit {
				return true
			}
			if golem {
				// x87 tests C0|C3: greater rejects, unordered/equal accept.
				if h.Facing(unit.PosVec, int16(unit.Direction1), target.PosVec)&1 == 0 ||
					float64(h.Distance(unit, target)) > float64(cachedUpdate.MonsterDef.MeleeAttackRange112) {
					return true
				}
			} else {
				dx := float64(target.PosVec.X) - float64(unit.PosVec.X)
				dy := float64(target.PosVec.Y) - float64(unit.PosVec.Y)
				storedY := float32(dy)
				distance := float32(math.Sqrt(dy*float64(storedY)+dx*dx) + float64(float32(0.01)))
				edge := float64(distance) - (float64(target.Shape.Circle.R) + float64(unit.Shape.Circle.R))
				// The PE x87 range branch accepts unordered; the facing branch
				// then rejects unordered. No allegiance filter exists here.
				if edge > float64(cachedUpdate.MonsterDef.MeleeAttackRange112) {
					return true
				}
				facing := h.Direction(unit.Direction1)
				dot := float64(storedY)/float64(distance)*float64(facing.Y) + dx/float64(distance)*float64(facing.X)
				if !(dot > float64(float32(0.4))) {
					return true
				}
			}
			if !h.trace(unit.PosVec, target.PosVec, MapTraceFlags(5)) {
				return true
			}
			monsterStrikeSpecialDamage549220(unit, target, cachedUpdate, h)
			if golem && target.Class().HasAny(object.ClassPlayer|object.ClassMonster) {
				h.SetAreaResult(true, 1)
			}
			monsterStrikeSpecialForce549220(unit, target, cachedUpdate, h, !golem)
			if !golem {
				h.SetAreaResult(false, 1)
			}
			return true
		})
		if golem {
			h.quake(unit.PosVec, 30)
		}
		return int(h.AreaResult(golem))
	}

	cachedUpdate := unit.UpdateDataMonster()
	target := h.pick(unit)
	if target == nil {
		if kind == MonsterStrikeWasp549220 {
			return 0
		}
		return 1
	}
	if !h.trace(unit.PosVec, target.PosVec, MapTraceFlags(5)) {
		return 0
	}
	monsterStrikeSpecialDamage549220(unit, target, cachedUpdate, h)
	if kind == MonsterStrikeWasp549220 {
		monsterStrikePoison549690(unit, target, h, "aifunc.c:PoisonedByWasp")
	}
	monsterStrikeSpecialForce549220(unit, target, cachedUpdate, h, false)
	switch kind {
	case MonsterStrikeScorpion549220:
		monsterStrikePoison549690(unit, target, h, "aifunc.c:PoisonedByScorpion")
	case MonsterStrikeZombie549220:
		monsterStrikePoison549690(unit, target, h, "aifunc.c:PoisonedByZombie")
	case MonsterStrikeGhost549220:
		h.BuffApply(target, EnchantID(5), int(2*h.tickRate()), 3)
		if action := unit.MonsterPushAction(ai.ACTION_FACE_LOCATION); action != nil {
			action.SetArgs(target.PosVec)
		}
		if action := unit.MonsterPushAction(ai.DEPENDENCY_TIME); action != nil {
			fps := h.tickRate()
			roll := h.random(int(int32(2*fps)), int(int32(4*fps)))
			action.SetArgs(h.frame() + uint32(roll))
		}
		if action := unit.MonsterPushAction(ai.ACTION_FLEE); action != nil {
			action.SetArgs(target.PosVec, uint32(0))
		}
	}
	return 1
}

// MonsterStrikeSpecial549220 restores the distinct area, poison, and Ghost
// attack tails with native-width objects. It does not change AI timing.
func (s *Server) MonsterStrikeSpecial549220(unit *Object, kind MonsterStrikeKind549220, runtime MonsterStrikeSpecialRuntime549220) int {
	if kind == MonsterStrikeDefault549220 {
		return s.MonsterStrikeDefault549380(unit, runtime.MonsterStrikeDefaultRuntime549380)
	}
	if kind == MonsterStrikeSpider549220 {
		return s.MonsterStrikeSpider549BC0(unit, MonsterStrikeSpiderRuntime549BC0{
			Damage: runtime.Damage, ApplyForce: runtime.ApplyForce,
			ActivatePoison: runtime.ActivatePoison, PriorityMessage: runtime.PriorityMessage,
		})
	}
	return monsterStrikeSpecial549220(unit, kind, monsterStrikeSpecialHooks549220{
		MonsterStrikeSpecialRuntime549220: runtime,
		pick: func(unit *Object) *Object {
			return monsterPickMeleeTarget549440(unit, false, monsterPickMeleeTargetHooks549440{eachInRect: s.Map.EachObjInRect, isEnemy: s.IsEnemyTo})
		},
		trace:  s.MapTraceRay,
		each:   s.Map.EachObjInCircle,
		quake:  s.Nox_xxx_earthquakeSend_4D9110,
		random: s.Rand.Logic.IntClamp, tickRate: s.TickRate, frame: s.Frame,
	})
}
