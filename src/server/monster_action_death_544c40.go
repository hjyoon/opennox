package server

import (
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
)

// MonsterActionDyingRuntime544C40 contains the engine calls made by the
// ACTION_DYING start and end routines at GAME.EXE 00544C40-00544CA0.
type MonsterActionDyingRuntime544C40 struct {
	AudioEvent       func(uint32, *Object)
	ScriptCallback   func(*ScriptCallback, *Object, *Object, ScriptEventType)
	CanDieFunc       func(unsafe.Pointer) bool
	DieFunc          func(unsafe.Pointer, *Object)
	IsZombie         func(*Object) bool
	ZombieBurnDelete func(*Object)
	Unsupported      func(string, *Object)
}

func monsterActionDeathUnsupported544C40(runtime MonsterActionDyingRuntime544C40, reason string, unit *Object) bool {
	if runtime.Unsupported != nil {
		runtime.Unsupported(reason, unit)
	}
	return false
}

// MonsterActionDyingStart544C40 restores GAME.EXE 00544C40 without loading
// UpdateData or MonsterDef through their PE32 pointer slots.
func (s *Server) MonsterActionDyingStart544C40(unit *Object, runtime MonsterActionDyingRuntime544C40) bool {
	if unit == nil || unit.UpdateData == nil || !unit.Class().Has(object.ClassMonster) {
		return false
	}
	update := unit.UpdateDataMonster()
	if update.MonsterDef == nil {
		return monsterActionDeathUnsupported544C40(runtime, "missing monster definition", unit)
	}
	if update.SoundSet122 != nil && runtime.AudioEvent == nil {
		return monsterActionDeathUnsupported544C40(runtime, "death sound", unit)
	}
	if runtime.ScriptCallback == nil {
		return monsterActionDeathUnsupported544C40(runtime, "death script callback", unit)
	}
	dieFunc := update.MonsterDef.DieFunc228
	if dieFunc != nil && (runtime.CanDieFunc == nil || !runtime.CanDieFunc(dieFunc) || runtime.DieFunc == nil) {
		return monsterActionDeathUnsupported544C40(runtime, "monster die function", unit)
	}

	if update.SoundSet122 != nil {
		runtime.AudioEvent(*(*uint32)(unsafe.Add(update.SoundSet122, 15*4)), unit)
	}
	runtime.ScriptCallback(&update.ScriptDeath, nil, unit, NoxEventMonsterDead)
	if dieFunc != nil {
		runtime.DieFunc(dieFunc, unit)
	}
	return true
}

// MonsterActionDyingUpdate544D60 restores GAME.EXE 00544D60.
func (s *Server) MonsterActionDyingUpdate544D60(unit *Object) bool {
	if unit == nil || unit.UpdateData == nil || !unit.Class().Has(object.ClassMonster) {
		return false
	}
	if unit.UpdateDataMonster().Field120_3 != 0 {
		unit.MonsterPopAction()
	}
	return true
}

// MonsterActionDyingEnd544CA0 restores GAME.EXE 00544CA0. The ordinary
// non-zombie path is intentionally a no-op in the original executable.
func (s *Server) MonsterActionDyingEnd544CA0(unit *Object, runtime MonsterActionDyingRuntime544C40) bool {
	if unit == nil || unit.UpdateData == nil || !unit.Class().Has(object.ClassMonster) || runtime.IsZombie == nil {
		return false
	}
	if !runtime.IsZombie(unit) || !unit.UpdateDataMonster().StatusFlags.Has(object.MonStatusOnFire) {
		return true
	}
	if runtime.ZombieBurnDelete == nil {
		return monsterActionDeathUnsupported544C40(runtime, "burning zombie deletion", unit)
	}
	runtime.ZombieBurnDelete(unit)
	return true
}

