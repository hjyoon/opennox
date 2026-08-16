package server

import (
	"image"
	"math"

	"github.com/opennox/libs/types"
)

func randomReachablePointServerDeps4ED970(s *Server) randomReachablePointNativeDeps4ED970 {
	return randomReachablePointNativeDeps4ED970{
		randomFloat: func(min, max float32, _ string, _ int32) float64 {
			return logicRandomFloat416030(s.Rand.Logic, min, max)
		},
		cosine: math.Cos,
		sine:   math.Sin,
		mapTrace: func(ray *randomReachablePointRay4ED970, outPoint *types.Pointf, outGrid *image.Point, flags uint8) int32 {
			if s.MapTraceRayAt(ray.Origin, ray.Destination, outPoint, outGrid, MapTraceFlags(flags)) {
				return 1
			}
			return 0
		},
	}
}

// RandomReachablePointAroundInto4ED970 binds the restored helper to native
// pointer-width points. The center remains live across trace callbacks, output
// may alias it, and the exact output pointer is returned.
func (s *Server) RandomReachablePointAroundInto4ED970(
	radius float32,
	center, output *types.Pointf,
) *types.Pointf {
	return randomReachablePointNative4ED970(
		radius,
		center,
		output,
		randomReachablePointServerDeps4ED970(s),
	)
}

// RandomReachablePointAround preserves the existing value-oriented Go API on
// top of the pointer-accurate 004ED970 implementation.
func (s *Server) RandomReachablePointAround(radius float32, center types.Pointf) types.Pointf {
	var output types.Pointf
	s.RandomReachablePointAroundInto4ED970(radius, &center, &output)
	return output
}
