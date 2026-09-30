package server

const (
	mimicCheckMorphMoveTo534950        = uint32(4)
	mimicCheckMorphChest534950         = uint32(33)
	mimicCheckMorphMonster534950       = uint32(34)
	mimicCheckMorphUninterrupt534950   = uint32(61)
	mimicCheckMorphChestFlag534950     = uint32(0x40000)
	mimicCheckMorphSound534950         = uint32(460)
	mimicCheckMorphDistanceLimit534950 = float32(64) // GAME.EXE 00583C7C
)

type mimicCheckMorphHooks534950[O, U, A any] struct {
	update     func(O) U
	stackIndex func(U) int8
	head       func(U, int8) A
	action     func(A) uint32
	targetX    func(A) float32
	posX       func(O) float32
	targetY    func(A) float32
	posY       func(O) float32
	status     func(U) uint32
	frame      func() uint32
	timer      func(U) uint32
	tickRate   func() uint32
	pushAction func(O, uint32) A
	audio      func(uint32, O, int32, uint32)
}

// mimicCheckMorph534950 preserves the whole GAME.EXE 00534950 body. The
// update pointer, signed index, stack address and action are cached in that
// order. Only MOVE_TO reads coordinates. Its unspilled x87 subtraction,
// squares (Y first) and sum retain 53-bit precision without binary32 delta
// stores or contraction; an unordered distance takes the idle branch.
// Neither a nil push result nor a full stack suppresses the second push or
// sound. There are no added object/class/stack guards in the original body.
func mimicCheckMorph534950[O, U, A any](obj O, h mimicCheckMorphHooks534950[O, U, A]) {
	update := h.update(obj)
	index := h.stackIndex(update)
	head := h.head(update, index)
	action := h.action(head)
	active := action != 0
	if action == mimicCheckMorphMoveTo534950 {
		x := h.targetX(head)
		objectX := h.posX(obj)
		dx := logicRandomFloatSub64_416030(float64(x), float64(objectX))
		y := h.targetY(head)
		objectY := h.posY(obj)
		dy := logicRandomFloatSub64_416030(float64(y), float64(objectY))
		ySquared := logicRandomFloatMul64_416030(dy, dy)
		xSquared := logicRandomFloatMul64_416030(dx, dx)
		distance := logicRandomFloatAdd64_416030(ySquared, xSquared)
		active = distance > float64(mimicCheckMorphDistanceLimit534950)
	}

	morph := mimicCheckMorphMonster534950
	if active {
		if action == mimicCheckMorphMonster534950 {
			return
		}
		if h.status(update)&mimicCheckMorphChestFlag534950 == 0 {
			return
		}
	} else {
		if h.status(update)&mimicCheckMorphChestFlag534950 != 0 {
			return
		}
		frame := h.frame()
		timer := h.timer(update)
		tickRate := h.tickRate()
		if frame-timer <= tickRate { // original SUB / unsigned JBE, including wrap
			return
		}
		morph = mimicCheckMorphChest534950
	}
	h.pushAction(obj, mimicCheckMorphUninterrupt534950)
	h.pushAction(obj, morph)
	h.audio(mimicCheckMorphSound534950, obj, 0, 0)
}
