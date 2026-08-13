package opennox

import (
	"encoding/binary"

	"github.com/opennox/libs/noxnet/netmsg"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/ntype"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/server"
)

func (s *Server) playerLeaveObserver_4E6AA0(pl *server.Player) {
	playerLeaveObserver_4E6AA0(pl, playerLeaveObserverHooks_4E6AA0{
		isMonsterBot: func(unit *server.Object) bool {
			return unit.Update == legacy.Get_nox_xxx_updatePlayerMonsterBot_4FAB20()
		},
		unsetStatus: func(player *server.Player, status uint32) {
			legacy.Nox_xxx_playerUnsetStatus_417530(player, int(status))
		},
		disableEnchant: legacy.Nox_xxx_spellBuffOff_4FF5B0,
		setPlayerUpdate: func(unit *server.Object) {
			_ = nox_xxx_updatePlayer_4F8100
			unit.Update = legacy.Get_nox_xxx_updatePlayer_4F8100()
		},
		markUpdate: func(unit *server.Object) {
			unit.Nox_xxx_monsterMarkUpdate_4E8020()
		},
		gameFlag:     noxflags.HasGame,
		gameplayFlag: noxflags.HasGamePlay,
		teamFlag: func(unit *server.Object) *server.Object {
			return s.Teams.TeamFlag(s.Teams.ByID(unit.TeamVal.ID))
		},
		pickupTeamFlag: func(unit, flag *server.Object) {
			flag.CallPickup(unit, 1, 1)
		},
		questListed:     legacy.Sub_509D80,
		rememberQuest:   legacy.Sub_509C30,
		firstPlayerUnit: s.Players.FirstUnit,
		nextPlayerUnit:  s.Players.NextUnit,
		reportEnchant:   s.netReportEnchant_4D8F90,
	})
}

func (s *Server) netReportEnchant_4D8F90(ind ntype.PlayerInd, unit *server.Object) {
	var buf [7]byte
	buf[0] = byte(netmsg.MSG_REPORT_ENCHANTMENT)
	binary.LittleEndian.PutUint16(buf[1:], uint16(s.GetUnitNetCode(unit)))
	binary.LittleEndian.PutUint32(buf[3:], unit.Buffs)
	s.NetSendPacketXxx1(int(ind), buf[:], nil, 1)
}
