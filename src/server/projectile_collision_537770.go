package server

import (
	"image"
	"math"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
)

const (
	projectileTraceShortDistanceSquared537850 = float64(36)
	projectileTraceStepScaleBits537850        = uint32(0x3ce38e39)
	projectileTraceMapFlags537850             = MapTraceFlag1 | MapTraceFlag3
)

type projectileCanCollideHooks54E730[O comparable] struct {
	loadClassLow    func(O) uint8
	loadSubclassLow func(O) uint8
	loadFlags       func(O) object.Flags
	hasCollide      func(O) bool
	loadOwner       func(O) O
	sameTeam        func(O, O) bool
	isEnemy         func(O, O) bool
}

// projectileCanCollide54E730 preserves the candidate filter used by
// GAME.EXE 0054E730. In particular, MissileHit bypasses the airborne, team,
// and monster-owner exclusions, but not the earlier class, destruction,
// callback, or NoCollide checks.
func projectileCanCollide54E730[O comparable](
	first, second O,
	hooks projectileCanCollideHooks54E730[O],
) bool {
	if hooks.loadClassLow(second)&uint8(object.ClassMissile) != 0 {
		return false
	}
	firstFlags := hooks.loadFlags(first)
	if firstFlags.Has(object.FlagDestroyed) {
		return false
	}
	secondFlags := hooks.loadFlags(second)
	if secondFlags.Has(object.FlagDestroyed) ||
		!hooks.hasCollide(first) ||
		!hooks.hasCollide(second) ||
		secondFlags.Has(object.FlagNoCollide) {
		return false
	}
	if secondFlags.Has(object.FlagMissileHit) {
		return true
	}
	if firstFlags.HasAny(object.FlagBelow|object.FlagShort) && secondFlags.Has(object.FlagAirborne) {
		return false
	}
	if secondFlags.HasAny(object.FlagBelow|object.FlagShort) && firstFlags.Has(object.FlagAirborne) {
		return false
	}
	if (firstFlags.Has(object.FlagNoCollideOwner) || secondFlags.Has(object.FlagNoCollideOwner)) &&
		hooks.sameTeam(second, first) {
		return false
	}
	owner := hooks.loadOwner(first)
	var zero O
	if owner != zero &&
		hooks.loadClassLow(first)&uint8(object.ClassMissile) != 0 &&
		hooks.loadSubclassLow(first)&2 == 0 &&
		hooks.loadClassLow(owner)&uint8(object.ClassMonster) != 0 &&
		hooks.loadClassLow(second)&uint8(object.ClassMonster) != 0 {
		if !hooks.isEnemy(owner, second) {
			return false
		}
		// IsEnemy can run script-visible logic. GAME.EXE reloads the owner
		// field before the final team comparison rather than reusing owner.
		if hooks.sameTeam(second, hooks.loadOwner(first)) {
			return false
		}
	}
	return true
}

type projectileSampleObjectHooks54E810[O, U comparable] struct {
	eachAt            func(types.Pointf, func(O))
	loadClassLow      func(O) uint8
	loadSubclassLow   func(O) uint8
	loadDoorUpdate    func(O) U
	loadDoorDirection func(U) int32
	loadPosX          func(O) float32
	loadPosY          func(O) float32
	doorSize          func(byte) image.Point
	lineTrace         func(types.Rectf, types.Rectf) bool
	canCollide        func(O, O) bool
	intersects        func(O, *types.Pointf) bool
}

// projectileSampleObject54E810 preserves GAME.EXE 0054E810/0054E850. The
// map traversal deliberately keeps going after a match, so the last matching
// candidate wins. A Door hit rolls current back to previous immediately;
// later candidates in the same traversal observe that mutated sample.
func projectileSampleObject54E810[O, U comparable](
	source O,
	current, previous *types.Pointf,
	hooks projectileSampleObjectHooks54E810[O, U],
) O {
	var found O
	hooks.eachAt(*current, func(candidate O) {
		if int8(hooks.loadClassLow(candidate)) >= 0 {
			if hooks.canCollide(source, candidate) && hooks.intersects(candidate, current) {
				found = candidate
			}
			return
		}

		// The original caches the low subclass byte, then loads UpdateData,
		// and only then tests the cached Door subclass bit.
		subclass := hooks.loadSubclassLow(candidate)
		update := hooks.loadDoorUpdate(candidate)
		if subclass&4 != 0 {
			return
		}
		direction := hooks.loadDoorDirection(update)
		position := types.Ptf(hooks.loadPosX(candidate), hooks.loadPosY(candidate))
		size := hooks.doorSize(byte(direction))
		doorEnd := position.Add(types.Ptf(float32(size.X), float32(size.Y)))
		if hooks.lineTrace(
			types.Rectf{Min: *previous, Max: *current},
			types.Rectf{Min: position, Max: doorEnd},
		) {
			*current = *previous
			found = candidate
		}
	})
	return found
}

