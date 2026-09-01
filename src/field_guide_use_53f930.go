package opennox

import (
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/server"
)

func beastGuideAwardRuntime4FAE80() server.BeastGuideAwardRuntime4FAE80 {
	return server.BeastGuideAwardRuntime4FAE80{
		AwardProtection: func(token uint32, guide, level int32) {
			legacy.Nox_xxx_playerAwardSpellProtectionCRC_56FCE0(token, int(guide), int(level))
		},
		RelatedGuides:   legacy.PlayerGuideRelationsNative4FAE80,
		SendLineMessage: legacy.Nox_xxx_netSendLineMessage_4D9EB0,
	}
}

// AwardBeastGuide4FAE80 supplies root-owned legacy services to the restored
// native-width model of GAME.EXE 004FAE80.
func (s *Server) AwardBeastGuide4FAE80(
	unit *server.Object,
	guide, notify int32,
) int32 {
	return s.Server.AwardBeastGuide4FAE80(
		unit,
		guide,
		notify,
		beastGuideAwardRuntime4FAE80(),
	)
}

// UseFieldGuide53F930 keeps both registered-use Object arguments native width
// and routes its nested guide award through the same restored service.
func (s *Server) UseFieldGuide53F930(owner, item *server.Object) int32 {
	return s.Server.UseFieldGuide53F930(
		owner,
		item,
		server.UseFieldGuideRuntime53F930{
			BeastGuide:    beastGuideAwardRuntime4FAE80(),
			DelayedDelete: s.DelayedDelete,
		},
	)
}
