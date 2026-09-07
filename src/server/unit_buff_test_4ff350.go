package server

// UnitBuffTestHooks4FF350 exposes the two observable non-null-path loads in
// GAME.EXE 004FF350. Comparable object tokens preserve canonical null and
// full native pointer identities without inheriting PE32's pointer width.
type UnitBuffTestHooks4FF350[Object comparable] struct {
	LoadBuffArg func() int32
	LoadBuffs   func(Object) uint32
}

// UnitBuffTest4FF350 preserves GAME.EXE 004FF350's null gate, access order,
// x86 shift-count semantics, and canonical result. A null unit returns before
// loading either the buff argument or Buffs154. Otherwise the signed dword
// argument is loaded first and SHL uses only CL's low five bits, so every
// integer aliases modulo 32 (for example, -1 selects 31 and 32 selects 0).
// The complete object identity is passed to LoadBuffs without narrowing.
func UnitBuffTest4FF350[Object comparable](
	unit Object,
	h UnitBuffTestHooks4FF350[Object],
) int32 {
	var nilObject Object
	if unit == nilObject {
		return 0
	}

	buff := h.LoadBuffArg()
	buffs := h.LoadBuffs(unit)
	mask := uint32(1) << (uint32(buff) & 31)
	if buffs&mask != 0 {
		return 1
	}
	return 0
}
