package legacy

import (
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/server"
)

func crownDropRuntime4ED5E0() server.CrownDropRuntime4ED5E0 {
	return server.CrownDropRuntime4ED5E0{
		DefaultDrop: defaultDropCall4ED290,
		TeamContains: func(team *server.ObjectTeam, id server.TeamID) int32 {
			return int32(Nox_xxx_teamCompare2_419180(team, id))
		},
		BuffOff: Nox_xxx_spellBuffOff_4FF5B0,
	}
}

func crownDropCall4ED5E0(
	owner, crown *server.Object,
	point *types.Pointf,
) int32 {
	s := GetServer().S()
	return s.CrownDrop4ED5E0(owner, crown, point, crownDropRuntime4ED5E0())
}
