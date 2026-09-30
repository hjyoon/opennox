package legacy

const monsterGeneratorDrawSource4BC750 = `C:\NoxPost\src\client\Draw\MGenDraw.c`

type monsterGeneratorDrawHooks4BC750[D, P, I any] struct {
	flags      func() uint32
	data       func() D
	images     func(D, int) P
	count      func(D, int) uint8
	kind       func(D, int) uint32
	delay      func(D, int) uint8
	delayIndex func(D, int) int32
	netcode    func() uint32
	frame      func() uint32
	slave      func() uint32
	random     func(int32, int32, string, int32) int32
	class      func() uint32
	setClass   func(uint32)
	objFlags   func() uint32
	setFlags   func(uint32)
	timer      func() uint32
	setTimer   func(uint32)
	setState   func(uint32)
	image      func(P, int32) I
	draw       func(I)
}

// monsterGeneratorDraw4BC750 is the complete sealed 004BC750..004BC8F8
// body. Pointer tables stay native; frame words retain x86 dword arithmetic.
func monsterGeneratorDraw4BC750[D, P, I any](h monsterGeneratorDrawHooks4BC750[D, P, I]) int32 {
	flags := h.flags()
	data := h.data()
	state := 0
	switch {
	case flags&0x100 != 0:
		state = 1
	case flags&0x200 != 0:
		state = 2
	case flags&0xc00 != 0:
		state = 3
	}
	main := h.images(data, state) // cached before count/kind/delay
	count := int32(h.count(data, state))
	kind := h.kind(data, state)
	delay := uint32(h.delay(data, state))
	var frame int32
	switch kind {
	case 0:
		// This is an address used as a scalar frame index in the original.
		// Do not reject it: the terminal/timer paths can replace the index.
		frame = h.delayIndex(data, state)
	case 2:
		netcode := h.netcode()
		tick := h.frame()
		frame = int32((netcode + tick) / (delay + 1))
		if frame >= count { // signed CMP/JL, followed by signed IDIV
			frame %= count
		}
	case 4:
		// The upper bound really is inclusive count, not count-1.
		frame = h.random(0, count, monsterGeneratorDrawSource4BC750, 86)
	case 5:
		frame = int32(h.slave())
	default:
		return 0
	}
	if h.flags()&0x800 != 0 {
		class := h.class()
		objFlags := h.objFlags()
		frame = count - 1
		h.setClass(class &^ 0x80000)
		h.setFlags(objFlags &^ 0x20000000)
	}
	timer := h.timer()
	if timer != 0 {
		delay := uint32(h.delay(data, state)) + 1
		liveCount := uint32(h.count(data, state))
		frame = int32((delay*liveCount - timer) / delay)
		if frame >= count {
			frame = count - 1
		}
		if frame < 0 {
			frame = 0
		}
		timer--
		h.setTimer(timer)
		if timer == 0 {
			h.setState(h.flags()&^0x400 | 0x800)
		}
	}
	h.draw(h.image(main, frame))
	// The main callback can change state and the cached data's overlay.
	if h.flags()&0xc00 == 0 {
		netcode := h.netcode()
		tick := h.frame()
		overlay := h.images(data, 4)
		delay := uint32(h.delay(data, 4))
		count := int32(h.count(data, 4))
		frame := int32((netcode + tick) / (delay + 1))
		if frame >= count {
			frame %= count
		}
		h.draw(h.image(overlay, frame))
	}
	if h.flags()&0x800 != 0 {
		h.setFlags(h.objFlags() | 1)
	}
	return 1
}
