package legacy

const (
	objectEnabledFlag4E75B0   uint32 = 0x01000000
	objectNoCollideFlag4E75B0 uint32 = 0x00000040
	objectMissileClass4E75B0  uint32 = 0x00000001
	objectElevatorClass4E75B0 uint32 = 0x00004000
	objectClearClass4E75B0    uint32 = 0x10042000
)

type objectSetOnHooks4E75B0[O comparable] struct {
	flags              func(O) uint32
	class              func(O) uint32
	audio              func(O)
	setOnOff           func(O, bool)
	clearFlags         func(O, uint32)
	hasCollideOrUpdate func(O) byte
}

// objectSetOn4E75B0 is the pointer-width-independent contract for GAME.EXE
// 004E75B0. The class is deliberately reloaded after setOnOff and then cached
// across the NoCollide update and Missile decision.
func objectSetOn4E75B0[O comparable](obj O, h objectSetOnHooks4E75B0[O]) byte {
	if h.flags(obj)&objectEnabledFlag4E75B0 == 0 {
		if h.class(obj)&objectElevatorClass4E75B0 != 0 {
			h.audio(obj)
		}
	}

	h.setOnOff(obj, true)
	class := h.class(obj)
	if class&objectClearClass4E75B0 != 0 {
		h.clearFlags(obj, objectNoCollideFlag4E75B0)
	}
	if byte(class)&byte(objectMissileClass4E75B0) != 0 {
		return byte(class)
	}
	return h.hasCollideOrUpdate(obj)
}
