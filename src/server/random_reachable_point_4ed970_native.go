package server

import (
	"image"

	"github.com/opennox/libs/types"
)

type randomReachablePointNativeDeps4ED970 struct {
	randomFloat func(float32, float32, string, int32) float64
	cosine      func(float64) float64
	sine        func(float64) float64
	mapTrace    func(*randomReachablePointRay4ED970, *types.Pointf, *image.Point, uint8) int32
}

func randomReachablePointNative4ED970(
	radius float32,
	center, output *types.Pointf,
	deps randomReachablePointNativeDeps4ED970,
) *types.Pointf {
	return randomReachablePoint4ED970(randomReachablePointHooks4ED970[
		*types.Pointf,
		*types.Pointf,
	]{
		loadRadiusArg: func() float32 {
			return radius
		},
		loadCenterArg: func() *types.Pointf {
			return center
		},
		loadCenterX: func(point *types.Pointf) float32 {
			return point.X
		},
		loadCenterY: func(point *types.Pointf) float32 {
			return point.Y
		},
		randomFloat: deps.randomFloat,
		cosine:      deps.cosine,
		sine:        deps.sine,
		mapTrace:    deps.mapTrace,
		loadOutputArg: func() *types.Pointf {
			return output
		},
		storeOutputX: func(point *types.Pointf, value float32) {
			point.X = value
		},
		storeOutputY: func(point *types.Pointf, value float32) {
			point.Y = value
		},
	})
}
