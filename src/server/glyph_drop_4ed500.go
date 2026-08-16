package server

const glyphDropAudio4ED500 = uint32(825)

// glyphDropHooks4ED500 keeps GAME.EXE 004ED500's object, glyph-data, and
// point domains separate. Point and object coordinates are deliberately
// reloaded after the glyph-data stores because the original pointers may
// alias one another.
type glyphDropHooks4ED500[O, D, P comparable] struct {
	dropTrap func(O, O, P) int32

	loadInitData func(O) D
	loadPointX   func(P) float32
	storeGlyphX  func(D, float32)
	loadPointY   func(P) float32
	storeGlyphY  func(D, float32)

	loadObjectX     func(O) float32
	loadObjectY     func(O) float32
	vectorDirection func(float32, float32) int32
	storeDirection2 func(O, uint16)
	storeDirection1 func(O, uint16)
	audio           func(uint32, O, int32, uint32)
}

// glyphDropSubtract4ED500 models an x87 single-precision load/subtract followed
// by FSTP to the local binary32 vector. float64 represents the subtraction
// exactly for two binary32 operands before the explicit binary32 spill.
func glyphDropSubtract4ED500(left, right float32) float32 {
	return float32(float64(left) - float64(right))
}

// glyphDrop4ED500 preserves GAME.EXE 004ED500. TrapDrop owns the entry guards;
// a zero result returns without touching any object or point field. On success
// the GlyphInitData point is written before the live position reloads, and the
// low AX word of the direction result is stored to Direction2 then Direction1.
func glyphDrop4ED500[O, D, P comparable](
	owner, glyph O,
	point P,
	hooks glyphDropHooks4ED500[O, D, P],
) int32 {
	if hooks.dropTrap(owner, glyph, point) == 0 {
		return 0
	}

	data := hooks.loadInitData(glyph)
	x := hooks.loadPointX(point)
	hooks.storeGlyphX(data, x)
	y := hooks.loadPointY(point)
	hooks.storeGlyphY(data, y)

	vectorX := glyphDropSubtract4ED500(hooks.loadObjectX(owner), hooks.loadPointX(point))
	vectorY := glyphDropSubtract4ED500(hooks.loadObjectY(owner), hooks.loadPointY(point))
	direction := uint16(hooks.vectorDirection(vectorX, vectorY))
	hooks.storeDirection2(glyph, direction)
	hooks.storeDirection1(glyph, direction)
	hooks.audio(glyphDropAudio4ED500, glyph, 0, 0)
	return 1
}
