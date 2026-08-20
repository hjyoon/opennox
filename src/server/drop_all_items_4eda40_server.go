package server

import (
	"image"

	"github.com/opennox/libs/types"
)

// DropAllItemsRuntime4EDA40 supplies the restored 004ED790 dispatcher while
// RNG and map tracing are owned by Server. Object and point pointers remain
// native-width through the whole call.
type DropAllItemsRuntime4EDA40 struct {
	Dispatch func(*Object, *Object, *types.Pointf) int32
}

func dropAllItemsServerDeps4EDA40(
	s *Server,
	runtime DropAllItemsRuntime4EDA40,
) dropAllItemsNativeDeps4EDA40 {
	return dropAllItemsNativeDeps4EDA40{
		randomFloat: func(min, max float32, _ string, _ int32) float64 {
			return logicRandomFloat416030(s.Rand.Logic, min, max)
		},
		mapTrace: func(ray *dropAllItemsRay4EDA40, outPoint *types.Pointf, outGrid *image.Point, flags uint8) int32 {
			if s.MapTraceRayAt(ray.Origin, ray.Destination, outPoint, outGrid, MapTraceFlags(flags)) {
				return 1
			}
			return 0
		},
		drop: runtime.Dispatch,
	}
}

// DropAllItems4EDA40 drops every eligible inventory item using the original
// radius-derived spacing, randomized square spiral, and live-position
// fallback. Its int32 result preserves the otherwise caller-ignored EAX value.
func (s *Server) DropAllItems4EDA40(owner *Object, runtime DropAllItemsRuntime4EDA40) int32 {
	return dropAllItemsNative4EDA40(owner, dropAllItemsServerDeps4EDA40(s, runtime))
}
