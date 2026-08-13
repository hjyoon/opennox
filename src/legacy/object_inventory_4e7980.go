package legacy

// objectInventoryFirst4E7980 preserves the unconditional first-item load in
// GAME.EXE 004E7980. In particular, a nil object is not handled here: the
// supplied load has the same faulting contract as the original field read.
func objectInventoryFirst4E7980[O comparable](obj O, loadFirst func(O) O) O {
	return loadFirst(obj)
}

// objectInventoryNext4E7990 preserves the null guard in GAME.EXE 004E7990.
// The next-item field must not be read when obj is the zero pointer value.
func objectInventoryNext4E7990[O comparable](obj O, loadNext func(O) O) O {
	var zero O
	if obj == zero {
		return zero
	}
	return loadNext(obj)
}
