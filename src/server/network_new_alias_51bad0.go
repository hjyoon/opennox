package server

const (
	networkNewAliasPacketSize51BAD0 = 10
	networkNewAliasCount51BAD0      = 255
)

// networkNewAliasHooks51BAD0 exposes the packet reads and native-width table
// stores in the MSG_NEW_ALIAS branch at GAME.EXE 0051C36B..0051C3B6.
type networkNewAliasHooks51BAD0[P comparable] struct {
	loadAlias     func() byte
	loadCode      func() uint16
	loadTypeID    func() uint16
	loadDeadline  func() uint32
	storeCode     func(P, byte, uint16)
	storeTypeID   func(P, byte, uint16)
	storeDeadline func(P, byte, uint32)
}

// networkNewAlias51BAD0 updates one of the 255 per-player aliases without
// converting the Player identity to a PE32 integer. Alias 0xff is the wire
// escape value, not a table slot, so malformed reports are consumed without
// overwriting the Player fields that follow the table.
func networkNewAlias51BAD0[P comparable](player P, hooks networkNewAliasHooks51BAD0[P]) int32 {
	alias := hooks.loadAlias()
	if int(alias) >= networkNewAliasCount51BAD0 {
		return networkNewAliasPacketSize51BAD0
	}
	hooks.storeCode(player, alias, hooks.loadCode())
	hooks.storeTypeID(player, alias, hooks.loadTypeID())
	hooks.storeDeadline(player, alias, hooks.loadDeadline())
	return networkNewAliasPacketSize51BAD0
}
