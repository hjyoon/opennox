package server

import "math"

const (
	poisonGasTrapCloudType4EB910 = "ToxicCloud"
	poisonGasTrapLifetime4EB910  = "ToxicCloudLifetime"
	poisonGasTrapTriggered4EB910 = uint32(847)
)

type poisonGasTrapCollideHooks4EB910[O comparable, D any] struct {
	allowed        func(O, O) int32
	newObject      func(string) O
	loadPosY       func(O) float32
	loadPosX       func(O) float32
	loadOwner      func(O) O
	createAt       func(O, O, float32, float32, uint32)
	loadUpdateData func(O) D
	loadLifetime   func(string) float32
	loadFPS        func() uint32
	multiply       func(float32, uint32) float32
	floatToInt     func(float32) int32
	storeDuration  func(D, int32)
	audio          func(uint32, O, int32, uint32)
	delayedDelete  func(O)
}

// poisonGasTrapCollide4EB910 preserves GAME.EXE 004EB910. A nil target
// returns before the source is used. The shared glyph gate and ToxicCloud
// allocation must both succeed before source Y, X, and owner are read in that
// order. CreateAt receives the original call site's reserved zero slot, after
// which the cloud's live UpdateData pointer is cached. Balance is loaded
// before the signed FPS dword; their product is spilled once to binary32 and
// converted with x87 round-to-nearest-even semantics. The duration store
// precedes source audio and delayed deletion. Collision is never read.
func poisonGasTrapCollide4EB910[O comparable, D any](
	source, target O,
	_ any,
	hooks poisonGasTrapCollideHooks4EB910[O, D],
) {
	var zero O
	if target == zero {
		return
	}
	if hooks.allowed(source, target) == 0 {
		return
	}
	cloud := hooks.newObject(poisonGasTrapCloudType4EB910)
	if cloud == zero {
		return
	}

	y := hooks.loadPosY(source)
	x := hooks.loadPosX(source)
	owner := hooks.loadOwner(source)
	hooks.createAt(cloud, owner, x, y, 0)
	data := hooks.loadUpdateData(cloud)
	lifetime := hooks.loadLifetime(poisonGasTrapLifetime4EB910)
	fps := hooks.loadFPS()
	duration := hooks.floatToInt(hooks.multiply(lifetime, fps))
	hooks.storeDuration(data, duration)
	hooks.audio(poisonGasTrapTriggered4EB910, source, 0, 0)
	hooks.delayedDelete(source)
}

//go:noinline
func poisonGasTrapMultiply4EB910(lifetime float32, fps uint32) float32 {
	// GAME.EXE uses FIMUL with a signed dword while the x87 precision control
	// is 53 bits, then spills the product once to binary32.
	return float32(float64(lifetime) * float64(int32(fps)))
}

func poisonGasTrapRound4EB910(value float32) int32 {
	// nox_float2int at 00419A70 uses FISTP in the default
	// round-to-nearest-even mode. Invalid conversions yield 0x80000000.
	if math.IsNaN(float64(value)) || value >= 2147483648 || value < -2147483648 {
		return math.MinInt32
	}
	return int32(math.RoundToEven(float64(value)))
}
