package opennox

import (
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/server"
)

func abilityRewardRuntime4FB9C0() server.AbilityRewardRuntime4FB9C0 {
	return server.AbilityRewardRuntime4FB9C0{
		AwardProtection: func(token uint32, ability, level int32) {
			legacy.Nox_xxx_playerAwardSpellProtectionCRC_56FCE0(token, int(ability), int(level))
		},
		SendLineMessage: legacy.Nox_xxx_netSendLineMessage_4D9EB0,
	}
}

// AbilityRewardServ4FB9C0 supplies root-owned legacy services to the restored
// native-width model of GAME.EXE 004FB9C0.
func (s *Server) AbilityRewardServ4FB9C0(
	unit *server.Object,
	ability, rewardArg int32,
) int32 {
	return s.Server.AbilityRewardServ4FB9C0(
		unit,
		ability,
		rewardArg,
		abilityRewardRuntime4FB9C0(),
	)
}

// UseAbilityReward53FAE0 keeps both registered-use Object arguments native
// width and routes its nested reward through the same restored service.
func (s *Server) UseAbilityReward53FAE0(owner, item *server.Object) int32 {
	return s.Server.UseAbilityReward53FAE0(
		owner,
		item,
		server.UseAbilityRewardRuntime53FAE0{
			AbilityReward: abilityRewardRuntime4FB9C0(),
			DelayedDelete: s.DelayedDelete,
		},
	)
}
