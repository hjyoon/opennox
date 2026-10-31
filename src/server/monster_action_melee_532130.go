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
	AudioEvent   func(uint32, *Object)
	BuffOff      func(*Object, EnchantID)
	PlayerAttack func(*Object) int
	CanStrike    func(unsafe.Pointer) bool
	Strike       func(*Object, unsafe.Pointer) int
}

type monsterActionMeleeStartHooks532130 struct {
	frame         func() uint32
	tickRate      func() uint32
	random        func(int, int) int
	audio         func(uint32, *Object)
	buffOff       func(*Object, EnchantID)
	weaponStamina func(uint32) int32
	eachInRect    func(types.Rectf, func(*Object) bool)
	isEnemy       func(*Object, *Object) bool
	pop           func() int
	push          func(ai.ActionType, ...any) *AIStackItem
}

const monsterMeleeFriendlyRadius532390 = float32(50)

// monsterMeleeFriendlyBlocker532390 restores the unit query used by NPC melee
// attacks. Unlike ordinary melee target selection, it deliberately considers
// both friends and enemies and keeps the closest living unit in a 60-degree
// forward cone. Allegiance is checked only after the closest unit is chosen.
func monsterMeleeFriendlyBlocker532390(unit *Object, eachInRect func(types.Rectf, func(*Object) bool)) *Object {
	if unit == nil || eachInRect == nil {
		return nil
	}
	var selected unsafe.Pointer
	distance := float32(51)
	state := MonsterMeleeBlockerState532390{Selected: &selected, Distance: &distance}
	eachInRect(monsterMeleeSearchRect532130(unit), func(candidate *Object) bool {
		if candidate != nil {
			MonsterMeleeBlocker532390(candidate, unit, state)
		}
		return true
	})
	return (*Object)(selected)
}

// monsterActionMeleeStart532130 restores GAME.EXE 00532130, including the
// NPC stamina and friendly-hammer avoidance path that cannot safely execute
// through the original PE32 callback on a native-width build.
func monsterActionMeleeStart532130(unit *Object, hooks monsterActionMeleeStartHooks532130) bool {
	if unit == nil || unit.UpdateData == nil || hooks.frame == nil || hooks.random == nil || hooks.push == nil {
		return false
	}
	monsterMeleeStartNative532130(unit, MonsterMeleeGlobalState532390(), hooks)
	return true
}

// MonsterActionMeleeStart532130 binds the restored attack start to native-width
// inventory, world-query, allegiance, and action-stack state.
func (s *Server) MonsterActionMeleeStart532130(unit *Object, runtime MonsterActionMeleeRuntime532130) bool {
	if unit == nil || unit.UpdateData == nil {
		return false
	}
	_, handled := s.monsterMeleeNative532130(0x532130, unit, MonsterMeleeGlobalState532390(), monsterMeleeRuntimeFromAction532130(runtime))
	return handled
}

type monsterActionMeleeUpdateHooks532440 struct {
	audio        func(uint32, *Object)
	pop          func() int
	playerAttack func(*Object) int
	canStrike    func(unsafe.Pointer) bool
	strike       func(*Object, unsafe.Pointer) int
}

// monsterActionMeleeUpdate532440 restores GAME.EXE 00532440. NPC attacks use
// the typed player-attack boundary, while ordinary monsters reject unsupported
// strike callback identities before changing state.
func monsterActionMeleeUpdate532440(unit *Object, hooks monsterActionMeleeUpdateHooks532440) bool {
	if unit == nil || unit.UpdateData == nil || hooks.pop == nil {
		return false
	}
	if uint8(unit.ObjSubClass)&16 == 0 && (*MonsterUpdateData)(unit.UpdateData).MonsterDef == nil {
		return false
	}
	_, handled := monsterMeleeUpdateNative532440(unit, hooks)
	return handled
}

// MonsterActionMeleeUpdate532440 binds the restored attack animation update
// to a native strike resolver supplied by the legacy compatibility boundary.
func (s *Server) MonsterActionMeleeUpdate532440(unit *Object, runtime MonsterActionMeleeRuntime532130) bool {
	if unit == nil || unit.UpdateData == nil {
		return false
	}
	_, handled := s.monsterMeleeNative532130(0x532440, unit, MonsterMeleeGlobalState532390(), monsterMeleeRuntimeFromAction532130(runtime))
	return handled
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
	// 00549459..00549495 retains the radius at precision 53 / ToZero;
	// only the four rectangle coordinates are stored as binary32.
	add := pointOnLineAddChop53_57C8A0
	mul := pointOnLineMulChop53_57C8A0
	div := pointOnLineDivChop53_57C8A0
	spill := pointOnLineSpill57C8A0
	searchRadius := add(add(float64(def.MeleeAttackRange112), float64(unit.Shape.Circle.R)), float64(monsterMeleeTargetPadding549440))
	rect := types.Rectf{
		Min: types.Ptf(spill(add(float64(unit.PosVec.X), -searchRadius)), spill(add(float64(unit.PosVec.Y), -searchRadius))),
		Max: types.Ptf(spill(add(searchRadius, float64(unit.PosVec.X))), spill(add(searchRadius, float64(unit.PosVec.Y)))),
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
		// 00549508 calls allegiance even when friendly targets are allowed.
		// Candidate positions are live loads after that callback.
		if !hooks.isEnemy(unit, candidate) && !allowFriendly {
			return true
		}
		// X remains in x87; Y and sqrt(square)+the original binary32
		// epsilon have FSTP stores. Do not round each delta/product/dot
		// to binary32 or change the calling thread's floating-point mode.
		dx := add(float64(candidate.PosVec.X), -float64(unit.PosVec.X))
		dy := float64(spill(add(float64(candidate.PosVec.Y), -float64(unit.PosVec.Y))))
		square := add(mul(dy, dy), mul(dx, dx))
		distance := float64(spill(add(monsterMoveForceSqrtChop53_50D50C(square), float64(math.Float32frombits(0x3a83126f)))))
		facing := unit.Direction1.Vec()
		dot := add(mul(div(dy, distance), float64(facing.Y)), mul(div(dx, distance), float64(facing.X)))
		if !(dot > 0.5) {
			return true
		}
		edgeDistance := add(distance, -add(float64(candidate.Shape.Circle.R), float64(unit.Shape.Circle.R)))
		// 00549590 tests C0 alone: an unordered edge comparison is also
		// accepted. The cone above instead requires ordered strict >.
		if !(edgeDistance >= float64(bestRange)) {
			bestRange = spill(edgeDistance)
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
