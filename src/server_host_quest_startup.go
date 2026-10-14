package opennox

import (
	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/legacy"
)

// RunServerArgs bypasses startServer's menu path. Preserve its server-side
// Quest selection and fresh-host resets before accepting players or commands;
// do not initialize the graphical tubes, minimap, or character picker here.
func (s *Server) initDedicatedHostQuest(quest bool) {
	legacy.Set_nox_game_createOrJoin_815048(1)
	if quest {
		noxflags.UnsetGame(noxflags.GameNotQuest)
	} else {
		noxflags.SetGame(noxflags.GameNotQuest)
	}
	noxflags.UnsetGame(noxflags.GameModeCoop)
	legacy.Sub_461440(0)
	sub4D6F40(false)
	sub_4D6F90(0)
	s.nox_xxx_setQuest_4D6F60(bool2int(quest))
	legacy.Sub_4D6F80(0)
}
