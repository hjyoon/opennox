package server

import (
	"image"
	"math"

	"github.com/opennox/libs/types"
)

const (
	chestOpenDirectionNorthWest4EDF00 = uint32(0x00000100)
	chestOpenDirectionNorthEast4EDF00 = uint32(0x00000200)
	chestOpenDirectionSouthEast4EDF00 = uint32(0x00000400)
	chestOpenDirectionSouthWest4EDF00 = uint32(0x00000800)

	chestOpenShapeMargin4EDF00      = float32(4)
	chestOpenForwardClearance4EDF00 = float32(15)
	chestOpenLateralOffset4EDF00    = float32(30)
	chestOpenTraceFlag4EDF00        = uint8(1)
	chestOpenMonsterClass4EDF00     = uint8(0x02)
	chestOpenDropFlag4EDF00         = uint32(0x40)
	chestOpenInvalidWeight4EDF00    = uint8(0xff)

	chestShapeCircle4EE2A0 = uint32(2)
	chestShapeBox4EE2A0    = uint32(3)
	chestShapeHalf4EE2A0   = float32(0.5)
)

type chestShapeExtentHooks4EE2A0[O any] struct {
	loadShapeKind  func(O) uint32
	loadCircleR    func(O) float32
	loadBoxExtentW func(O) float32
	loadBoxExtentH func(O) float32
}

// Keep each original x87 arithmetic instruction at a binary64 boundary.
// Nox uses 53-bit precision control; explicit float32 conversion models FSTP.
//
//go:noinline
func chestOpenAdd64_4EDF00(a, b float64) float64 { return a + b }

//go:noinline
func chestOpenSub64_4EDF00(a, b float64) float64 { return a - b }

//go:noinline
func chestOpenMul64_4EDF00(a, b float64) float64 { return a * b }

//go:noinline
func chestOpenSpill32_4EDF00(value float64) float32 { return float32(value) }

func chestOpenNeg32_4EDF00(value float32) float32 {
	return math.Float32frombits(math.Float32bits(value) ^ uint32(1<<31))
}

func chestOpenLessOrUnordered4EDF00(a, b float64) bool {
	return math.IsNaN(a) || math.IsNaN(b) || a < b
}

// chestShapeExtent4EE2A0 preserves GAME.EXE 004EE2A0. Shape kind is read
// first. Circle radius is loaded once. A box compare reads W then H; W is
// selected only for an ordered strict-greater result, and the selected field
// is loaded live a second time before multiplication by 0.5. Every other kind
// returns exact positive zero without reading shape payload fields.
func chestShapeExtent4EE2A0[O any](obj O, hooks chestShapeExtentHooks4EE2A0[O]) float64 {
	switch hooks.loadShapeKind(obj) {
	case chestShapeCircle4EE2A0:
		return float64(hooks.loadCircleR(obj))
	case chestShapeBox4EE2A0:
		first := hooks.loadBoxExtentW(obj)
		second := hooks.loadBoxExtentH(obj)
		if !math.IsNaN(float64(first)) && !math.IsNaN(float64(second)) && first > second {
			return chestOpenMul64_4EDF00(float64(hooks.loadBoxExtentW(obj)), float64(chestShapeHalf4EE2A0))
		}
		return chestOpenMul64_4EDF00(float64(hooks.loadBoxExtentH(obj)), float64(chestShapeHalf4EE2A0))
	default:
		return float64(0)
	}
}

type chestOpenRay4EDF00 struct {
	Origin      types.Pointf
	Destination types.Pointf
}

type chestOpenHooks4EDF00[O comparable] struct {
	loadChestArg func() O
	loadUnitArg  func() O

	countInventory func(O, int32) int32
	loadSubClass   func(O) uint32
	loadPosX       func(O) float32
	loadPosY       func(O) float32
	normalize      func(*types.Pointf)
	shapeExtent    func(O) float64
	mapTrace       func(*chestOpenRay4EDF00, *types.Pointf, *image.Point, uint8) int32

	firstItem    func(O) O
	nextItem     func(O) O
	loadWeight   func(O) uint8
	loadClassLow func(O) uint8
	loadFlags    func(O) uint32
	storeFlags   func(O, uint32)
	refresh      func(O)
	drop         func(O, O, *types.Pointf) int32
}

