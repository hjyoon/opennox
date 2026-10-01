package opennox

type optionsShowHooks4AA6B0[W comparable, A comparable, I any] struct {
	addState       func(int32)
	newWindow      func(string) W
	storeRoot      func(W)
	advanced       func(W) int32
	root           func() W
	setProc        func(W)
	tabWidth       func(int32)
	newAnimation   func(W, [8]int32) A
	storeAnimation func(A)
	animation      func() A
	stateID        func(A, int32)
	startOut       func(A)
	doneOut        func(A)
	child          func(W, int32) W
	thumb          func(W) W
	width          func(W, int32)
	height         func(W, int32)
	loadImage      func(string) I
	images         func(W, I, I, I)
	event          func(W, int32, uint32, uint32)
	current        func(int32) uint32
	storeCheckbox  func(int32, W)
	enabled        func(int32) int32
	checkbox       func(int32) W
	flags          func(W) uint32
	storeFlags     func(W, uint32)
	returnNull     func(W)
	backText       func(string)
	backEnabled    func(int32)
	video          func()
}

// optionsShow4AA6B0 restores the complete GAME.EXE 004AA6B0 constructor.
// Publish before each failure check; reload globals at the original accesses.
// Only the returned root (for advanced options), animation (for its state ID),
// and each slider are cached. The three image lookups are not coalesced.
func optionsShow4AA6B0[W comparable, A comparable, I any](h optionsShowHooks4AA6B0[W, A, I]) int32 {
	h.addState(300)
	root := h.newWindow("Options.wnd")
	h.storeRoot(root)
	var noWindow W
	if root == noWindow {
		return 0
	}
	if h.advanced(root) == 0 {
		return 0
	}
	h.setProc(h.root())
	h.tabWidth(15)
	anim := h.newAnimation(h.root(), [8]int32{0, 0, 0, -480, 0, 20, 0, -40})
	h.storeAnimation(anim)
	var noAnimation A
	if anim == noAnimation {
		return 0
	}
	h.stateID(anim, 300)
	h.startOut(h.animation())
	h.doneOut(h.animation())
	for channel := int32(0); channel < 3; channel++ {
		slider := h.child(h.root(), 351+channel)
		h.width(h.thumb(slider), 24)
		h.height(h.thumb(slider), 20)
		highlight := h.loadImage("OptionsVolumeSliderLit")
		selected := h.loadImage("OptionsVolumeSliderLit")
		enabled := h.loadImage("OptionsVolumeSlider")
		h.images(slider, enabled, selected, highlight)
		h.event(slider, 0x400b, 0, 0x4000)
		current := h.current(channel)
		h.event(slider, 0x400a, current>>16, 0)
		checkbox := h.child(h.root(), 361+channel)
		h.storeCheckbox(channel, checkbox)
		checked := h.enabled(channel)
		checkbox = h.checkbox(channel)
		flags := h.flags(checkbox)
		if checked == 1 {
			flags |= 4
		} else {
			flags &^= 4
		}
		h.storeFlags(checkbox, flags)
	}
	h.returnNull(h.root())
	h.backText("OptsBack.wnd:Back")
	h.backEnabled(0)
	h.video()
	return 1
}
