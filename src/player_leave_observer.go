package opennox

import (
	"github.com/opennox/libs/object"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/ntype"
	"github.com/opennox/opennox/v1/server"
)

type playerLeaveObserverHooks_4E6AA0 struct {
	isMonsterBot    func(*server.Object) bool
	unsetStatus     func(*server.Player, uint32)
	disableEnchant  func(*server.Object, server.EnchantID)
	setPlayerUpdate func(*server.Object)
	markUpdate      func(*server.Object)
	gameFlag        func(noxflags.GameFlag) bool
	gameplayFlag    func(noxflags.GameplayFlag) bool
	teamFlag        func(*server.Object) *server.Object
	pickupTeamFlag  func(*server.Object, *server.Object)
	questListed     func(*server.Player) int
	rememberQuest   func(*server.Player)
	firstPlayerUnit func() *server.Object
	nextPlayerUnit  func(*server.Object) *server.Object
	reportEnchant   func(ntype.PlayerInd, *server.Object)
}

// playerLeaveObserver_4E6AA0 preserves GAME.EXE's initial unit cache and the
// deliberate PlayerUnit and PlayerInd reloads around callbacks.
func playerLeaveObserver_4E6AA0(pl *server.Player, h playerLeaveObserverHooks_4E6AA0) {
	if pl == nil {
		return
	}
	unit := pl.PlayerUnit
	if unit == nil || h.isMonsterBot(unit) {
		return
	}

	h.unsetStatus(pl, 0x121)
	h.disableEnchant(unit, server.ENCHANT_INVISIBLE)
	flags := unit.ObjFlags
	h.setPlayerUpdate(unit)
	unit.ObjFlags = flags &^ object.FlagNoCollide
	h.markUpdate(pl.PlayerUnit)

	if h.gameFlag(noxflags.GameModeKOTR) && h.gameplayFlag(noxflags.GameplayFlag4) {
		flag := h.teamFlag(pl.PlayerUnit)
		if flag != nil && flag.InvHolder == nil {
			h.pickupTeamFlag(pl.PlayerUnit, flag)
		}
	}
	if h.gameFlag(noxflags.GameFlag15|noxflags.GameFlag16) && h.questListed(pl) == 0 {
		h.rememberQuest(pl)
	}
	if h.gameFlag(noxflags.GameModeQuest) {
		for it := h.firstPlayerUnit(); it != nil; it = h.nextPlayerUnit(it) {
			if it.UpdateDataPlayer().Player.Field4792 == 1 {
				h.reportEnchant(ntype.PlayerInd(pl.PlayerInd), it)
			}
		}
	}
}
