package server

import (
	"math"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/unit/ai"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

// MonsterActionMissileRuntime532540 supplies engine services that still live
// above the server package while keeping every object pointer native-width.
type MonsterActionMissileRuntime532540 struct {
	AudioEvent     func(uint32, *Object)
	BuffOff        func(*Object, EnchantID)
	PlayerAttack   func(*Object) int
	CreateObjectAt func(*Object, *Object, types.Pointf)
	DelayedDelete  func(*Object)
}

type monsterActionMissileStartHooks532540 struct {
	frame   func() uint32
	random  func(int, int) int
	audio   func(uint32, *Object)
	buffOff func(*Object, EnchantID)
	push    func(ai.ActionType, ...any) *AIStackItem
}

// monsterActionMissileStart532540 restores GAME.EXE 00532540. In particular,
// it preserves the three independent frame reads and the original order of
// the cooldown RNG call between the second and third reads.
func monsterActionMissileStart532540(unit *Object, hooks monsterActionMissileStartHooks532540) bool {
	if unit == nil || unit.UpdateData == nil || !unit.Class().Has(object.ClassMonster) ||
		hooks.frame == nil || hooks.random == nil || hooks.push == nil {
		return false
	}
	update := unit.UpdateDataMonster()
	def := update.MonsterDef
	if def == nil {
		return false
	}
	if hooks.frame() >= update.Field128 {
		if hooks.buffOff != nil {
			hooks.buffOff(unit, EnchantID(0))
			hooks.buffOff(unit, EnchantID(23))
		}
		unit.Field34 = hooks.frame()
		delay := hooks.random(int(def.MissileAttackDelayMin220), int(def.MissileAttackDelayMax224))
		update.Field128 = hooks.frame() + uint32(delay)
		if hooks.audio != nil && update.SoundSet122 != nil {
			hooks.audio(*(*uint32)(unsafe.Add(update.SoundSet122, 10*4)), unit)
		}
		return true
	}
	hooks.push(ai.DEPENDENCY_OBJECT_CLOSER_THAN, def.MissileAttackRange212*1.2, uint32(0), update.CurrentEnemy)
	hooks.push(ai.ACTION_WAIT, update.Field128)
	return true
}

// MonsterActionMissileStart532540 binds the restored start action to the live
// frame clock, deterministic logic RNG, and native-width action stack.
func (s *Server) MonsterActionMissileStart532540(unit *Object, runtime MonsterActionMissileRuntime532540) bool {
	return monsterActionMissileStart532540(unit, monsterActionMissileStartHooks532540{
		frame:   s.Frame,
		random:  s.Rand.Logic.IntClamp,
		audio:   runtime.AudioEvent,
		buffOff: runtime.BuffOff,
		push:    unit.MonsterPushAction,
	})
}

type monsterActionMissileUpdateHooks532610 struct {
	audio         func(uint32, *Object)
	playerAttack  func(*Object) int
	newObject     func(string) *Object
	trace         func(types.Pointf, types.Pointf, MapTraceFlags) bool
	createAt      func(*Object, *Object, types.Pointf)
	delayedDelete func(*Object)
	pop           func() int
}

// monsterProjectileLead533080 restores GAME.EXE 00533080. Inputs originate as
// binary32 values, while the original x87 computation retains extended
// intermediates until each output coordinate is stored back to binary32.
func monsterProjectileLead533080(unit, target *Object, speed float32) types.Pointf {
	dx := float64(target.PosVec.X) - float64(unit.PosVec.X)
	dy := float64(target.PosVec.Y) - float64(unit.PosVec.Y)
	travel := math.Sqrt(dx*dx+dy*dy) / float64(speed)
	return types.Pointf{
		X: float32(travel*float64(target.VelVec.X) + float64(target.PosVec.X)),
		Y: float32(travel*float64(target.VelVec.Y) + float64(target.PosVec.Y)),
	}
}

func monsterMissileTrajectory532610(unit, projectile *Object, targetPos types.Pointf) (spawn, velocity, traceTo types.Pointf) {
	dx := float64(targetPos.X) - float64(unit.PosVec.X)
	dyExtended := float64(targetPos.Y) - float64(unit.PosVec.Y)
	dy := float32(dyExtended)
	denominator := float32(math.Sqrt(dx*dx+dyExtended*float64(dy)) + 0.1)
	velocity.X = float32(dx * float64(projectile.SpeedCur) / float64(denominator))
	velocity.Y = float32(float64(dy) * float64(projectile.SpeedCur) / float64(denominator))

	spawnDistance := float32(float64(unit.Shape.Circle.R) + 4.0)
	direction := unit.Direction1.Vec()
	spawn.X = float32(float64(spawnDistance)*float64(direction.X) + float64(unit.PosVec.X))
	spawn.Y = float32(float64(spawnDistance)*float64(direction.Y) + float64(unit.PosVec.Y))
	traceTo.X = spawn.X + velocity.X
	traceTo.Y = spawn.Y + velocity.Y
	return spawn, velocity, traceTo
}

// monsterActionMissileUpdate532610 restores GAME.EXE 00532610. Invalid action
// metadata is contained by completing the action instead of dereferencing a
// stale PE32 pointer or leaving a permanently stuck native action.
func monsterActionMissileUpdate532610(unit *Object, hooks monsterActionMissileUpdateHooks532610) bool {
	if unit == nil || unit.UpdateData == nil || !unit.Class().Has(object.ClassMonster) {
		return false
	}
	update := unit.UpdateDataMonster()
	if unit.SubClass().Has(object.SubClass(object.MonsterNPC)) {
		attack := 0
		if hooks.playerAttack != nil {
			attack = hooks.playerAttack(unit)
		}
		if attack == 0 && hooks.pop != nil {
			hooks.pop()
		}
		return true
	}

	index := int(update.AIStackInd)
	if index < 0 || index >= len(update.AIStack) || update.MonsterDef == nil {
		if hooks.pop != nil {
			hooks.pop()
		}
		return true
	}
	head := &update.AIStack[index]
	def := update.MonsterDef
	if uint32(update.Field120_1) == def.MissileAttackFrame216 && update.Field120_2 == 0 {
		var projectile *Object
		if hooks.newObject != nil {
			projectile = hooks.newObject(alloc.GoStringS(def.MissileName148[:]))
		}
		if projectile != nil {
			targetPos := head.ArgPos(0)
			if target := head.ArgObj(2); target != nil {
				targetPos = monsterProjectileLead533080(unit, target, projectile.SpeedCur)
			}
			spawn, velocity, traceTo := monsterMissileTrajectory532610(unit, projectile, targetPos)
			if hooks.trace != nil && hooks.trace(unit.PosVec, traceTo, MapTraceFlags(5)) {
				if hooks.createAt != nil {
					hooks.createAt(projectile, unit, spawn)
					projectile.VelVec = velocity
					projectile.Direction1 = unit.Direction1
					projectile.Direction2 = unit.Direction1
				} else if hooks.delayedDelete != nil {
					hooks.delayedDelete(projectile)
				}
			} else if hooks.delayedDelete != nil {
				hooks.delayedDelete(projectile)
			}
		}
		if hooks.audio != nil && update.SoundSet122 != nil {
			hooks.audio(*(*uint32)(unsafe.Add(update.SoundSet122, 11*4)), unit)
		}
	}
	if update.Field120_3 != 0 && hooks.pop != nil {
		hooks.pop()
	}
	return true
}

// MonsterActionMissileUpdate532610 binds projectile allocation, tracing,
// creation, deletion, and action completion to native-width engine services.
func (s *Server) MonsterActionMissileUpdate532610(unit *Object, runtime MonsterActionMissileRuntime532540) bool {
	return monsterActionMissileUpdate532610(unit, monsterActionMissileUpdateHooks532610{
		audio:         runtime.AudioEvent,
		playerAttack:  runtime.PlayerAttack,
		newObject:     s.NewObjectByTypeID,
		trace:         s.MapTraceRay,
		createAt:      runtime.CreateObjectAt,
		delayedDelete: runtime.DelayedDelete,
		pop:           unit.MonsterPopAction,
	})
}
