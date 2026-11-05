package server

import "unsafe"

// RecordMonsterGeneratorCreditRaw4D61B0 preserves the raw 004D61B0 entry's
// admission and load order. The ordinary dispatcher performs its own checks.
func (unit *Object) RecordMonsterGeneratorCreditRaw4D61B0() {
	if unit == nil || generatorCreditFlags4D61B0(unit)&0x20 != 0 {
		return
	}
	update := generatorCreditUpdate4D61B0(unit)
	generatorCreditIncrement4D61B0(generatorCreditPlayer4D61B0(update))
	// The counter can alias update.Player. Reload from the entry-cached update
	// allocation after the increment, as the original does.
	generatorCreditSetFlags4D61B0(generatorCreditPlayer4D61B0(update))
}

// Separate accesses preserve original fault sites and the pointer reload across
// the potentially aliased counter write.

//go:noinline
func generatorCreditFlags4D61B0(unit *Object) uint8 {
	return *(*uint8)(unsafe.Add(unsafe.Pointer(unit), unsafe.Offsetof(unit.ObjFlags)))
}

//go:noinline
func generatorCreditUpdate4D61B0(unit *Object) *PlayerUpdateData {
	return (*PlayerUpdateData)(unit.UpdateData)
}

//go:noinline
func generatorCreditPlayer4D61B0(update *PlayerUpdateData) *Player {
	return update.Player
}

//go:noinline
func generatorCreditIncrement4D61B0(player *Player) {
	player.field4668++
}

//go:noinline
func generatorCreditSetFlags4D61B0(player *Player) {
	player.field4692 |= 8
}
