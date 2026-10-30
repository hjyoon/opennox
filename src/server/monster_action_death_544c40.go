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
	// The original reloads the definition through the cached entry update after
	// audio and ScriptDeath. Replacing Object.UpdateData does not replace
	// that cached base, but either callback may replace its definition/slot.
	if update.MonsterDef == nil {
		return monsterActionDeathUnsupported544C40(runtime, "live monster definition", unit)
	}
	dieFunc = update.MonsterDef.DieFunc228
	if dieFunc != nil {
		if runtime.CanDieFunc == nil || !runtime.CanDieFunc(dieFunc) || runtime.DieFunc == nil {
			return monsterActionDeathUnsupported544C40(runtime, "live monster die function", unit)
		}
		runtime.DieFunc(dieFunc, unit)
	}
	return true
}

// MonsterActionDyingUpdate544D60 restores GAME.EXE 00544D60.
func (s *Server) MonsterActionDyingUpdate544D60(unit *Object) bool {
	return monsterActionBlockTerminal5320E0(unit, unit.MonsterPopAction)
}

// MonsterActionDyingEnd544CA0 restores GAME.EXE 00544CA0. The ordinary
// non-zombie path is intentionally a no-op in the original executable.
func (s *Server) MonsterActionDyingEnd544CA0(unit *Object, runtime MonsterActionDyingRuntime544C40) bool {
	if unit == nil || unit.UpdateData == nil || !unit.Class().Has(object.ClassMonster) || runtime.IsZombie == nil {
		return false
	}
	update := unit.UpdateDataMonster()
	if !runtime.IsZombie(unit) || !update.StatusFlags.Has(object.MonStatusOnFire) {
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
	ZombieDeadDuration func(int) float32
	// ConvertDuration selects an explicit original FPU control profile.
	// Nil uses the production startup signed-DWORD chop conversion.
	ConvertDuration  func(float32) int32
	RandomInt        func(int, int) int
	SparkExplosion   func(types.Pointf, byte)
	ZombieBurnDelete func(*Object)
	RaiseZombie      func(*Object)
	RemoveUpdatable  func(*Object)
	DelayedDelete    func(*Object)
	Unsupported      func(string, *Object)
}

// MonsterReleasedSoulRuntime544E60 contains the object operations used by
// GAME.EXE 00544E60. Keeping these operations native-width avoids sending the
// dying monster or the newly allocated soul through PE32 integer arguments.
type MonsterReleasedSoulRuntime544E60 struct {
	NewObjectByTypeID func(string) *Object
	CreateObjectAt    func(*Object, *Object, types.Pointf)
}

// MonsterCreateReleasedSoul544E60 restores GAME.EXE 00544E60. Allocation
// failure is an original, successfully handled outcome.
func MonsterCreateReleasedSoul544E60(unit *Object, runtime MonsterReleasedSoulRuntime544E60) bool {
	if unit == nil || runtime.NewObjectByTypeID == nil || runtime.CreateObjectAt == nil {
		return false
	}
	soul := runtime.NewObjectByTypeID("ReleasedSoul")
	if soul == nil {
		return true
	}
	runtime.CreateObjectAt(soul, nil, unit.PosVec)
	soul.Direction2 = unit.Direction1
	soul.Direction1 = unit.Direction1
	return true
}

func monsterActionDeadUnsupported544D80(runtime MonsterActionDeadRuntime544D80, reason string, unit *Object) bool {
	if runtime.Unsupported != nil {
		runtime.Unsupported(reason, unit)
	}
	return false
}

// MonsterActionDeadStart544D80 restores GAME.EXE 00544D80. Unsupported
// callbacks are rejected before motion or object flags are changed.
func (s *Server) MonsterActionDeadStart544D80(unit *Object, runtime MonsterActionDeadRuntime544D80) bool {
	if unit == nil || unit.UpdateData == nil || !unit.Class().Has(object.ClassMonster) || runtime.IsZombie == nil {
		return false
	}
	update := unit.UpdateDataMonster()
	if update.MonsterDef == nil {
		return monsterActionDeadUnsupported544D80(runtime, "missing monster definition", unit)
	}
	// Preserve the established read-only preflight for partial runtimes.
	// Fully bound production runtimes query only after the live dead callback.
	if (runtime.ZombieDeadDuration == nil || runtime.RandomInt == nil) && runtime.IsZombie(unit) {
		return monsterActionDeadUnsupported544D80(runtime, "zombie dead duration", unit)
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
	// ReleasedSoul may change the definition or callback in the entry update.
	// The original reads it after zeroing motion, even if current UpdateData changed.
	if update.MonsterDef == nil {
		return monsterActionDeadUnsupported544D80(runtime, "live monster definition", unit)
	}
	deadFunc = update.MonsterDef.DeadFunc232
	if deadFunc != nil {
		if runtime.CanDeadFunc == nil || !runtime.CanDeadFunc(deadFunc) || runtime.DeadFunc == nil {
			return monsterActionDeadUnsupported544D80(runtime, "live monster dead function", unit)
		}
		runtime.DeadFunc(deadFunc, unit)
	}
	if runtime.IsZombie(unit) {
		if runtime.ZombieDeadDuration == nil || runtime.RandomInt == nil {
			return monsterActionDeadUnsupported544D80(runtime, "live zombie dead duration", unit)
		}
		convert := runtime.ConvertDuration
		if convert == nil {
			convert = netFXTruncateSignedDWORD419A70
		}
		minimum := int(convert(runtime.ZombieDeadDuration(0)))
		maximum := int(convert(runtime.ZombieDeadDuration(1)))
		update.Field123 = uint32(runtime.RandomInt(minimum, maximum))
		unit.ObjFlags |= object.FlagShort
	} else {
		// The original queries again after a false first result. Both calls can
		// observe state changed by the live dead callback or preceding query.
		overlap := !runtime.IsZombie(unit)
		unit.ObjFlags |= object.FlagShort
		if overlap {
			unit.ObjFlags |= object.FlagAllowOverlap
		}
	}
	return true
}

// MonsterActionDeadUpdate544EC0 restores GAME.EXE 00544EC0 and the
// pointer-bearing ordinary-monster cleanup at 00544F70.
func (s *Server) MonsterActionDeadUpdate544EC0(unit *Object, runtime MonsterActionDeadRuntime544D80) bool {
	if unit == nil || unit.UpdateData == nil || !unit.Class().Has(object.ClassMonster) ||
		runtime.IsZombie == nil {
		return false
	}
	update := unit.UpdateDataMonster()
	if runtime.IsZombie(unit) {
		status := update.StatusFlags
		if status.Has(object.MonStatusOnFire) {
			if runtime.SparkExplosion == nil || runtime.ZombieBurnDelete == nil {
				return monsterActionDeadUnsupported544D80(runtime, "burning zombie deletion", unit)
			}
			runtime.SparkExplosion(unit.PosVec, 100)
			runtime.ZombieBurnDelete(unit)
			return true
		}
		if s.Frame()-update.Field137 > update.Field123 &&
			!status.Has(object.MonStatusStayDead) && update.CurrentEnemy != nil {
			if runtime.RaiseZombie == nil {
				return monsterActionDeadUnsupported544D80(runtime, "zombie raise", unit)
			}
			runtime.RaiseZombie(unit)
		}
		return true
	}
	if runtime.RemoveUpdatable == nil {
		return false
	}
	runtime.RemoveUpdatable(unit)
	unit.Update = nil
	if !monsterDeadCleanup544F70(unit, runtime) {
		return false
	}
	// The cleanup helper reloads Object.UpdateData, but the final definition
	// lookup still uses the update cached on entry to ACTION_DEAD.
	if update.MonsterDef == nil {
		return monsterActionDeadUnsupported544D80(runtime, "missing live monster definition", unit)
	}
	if uint32(update.MonsterDef.StatusFlags92)&1 != 0 {
		if runtime.DelayedDelete == nil {
			return monsterActionDeadUnsupported544D80(runtime, "dead monster deletion", unit)
		}
		runtime.DelayedDelete(unit)
	}
	return true
}

// monsterDeadCleanup544F70 clears only the original counted prefixes. Slots
// beyond the counts and adjacent byte/word fields are left intact. Malformed
// counts that would cross the native arrays stop at their store boundary.
func monsterDeadCleanup544F70(unit *Object, runtime MonsterActionDeadRuntime544D80) bool {
	if unit.UpdateData == nil {
		return monsterActionDeadUnsupported544D80(runtime, "missing live monster dead update", unit)
	}
	// Admission happened before the callbacks in ACTION_DEAD. The original
	// reloads this link without rechecking class, which a callback may change.
	update := (*MonsterUpdateData)(unit.UpdateData)
	waypoints := int32(update.Field74)
	update.Field523_2 = 0
	if waypoints > int32(len(update.Waypoints)) {
		return monsterActionDeadUnsupported544D80(runtime, "monster dead waypoint count", unit)
	}
	for i := int32(0); i < waypoints; i++ {
		update.Waypoints[i] = nil
	}
	update.Field74 = 0
	update.Field91 = 0
	seen := int(update.Field282_1)
	if seen > len(update.SeenEnemies) {
		return monsterActionDeadUnsupported544D80(runtime, "monster dead seen-enemy count", unit)
	}
	for i := 0; i < seen; i++ {
		update.SeenEnemies[i] = nil
	}
	update.Field282_1 = 0
	update.CurrentEnemy = nil
	return true
}

// MonsterDeadCleanupNative544F70 exposes the original cleanup pointer result
// without narrowing the update address to the PE32 return word.
func MonsterDeadCleanupNative544F70(unit *Object, runtime MonsterActionDeadRuntime544D80) unsafe.Pointer {
	if unit == nil || !monsterDeadCleanup544F70(unit, runtime) {
		return nil
	}
	return unit.UpdateData
}
