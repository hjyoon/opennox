package server

import "github.com/opennox/libs/types"

type mapPushUnitsAroundHooks52E040[O comparable] struct {
	eachInRect  func(types.Rectf, func(O) bool)
	isMovable   func(O) bool
	position    func(O) types.Pointf
	traceRay    func(types.Pointf, types.Pointf) bool
	mass        func(O) float32
	isMissile   func(O) bool
	loadForce   func(O) types.Pointf
	storeForce  func(O, types.Pointf, types.Pointf)
	beforeForce func(O, float32)
	wake        func(O)
}

// MapPushUnitArguments52E0E0 views the original callback record through live
// native-width providers. The radius and force are already prepared by the
// root; direct C callback callers retain mutable record fields.
type MapPushUnitArguments52E0E0 struct {
	Origin      func() types.Pointf
	Radius      func() float32
	Inner       func() float32
	ScaledForce func() float32
}

func mapPushDependencies52E0E0[O comparable](h mapPushUnitsAroundHooks52E040[O]) bool {
	return h.isMovable != nil && h.position != nil && h.traceRay != nil && h.mass != nil && h.isMissile != nil && h.loadForce != nil && h.storeForce != nil && h.wake != nil
}

func mapPushUnit52E0E0[O comparable](unit O, args MapPushUnitArguments52E0E0, h mapPushUnitsAroundHooks52E040[O]) {
	if !h.isMovable(unit) {
		return
	}
	if !h.traceRay(args.Origin(), h.position(unit)) {
		return
	}
	// Vision may change both the origin record and the candidate position.
	origin, position := args.Origin(), h.position(unit)
	dx := experienceSpill32TowardZero(objectForceSub53_52DF80(float64(position.X), objectForceMemory32_52DF80(origin.X)))
	dyWide := objectForceSub53_52DF80(float64(position.Y), objectForceMemory32_52DF80(origin.Y))
	dy := experienceSpill32TowardZero(dyWide)
	squareY := objectForceMul53_52DF80(dyWide, float64(dy))
	squareX := objectForceMul53_52DF80(float64(dx), float64(dx))
	distanceWide := objectForceAdd53_52DF80(monsterMoveForceSqrtChop53_50D50C(objectForceAdd53_52DF80(squareY, squareX)), float64(float32(0.1)))
	distance := experienceSpill32TowardZero(distanceWide)
	// FST retains the wide distance for the outer C0|C3 comparison.
	if distanceWide > float64(args.Radius()) {
		return
	}
	var scaled float32
	if distance > args.Inner() {
		numerator := objectForceSub53_52DF80(float64(distance), objectForceMemory32_52DF80(args.Inner()))
		denominator := objectForceSub53_52DF80(float64(args.Radius()), objectForceMemory32_52DF80(args.Inner()))
		fraction := objectForceDiv53_52DF80(numerator, denominator)
		ratio := objectForceSub53_52DF80(1, fraction)
		scaled = experienceSpill32TowardZero(objectForceMul53_52DF80(ratio, objectForceMemory32_52DF80(args.ScaledForce())))
	} else {
		scaled = args.ScaledForce()
	}
	factor := experienceSpill32TowardZero(objectForceDiv53_52DF80(objectForceMemory32_52DF80(scaled), float64(h.mass(unit))))
	if h.beforeForce != nil {
		h.beforeForce(unit, distance)
	}
	// The callback can change class and accumulated force, but the geometry,
	// mass quotient and the class decision survive the subsequent two stores.
	missile := h.isMissile(unit)
	accumulated := h.loadForce(unit)
	x := objectForceDiv53_52DF80(objectForceMul53_52DF80(float64(factor), float64(dx)), float64(distance))
	y := objectForceDiv53_52DF80(objectForceMul53_52DF80(float64(factor), float64(dy)), float64(distance))
	result := types.Ptf(experienceSpill32TowardZero(objectForceAdd53_52DF80(x, objectForceMemory32_52DF80(accumulated.X))), experienceSpill32TowardZero(objectForceAdd53_52DF80(y, objectForceMemory32_52DF80(accumulated.Y))))
	h.storeForce(unit, origin, result)
	if !missile {
		h.wake(unit)
	}
}