func chestOpenDirection4EDF00[O comparable](
	chest, unit O,
	hooks chestOpenHooks4EDF00[O],
) types.Pointf {
	subClass := hooks.loadSubClass(chest)
	var direction types.Pointf
	switch {
	case subClass&chestOpenDirectionNorthWest4EDF00 != 0:
		direction.X = -1
		direction.Y = -1
	case subClass&chestOpenDirectionNorthEast4EDF00 != 0:
		direction.X = 1
		direction.Y = -1
	case subClass&chestOpenDirectionSouthEast4EDF00 != 0:
		direction.X = 1
		direction.Y = 1
	case subClass&chestOpenDirectionSouthWest4EDF00 != 0:
		direction.X = -1
		direction.Y = 1
	default:
		unitX := hooks.loadPosX(unit)
		chestX := hooks.loadPosX(chest)
		direction.X = chestOpenSpill32_4EDF00(chestOpenSub64_4EDF00(float64(unitX), float64(chestX)))
		unitY := hooks.loadPosY(unit)
		chestY := hooks.loadPosY(chest)
		direction.Y = chestOpenSpill32_4EDF00(chestOpenSub64_4EDF00(float64(unitY), float64(chestY)))
	}
	hooks.normalize(&direction)
	return direction
}

func chestOpenCandidates4EDF00[O comparable](
	chest, unit O,
	hooks chestOpenHooks4EDF00[O],
) [3]types.Pointf {
	direction := chestOpenDirection4EDF00(chest, unit, hooks)
	forward := hooks.shapeExtent(chest)
	forward = chestOpenAdd64_4EDF00(forward, float64(chestOpenShapeMargin4EDF00))
	forward = chestOpenAdd64_4EDF00(forward, float64(chestOpenForwardClearance4EDF00))

	var points [3]types.Pointf
	xScaled := chestOpenMul64_4EDF00(forward, float64(direction.X))
	points[0].X = chestOpenSpill32_4EDF00(
		chestOpenAdd64_4EDF00(xScaled, float64(hooks.loadPosX(chest))),
	)
	yScaled := chestOpenMul64_4EDF00(forward, float64(direction.Y))
	points[0].Y = chestOpenSpill32_4EDF00(
		chestOpenAdd64_4EDF00(yScaled, float64(hooks.loadPosY(chest))),
	)

	negativeY := chestOpenNeg32_4EDF00(direction.Y)
	positiveY := chestOpenNeg32_4EDF00(negativeY)
	negativeX := chestOpenNeg32_4EDF00(direction.X)
	points[1].X = chestOpenSpill32_4EDF00(chestOpenAdd64_4EDF00(
		chestOpenMul64_4EDF00(float64(negativeY), float64(chestOpenLateralOffset4EDF00)),
		float64(points[0].X),
	))
	points[1].Y = chestOpenSpill32_4EDF00(chestOpenAdd64_4EDF00(
		chestOpenMul64_4EDF00(float64(direction.X), float64(chestOpenLateralOffset4EDF00)),
		float64(points[0].Y),
	))
	points[2].X = chestOpenSpill32_4EDF00(chestOpenAdd64_4EDF00(
		chestOpenMul64_4EDF00(float64(positiveY), float64(chestOpenLateralOffset4EDF00)),
		float64(points[0].X),
	))
	points[2].Y = chestOpenSpill32_4EDF00(chestOpenAdd64_4EDF00(
		chestOpenMul64_4EDF00(float64(negativeX), float64(chestOpenLateralOffset4EDF00)),
		float64(points[0].Y),
	))

	distance0 := chestOpenSpill32_4EDF00(chestOpenSquaredDistance4EDF00(points[0], unit, hooks))
	distance1 := chestOpenSquaredDistance4EDF00(points[1], unit, hooks)
	distance2 := chestOpenSpill32_4EDF00(chestOpenSquaredDistance4EDF00(points[2], unit, hooks))
	chestOpenSortCandidates4EDF00(&points, distance0, distance1, distance2)
	return points
}

