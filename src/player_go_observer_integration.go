package opennox

import (
	"github.com/opennox/libs/types"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/ntype"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/server"
)

func (s *Server) playerGoObserver_4E6860(pl *server.Player, notify, keep int) int {
	return playerGoObserver_4E6860(pl, notify, keep, playerGoObserverHooks_4E6860{
		abilityActive: func(unit *server.Object) int {
			return bool2int(s.Abils.IsAnyActive(unit))
		},
		isMonsterBot: func(unit *server.Object) bool {
			return unit.Update == legacy.Get_nox_xxx_updatePlayerMonsterBot_4FAB20()
		},
		gameFlag: noxflags.HasGame,
		ensureCrownID: func() {
			_ = s.Types.CrownID()
		},
		ensureGameBallID: func() {
			_ = s.Types.GameBallID()
		},
		crownID:    func() uint32 { return uint32(s.Types.CrownIDCached()) },
		gameBallID: func() uint32 { return uint32(s.Types.GameBallIDCached()) },
		dropCrown: func(owner, item *server.Object, pos *types.Pointf) {
			legacy.Nox_xxx_drop_4ED790(owner, item, *pos)
		},
		clearOwner: func(item *server.Object) {
			item.SetOwner(nil)
		},
		gameBallDropped: func() {
			s.setGameBallStatus4E8290(1, 0)
		},
		dropFlag: func(owner, item *server.Object, pos *types.Pointf) {
			asObjectS(owner).forceDropAt(item, *pos)
		},
		getPossess: nox_xxx_playerGetPossess_4DDF30,
		clearObserve: func(unit *server.Object) {
			asObjectS(unit).observeClear()
		},
		needTimestamp: func(player *server.Player) {
			legacy.Nox_xxx_netNeedTimestampStatus_4174F0(player, 1)
		},
		anyPlayers: legacy.Nox_xxx_gamePlayIsAnyPlayers_40A8A0,
		resetState: func() {
			legacy.Sub_40A1F0(0)
		},
		forceLessons: func() {
			legacy.Nox_xxx_playerForceSendLessons_416E50(1)
		},
		resetTeams: func() {
			s.TeamsResetYyy()
		},
		finishReset: legacy.Sub_40A970,
		inform: func(ind ntype.PlayerInd, code byte, value uint32) {
			s.NetInformTextMsg(ind, code, int(value))
		},
		applyInvisible: func(unit *server.Object) {
			asObjectS(unit).ApplyEnchant(server.ENCHANT_INVISIBLE, 0, 5)
		},
		unlockCamera:  playerCameraUnlock_4E6040,
		leaveObserver: s.PlayerLeaveMonsterObserver,
		removeSpawned: legacy.Nox_xxx_playerRemoveSpawnedStuff_4E5AD0,
		setObserverUpdate: func(unit *server.Object) {
			_ = nox_xxx_updatePlayerObserver_4E62F0
			unit.Update = legacy.Get_nox_xxx_updatePlayerObserver_4E62F0()
		},
		resetCamping: s.Sub_4D7E50,
	})
}
