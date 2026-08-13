package legacy

type objectNormalizeDirectionHooks4E7680[O comparable] struct {
	direction    func(O) int16
	addDirection func(O, int16)
}

// objectNormalizeDirection4E7680 is the pointer-width-independent contract
// for GAME.EXE 004E7680. Comparisons interpret the 16-bit direction as signed,
// while each add or subtract wraps in the original 16-bit storage.
func objectNormalizeDirection4E7680[O comparable](obj O, h objectNormalizeDirectionHooks4E7680[O]) O {
	for h.direction(obj) < 0 {
		h.addDirection(obj, 256)
	}
	for h.direction(obj) >= 256 {
		h.addDirection(obj, -256)
	}
	return obj
}
