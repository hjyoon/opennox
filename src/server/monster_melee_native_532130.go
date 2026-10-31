package server

import (
	"math"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/common/unit/ai"
)

func MonsterMeleeGlobalState532390() MonsterMeleeBlockerState532390 {
	return MonsterMeleeBlockerState532390{
		Selected: memmap.PtrPtr(0x5d4594, 2487948),
		Distance: memmap.PtrFloat32(0x5d4594, 2487952),
	}
}

// MonsterMeleeBlocker532390 is the original callback's shared state update.
// X and Y remain at precision53/ToZero; only Y's memory operand and the
// denominator are spilled. In particular, the Y product is retainedY*spillY.
func MonsterMeleeBlocker532390(candidate, unit *Object, state MonsterMeleeBlockerState532390) {
	if uint8(candidate.ObjClass)&6 == 0 || candidate.ObjFlags.Has(object.FlagDead) {
		return
	}
	add := pointOnLineAddChop53_57C8A0
	mul := pointOnLineMulChop53_57C8A0
	div := pointOnLineDivChop53_57C8A0
	spill := pointOnLineSpill57C8A0
	dx := add(float64(candidate.PosVec.X), -float64(unit.PosVec.X))
	dy := add(float64(candidate.PosVec.Y), -float64(unit.PosVec.Y))
	spillY := spill(dy)
	square := add(mul(dy, float64(spillY)), mul(dx, dx))
	denominator := spill(add(monsterMoveForceSqrtChop53_50D50C(square), float64(math.Float32frombits(0x38d1b717))))
	// FCOMP tests C0 only, admitting unordered distance comparisons.
	if denominator >= *state.Distance {
		return
	}
	facing := unit.Direction1.Vec()
	dot := add(mul(div(float64(spillY), float64(denominator)), float64(facing.Y)),
		mul(div(dx, float64(denominator)), float64(facing.X)))
	if dot > 0.5 {
		*state.Selected = unsafe.Pointer(candidate)
		*state.Distance = denominator
	}
}

func monsterMeleeSearchRect532130(unit *Object) types.Rectf {
	add, spill := pointOnLineAddChop53_57C8A0, pointOnLineSpill57C8A0
	radius := float64(monsterMeleeFriendlyRadius532390)
	return types.Rectf{
		Min: types.Ptf(spill(add(float64(unit.PosVec.X), -radius)), spill(add(float64(unit.PosVec.Y), -radius))),
		Max: types.Ptf(spill(add(radius, float64(unit.PosVec.X))), spill(add(radius, float64(unit.PosVec.Y)))),
	}
}

func monsterMeleeSoundSet424300(unit *Object) unsafe.Pointer {
	if unit == nil || !unit.ObjClass.Has(object.ClassMonster) {
		return nil
	}
	return (*MonsterUpdateData)(unit.UpdateData).SoundSet122
}

func monsterMeleeStartNative532130(unit *Object, state MonsterMeleeBlockerState532390, hooks monsterActionMeleeStartHooks532130) uintptr {
	frame := hooks.frame()
	// This entry's original record read has no class admission. Keep the
	// global typed UpdateDataMonster invariant separate from this raw access.
	update := (*MonsterUpdateData)(unit.UpdateData)
	if frame < update.Field128 {
		item := hooks.push(ai.DEPENDENCY_OBJECT_CLOSER_THAN)
		if item != nil {
			rangeValue := pointOnLineSpill57C8A0(float64(update.MonsterDef.MeleeAttackRange112) * float64(float32(1.2)))
			item.Args[0] = uintptr(math.Float32bits(rangeValue))
			item.Args[2] = uintptr(unsafe.Pointer(update.CurrentEnemy))
		}
		item = hooks.push(ai.ACTION_WAIT)
		if item != nil {
			item.Args[0] = uintptr(update.Field128)
		}
		return uintptr(unsafe.Pointer(item))
	}
	if uint8(unit.ObjSubClass)&16 != 0 {
		// Seed an older inventory-only load before callbacks replace live data.
		unit.NPCWeaponForMeleeUpdate532130(update)
		if hooks.weaponStamina != nil {
			cost := hooks.weaponStamina(update.WeaponEquipFlags)
			if cost > int32(update.Stamina) {
				update.Stamina -= byte(cost)
			} else {
				update.Stamina = 0
			}
		}
	}
	rect := monsterMeleeSearchRect532130(unit)
	*state.Selected, *state.Distance = nil, 51
	if hooks.eachInRect != nil {
		hooks.eachInRect(rect, func(candidate *Object) bool {
			MonsterMeleeBlocker532390(candidate, unit, state)
			return true
		})
	}
	blocker := (*Object)(*state.Selected)
	if blocker != nil && hooks.isEnemy != nil && !hooks.isEnemy(unit, blocker) && uint8(unit.ObjSubClass)&16 != 0 {
		weapon := unit.NPCWeaponForMeleeUpdate532130(update)
		if weapon != nil && uint32(weapon.ObjSubClass)&0x4000 != 0 {
			hooks.pop()
			item := hooks.push(ai.ACTION_FACE_ANGLE)
			if item != nil {
				item.Args[0] = uintptr(uint32(int32(int16(unit.Direction1))))
			}
			item = hooks.push(ai.DEPENDENCY_TIME)
			if item != nil {
				fps := hooks.tickRate()
				delay := hooks.random(int(fps>>2), int(fps>>1))
				item.Args[0] = uintptr(hooks.frame() + uint32(delay))
			}
			item = hooks.push(ai.ACTION_FLEE)
			if item != nil {
				// A push/cancel callback may replace the shared selected object.
				position := (*Object)(*state.Selected).PosVec
				item.Args[0] = uintptr(math.Float32bits(position.X))
				item.Args[1] = uintptr(math.Float32bits(position.Y))
				item.Args[2] = 0
			}
			return uintptr(unsafe.Pointer(item))
		}
	}
	if hooks.buffOff != nil {
		hooks.buffOff(unit, 0)
		hooks.buffOff(unit, 23)
	}
	unit.Field34 = hooks.frame()
	definition := update.MonsterDef
	delay := hooks.random(int(int32(definition.MeleeAttackDelayMin128)), int(int32(definition.MeleeAttackDelayMax132)))
	update.Field128 = hooks.frame() + uint32(delay)
	set := monsterMeleeSoundSet424300(unit)
	if set != nil && hooks.audio != nil {
		hooks.audio(*(*uint32)(unsafe.Add(set, 6*4)), unit)
		return 0 // The declared audio service has a void return convention.
	}
	return uintptr(set)
}

