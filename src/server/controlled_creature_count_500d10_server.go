package server

import "unsafe"

func controlledCreatureCountNative500D10(
	owner *Object,
	isMonitored func(owner, unit *Object) bool,
) int32 {
	return controlledCreatureCount500D10(owner, controlledCreatureCountHooks500D10[*Object]{
		loadFirstOwned: func(owner *Object) *Object {
			return owner.Field129
		},
		isMonitored: isMonitored,
		loadSubclassLow: func(unit *Object) uint8 {
			return uint8(unit.ObjSubClass)
		},
		loadNextOwned: func(unit *Object) *Object {
			return unit.Field128
		},
	})
}

// Nox_xxx_countControlledCreatures_500D10 binds the exact PE32 traversal and
// wrapping accumulator to native-width Object links.
func (obj *Object) Nox_xxx_countControlledCreatures_500D10() int {
	return int(controlledCreatureCountNative500D10(obj, Nox_xxx_creatureIsMonitored_500CC0))
}

var (
	_ = [1]struct{}{}[4-unsafe.Sizeof(Object{}.ObjSubClass)]
	_ = [1]struct{}{}[unsafe.Sizeof(uintptr(0))-unsafe.Sizeof(Object{}.Field128)]
	_ = [1]struct{}{}[unsafe.Sizeof(uintptr(0))-unsafe.Sizeof(Object{}.Field129)]
)
