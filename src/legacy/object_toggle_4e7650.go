package legacy

const objectEnabledFlag4E7650 uint32 = 0x01000000

type objectToggleHooks4E7650[O comparable] struct {
	flags  func(O) uint32
	setOff func(O) uint32
	setOn  func(O) byte
}

// objectToggle4E7650 is the pointer-width-independent contract for GAME.EXE
// 004E7650. It reads flags once, dispatches to exactly one state transition,
// and returns AL. The extra boolean preserves the modern API's prior-state
// result without introducing a second flags read.
func objectToggle4E7650[O comparable](obj O, h objectToggleHooks4E7650[O]) (result byte, wasEnabled bool) {
	if h.flags(obj)&objectEnabledFlag4E7650 == objectEnabledFlag4E7650 {
		return byte(h.setOff(obj)), true
	}
	return h.setOn(obj), false
}