// MonsterActionDeadRuntime544D80 contains callbacks used by the ACTION_DEAD
// start and update routines at GAME.EXE 00544D80-00544EC0.
type MonsterActionDeadRuntime544D80 struct {
	IsZombie           func(*Object) bool
	CreateReleasedSoul func(*Object)
	CanDeadFunc        func(unsafe.Pointer) bool
	DeadFunc           func(unsafe.Pointer, *Object)
	RemoveUpdatable    func(*Object)
	DelayedDelete      func(*Object)
	Unsupported        func(string, *Object)
}

func monsterActionDeadUnsupported544D80(runtime MonsterActionDeadRuntime544D80, reason string, unit *Object) bool {
	if runtime.Unsupported != nil {
		runtime.Unsupported(reason, unit)
	}
	return false
}

// MonsterActionDeadStart544D80 restores the ordinary non-zombie branch of
// GAME.EXE 00544D80. Unsupported callbacks are rejected before motion or
// object flags are changed.
func (s *Server) MonsterActionDeadStart544D80(unit *Object, runtime MonsterActionDeadRuntime544D80) bool {
	if unit == nil || unit.UpdateData == nil || !unit.Class().Has(object.ClassMonster) || runtime.IsZombie == nil {
		return false
	}
	update := unit.UpdateDataMonster()
	if update.MonsterDef == nil {
		return monsterActionDeadUnsupported544D80(runtime, "missing monster definition", unit)
	}
	if runtime.IsZombie(unit) {
		return monsterActionDeadUnsupported544D80(runtime, "zombie dead start", unit)
	}
	needsReleasedSoul := unit.Field131 == 14 && uint32(unit.SubClass())&0x10000 != 0
	if needsReleasedSoul && runtime.CreateReleasedSoul == nil {
		return monsterActionDeadUnsupported544D80(runtime, "released soul creation", unit)
	}
	deadFunc := update.MonsterDef.DeadFunc232
	if deadFunc != nil && (runtime.CanDeadFunc == nil || !runtime.CanDeadFunc(deadFunc) || runtime.DeadFunc == nil) {
		return monsterActionDeadUnsupported544D80(runtime, "monster dead function", unit)
	}

	if needsReleasedSoul {
		runtime.CreateReleasedSoul(unit)
	}
	unit.VelVec = types.Pointf{}
	unit.ForceVec = types.Pointf{}
	unit.Pos24 = types.Pointf{}
	if deadFunc != nil {
		runtime.DeadFunc(deadFunc, unit)
	}
	unit.ObjFlags |= object.FlagAllowOverlap | object.FlagShort
	return true
}

// MonsterActionDeadUpdate544EC0 restores the ordinary non-zombie branch of
// GAME.EXE 00544EC0 and the pointer-bearing cleanup at 00544F70.
func (s *Server) MonsterActionDeadUpdate544EC0(unit *Object, runtime MonsterActionDeadRuntime544D80) bool {
	if unit == nil || unit.UpdateData == nil || !unit.Class().Has(object.ClassMonster) ||
		runtime.IsZombie == nil || runtime.RemoveUpdatable == nil {
		return false
	}
	if runtime.IsZombie(unit) {
		return monsterActionDeadUnsupported544D80(runtime, "zombie dead update", unit)
	}
	update := unit.UpdateDataMonster()
	if update.MonsterDef == nil {
		return monsterActionDeadUnsupported544D80(runtime, "missing monster definition", unit)
	}
	deleteAfterUpdate := uint32(update.MonsterDef.StatusFlags92)&1 != 0
	if deleteAfterUpdate && runtime.DelayedDelete == nil {
		return monsterActionDeadUnsupported544D80(runtime, "dead monster deletion", unit)
	}

	runtime.RemoveUpdatable(unit)
	unit.Update = nil
	update.Field523_2 = 0
	update.Field74 = 0
	for i := range update.Waypoints {
		update.Waypoints[i] = nil
	}
	update.Field91 = 0
	update.Field282_1 = 0
	for i := range update.SeenEnemies {
		update.SeenEnemies[i] = nil
	}
	update.CurrentEnemy = nil
	if deleteAfterUpdate {
		runtime.DelayedDelete(unit)
	}
	return true
}