// mapPushUnitsAround52E040 preserves the original root and calls its native
// callback directly, retaining live origin aliases across all candidates.
func mapPushUnitsAround52E040[O comparable](origin *types.Pointf, outer, inner, force float32, h mapPushUnitsAroundHooks52E040[O]) bool {
	if origin == nil || h.eachInRect == nil || !mapPushDependencies52E0E0(h) {
		return false
	}
	rect := types.Rectf{
		Min: types.Ptf(experienceSpill32TowardZero(objectForceSub53_52DF80(float64(origin.X), objectForceMemory32_52DF80(outer))), experienceSpill32TowardZero(objectForceSub53_52DF80(float64(origin.Y), objectForceMemory32_52DF80(outer)))),
		Max: types.Ptf(experienceSpill32TowardZero(objectForceAdd53_52DF80(float64(outer), objectForceMemory32_52DF80(origin.X))), experienceSpill32TowardZero(objectForceAdd53_52DF80(float64(outer), objectForceMemory32_52DF80(origin.Y)))),
	}
	scaled := experienceSpill32TowardZero(objectForceMul53_52DF80(float64(force), 10))
	radius := outer
	if !(outer >= inner) {
		radius = inner
	}
	args := MapPushUnitArguments52E0E0{Origin: func() types.Pointf { return *origin }, Radius: func() float32 { return radius }, Inner: func() float32 { return inner }, ScaledForce: func() float32 { return scaled }}
	h.eachInRect(rect, func(unit O) bool { mapPushUnit52E0E0(unit, args, h); return true })
	return true
}

// MapPushUnitsAroundRuntime52E040 supplies the optional original callback and
// the final activation operation. Force arithmetic and stores remain native.
type MapPushUnitsAroundRuntime52E040 struct {
	BeforeForce func(*Object, float32)
	Wake        func(*Object)
}

func (s *Server) mapPushHooks52E0E0(runtime MapPushUnitsAroundRuntime52E040) mapPushUnitsAroundHooks52E040[*Object] {
	return mapPushUnitsAroundHooks52E040[*Object]{
		eachInRect: s.Map.EachObjAndMissileInRect, isMovable: (*Object).IsMovable, position: (*Object).Pos,
		traceRay: func(from, to types.Pointf) bool { return s.MapTraceRay(from, to, 0) },
		mass:     func(o *Object) float32 { return o.Mass }, isMissile: func(o *Object) bool { return uint32(o.ObjClass)&1 != 0 },
		loadForce: func(o *Object) types.Pointf { return o.ForceVec }, storeForce: func(o *Object, _ types.Pointf, v types.Pointf) { o.ForceVec = v }, beforeForce: runtime.BeforeForce, wake: runtime.Wake,
	}
}

// MapPushUnitsAround52E040 keeps the value API for callers with a local point.
func (s *Server) MapPushUnitsAround52E040(origin types.Pointf, outer, inner, force float32, runtime MapPushUnitsAroundRuntime52E040) bool {
	return s.MapPushUnitsAroundPtr52E040(&origin, outer, inner, force, runtime)
}

// MapPushUnitsAroundPtr52E040 retains the caller's native point reference.
func (s *Server) MapPushUnitsAroundPtr52E040(origin *types.Pointf, outer, inner, force float32, runtime MapPushUnitsAroundRuntime52E040) bool {
	if s == nil {
		return false
	}
	return mapPushUnitsAround52E040(origin, outer, inner, force, s.mapPushHooks52E0E0(runtime))
}

func (s *Server) MapPushUnit52E0E0(unit *Object, args MapPushUnitArguments52E0E0, runtime MapPushUnitsAroundRuntime52E040) bool {
	if s == nil || unit == nil || args.Origin == nil || args.Radius == nil || args.Inner == nil || args.ScaledForce == nil {
		return false
	}
	h := s.mapPushHooks52E0E0(runtime)
	if !mapPushDependencies52E0E0(h) {
		return false
	}
	mapPushUnit52E0E0(unit, args, h)
	return true
}
