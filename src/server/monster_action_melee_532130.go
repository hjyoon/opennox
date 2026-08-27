package server

import (
	"math"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/strman"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/unit/ai"
)

// MonsterActionMeleeRuntime532130 supplies services whose implementations
// still live above the server package. CanStrike must reject a legacy strike
// callback unless Strike has a native-width implementation for it.
type MonsterActionMeleeRuntime532130 struct {
	AudioEvent func(uint32, *Object)
	BuffOff    func(*Object, EnchantID)
	CanStrike  func(unsafe.Pointer) bool
	Strike     func(*Object, unsafe.Pointer) int
}

type monsterActionMeleeStartHooks532130 struct {
	frame   func() uint32
	random  func(int, int) int
	audio   func(uint32, *Object)
	buffOff func(*Object, EnchantID)
	push    func(ai.ActionType, ...any) *AIStackItem
}

// monsterActionMeleeStart532130 restores the non-NPC branch of GAME.EXE
// 00532130. The NPC stamina/friendly-fire branch remains a separate admission
// gate because it requires equipped-item state and its own avoidance actions.
func monsterActionMeleeStart532130(unit *Object, hooks monsterActionMeleeStartHooks532130) bool {
	if unit == nil || unit.UpdateData == nil || !unit.Class().Has(object.ClassMonster) ||
		hooks.frame == nil || hooks.random == nil || hooks.push == nil {
		return false
	}
	if unit.SubClass().Has(object.SubClass(object.MonsterNPC)) {
		return false
	}
	update := unit.UpdateDataMonster()
	def := update.MonsterDef
	if def == nil {
		return false
	}
	frame := hooks.frame()
	if frame >= update.Field128 {
		if hooks.buffOff != nil {
			hooks.buffOff(unit, EnchantID(0))
			hooks.buffOff(unit, EnchantID(23))
		}
		unit.Field34 = frame
		update.Field128 = frame + uint32(hooks.random(int(def.MeleeAttackDelayMin128), int(def.MeleeAttackDelayMax132)))
		if hooks.audio != nil && update.SoundSet122 != nil {
			hooks.audio(*(*uint32)(unsafe.Add(update.SoundSet122, 6*4)), unit)
		}
		return true
	}
	hooks.push(ai.DEPENDENCY_OBJECT_CLOSER_THAN, def.MeleeAttackRange112*1.2, uint32(0), update.CurrentEnemy)
	hooks.push(ai.ACTION_WAIT, update.Field128)
	return true
}

// MonsterActionMeleeStart532130 binds the restored non-NPC attack start to
// the live frame clock, RNG, and action stack.
func (s *Server) MonsterActionMeleeStart532130(unit *Object, runtime MonsterActionMeleeRuntime532130) bool {
	return monsterActionMeleeStart532130(unit, monsterActionMeleeStartHooks532130{
		frame:   s.Frame,
		random:  s.Rand.Logic.IntClamp,
		audio:   runtime.AudioEvent,
		buffOff: runtime.BuffOff,
		push:    unit.MonsterPushAction,
	})
}

type monsterActionMeleeUpdateHooks532440 struct {
	audio     func(uint32, *Object)
	pop       func() int
	canStrike func(unsafe.Pointer) bool
	strike    func(*Object, unsafe.Pointer) int
}

// monsterActionMeleeUpdate532440 restores the non-NPC branch of GAME.EXE
// 00532440. It rejects unsupported strike callback identities before changing
// state, preventing a native-width object from entering a PE32 callback.
func monsterActionMeleeUpdate532440(unit *Object, hooks monsterActionMeleeUpdateHooks532440) bool {
	if unit == nil || unit.UpdateData == nil || !unit.Class().Has(object.ClassMonster) || hooks.pop == nil {
		return false
	}
	if unit.SubClass().Has(object.SubClass(object.MonsterNPC)) {
		return false
	}
	update := unit.UpdateDataMonster()
	def := update.MonsterDef
	if def == nil {
		return false
	}
	strike := def.MeleeStrikeFunc236
	if strike == nil {
		hooks.pop()
		return true
	}
	if hooks.canStrike == nil || !hooks.canStrike(strike) || hooks.strike == nil {
		return false
	}
	if uint32(update.Field120_1) == def.MeleeAttackFrame108 && update.Field120_2 == 0 {
		hit := hooks.strike(unit, strike)
		if hooks.audio != nil && update.SoundSet122 != nil {
			index := uintptr(9)
			if hit != 0 {
				index = 8
			}
			hooks.audio(*(*uint32)(unsafe.Add(update.SoundSet122, index*4)), unit)
		}
	}
	if update.Field120_3 != 0 {
		hooks.pop()
	}
	return true
}

// MonsterActionMeleeUpdate532440 binds the restored attack animation update
// to a native strike resolver supplied by the legacy compatibility boundary.
func (s *Server) MonsterActionMeleeUpdate532440(unit *Object, runtime MonsterActionMeleeRuntime532130) bool {
	return monsterActionMeleeUpdate532440(unit, monsterActionMeleeUpdateHooks532440{
		audio:     runtime.AudioEvent,
		pop:       unit.MonsterPopAction,
		canStrike: runtime.CanStrike,
		strike:    runtime.Strike,
	})
}

