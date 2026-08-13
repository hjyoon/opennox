package legacy

// objectIsPlayer4E7BC0 preserves GAME.EXE 004E7BC0. A zero object returns
// before the class load; a nonzero object loads one complete 32-bit class
// value and returns its Player bit as exactly zero or one.
func objectIsPlayer4E7BC0[O comparable](obj O, loadClass func(O) uint32) uint32 {
	var zero O
	if obj == zero {
		return 0
	}
	class := loadClass(obj)
	return (class >> 2) & 1
}
