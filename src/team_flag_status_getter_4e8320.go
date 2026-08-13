package opennox

import "unsafe"

// teamFlagStatusGetter4E8320 preserves GAME.EXE 004E8320. The original masks
// the full stack argument to its low byte, computes base + 6*index, and returns
// that address without reading the selected record.
func teamFlagStatusGetter4E8320(
	base *teamFlagStatusRecord4E82C0,
	teamID uint32,
) *teamFlagStatusRecord4E82C0 {
	index := uint8(teamID)
	return (*teamFlagStatusRecord4E82C0)(unsafe.Add(
		unsafe.Pointer(base),
		uintptr(index)*teamFlagStatusRecordSize4E82C0,
	))
}
