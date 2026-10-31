package server

import (
	"math"
	"unsafe"

	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/common/unit/ai"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

func monsterMissileStartNative532540(unit *Object, hooks monsterActionMissileStartHooks532540) uintptr {
	frame := hooks.frame()
	update := (*MonsterUpdateData)(unit.UpdateData)
	if frame < update.Field128 {
		item := hooks.push(ai.DEPENDENCY_OBJECT_CLOSER_THAN)
		if item != nil {
			rangeValue := pointOnLineSpill57C8A0(float64(update.MonsterDef.MissileAttackRange212) * float64(float32(1.2)))
			item.Args[0] = uintptr(math.Float32bits(rangeValue))
			item.Args[2] = uintptr(unsafe.Pointer(update.CurrentEnemy))
		}
		item = hooks.push(ai.ACTION_WAIT)
		if item != nil {
			item.Args[0] = uintptr(update.Field128)
		}
		return uintptr(unsafe.Pointer(item))
	}
	if hooks.buffOff != nil {
		hooks.buffOff(unit, 0)
		hooks.buffOff(unit, 23)
	}
	unit.Field34 = hooks.frame()
	definition := update.MonsterDef
	delay := hooks.random(int(int32(definition.MissileAttackDelayMin220)), int(int32(definition.MissileAttackDelayMax224)))
	update.Field128 = hooks.frame() + uint32(delay)
	set := monsterMeleeSoundSet424300(unit)
	if set != nil && hooks.audio != nil {
		hooks.audio(*(*uint32)(unsafe.Add(set, 10*4)), unit)
		return 0 // The declared audio service has a void return convention.
	}
	return uintptr(set)
}

func monsterMissileUpdateNative532610(unit *Object, hooks monsterActionMissileUpdateHooks532610) uintptr {
	update := (*MonsterUpdateData)(unit.UpdateData)
	if uint8(unit.ObjSubClass)&16 != 0 {
		attack := 0
		if hooks.playerAttack != nil {
			attack = hooks.playerAttack(unit)
		}
		if attack == 0 {
			return uintptr(byte(hooks.pop()))
		}
		return uintptr(byte(attack))
	}
	// The original computes this address before allocation. Callbacks can
	// replace the live head or update pointer without replacing this slot.
	index := int(update.AIStackInd)
	head := (*AIStackItem)(unsafe.Add(unsafe.Pointer(&update.AIStack[0]), index*int(unsafe.Sizeof(AIStackItem{}))))
	definition := update.MonsterDef
	if uint32(update.Field120_1) == definition.MissileAttackFrame216 && update.Field120_2 == 0 {
		projectile := hooks.newObject(alloc.GoStringS(definition.MissileName148[:]))
		if projectile != nil {
			var targetPos types.Pointf
			if target := head.ArgObj(2); target != nil {
				MonsterProjectileLead533080(unit, target, projectile.SpeedCur, &targetPos)
			} else {
				targetPos = head.ArgPos(0)
			}
			spawn, velocity, traceTo := monsterMissileTrajectoryNative532610(unit, projectile, targetPos)
			from := unit.PosVec
			if hooks.trace(from, traceTo, MapTraceFlags(5)) {
				hooks.createAt(projectile, unit, spawn)
				projectile.VelVec = velocity
				direction := unit.Direction1
				projectile.Direction1 = direction
				projectile.Direction2 = direction
			} else {
				hooks.delayedDelete(projectile)
			}
		}
		set := monsterMeleeSoundSet424300(unit)
		if set != nil && hooks.audio != nil {
			hooks.audio(*(*uint32)(unsafe.Add(set, 11*4)), unit)
		}
	}
	result := update.Field120_3
	if result != 0 {
		result = byte(hooks.pop())
	}
	return uintptr(result)
}

func monsterMissileTrajectoryNative532610(unit, projectile *Object, target types.Pointf) (spawn, velocity, traceTo types.Pointf) {
	add, sub := monsterMissileAdd533080, monsterMissileSub533080
	mul, div, spill := monsterMissileMul533080, monsterMissileDiv533080, monsterMissileSpill533080
	load, memory := monsterMissileLoadF32_533080, monsterMissileMemoryF32_533080
	dx := sub(load(target.X), memory(unit.PosVec.X))
	dy := sub(load(target.Y), memory(unit.PosVec.Y))
	spillY := spill(dy)
	denominator := spill(add(monsterMissileSqrt533080(add(mul(dy, memory(spillY)), mul(dx, dx))), memory(0.1)))
	spawnDistance := spill(add(load(unit.Shape.Circle.R), 4))
	velocity.X = spill(div(mul(dx, memory(projectile.SpeedCur)), memory(denominator)))
	velocity.Y = spill(div(mul(load(spillY), memory(projectile.SpeedCur)), memory(denominator)))
	direction := unit.Direction1.Vec()
	spawn.X = spill(add(mul(load(spawnDistance), memory(direction.X)), memory(unit.PosVec.X)))
	spawn.Y = spill(add(mul(load(spawnDistance), memory(direction.Y)), memory(unit.PosVec.Y)))
	traceTo.X = spill(add(load(velocity.X), memory(spawn.X)))
	traceTo.Y = spill(add(load(velocity.Y), memory(spawn.Y)))
	return
}

func (s *Server) MonsterMissileNative532540(entry uint32, unit *Object, runtime MonsterMissileNativeRuntime532540) uintptr {
	reset := func() { monsterActionResetRecord50A110((*MonsterUpdateData)(unit.UpdateData), s.Frame) }
	pop := func() int {
		return unit.monsterPopActionRecord50A160(s, (*MonsterUpdateData)(unit.UpdateData), runtime.EndAction, reset)
	}
	push := func(action ai.ActionType, _ ...any) *AIStackItem {
		cached := (*MonsterUpdateData)(unit.UpdateData)
		item := unit.monsterPushActionWithReset50A260(action, "go", 0, runtime.CancelAction, reset)
		if runtime.ObservePush != nil {
			runtime.ObservePush(action, cached, item)
		}
		return item
	}
	if entry == 0x532540 {
		return monsterMissileStartNative532540(unit, monsterActionMissileStartHooks532540{frame: s.Frame, random: s.Rand.Logic.IntClamp, audio: runtime.AudioEvent, buffOff: runtime.BuffOff, push: push})
	}
	newObject, trace := runtime.NewObjectByTypeID, runtime.MapTraceRay
	if newObject == nil {
		newObject = s.NewObjectByTypeID
	}
	if trace == nil {
		trace = s.MapTraceRay
	}
	return monsterMissileUpdateNative532610(unit, monsterActionMissileUpdateHooks532610{audio: runtime.AudioEvent, playerAttack: runtime.PlayerAttack, newObject: newObject, trace: trace, createAt: runtime.CreateObjectAt, delayedDelete: runtime.DelayedDelete, pop: pop})
}
