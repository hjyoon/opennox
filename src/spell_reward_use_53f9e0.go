package opennox

import "github.com/opennox/opennox/v1/server"

// UseSpellReward53F9E0 keeps both registered-use Object arguments native
// width and routes its nested grant through the restored spell service.
func (s *Server) UseSpellReward53F9E0(owner, item *server.Object) int32 {
	return s.Server.UseSpellReward53F9E0(
		owner,
		item,
		server.UseSpellRewardRuntime53F9E0{
			SpellGrant:    s.spellGrantRuntime4FB550(),
			DelayedDelete: s.DelayedDelete,
		},
	)
}
