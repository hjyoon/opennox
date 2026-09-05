package server

const positionDeltaLimit4FEA70 = float32(5)

// positionDeltaHooks4FEA70 exposes the four observable coordinate loads in
// GAME.EXE 004FEA70. Keeping the object and point handles generic lets tests
// model native-width pointers without ever narrowing them to a PE32 integer.
type positionDeltaHooks4FEA70[O, P any] struct {
	loadPointX  func(P) float32
	loadObjectX func(O) float32
	loadPointY  func(P) float32
	loadObjectY func(O) float32
}

// GAME.EXE evaluates the subtractions in x87 registers. The original inputs
// are binary32, so binary64 represents each subtraction exactly while also
// giving every supported target one explicit, architecture-independent
// precision boundary.
//
//go:noinline
func positionDeltaSub64_4FEA70(left, right float32) float64 {
	return float64(left) - float64(right)
}

// positionDeltaNormalizeX87_4FEA70 models FCOM +0.0 followed by a C0 test and
// conditional FCHS. Thus negative and unordered values have their sign bit
// toggled, while negative zero compares equal and remains negative zero.
func positionDeltaNormalizeX87_4FEA70(value float64) float64 {
	if !(value >= 0) {
		return -value
	}
	return value
}

// positionDelta4FEA70 preserves GAME.EXE 004FEA70. X is normalized and
// spilled once to binary32. Y is then loaded and normalized but remains in
// binary64, modeling its live x87 value. Both coordinate pairs are therefore
// read before the X threshold can short-circuit. Ordered deltas of at least
// five return canonical one; unordered deltas never qualify.
func positionDelta4FEA70[O, P any](
	object O,
	point P,
	hooks positionDeltaHooks4FEA70[O, P],
) int32 {
	pointX := hooks.loadPointX(point)
	objectX := hooks.loadObjectX(object)
	x := float32(positionDeltaNormalizeX87_4FEA70(positionDeltaSub64_4FEA70(pointX, objectX)))

	pointY := hooks.loadPointY(point)
	objectY := hooks.loadObjectY(object)
	y := positionDeltaNormalizeX87_4FEA70(positionDeltaSub64_4FEA70(pointY, objectY))

	if x >= positionDeltaLimit4FEA70 || y >= float64(positionDeltaLimit4FEA70) {
		return 1
	}
	return 0
}