type projectileMapTraceResult537850 struct {
	clear bool
	point types.Pointf
	grid  image.Point
}

type projectileTraceHitHooks537850[O comparable] struct {
	loadPosX       func(O) float32
	loadPosY       func(O) float32
	loadNewPosX    func(O) float32
	loadNewPosY    func(O) float32
	storeNewPosX   func(O, float32)
	storeNewPosY   func(O, float32)
	loadObjectPosX func(O) float32
	loadObjectPosY func(O) float32
	sampleObject   func(O, *types.Pointf, *types.Pointf) O
	mapTrace       func(types.Pointf, types.Pointf, MapTraceFlags) projectileMapTraceResult537850
	storeTraceGrid func(image.Point)
	setTraceReady  func(uint32)
	wallNormal     func(image.Point, types.Rectf) (types.Pointf, bool)
}

// Keep the Win32 x87 53-bit arithmetic and binary32 spill boundaries
// explicit. Separate helpers also prevent target compilers from contracting
// multiply-plus-add sequences to FMA.

//go:noinline
func projectileAdd64_537850(a, b float64) float64 { return a + b }

//go:noinline
func projectileSub64_537850(a, b float64) float64 { return a - b }

//go:noinline
func projectileMul64_537850(a, b float64) float64 { return a * b }

//go:noinline
func projectileDiv64_537850(a, b float64) float64 { return a / b }

//go:noinline
func projectileSqrt64_537850(value float64) float64 { return math.Sqrt(value) }

//go:noinline
func projectileSpill32_537850(value float64) float32 { return float32(value) }

// projectileDoubleToIntLow32_419B10 models nox_double2int's FISTP qword
// under the default nearest-even x87 rounding mode, followed by EAX's low
// dword return. The x87 indefinite integer has a zero low dword.
func projectileDoubleToIntLow32_419B10(value float64) int32 {
	if math.IsNaN(value) || value >= 0x1p63 || value < -0x1p63 {
		return 0
	}
	return int32(int64(math.RoundToEven(value)))
}

func projectileObjectNormal537850[O comparable](
	previous types.Pointf,
	hit O,
	hooks projectileTraceHitHooks537850[O],
) types.Pointf {
	return types.Ptf(
		projectileSpill32_537850(projectileSub64_537850(
			float64(previous.X), float64(hooks.loadObjectPosX(hit)),
		)),
		projectileSpill32_537850(projectileSub64_537850(
			float64(previous.Y), float64(hooks.loadObjectPosY(hit)),
		)),
	)
}

