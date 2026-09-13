package server

// SpellTagUpdate530250 implements GAME.EXE 00530250 using native-width
// duration and target objects. A nil target ends the duration; otherwise the
// result is bit 5 of the target's low object-flags byte.
func SpellTagUpdate530250(record *DurSpell) int32 {
	target := record.Target48
	if target == nil {
		return 1
	}
	return int32((uint32(target.ObjFlags) & 0xff) >> 5 & 1)
}
