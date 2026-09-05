package server

const lifetimeUpdateDeadFlag53B8F0 = uint32(0x8000)

type lifetimeUpdateHooks53B8F0[O, U any, D comparable] struct {
	frame             func() uint32
	loadCreationFrame func(O) uint32
	loadUpdateData    func(O) U
	loadDuration      func(U) uint32
	loadFlags         func(O) uint32
	loadDeath         func(O) D
	storeFlags        func(O, uint32)
	callDeath         func(D, O)
	delayedDelete     func(O)
}

// lifetimeUpdate53B8F0 preserves GAME.EXE 0053B8F0. Its expiry test is a
// strict unsigned comparison: an object lives while frame-creationFrame is
// less than or equal to its configured duration. On expiry, the Dead bit is
// stored before the cached death callback or delayed-delete fallback runs.
func lifetimeUpdate53B8F0[O, U any, D comparable](
	source O,
	hooks lifetimeUpdateHooks53B8F0[O, U, D],
) {
	frame := hooks.frame()
	creationFrame := hooks.loadCreationFrame(source)
	age := frame - creationFrame
	updateData := hooks.loadUpdateData(source)
	duration := hooks.loadDuration(updateData)
	if age <= duration {
		return
	}

	flags := hooks.loadFlags(source)
	death := hooks.loadDeath(source)
	hooks.storeFlags(source, flags|lifetimeUpdateDeadFlag53B8F0)

	var zeroDeath D
	if death != zeroDeath {
		hooks.callDeath(death, source)
		return
	}
	hooks.delayedDelete(source)
}
