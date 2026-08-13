package legacy

const (
	objectEnabledFlag4E7600    uint32 = 0x01000000
	objectNoCollideFlag4E7600  uint32 = 0x00000040
	objectElevatorClass4E7600  uint32 = 0x00004000
	objectNoCollideClass4E7600 uint32 = 0x10042000
)

type objectSetOffHooks4E7600[O comparable] struct {
	flags    func(O) uint32
	class    func(O) uint32
	audio    func(O)
	setOnOff func(O, bool)
	setFlags func(O, uint32)
}

// objectSetOff4E7600 is the pointer-width-independent contract for GAME.EXE
// 004E7600. The post-callback class decides whether EAX returns that class or
// the full flags word after setting NoCollide through its low byte.
func objectSetOff4E7600[O comparable](obj O, h objectSetOffHooks4E7600[O]) uint32 {
	if h.flags(obj)&objectEnabledFlag4E7600 != 0 {
		if h.class(obj)&objectElevatorClass4E7600 != 0 {
			h.audio(obj)
		}
	}

	h.setOnOff(obj, false)
	class := h.class(obj)
	if class&objectNoCollideClass4E7600 == 0 {
		return class
	}
	flags := h.flags(obj) | objectNoCollideFlag4E7600
	h.setFlags(obj, flags)
	return flags
}
