package legacy

const (
	unitFreezeFlag4E79C0   uint32 = 0x00000002
	unitDeadFlag4E79C0     uint32 = 0x00008000
	unitMonsterClass4E79C0 uint32 = 0x00000002
	unitPlayerClass4E79C0  uint32 = 0x00000004
)

type unitFreezeHooks4E79C0[O comparable] struct {
	flags         func(O) uint32
	setFlags      func(O, uint32)
	class         func(O) uint32
	gate          func() uint32
	setGate       func(uint32)
	reportStatus  func(O) byte
	setPlayerIdle func(O)
	raiseZero     func(O)
	resetPaths    func()
	firstOwned    func(O) O
	nextOwned     func(O) O
	// monsterStatus returns the low byte left by the original update-data
	// pointer load as well as the Summoned status bit. Only unfreeze exposes
	// that register artifact, and none of the direct callers consumes it.
	monsterStatus func(O) (loadByte byte, summoned bool)
	pushIdle      func(O) byte
	popAction     func(O) byte
}

// unitFreeze4E79C0 preserves the load, callback, recursive traversal and
// post-callback reload order of GAME.EXE 004E79C0 without assuming a pointer
// width for O.
func unitFreeze4E79C0[O comparable](obj O, source uint32, h unitFreezeHooks4E79C0[O]) byte {
	flags := h.flags(obj)
	if flags&unitFreezeFlag4E79C0 != 0 {
		return byte(flags)
	}
	h.setFlags(obj, flags|unitFreezeFlag4E79C0)

	if byte(h.class(obj))&byte(unitPlayerClass4E79C0) != 0 {
		if h.gate() == 0 {
			h.setGate(source)
		}
		h.reportStatus(obj)
		h.setPlayerIdle(obj)
		h.raiseZero(obj)
		h.resetPaths()

		var zero O
		for it := h.firstOwned(obj); it != zero; it = h.nextOwned(it) {
			if byte(h.class(it))&byte(unitMonsterClass4E79C0) == 0 {
				continue
			}
			_, summoned := h.monsterStatus(it)
			if summoned {
				unitFreeze4E79C0(it, source, h)
			}
		}
	}

	class := byte(h.class(obj))
	if class&byte(unitMonsterClass4E79C0) == 0 {
		return class
	}
	flags = h.flags(obj)
	if flags&unitDeadFlag4E79C0 != 0 {
		return byte(flags)
	}
	return h.pushIdle(obj)
}

// unitUnfreeze4E7A60 preserves GAME.EXE 004E7A60. In particular, a player
// with a nonzero gate and force == 0 returns before clearing any state, while
// the forced path reloads flags before clearing them and reloads each owned
// successor after recursive callbacks.
func unitUnfreeze4E7A60[O comparable](obj O, force uint32, h unitFreezeHooks4E79C0[O]) byte {
	flags := h.flags(obj)
	if flags&unitFreezeFlag4E79C0 == 0 {
		return byte(flags)
	}

	var result byte
	if byte(h.class(obj))&byte(unitPlayerClass4E79C0) != 0 {
		gate := h.gate()
		result = byte(gate)
		if gate != 0 && force == 0 {
			return result
		}
		h.setGate(0)
		flags = h.flags(obj)
		h.setFlags(obj, flags&^unitFreezeFlag4E79C0)
		result = h.reportStatus(obj)

		var zero O
		for it := h.firstOwned(obj); it != zero; it = h.nextOwned(it) {
			if byte(h.class(it))&byte(unitMonsterClass4E79C0) == 0 {
				continue
			}
			loadByte, summoned := h.monsterStatus(it)
			result = loadByte
			if summoned {
				result = unitUnfreeze4E7A60(it, force, h)
			}
		}
	} else {
		flags &^= unitFreezeFlag4E79C0
		h.setFlags(obj, flags)
		result = byte(flags)
	}

	if byte(h.class(obj))&byte(unitMonsterClass4E79C0) == 0 {
		return result
	}
	flags = h.flags(obj)
	if flags&unitDeadFlag4E79C0 != 0 {
		return byte(flags)
	}
	return h.popAction(obj)
}
