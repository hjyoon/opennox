package server

import (
	"image"

	"github.com/opennox/libs/types"
)

// ChestOpenRuntime4EDF00 supplies the two effects that remain owned by the
// legacy-facing runtime. Inventory, shape and position fields stay native;
// map tracing is owned by Server.
type ChestOpenRuntime4EDF00 struct {
	RefreshUnit func(*Object)
	Dispatch    func(*Object, *Object, *types.Pointf) int32
}

func chestOpenServerDeps4EDF00(
	s *Server,
	runtime ChestOpenRuntime4EDF00,
) chestOpenNativeDeps4EDF00 {
	return chestOpenNativeDeps4EDF00{
		mapTrace: func(ray *chestOpenRay4EDF00, outPoint *types.Pointf, outGrid *image.Point, flags uint8) int32 {
			if s.MapTraceRayAt(ray.Origin, ray.Destination, outPoint, outGrid, MapTraceFlags(flags)) {
				return 1
			}
			return 0
		},
		refresh: runtime.RefreshUnit,
		drop:    runtime.Dispatch,
	}
}

// ChestOpen4EDF00 drops every eligible chest inventory item through the
// original ranked candidate and live trace-fallback algorithm.
func (s *Server) ChestOpen4EDF00(
	chest, unit *Object,
	runtime ChestOpenRuntime4EDF00,
) {
	chestOpenNative4EDF00(chest, unit, chestOpenServerDeps4EDF00(s, runtime))
}
