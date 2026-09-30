package legacy

const questStartBriefingSource450A30 = `C:\NoxPost\src\client\Gui\GUIBrief.c`

type questStartBriefingHooks450A30[P, I, T any] struct {
	storeState     func(uint32)
	resetParticles func()
	hideBook       func(int32) int32
	prepare        func()
	offset         func(P, uintptr) P
	loadImage      func(P) I
	setImage       func(I)
	stringLength   func(P) uintptr
	loadText       func(P, string, int32) T
	emptyText      func() T
	setText        func(T)
	stage          func(P) uint16
	setStage       func(uint32)
	lock           func(int32, int32, int8) int32
}

// questStartBriefing450A30 implements the entire sealed GAME.EXE body
// 00450A30..00450AC7. Packet fields are read at their original call boundaries,
// not parsed up front: GUI/image/text callbacks precede the live stage load.
func questStartBriefing450A30[P, I, T any](packet P, show int32, h questStartBriefingHooks450A30[P, I, T]) int32 {
	h.storeState(0)
	h.resetParticles()
	h.hideBook(1) // Its return, like both setter returns, is discarded in PE32.
	h.prepare()
	h.setImage(h.loadImage(h.offset(packet, 5)))
	key := h.offset(packet, 37)
	if h.stringLength(key) != 0 {
		h.setText(h.loadText(key, questStartBriefingSource450A30, 1756))
	} else {
		h.setText(h.emptyText())
	}
	h.setStage(uint32(h.stage(packet)))
	if show != 0 {
		return h.lock(254, 1, 4)
	}
	return show
}
