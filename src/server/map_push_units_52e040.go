package server

import (
	"math"

	"github.com/opennox/libs/types"
)

type mapPushUnitsAroundHooks52E040[O comparable] struct {
	eachInRect func(types.Rectf, func(O) bool)
	isMovable  func(O) bool
	position   func(O) types.Pointf
	traceRay   func(types.Pointf, types.Pointf) bool
	applyForce func(O, types.Pointf, float64)
}

// mapPushUnitsAround52E040 restores GAME.EXE 0052E040-0052E0E0. The caller
// supplies native-width object accessors so the PE32 callback data record is
// never materialized on a wide host.
func mapPushUnitsAround52E040[O comparable](
	origin types.Pointf,
	outerRadius, innerRadius, force float32,
	hooks mapPushUnitsAroundHooks52E040[O],
) bool {
	if hooks.eachInRect == nil || hooks.isMovable == nil || hooks.position == nil ||
		hooks.traceRay == nil || hooks.applyForce == nil {
		return false
	}
	radius := outerRadius
	if radius < innerRadius {
		radius = innerRadius
	}
	rect := types.Rectf{
		Min: origin.Sub(types.Ptf(outerRadius, outerRadius)),
		Max: origin.Add(types.Ptf(outerRadius, outerRadius)),
	}
	hooks.eachInRect(rect, func(candidate O) bool {
		if !hooks.isMovable(candidate) {
			return true
		}
		position := hooks.position(candidate)
		if !hooks.traceRay(origin, position) {
			return true
		}
		delta := position.Sub(origin)
		distance := float32(math.Sqrt(float64(delta.X*delta.X+delta.Y*delta.Y)) + 0.1)
		if distance > radius {
			return true
		}
		push := force
		if distance > innerRadius {
			push *= 1 - (distance-innerRadius)/(radius-innerRadius)
		}
		hooks.applyForce(candidate, origin, float64(push))
		return true
	})
	return true
}

// MapPushUnitsAroundRuntime52E040 supplies the final force application, which
// remains owned by the outer server because it also wakes collision/update
// callbacks.
type MapPushUnitsAroundRuntime52E040 struct {
	ApplyForce func(*Object, types.Pointf, float64)
}

// MapPushUnitsAround52E040 binds the original radial push to the native object
// map and ray tracer. It includes missiles, matching getUnitsInRectAdv.
func (s *Server) MapPushUnitsAround52E040(
	origin types.Pointf,
	outerRadius, innerRadius, force float32,
	runtime MapPushUnitsAroundRuntime52E040,
) bool {
	if s == nil {
		return false
	}
	return mapPushUnitsAround52E040(origin, outerRadius, innerRadius, force, mapPushUnitsAroundHooks52E040[*Object]{
		eachInRect: s.Map.EachObjAndMissileInRect,
		isMovable:  (*Object).IsMovable,
		position:   (*Object).Pos,
		traceRay: func(from, to types.Pointf) bool {
			return s.MapTraceRay(from, to, 0)
		},
		applyForce: runtime.ApplyForce,
	})
}