// projectileTraceHit537850 preserves GAME.EXE 00537850. Object sampling is
// performed before the live wall trace. A blocked wall always records the
// grid, marks the trace ready, and resets NewPos to the live Pos even when the
// wall-normal helper fails. Object distance must be strictly less than wall
// distance; a tie selects the wall, while an unordered x87 comparison selects
// the object.
func projectileTraceHit537850[O comparable](
	source O,
	hooks projectileTraceHitHooks537850[O],
) (O, types.Pointf, bool) {
	deltaX := projectileSpill32_537850(projectileSub64_537850(
		float64(hooks.loadNewPosX(source)), float64(hooks.loadPosX(source)),
	))
	deltaY := projectileSpill32_537850(projectileSub64_537850(
		float64(hooks.loadNewPosY(source)), float64(hooks.loadPosY(source)),
	))
	distanceSquared := projectileAdd64_537850(
		projectileMul64_537850(float64(deltaY), float64(deltaY)),
		projectileMul64_537850(float64(deltaX), float64(deltaX)),
	)

	var hit O
	var objectNormal types.Pointf
	if !(distanceSquared > projectileTraceShortDistanceSquared537850) {
		// Preserve the non-spatial field-load order from GAME.EXE: NewPosY,
		// PosX, NewPosX, then PosY.
		newY := hooks.loadNewPosY(source)
		posX := hooks.loadPosX(source)
		newX := hooks.loadNewPosX(source)
		current := types.Ptf(newX, newY)
		previous := types.Ptf(posX, hooks.loadPosY(source))
		hit = hooks.sampleObject(source, &current, &previous)
		var zero O
		if hit != zero {
			objectNormal = projectileObjectNormal537850(previous, hit, hooks)
		}
	} else {
		scale := math.Float32frombits(projectileTraceStepScaleBits537850)
		steps := projectileDoubleToIntLow32_419B10(projectileSqrt64_537850(
			projectileMul64_537850(distanceSquared, float64(scale)),
		)) + 1
		previous := types.Ptf(hooks.loadPosX(source), hooks.loadPosY(source))
		current := previous
		stepX := projectileSpill32_537850(projectileDiv64_537850(float64(deltaX), float64(steps)))
		stepY := projectileSpill32_537850(projectileDiv64_537850(float64(deltaY), float64(steps)))
		for index := int32(0); index < steps; index++ {
			current.X = projectileSpill32_537850(projectileAdd64_537850(float64(current.X), float64(stepX)))
			current.Y = projectileSpill32_537850(projectileAdd64_537850(float64(current.Y), float64(stepY)))
			hit = hooks.sampleObject(source, &current, &previous)
			var zero O
			if hit != zero {
				objectNormal = projectileObjectNormal537850(previous, hit, hooks)
				break
			}
			previous = current
		}
	}

	from := types.Ptf(hooks.loadPosX(source), hooks.loadPosY(source))
	to := types.Ptf(hooks.loadNewPosX(source), hooks.loadNewPosY(source))
	trace := hooks.mapTrace(from, to, projectileTraceMapFlags537850)
	var wallNormal types.Pointf
	wallValid := false
	if !trace.clear {
		hooks.storeTraceGrid(trace.grid)
		hooks.setTraceReady(1)
		wallNormal, wallValid = hooks.wallNormal(
			trace.grid,
			types.Rectf{Min: from, Max: trace.point},
		)
		// GAME.EXE loads PosX before PosY, then performs both stores.
		positionX := hooks.loadPosX(source)
		positionY := hooks.loadPosY(source)
		hooks.storeNewPosX(source, positionX)
		hooks.storeNewPosY(source, positionY)
	}

	var zero O
	if hit != zero {
		if !wallValid {
			return hit, objectNormal, true
		}
		objectDX := projectileSub64_537850(
			float64(hooks.loadPosX(source)), float64(hooks.loadObjectPosX(hit)),
		)
		objectDY := projectileSub64_537850(
			float64(hooks.loadPosY(source)), float64(hooks.loadObjectPosY(hit)),
		)
		objectDistanceSquared := projectileAdd64_537850(
			projectileMul64_537850(objectDY, objectDY),
			projectileMul64_537850(objectDX, objectDX),
		)
		wallDX := projectileSub64_537850(float64(hooks.loadPosX(source)), float64(trace.point.X))
		wallDY := projectileSub64_537850(float64(hooks.loadPosY(source)), float64(trace.point.Y))
		wallDistanceSquared := projectileAdd64_537850(
			projectileMul64_537850(wallDY, wallDY),
			projectileMul64_537850(wallDX, wallDX),
		)
		// FCOMPP followed by the C0 condition bit selects the object both
		// when it is strictly closer and when either operand is unordered.
		if objectDistanceSquared < wallDistanceSquared ||
			math.IsNaN(objectDistanceSquared) || math.IsNaN(wallDistanceSquared) {
			return hit, objectNormal, true
		}
		return zero, wallNormal, true
	}
	if wallValid {
		return zero, wallNormal, true
	}
	return zero, types.Pointf{}, false
}

type projectileCollisionDispatchHooks537770[O, C comparable] struct {
	loadSmallCache   func() uint32
	storeSmallCache  func(uint32)
	loadMediumCache  func() uint32
	storeMediumCache func(uint32)
	loadLargeCache   func() uint32
	storeLargeCache  func(uint32)
	lookupType       func(string) uint32
	loadFlagsLow     func(O) uint8
	traceHit         func(O) (O, types.Pointf, bool)
	loadTypeIndex    func(O) uint16
	loadCollide      func(O) C
	callCollide      func(C, O, O, *types.Pointf)
	setTraceReady    func(uint32)
}

// projectileCollisionDispatch537770 preserves GAME.EXE 00537770, including
// its asymmetric SmallFist cache gate and live callback loads. A fist target
// suppresses both callbacks without clearing a trace-ready value set by the
// helper. Otherwise trace-ready is cleared after the source callback, and the
// target callback is loaded only after that callback returns.
func projectileCollisionDispatch537770[O, C comparable](
	source O,
	hooks projectileCollisionDispatchHooks537770[O, C],
) {
	if hooks.loadSmallCache() == 0 {
		hooks.storeSmallCache(hooks.lookupType("SmallFist"))
		hooks.storeMediumCache(hooks.lookupType("MediumFist"))
		hooks.storeLargeCache(hooks.lookupType("LargeFist"))
	}
	if hooks.loadFlagsLow(source)&uint8(object.FlagDestroyed|object.FlagNoCollide) != 0 {
		return
	}

	hooks.setTraceReady(0)
	hit, normal, ok := hooks.traceHit(source)
	if !ok {
		return
	}
	var zero O
	if hit != zero {
		small := hooks.loadSmallCache()
		typeIndex := uint32(hooks.loadTypeIndex(hit))
		if typeIndex == small ||
			typeIndex == hooks.loadMediumCache() ||
			typeIndex == hooks.loadLargeCache() {
			return
		}
	}

	sourceCollide := hooks.loadCollide(source)
	hooks.callCollide(sourceCollide, source, hit, &normal)
	hooks.setTraceReady(0)
	if hit == zero {
		return
	}
	normal.X = -normal.X
	normal.Y = -normal.Y
	targetCollide := hooks.loadCollide(hit)
	hooks.callCollide(targetCollide, hit, source, &normal)
}
