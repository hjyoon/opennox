package server

import (
	"image"

	"github.com/opennox/opennox/v1/internal/noxgroupwalk"
)

// ForEachGroup502670 binds GAME.EXE 00502670 to native-width group, item,
// object, waypoint, and wall pointers. Group and extent IDs remain ABI32 words.
// Unlike the modern Each*Recursive helpers, the original callback cannot stop
// traversal, and a wall group also tries its items as child-group IDs.
func (s *Server) ForEachGroup502670(g *MapGroup, expected MapGroupKind, visit func(any)) {
	noxgroupwalk.Walk(g, int32(expected), noxgroupwalk.Hooks[*MapGroup, *MapGroupItem]{
		Kind:         func(g *MapGroup) uint8 { return g.typ },
		First:        func(g *MapGroup) *MapGroupItem { return g.List },
		Next:         func(it *MapGroupItem) *MapGroupItem { return it.Next8 },
		IDs:          func(it *MapGroupItem) (uint32, uint32) { return it.Raw0, it.Raw4 },
		ResolveGroup: func(id uint32) *MapGroup { return s.MapGroups.GroupByInd(int(int32(id))) },
		Visit: func(kind uint8, first, second uint32) {
			switch MapGroupKind(kind) {
			case MapGroupObjects:
				if obj := s.Objs.GetObjectByInd(int(int32(first))); obj != nil {
					visit(obj)
				}
			case MapGroupWaypoints:
				if wp := s.WPs.ByInd(int(int32(first))); wp != nil {
					visit(wp)
				}
			case MapGroupWalls:
				if w := s.Walls.GetWallAtGrid(image.Pt(int(int32(first)), int(int32(second)))); w != nil {
					visit(w)
				}
			}
		},
	})
}
