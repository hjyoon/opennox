package server

import "math"

type monsterLookAtHooks5125A0[O any, A comparable] struct {
	directionToAngle func(int32) uint32
	loadClassLow     func(O) uint8
	loadFlags        func(O) uint32
	loadCosine       func(uint32) float32
	loadSine         func(uint32) float32
	loadX            func(O) float32
	loadY            func(O) float32
	pushAction       func(O, uint32) A
	storeArgBits     func(A, int, uint32)
	actionResult     func(A) uintptr
}

// GAME.EXE 005125A0 maps the script direction before either object gate,
// spills both x87 coordinate results before pushing FACE_LOCATION, and never
// clears the existing stack. Its residual result is the angle at a gate, or
// the pushed action address (zero on rejection), not a floating-point value.
func monsterLookAt5125A0[O any, A comparable](unit O, direction int32, hooks monsterLookAtHooks5125A0[O, A]) uintptr {
	angle := hooks.directionToAngle(direction)
	if hooks.loadClassLow(unit)&0x02 == 0 {
		return uintptr(angle)
	}
	if hooks.loadFlags(unit)&0x00008000 != 0 {
		return uintptr(angle)
	}
	x := float32(float64(hooks.loadCosine(angle))*10 + float64(hooks.loadX(unit)))
	y := float32(float64(hooks.loadSine(angle))*10 + float64(hooks.loadY(unit)))
	action := hooks.pushAction(unit, 25)
	var nilAction A
	if action != nilAction {
		hooks.storeArgBits(action, 0, math.Float32bits(x))
		hooks.storeArgBits(action, 1, math.Float32bits(y))
	}
	return hooks.actionResult(action)
}
