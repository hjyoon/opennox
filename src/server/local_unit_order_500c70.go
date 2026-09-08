package server

import "github.com/opennox/opennox/v1/common/ntype"

type localUnitOrderHooks500C70[Player any] struct {
	playerByIndex       func(ntype.PlayerInd) Player
	storeOrder          func(Player, uint32)
	sendCreatureCommand func(ntype.PlayerInd, byte) int
}

func localUnitOrder500C70[Player any](owner ntype.PlayerInd, orderType uint32, hooks localUnitOrderHooks500C70[Player]) int {
	player := hooks.playerByIndex(owner)
	hooks.storeOrder(player, orderType)
	return hooks.sendCreatureCommand(owner, byte(orderType))
}
