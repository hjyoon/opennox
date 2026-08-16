package server

import (
	"image"

	"github.com/opennox/libs/types"
)

const (
	randomReachablePointPi4ED970           = float32(0x1.921fb6p+1)
	randomReachablePointAngleStep4ED970    = float32(0x1.e28c76p+0)
	randomReachablePointRadiusStep4ED970   = float32(0x1p-6)
	randomReachablePointAttemptLimit4ED970 = 64
	randomReachablePointSource4ED970       = `C:\NoxPost\src\Server\Object\pickdrop\drop.c`
	randomReachablePointSourceLine4ED970   = int32(728)
	randomReachablePointTraceFlag4ED970    = uint8(1)
)

// randomReachablePointRay4ED970 is the local float4 passed by address to the
// original map trace. The same local survives every attempt, so a trace-side
// mutation of the origin remains visible to later trace calls.
type randomReachablePointRay4ED970 struct {
	Origin      types.Pointf
	Destination types.Pointf
}

// randomReachablePointHooks4ED970 exposes the original argument, field,
// random, trigonometric, trace, and output access order. The center pointer is
// cached once, but its coordinates are reloaded for every candidate and again
// on the 64-failure fallback. The output pointer is deliberately delayed until
// one of the two terminal copy paths.
type randomReachablePointHooks4ED970[C, O any] struct {
	loadRadiusArg func() float32
	loadCenterArg func() C
	loadCenterX   func(C) float32
	loadCenterY   func(C) float32

	randomFloat func(min, max float32, source string, line int32) float64
	cosine      func(float64) float64
	sine        func(float64) float64
	mapTrace    func(*randomReachablePointRay4ED970, *types.Pointf, *image.Point, uint8) int32

	loadOutputArg func() O
	storeOutputX  func(O, float32)
	storeOutputY  func(O, float32)
}

// Keep each GAME.EXE x87 arithmetic instruction at an explicit binary64
// boundary. Nox runs this code with 53-bit precision control, and separate
// helpers prevent a target compiler from contracting multiply-plus-add into
// an FMA. Explicit float32 spills model FSTP/FSTS boundaries.
//
//go:noinline
func randomReachablePointAdd64_4ED970(a, b float64) float64 { return a + b }

//go:noinline
func randomReachablePointSub64_4ED970(a, b float64) float64 { return a - b }

//go:noinline
func randomReachablePointMul64_4ED970(a, b float64) float64 { return a * b }

//go:noinline
func randomReachablePointSpill32_4ED970(value float64) float32 { return float32(value) }

// randomReachablePoint4ED970 preserves GAME.EXE 004ED970. The cosine receives
// the unspilled x87 sum, whereas the next angle and sine receive its binary32
// spill. Candidate coordinates retain the trigonometric result through the
// multiply and add before one binary32 destination spill. MapTrace receives
// the address of the persistent local ray, nil optional outputs, and flag 1;
// every whole nonzero result succeeds. Each failure performs the radius
// subtraction and spill, including failure 64. Success copies the trace-live
// destination; exhaustion instead reloads the live center. Both return the
// exact output handle after X-then-Y stores.
func randomReachablePoint4ED970[C, O any](hooks randomReachablePointHooks4ED970[C, O]) O {
	radius := hooks.loadRadiusArg()
	stepExtended := randomReachablePointMul64_4ED970(
		float64(radius),
		float64(randomReachablePointRadiusStep4ED970),
	)
	center := hooks.loadCenterArg()
	ray := randomReachablePointRay4ED970{
		Origin: types.Pointf{
			X: hooks.loadCenterX(center),
			Y: hooks.loadCenterY(center),
		},
	}
	step := randomReachablePointSpill32_4ED970(stepExtended)
	angle := randomReachablePointSpill32_4ED970(hooks.randomFloat(
		-randomReachablePointPi4ED970,
		randomReachablePointPi4ED970,
		randomReachablePointSource4ED970,
		randomReachablePointSourceLine4ED970,
	))

	for attempt := 0; ; attempt++ {
		angleExtended := randomReachablePointAdd64_4ED970(
			float64(angle),
			float64(randomReachablePointAngleStep4ED970),
		)
		angle = randomReachablePointSpill32_4ED970(angleExtended)

		cosine := hooks.cosine(angleExtended)
		centerX := hooks.loadCenterX(center)
		ray.Destination.X = randomReachablePointSpill32_4ED970(
			randomReachablePointAdd64_4ED970(
				randomReachablePointMul64_4ED970(cosine, float64(radius)),
				float64(centerX),
			),
		)

		sine := hooks.sine(float64(angle))
		centerY := hooks.loadCenterY(center)
		ray.Destination.Y = randomReachablePointSpill32_4ED970(
			randomReachablePointAdd64_4ED970(
				randomReachablePointMul64_4ED970(sine, float64(radius)),
				float64(centerY),
			),
		)

		if hooks.mapTrace(&ray, nil, nil, randomReachablePointTraceFlag4ED970) != 0 {
			output := hooks.loadOutputArg()
			x := ray.Destination.X
			y := ray.Destination.Y
			hooks.storeOutputX(output, x)
			hooks.storeOutputY(output, y)
			return output
		}

		radius = randomReachablePointSpill32_4ED970(
			randomReachablePointSub64_4ED970(float64(radius), float64(step)),
		)
		if attempt+1 >= randomReachablePointAttemptLimit4ED970 {
			output := hooks.loadOutputArg()
			x := hooks.loadCenterX(center)
			hooks.storeOutputX(output, x)
			y := hooks.loadCenterY(center)
			hooks.storeOutputY(output, y)
			return output
		}
	}
}
