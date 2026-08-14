package server

const pentagramCollideTriggered4EAB20 = uint32(1)

type pentagramCollideHooks4EAB20[O, U any] struct {
	loadUpdateData func(O) U
	storeTriggered func(U, uint32)
}

// pentagramCollide4EAB20 preserves GAME.EXE 004EAB20. The callback loads the
// source update-data pointer once, stores the fixed dword one at offset four,
// and leaves the original source in EAX. Registered target and collision
// arguments are not observed.
func pentagramCollide4EAB20[O, T, C, U any](
	source O,
	_ T,
	_ C,
	hooks pentagramCollideHooks4EAB20[O, U],
) O {
	data := hooks.loadUpdateData(source)
	hooks.storeTriggered(data, pentagramCollideTriggered4EAB20)
	return source
}
