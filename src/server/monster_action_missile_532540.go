package server

import (
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/unit/ai"
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
	if unit == nil || unit.UpdateData == nil || hooks.frame == nil || hooks.random == nil || hooks.push == nil {
		return false
	}
	if (*MonsterUpdateData)(unit.UpdateData).MonsterDef == nil {
		return false
	}
	monsterMissileStartNative532540(unit, hooks)
	return true
}

// MonsterActionMissileStart532540 binds the restored start action to the live
// frame clock, deterministic logic RNG, and native-width action stack.
func (s *Server) MonsterActionMissileStart532540(unit *Object, runtime MonsterActionMissileRuntime532540) bool {
	if unit == nil || unit.UpdateData == nil {
		return false
	}
	s.MonsterMissileNative532540(0x532540, unit, monsterMissileRuntimeFromAction532540(runtime))
	return true
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
// binary32 values, while x87 intermediates use precision53 / ToZero until
// each output coordinate is stored back to binary32.
func monsterProjectileLead533080(unit, target *Object, speed float32) types.Pointf {
	var output types.Pointf
	MonsterProjectileLead533080(unit, target, speed, &output)
	return output
}

func monsterMissileTrajectory532610(unit, projectile *Object, targetPos types.Pointf) (spawn, velocity, traceTo types.Pointf) {
	return monsterMissileTrajectoryNative532610(unit, projectile, targetPos)
}

// monsterActionMissileUpdate532610 restores GAME.EXE 00532610. Invalid action
// metadata is contained by completing the action instead of dereferencing a
// stale PE32 pointer or leaving a permanently stuck native action.
func monsterActionMissileUpdate532610(unit *Object, hooks monsterActionMissileUpdateHooks532610) bool {
	if unit == nil || unit.UpdateData == nil {
		return false
	}
	update := (*MonsterUpdateData)(unit.UpdateData)
	// The existing hook API contains invalid action metadata before invoking
	// services. The public native entry retains the original raw record access.
	if uint8(unit.ObjSubClass)&16 == 0 && (update.AIStackInd < 0 || int(update.AIStackInd) >= len(update.AIStack) || update.MonsterDef == nil) {
		if hooks.pop != nil {
			hooks.pop()
		}
		return true
	}
	monsterMissileUpdateNative532610(unit, hooks)
	return true
}

// MonsterActionMissileUpdate532610 binds projectile allocation, tracing,
// creation, deletion, and action completion to native-width engine services.
func (s *Server) MonsterActionMissileUpdate532610(unit *Object, runtime MonsterActionMissileRuntime532540) bool {
	if unit == nil || unit.UpdateData == nil {
		return false
	}
	s.MonsterMissileNative532540(0x532610, unit, monsterMissileRuntimeFromAction532540(runtime))
	return true
}

func monsterMissileRuntimeFromAction532540(runtime MonsterActionMissileRuntime532540) MonsterMissileNativeRuntime532540 {
	return MonsterMissileNativeRuntime532540{AudioEvent: runtime.AudioEvent, BuffOff: runtime.BuffOff, PlayerAttack: runtime.PlayerAttack, CreateObjectAt: runtime.CreateObjectAt, DelayedDelete: runtime.DelayedDelete}
}
