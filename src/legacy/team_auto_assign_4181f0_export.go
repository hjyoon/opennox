package legacy

/*
#include "GAME1.h"
*/
import "C"

import (
	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/server"
)

// TeamAutoAssign4181F0 binds the native-width server implementation to the
// remaining client, network, and team reset operations in the legacy layer.
func TeamAutoAssign4181F0(resetTeams bool) {
	srv := GetServer().S()
	srv.TeamAutoAssign4181F0(resetTeams, server.TeamAutoAssignRuntime4181F0{
		ResetTeam: Sub_418D80,
		ClientNetCode: func() uint32 {
			return uint32(ClientPlayerNetCode())
		},
		NoRendering: func() bool {
			return noxflags.HasEngine(noxflags.EngineNoRendering)
		},
		PreferConfiguredTeam: func() bool {
			return Sub_40A740() != 0
		},
		Attach: func(teamID server.TeamID, value *server.ObjectTeam, active int32, netCode uint32, flags int32) {
			Nox_xxx_createAtImpl_4191D0(teamID, value, int(active), int(netCode), int(flags))
		},
	})
}

//export sub_4181F0
func sub_4181F0(resetTeams C.int) {
	TeamAutoAssign4181F0(resetTeams != 0)
}

func Sub_4181F0(resetTeams int) {
	TeamAutoAssign4181F0(resetTeams != 0)
}
