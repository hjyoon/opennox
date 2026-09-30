package legacy

const questBriefingDrawSource44F300 = `C:\NoxPost\src\client\Gui\GUIBrief.c`

type questBriefingColor44F300 uint8

const (
	questBriefingBlack44F300 questBriefingColor44F300 = iota
	questBriefingOrange44F300
	questBriefingWhite44F300
)

type questBriefingDrawHooks44F300[V, D, F, T any] struct {
	viewport       func() V
	initPreviews   func()
	resetParticles func()
	hideBook       func(int32) int32
	prepare        func()
	dimensions     func() (int32, int32)
	loadText       func(string, string, int32) T
	titleFont      func() F
	normalFont     func() F
	measure        func(F, T) int32
	color          func(questBriefingColor44F300) uint32
	setColor       func(uint32)
	drawText       func(F, T, int32, int32) int32
	wrapText       func(F, T, int32, int32, int32, int32) int32
	loadDrawable   func(int) D
	position       func(V, D, int32, int32)
	drawDrawable   func(V, D)
	frame          func() uint32
	loadBlink      func() uint32
	storeBlink     func(uint32)
}

// questBriefingDraw44F300 implements the entire sealed 0044F300..00450154
// body. Coordinates and widths deliberately retain signed dword wrapping and
// division toward zero. Native pointers are never coordinates or PE32 offsets.
func questBriefingDraw44F300[V, D, F, T any](h questBriefingDrawHooks44F300[V, D, F, T]) int32 {
	vp := h.viewport()
	h.initPreviews()
	h.resetParticles()
	h.hideBook(1)
	h.prepare()
	width, height := h.dimensions()
	dx, dy := (width-640)/2, (height-480)/2
	load := func(key string, line int32) T {
		return h.loadText("GeneralPrint:QuestSplash"+key, questBriefingDrawSource44F300, line)
	}
	color := func(c questBriefingColor44F300) { h.setColor(h.color(c)) }
	draw := func(text T, x, y int32) { h.drawText(h.normalFont(), text, x, y) }
	wrap := func(text T, x, y, width int32) { h.wrapText(h.normalFont(), text, x, y, width, 0) }
	icon := func(slot int, x, y int32) {
		h.position(vp, h.loadDrawable(slot), dx+x, dy+y)
		// PE32 reloads the cache after mapping and before the indirect call.
		h.drawDrawable(vp, h.loadDrawable(slot))
	}
	title := load("1", 765)
	tw := h.measure(h.titleFont(), title)
	tx, ty := dx-tw/2+320, dy+20
	color(questBriefingBlack44F300)
	h.drawText(h.titleFont(), title, tx-1, ty-1)
	h.drawText(h.titleFont(), title, tx+1, ty-1)
	h.drawText(h.titleFont(), title, tx-1, ty+1)
	h.drawText(h.titleFont(), title, tx+1, ty+1)
	color(questBriefingOrange44F300)
	h.drawText(h.titleFont(), title, tx, ty)

	// The short orange prefix is drawn before measuring it, then the white
	// suffix is looked up after its color change. Font loads remain live.
	prefixWrap := func(aKey, bKey string, aLine, bLine, x, y, right int32) {
		a := load(aKey, aLine)
		color(questBriefingOrange44F300)
		draw(a, x, y)
		aw := h.measure(h.normalFont(), a)
		x += aw + 4
		color(questBriefingWhite44F300)
		b := load(bKey, bLine)
		wrap(b, x, y, right-x)
	}
	// Right-aligned pairs measure A before loading B. The signed sum alone
	// selects the unwrapped branch, with equality included.
	rightPair := func(aKey, bKey string, aLine, bLine, right, limit, left, y int32) {
		a := load(aKey, aLine)
		aw := h.measure(h.normalFont(), a)
		b := load(bKey, bLine)
		bw := h.measure(h.normalFont(), b)
		if bw+aw <= limit {
			color(questBriefingOrange44F300)
			draw(a, right-bw-aw-4, y)
			color(questBriefingWhite44F300)
			draw(b, right-bw, y)
		} else {
			color(questBriefingOrange44F300)
			draw(a, left, y)
			x := left + aw + 4
			color(questBriefingWhite44F300)
			wrap(b, x, y, right-x)
		}
	}
	icon(0, 73, 123)
	prefixWrap("2a", "2b", 792, 799, dx+109, dy+76, dx+520)
	icon(1, 565, 117)
	rightPair("3a", "3b", 809, 811, dx+520, 390, dx+199, dy+115)
	icon(3, 133, 192)
	// The original width calculation really uses dy, not dx (0044F7A4).
	prefixWrap("4a", "4b", 862, 869, dx+157, dy+156, dy+630)
	icon(2, 525, 222)
	rightPair("7a", "7b", 879, 881, dx+500, 215, dx+250, dy+198)
	icon(9, 182, 262)
	icon(11, 201, 251)
	icon(10, 185, 234)
	prefixWrap("5a", "5b", 942, 949, dx+221, dy+240, dx+470)
	icon(6, 484, 278)
	icon(7, 503, 303)
	rightPair("6a", "6b", 964, 966, dx+462, 350, dx+113, dy+286)
	icon(5, 186, 333)
	icon(4, 219, 345)
	icon(8, 220, 322)
	prefixWrap("8a", "8b", 1027, 1034, dx+241, dy+330, dx+550)

	// Footer pairs, unlike the pairs above, load both strings before either
	// measurement. The four-pixel suffix gap is not part of centering.
	footer := func(aKey, bKey string, aLine, bLine, y int32) {
		a, b := load(aKey, aLine), load(bKey, bLine)
		aw := h.measure(h.normalFont(), a)
		bw := h.measure(h.normalFont(), b)
		x := dx - (aw+bw)/2 + 320
		color(questBriefingOrange44F300)
		draw(a, x, y)
		x += aw + 4
		color(questBriefingWhite44F300)
		draw(b, x, y)
	}
	footer("9a", "9b", 1040, 1041, dy+370)
	footer("10a", "10b", 1055, 1056, dy+395)
	footer("11a", "11b", 1070, 1071, dy+420)

	// One unsigned frame load and division in the actual instructions, not
	// the three apparent gameFrame calls in the C decompiler.
	frame := h.frame()
	if frame%30 == 0 {
		if h.loadBlink() == 1 {
			h.storeBlink(0)
			return 1
		}
		h.storeBlink(1)
	} else if h.loadBlink() != 1 {
		return int32(frame / 30)
	}
	white := h.color(questBriefingWhite44F300) // cached before text callbacks
	prompt := load("12", 1097)
	pw := h.measure(h.normalFont(), prompt)
	h.setColor(white)
	// Original prompt measures normal font, but draws with the window font.
	return h.drawText(h.titleFont(), prompt, dx-pw/2+320, dy+450)
}