type MonsterStrikeSpiderRuntime549BC0 struct {
	Damage          func(target, source, attacker *Object, damage int, damageType object.DamageType) bool
	ApplyForce      func(target *Object, origin types.Pointf, force float64)
	ActivatePoison  func(target *Object, increment, maximum int32) int32
	PriorityMessage func(*Object, strman.ID, byte)
}

type monsterPickMeleeTargetHooks549440 struct {
	eachInRect func(types.Rectf, func(*Object) bool)
	isEnemy    func(*Object, *Object) bool
}

type monsterStrikeSpiderHooks549BC0 struct {
	pickTarget      func(*Object, bool) *Object
	trace           func(types.Pointf, types.Pointf, MapTraceFlags) bool
	damage          func(target, source, attacker *Object, damage int, damageType object.DamageType) bool
	applyForce      func(target *Object, origin types.Pointf, force float64)
	random          func(int, int) int
	activatePoison  func(target *Object, increment, maximum int32) int32
	priorityMessage func(*Object, strman.ID, byte)
}

const monsterMeleeTargetPadding549440 = float32(30)

func monsterPickMeleeTarget549440(unit *Object, allowFriendly bool, hooks monsterPickMeleeTargetHooks549440) *Object {
	if unit == nil || unit.UpdateData == nil || hooks.eachInRect == nil || hooks.isEnemy == nil {
		return nil
	}
	def := unit.UpdateDataMonster().MonsterDef
	if def == nil {
		return nil
	}
	searchRadius := def.MeleeAttackRange112 + unit.Shape.Circle.R + monsterMeleeTargetPadding549440
	rect := types.Rectf{
		Min: unit.PosVec.Sub(types.Ptf(searchRadius, searchRadius)),
		Max: unit.PosVec.Add(types.Ptf(searchRadius, searchRadius)),
	}
	bestRange := def.MeleeAttackRange112
	var best *Object
	hooks.eachInRect(rect, func(candidate *Object) bool {
		if candidate == nil || candidate == unit {
			return true
		}
		if uint8(candidate.ObjFlags)&0x11 != 0 && uint8(candidate.ObjClass)&0x6 == 0 {
			return true
		}
		if !candidate.Class().HasAny(object.ClassMonster|object.ClassPlayer) &&
			(candidate.HealthData == nil || candidate.HealthData.Max == 0) {
			return true
		}
		if !allowFriendly && !hooks.isEnemy(unit, candidate) {
			return true
		}
		delta := candidate.PosVec.Sub(unit.PosVec)
		distance := float32(math.Sqrt(float64(delta.X*delta.X+delta.Y*delta.Y))) + 0.001
		facing := unit.Direction1.Vec()
		if delta.Y/distance*facing.Y+delta.X/distance*facing.X <= 0.5 {
			return true
		}
		edgeDistance := distance - (candidate.Shape.Circle.R + unit.Shape.Circle.R)
		if edgeDistance < bestRange {
			bestRange = edgeDistance
			best = candidate
		}
		return true
	})
	return best
}

func monsterStrikeSpider549BC0(unit *Object, hooks monsterStrikeSpiderHooks549BC0) int {
	if unit == nil || unit.UpdateData == nil || !unit.Class().Has(object.ClassMonster) {
		return 0
	}
	update := unit.UpdateDataMonster()
	def := update.MonsterDef
	if def == nil {
		return 0
	}
	if hooks.pickTarget == nil || hooks.trace == nil {
		return 0
	}
	target := hooks.pickTarget(unit, false)
	if target == nil || !hooks.trace(unit.PosVec, target.PosVec, MapTraceFlags(5)) {
		return 0
	}
	if hooks.damage != nil {
		hooks.damage(target, unit, unit, int(def.MeleeAttackDamage116), object.DamageType(def.MeleeAttackDamageType124))
	}
	if def.MeleeAttackImpact120 > 0 && hooks.applyForce != nil {
		hooks.applyForce(target, unit.PosVec, float64(def.MeleeAttackImpact120))
	}
	if def.MeleeAttackPoisonChange136 != 0 && hooks.random != nil && hooks.activatePoison != nil &&
		hooks.random(1, 100) <= int(def.MeleeAttackPoisonChange136) &&
		hooks.activatePoison(target, int32(def.MeleeAttackPoisonStrength140), int32(def.MeleeAttackPoisonMax144)) != 0 &&
		hooks.priorityMessage != nil {
		hooks.priorityMessage(target, strman.ID("aifunc.c:Poisoned"), 0)
	}
	return 1
}

// MonsterStrikeSpider549BC0 restores the Spider/SpittingSpider melee strike
// shared by GAME.EXE 00549BC0 and 00549CA0.
func (s *Server) MonsterStrikeSpider549BC0(unit *Object, runtime MonsterStrikeSpiderRuntime549BC0) int {
	return monsterStrikeSpider549BC0(unit, monsterStrikeSpiderHooks549BC0{
		pickTarget: func(unit *Object, allowFriendly bool) *Object {
			return monsterPickMeleeTarget549440(unit, allowFriendly, monsterPickMeleeTargetHooks549440{
				eachInRect: s.Map.EachObjInRect,
				isEnemy:    s.IsEnemyTo,
			})
		},
		trace:           s.MapTraceRay,
		damage:          runtime.Damage,
		applyForce:      runtime.ApplyForce,
		random:          s.Rand.Logic.IntClamp,
		activatePoison:  runtime.ActivatePoison,
		priorityMessage: runtime.PriorityMessage,
	})
}
