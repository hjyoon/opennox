package server

import "unsafe"

// MonsterActionRoamStartNative545790 is the raw public-entry operation. Its
// ordinary caller applies class/null admission before sharing these stores.
func (unit *Object) MonsterActionRoamStartNative545790() *MonsterUpdateData {
	update := (*MonsterUpdateData)(unit.UpdateData)
	update.Waypoints[0] = nil
	update.Field91 = 0
	return update
}

// MonsterActionRoamCancelNative5457C0 retains the original signed, unchecked
// stack-index arithmetic with native item/argument widths. Ordinary callers
// still use AIStackHead admission; this raw entry adds no bounds/null guard.
func (unit *Object) MonsterActionRoamCancelNative5457C0() *MonsterUpdateData {
	update := (*MonsterUpdateData)(unit.UpdateData)
	item := (*AIStackItem)(unsafe.Add(unsafe.Pointer(&update.AIStack[0]), int(update.AIStackInd)*int(unsafe.Sizeof(AIStackItem{}))))
	item.Args[0] = 0
	return update
}