func monsterMeleeUpdateNative532440(unit *Object, hooks monsterActionMeleeUpdateHooks532440) (uintptr, bool) {
	update := (*MonsterUpdateData)(unit.UpdateData)
	if uint8(unit.ObjSubClass)&16 != 0 {
		if update.StatusFlags&object.MonsterStatus(0x20000) != 0 {
			MonsterMorphToPlayer4FAAF0(unit)
		}
		attack := 0
		if hooks.playerAttack != nil {
			attack = hooks.playerAttack(unit)
		}
		result := byte(update.StatusFlags)
		if update.StatusFlags&object.MonsterStatus(0x20000) != 0 {
			result = byte(MonsterMorphFromPlayer4FAAC0(unit))
		}
		if attack == 0 {
			result = byte(hooks.pop())
		}
		return uintptr(result), true
	}
	definition := update.MonsterDef
	strike := definition.MeleeStrikeFunc236
	if strike == nil {
		return uintptr(byte(hooks.pop())), true
	}
	if hooks.canStrike == nil || !hooks.canStrike(strike) || hooks.strike == nil {
		return 0, false
	}
	if uint32(update.Field120_1) == definition.MeleeAttackFrame108 && update.Field120_2 == 0 {
		hit := hooks.strike(unit, strike)
		set := monsterMeleeSoundSet424300(unit)
		if set != nil && hooks.audio != nil {
			index := uintptr(9)
			if hit != 0 {
				index = 8
			}
			hooks.audio(*(*uint32)(unsafe.Add(set, index*4)), unit)
		}
	}
	result := update.Field120_3
	if result != 0 {
		result = byte(hooks.pop())
	}
	return uintptr(result), true
}

func (s *Server) monsterMeleeNative532130(entry uint32, unit *Object, state MonsterMeleeBlockerState532390, runtime MonsterMeleeNativeRuntime532130) (uintptr, bool) {
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
	if entry == 0x532440 {
		return monsterMeleeUpdateNative532440(unit, monsterActionMeleeUpdateHooks532440{
			audio: runtime.AudioEvent, pop: pop, playerAttack: runtime.PlayerAttack,
			canStrike: runtime.CanStrike, strike: runtime.Strike,
		})
	}
	each, enemy := runtime.EachInRect, runtime.IsEnemy
	if each == nil {
		each = s.Map.EachObjInRect
	}
	if enemy == nil {
		enemy = s.IsEnemyTo
	}
	return monsterMeleeStartNative532130(unit, state, monsterActionMeleeStartHooks532130{
		frame: s.Frame, tickRate: s.TickRate, random: s.Rand.Logic.IntClamp,
		audio: runtime.AudioEvent, buffOff: runtime.BuffOff, weaponStamina: WeaponStaminaByType4F7E80,
		eachInRect: each, isEnemy: enemy, pop: pop, push: push,
	}), true
}

func (s *Server) MonsterMeleeNative532130(entry uint32, unit *Object, state MonsterMeleeBlockerState532390, runtime MonsterMeleeNativeRuntime532130) uintptr {
	result, _ := s.monsterMeleeNative532130(entry, unit, state, runtime)
	return result
}

func monsterMeleeRuntimeFromAction532130(runtime MonsterActionMeleeRuntime532130) MonsterMeleeNativeRuntime532130 {
	return MonsterMeleeNativeRuntime532130{
		AudioEvent: runtime.AudioEvent, BuffOff: runtime.BuffOff, PlayerAttack: runtime.PlayerAttack,
		CanStrike: runtime.CanStrike, Strike: runtime.Strike,
	}
}
