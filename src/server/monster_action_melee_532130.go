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

// MonsterStrikeDefaultRuntime549380 supplies the two side effects performed
// by GAME.EXE's ordinary monster strike after target selection and tracing.
type MonsterStrikeDefaultRuntime549380 struct {
	Damage     func(target, source, attacker *Object, damage int, damageType object.DamageType) bool
	ApplyForce func(target *Object, origin types.Pointf, force float64)
}

type monsterStrikeDefaultHooks549380 struct {
	loadUpdate     func(*Object) *MonsterUpdateData
	pickTarget     func(*Object, bool) *Object
	loadUnitY      func(*Object) float32
	loadUnitX      func(*Object) float32
	loadTargetX    func(*Object) float32
	loadTargetY    func(*Object) float32
	trace          func(types.Pointf, types.Pointf, MapTraceFlags) int32
	loadMonsterDef func(*MonsterUpdateData) *MonsterDef
	loadDamageType func(*MonsterDef) uint32
	loadDamage     func(*MonsterDef) uint32
	damage         func(*Object, *Object, *Object, uint32, uint32)
	loadImpact     func(*MonsterDef) float32
	applyForce     func(*Object, *Object, float32)
}

// monsterStrikeDefault549380 restores GAME.EXE 00549380. The entry update
// pointer stays cached across target selection and tracing, while MonsterDef
// is read live once for damage/type and again for impact. Coordinate loads
// preserve the original Y/X/X/Y order. A missing target returns one, a failed
// trace returns zero, and impact force accepts only ordered strict-positive
// binary32 values.
func monsterStrikeDefault549380(unit *Object, hooks monsterStrikeDefaultHooks549380) int32 {
	update := hooks.loadUpdate(unit)
	target := hooks.pickTarget(unit, false)
	if target == nil {
		return 1
	}
	unitY := hooks.loadUnitY(unit)
	unitX := hooks.loadUnitX(unit)
	targetX := hooks.loadTargetX(target)
	targetY := hooks.loadTargetY(target)
	if hooks.trace(
		types.Pointf{X: unitX, Y: unitY},
		types.Pointf{X: targetX, Y: targetY},
		MapTraceFlags(5),
	) == 0 {
		return 0
	}

	def := hooks.loadMonsterDef(update)
	damageType := hooks.loadDamageType(def)
	damage := hooks.loadDamage(def)
	hooks.damage(target, unit, unit, damage, damageType)

	def = hooks.loadMonsterDef(update)
	impact := hooks.loadImpact(def)
	if impact > 0 {
		hooks.applyForce(unit, target, impact)
	}
	return 1
}

// MonsterStrikeDefault549380 binds the default strike to native-width object,
// update-data, MonsterDef, and Damage callback boundaries.
func (s *Server) MonsterStrikeDefault549380(unit *Object, runtime MonsterStrikeDefaultRuntime549380) int {
	return int(monsterStrikeDefault549380(unit, monsterStrikeDefaultHooks549380{
		loadUpdate: func(unit *Object) *MonsterUpdateData {
			return unit.UpdateDataMonster()
		},
		pickTarget: func(unit *Object, allowFriendly bool) *Object {
			return monsterPickMeleeTarget549440(unit, allowFriendly, monsterPickMeleeTargetHooks549440{
				eachInRect: s.Map.EachObjInRect,
				isEnemy:    s.IsEnemyTo,
			})
		},
		loadUnitY: func(unit *Object) float32 {
			return unit.PosVec.Y
		},
		loadUnitX: func(unit *Object) float32 {
			return unit.PosVec.X
		},
		loadTargetX: func(target *Object) float32 {
			return target.PosVec.X
		},
		loadTargetY: func(target *Object) float32 {
			return target.PosVec.Y
		},
		trace: func(from, to types.Pointf, flags MapTraceFlags) int32 {
			if s.MapTraceRay(from, to, flags) {
				return 1
			}
			return 0
		},
		loadMonsterDef: func(update *MonsterUpdateData) *MonsterDef {
			return update.MonsterDef
		},
		loadDamageType: func(def *MonsterDef) uint32 {
			return def.MeleeAttackDamageType124
		},
		loadDamage: func(def *MonsterDef) uint32 {
			return def.MeleeAttackDamage116
		},
		damage: func(target, source, attacker *Object, damage, damageType uint32) {
			if runtime.Damage != nil {
				runtime.Damage(target, source, attacker, int(int32(damage)), object.DamageType(damageType))
			}
		},
		loadImpact: func(def *MonsterDef) float32 {
			return def.MeleeAttackImpact120
		},
		applyForce: func(unit, target *Object, impact float32) {
			if runtime.ApplyForce != nil {
				runtime.ApplyForce(target, unit.PosVec, float64(impact))
			}
		},
	}))
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
