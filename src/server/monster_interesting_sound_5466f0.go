package server

import (
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/common/unit/ai"
)

// MonsterInterestingSound5466F0 retains ordinary nil/class admission and
// supplies real clocks, RNG, trace and stack services to the raw entry.
func (s *Server) MonsterInterestingSound5466F0(unit *Object, tileAt func(types.Pointf) int, pathReachable func(*Object, *types.Pointf) bool) int {
	if unit == nil || unit.UpdateData == nil {
		return 0
	}
	_ = unit.UpdateDataMonster()
	return unit.MonsterInterestingSoundNative5466F0(MonsterReactiveRuntime545DA0{
		Frame:    s.Frame,
		TickRate: s.TickRate,
		Random:   s.Rand.Logic.IntClamp,
		Push: func(unit *Object, action ai.ActionType) *AIStackItem {
			return unit.MonsterPushAction(action)
		},
	}, func(point *types.Pointf) int { return tileAt(*point) }, pathReachable,
		func(from, to types.Pointf) bool { return s.MapTraceRayAt(from, to, nil, nil, 9) })
}
