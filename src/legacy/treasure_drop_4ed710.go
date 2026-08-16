package legacy

import (
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/server"
)

func treasureDropRuntime4ED710(s *server.Server) server.TreasureDropRuntime4ED710 {
	return server.TreasureDropRuntime4ED710{
		DefaultDrop: defaultDropCall4ED290,
		TreasureMax: Nox_xxx_scavengerTreasureMax_4D1600,
		Report:      s.ScavengerHuntReport4D8CD0,
	}
}

func treasureDropCall4ED710(
	owner, treasure *server.Object,
	point *types.Pointf,
) int32 {
	s := GetServer().S()
	return s.TreasureDrop4ED710(owner, treasure, point, treasureDropRuntime4ED710(s))
}