func chestOpenSquaredDistance4EDF00[O comparable](
	point types.Pointf,
	unit O,
	hooks chestOpenHooks4EDF00[O],
) float64 {
	unitX := hooks.loadPosX(unit)
	deltaX := chestOpenSub64_4EDF00(float64(point.X), float64(unitX))
	unitY := hooks.loadPosY(unit)
	deltaY := chestOpenSub64_4EDF00(float64(point.Y), float64(unitY))
	ySquared := chestOpenMul64_4EDF00(deltaY, deltaY)
	xSquared := chestOpenMul64_4EDF00(deltaX, deltaX)
	return chestOpenAdd64_4EDF00(ySquared, xSquared)
}

// chestOpenSortCandidates4EDF00 models the original three x87 comparisons.
// Distance 0 and 2 are binary32 spills; distance 1 remains binary64 until a
// swap forces it through memory. C0 alone controls every swap, so unordered
// comparisons take the same branch as ordered less-than comparisons.
func chestOpenSortCandidates4EDF00(
	points *[3]types.Pointf,
	distance0 float32,
	distance1 float64,
	distance2 float32,
) {
	if chestOpenLessOrUnordered4EDF00(float64(distance0), distance1) {
		points[0], points[1] = points[1], points[0]
		oldDistance0 := distance0
		distance0 = chestOpenSpill32_4EDF00(distance1)
		distance1 = float64(oldDistance0)
	}
	if chestOpenLessOrUnordered4EDF00(distance1, float64(distance2)) {
		points[1], points[2] = points[2], points[1]
		distance1 = float64(distance2)
	}
	if chestOpenLessOrUnordered4EDF00(float64(distance0), distance1) {
		points[0], points[1] = points[1], points[0]
	}
}

func chestOpenTraceCandidates4EDF00[O comparable](
	chest, unit O,
	points *[3]types.Pointf,
	hooks chestOpenHooks4EDF00[O],
) {
	chestY := hooks.loadPosY(chest)
	chestX := hooks.loadPosX(chest)
	ray := chestOpenRay4EDF00{Origin: types.Pointf{X: chestX, Y: chestY}}
	for _, index := range [...]int{2, 1, 0} {
		ray.Destination = points[index]
		if hooks.mapTrace(&ray, nil, nil, chestOpenTraceFlag4EDF00) != 0 {
			continue
		}
		unitX := hooks.loadPosX(unit)
		unitY := hooks.loadPosY(unit)
		points[index].X = unitX
		points[index].Y = unitY
	}
}

func chestOpenDropItem4EDF00[O comparable](
	chest, item O,
	point *types.Pointf,
	hooks chestOpenHooks4EDF00[O],
) bool {
	if hooks.loadWeight(item) == chestOpenInvalidWeight4EDF00 {
		return false
	}
	if hooks.loadClassLow(item)&chestOpenMonsterClass4EDF00 != 0 {
		return false
	}
	flags := hooks.loadFlags(item)
	hooks.storeFlags(item, flags|chestOpenDropFlag4EDF00)
	hooks.refresh(item)
	_ = hooks.drop(chest, item, point)
	return true
}

// chestOpen4EDF00 preserves GAME.EXE 004EDF00. Entry arguments and inventory
// count are gated in order. Three stack-local candidates are built, ranked,
// and traced in reverse order through one persistent ray. A failed trace uses
// the unit's live position; trace-side destination changes on success are not
// copied back. A cached count of exactly one selects the dedicated first-item
// path. Every other nonzero count uses successor-before-field iteration and
// advances the cyclic point index only after an eligible item is refreshed
// and dropped. Refresh and drop return values are ignored.
func chestOpen4EDF00[O comparable](hooks chestOpenHooks4EDF00[O]) {
	var zero O
	chest := hooks.loadChestArg()
	if chest == zero {
		return
	}
	unit := hooks.loadUnitArg()
	if unit == zero {
		return
	}

	count := hooks.countInventory(chest, 0)
	if count == 0 {
		return
	}

	points := chestOpenCandidates4EDF00(chest, unit, hooks)
	chestOpenTraceCandidates4EDF00(chest, unit, &points, hooks)

	if count == 1 {
		item := hooks.firstItem(chest)
		chestOpenDropItem4EDF00(chest, item, &points[0], hooks)
		return
	}

	item := hooks.firstItem(chest)
	pointIndex := 0
	for item != zero {
		next := hooks.nextItem(item)
		if chestOpenDropItem4EDF00(chest, item, &points[pointIndex], hooks) {
			pointIndex++
			if pointIndex >= len(points) {
				pointIndex = 0
			}
		}
		item = next
	}
}
