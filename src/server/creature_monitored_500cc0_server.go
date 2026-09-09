package server

import "unsafe"

func creatureIsMonitoredNative500CC0(owner, unit *Object, isZombie func(*Object) bool) bool {
	return creatureIsMonitored500CC0(owner, unit, creatureMonitoredHooks500CC0[*Object, *MonsterUpdateData]{
		loadClassLow: func(unit *Object) uint8 {
			return uint8(unit.ObjClass)
		},
		loadFlags: func(unit *Object) uint32 {
			return uint32(unit.ObjFlags)
		},
		isZombie: isZombie,
		loadUpdate: func(unit *Object) *MonsterUpdateData {
			// The oracle admits a zombie without requiring the Monster class
			// bit, so this must remain a raw update-data interpretation rather
			// than UpdateDataMonster's additional class assertion.
			return (*MonsterUpdateData)(unit.UpdateData)
		},
		loadStatusLow: func(update *MonsterUpdateData) uint8 {
			return uint8(update.StatusFlags)
		},
		hasOwner: func(unit, owner *Object) bool {
			return unitHasThatParentNative4EC4F0(unit, owner)
		},
	})
}

// Nox_xxx_creatureIsMonitored_500CC0 binds the exact 00500CC0 predicate to
// native-width server objects. The server lookup remains lazy because an
// alive Monster does not call the original zombie predicate.
func Nox_xxx_creatureIsMonitored_500CC0(owner, unit *Object) bool {
	return creatureIsMonitoredNative500CC0(owner, unit, func(unit *Object) bool {
		return unit.Server().IsZombie(unit)
	})
}

var (
	_ = [1]struct{}{}[4-unsafe.Sizeof(Object{}.ObjClass)]
	_ = [1]struct{}{}[4-unsafe.Sizeof(Object{}.ObjFlags)]
	_ = [1]struct{}{}[4-unsafe.Sizeof(MonsterUpdateData{}.StatusFlags)]
)
