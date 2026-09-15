package server

import (
	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/unit/ai"
)

// ZombieBurnDeleteRuntime544CE0 contains the effects emitted by GAME.EXE
// 00544CE0. Object and owner pointers remain native-width throughout.
type ZombieBurnDeleteRuntime544CE0 struct {
	NetFxShield           func(int, *Object)
	UnmarkMinimap         func(int, *Object, uint32)
	SoloMonsterKillReward func(*Object)
	MakeScorch            func(types.Pointf, int)
	DelayedDelete         func(*Object)
	Unsupported           func(string, *Object)
}

func zombieBurnDeleteUnsupported544CE0(runtime ZombieBurnDeleteRuntime544CE0, reason string, unit *Object) bool {
	if runtime.Unsupported != nil {
		runtime.Unsupported(reason, unit)
	}
	return false
}

// ZombieBurnDelete544CE0 restores GAME.EXE 00544CE0. The player record is
// reloaded after the shield callback just as it is by the two original
// address expressions.
func (s *Server) ZombieBurnDelete544CE0(unit *Object, runtime ZombieBurnDeleteRuntime544CE0) bool {
	if unit == nil {
		return false
	}
	if runtime.SoloMonsterKillReward == nil || runtime.MakeScorch == nil || runtime.DelayedDelete == nil {
		return zombieBurnDeleteUnsupported544CE0(runtime, "zombie burn effects", unit)
	}
	owner := unit.ObjOwner
	var ownerUpdate *PlayerUpdateData
	if owner != nil && uint8(owner.ObjClass)&uint8(object.ClassPlayer) != 0 {
		if owner.UpdateData == nil {
			return zombieBurnDeleteUnsupported544CE0(runtime, "zombie player owner update", unit)
		}
		ownerUpdate = owner.UpdateDataPlayer()
		if ownerUpdate.Player == nil || runtime.NetFxShield == nil || runtime.UnmarkMinimap == nil {
			return zombieBurnDeleteUnsupported544CE0(runtime, "zombie player owner effects", unit)
		}
	}

	if ownerUpdate != nil {
		unit.ObjSubClass &^= object.SubClass(0x80)
		playerIndex := int(ownerUpdate.Player.PlayerInd)
		runtime.NetFxShield(playerIndex, unit)
		if ownerUpdate.Player == nil {
			return zombieBurnDeleteUnsupported544CE0(runtime, "zombie player owner reload", unit)
		}
		playerIndex = int(ownerUpdate.Player.PlayerInd)
		runtime.UnmarkMinimap(playerIndex, unit, 1)
	}
	runtime.SoloMonsterKillReward(unit)
	runtime.MakeScorch(unit.PosVec, 1)
	runtime.DelayedDelete(unit)
	return true
}

// MonsterRaiseZombieRuntime534AB0 contains the external effects used by
// GAME.EXE 00534AB0.
type MonsterRaiseZombieRuntime534AB0 struct {
	IsZombie       func(*Object) bool
	AudioEvent     func(uint32, *Object)
	SetHealthToMax func(*Object)
	Unsupported    func(string, *Object)
}

func monsterRaiseZombieUnsupported534AB0(runtime MonsterRaiseZombieRuntime534AB0, reason string, unit *Object) bool {
	if runtime.Unsupported != nil {
		runtime.Unsupported(reason, unit)
	}
	return false
}

// MonsterRaiseZombie534AB0 restores GAME.EXE 00534AB0. It pops ACTION_DEAD
// before pushing the uninterruptible condition and ACTION_GET_UP, which is
// required by MonsterPushAction's dead-action guard.
func (s *Server) MonsterRaiseZombie534AB0(unit *Object, runtime MonsterRaiseZombieRuntime534AB0) bool {
	if s == nil || unit == nil || unit.UpdateData == nil ||
		!unit.Class().Has(object.ClassMonster) || runtime.IsZombie == nil {
		return false
	}
	update := unit.UpdateDataMonster()
	if update.AIStackInd < 0 || int(update.AIStackInd) >= len(update.AIStack) {
		return monsterRaiseZombieUnsupported534AB0(runtime, "zombie action stack", unit)
	}
	if !runtime.IsZombie(unit) || update.AIStackHead().Type() != ai.ACTION_DEAD {
		return true
	}
	if unit.Server() == nil || runtime.AudioEvent == nil || runtime.SetHealthToMax == nil {
		return monsterRaiseZombieUnsupported534AB0(runtime, "zombie raise effects", unit)
	}
	base := int(update.AIStackInd) - 1
	for base >= 0 && update.AIStack[base].Type().IsCondition() {
		base--
	}
	if base > len(update.AIStack)-3 {
		return monsterRaiseZombieUnsupported534AB0(runtime, "zombie raise action capacity", unit)
	}

	unit.MonsterPopAction()
	if unit.MonsterPushAction(ai.DEPENDENCY_UNINTERRUPTABLE) == nil ||
		unit.MonsterPushAction(ai.ACTION_GET_UP) == nil {
		return monsterRaiseZombieUnsupported534AB0(runtime, "zombie raise action push", unit)
	}
	runtime.AudioEvent(469, unit)
	runtime.SetHealthToMax(unit)
	unit.ObjFlags &^= object.FlagAllowOverlap | object.FlagShort | object.FlagNoCollide | object.FlagDead
	return true
}

// ScriptRaiseZombie516D00 restores the script-facing GAME.EXE 00516D00
// wrapper. StayDead is cleared before the raise helper is entered.
func (s *Server) ScriptRaiseZombie516D00(unit *Object, runtime MonsterRaiseZombieRuntime534AB0) bool {
	if unit == nil || uint8(unit.ObjClass)&uint8(object.ClassMonster) == 0 ||
		!unit.ObjFlags.Has(object.FlagDead) {
		return true
	}
	if unit.UpdateData == nil || runtime.IsZombie == nil {
		return false
	}
	unit.UpdateDataMonster().StatusFlags &^= object.MonStatusStayDead
	return s.MonsterRaiseZombie534AB0(unit, runtime)
}
