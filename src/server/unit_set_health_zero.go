package server

import (
	"unsafe"

	"github.com/opennox/libs/object"
)

// UnitSetHealthZero4EE5E0 assigns zero health and completes the native death
// lifecycle. This is a direct setter, not GAME.EXE 004EE5E0's damage entry:
// SetHealth already writes through GodMode and must not leave a living unit
// with zero HP. The global engine flags and ordinary damage gate are untouched.
func (s *Server) UnitSetHealthZero4EE5E0(unit *Object, runtime UnitDamageClearRuntime4EE5E0) {
	if unit == nil {
		return
	}
	if unit.HealthData == nil || unit.HealthData.Max == 0 {
		// Preserve the existing raw setter/report for nonliving health records;
		// the native lethal lifecycle likewise excludes nil/zero-Max health.
		runtime.SetHP(unit, 0)
		if unit.Class().Has(object.ClassMonster) {
			s.MobInformOwnerHP4EE4C0(unit)
		}
		return
	}
	// Cur is uint16 and may change when a Warrior's harpoon is broken. Use
	// the full health range so that the live post-callback record is zeroed.
	unitDamageClearNative4EE5E0(unit, 0xffff, unitDamageClearNativeDeps4EE5E0{
		engineFlag:   func(uint32) int32 { return 0 },
		breakHarpoon: runtime.BreakHarpoon,
		setHP:        runtime.SetHP,
		buffOff:      runtime.BuffOff,
		isZombie:     s.IsZombie,
		soloReward:   runtime.SoloReward,
		monsterDie:   runtime.MonsterDie,
		callDeath: func(death unsafe.Pointer, unit *Object) {
			CallObjectDeath(death, unit)
		},
		delayedDelete: runtime.DelayedDelete,
		informOwnerHP: s.MobInformOwnerHP4EE4C0,
	})
}
