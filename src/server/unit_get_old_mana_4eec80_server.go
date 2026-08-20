package server

func unitGetOldManaNative4EEC80(unit *Object) uint16 {
	return unitGetOldMana4EEC80(unitGetOldManaHooks4EEC80[*Object, *PlayerUpdateData]{
		loadUnitArg: func() *Object {
			return unit
		},
		loadClass: func(unit *Object) uint32 {
			return uint32(unit.ObjClass)
		},
		loadUpdateData: func(unit *Object) *PlayerUpdateData {
			return (*PlayerUpdateData)(unit.UpdateData)
		},
		loadCurrent: func(update *PlayerUpdateData) uint16 {
			return update.ManaCur
		},
	})
}

// UnitGetOldMana4EEC80 returns the exact AX-width value from GAME.EXE
// 004EEC80 using native-width object and update-data pointers. The historical
// symbol says "old" mana, but the original field is the live current-mana word.
func UnitGetOldMana4EEC80(unit *Object) uint16 {
	return unitGetOldManaNative4EEC80(unit)
}
