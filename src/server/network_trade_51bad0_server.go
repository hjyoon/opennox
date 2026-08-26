package server

import "encoding/binary"

// NetworkTradeStartPacketSize51BAD0 is the exact MSG_TRADE/0x15 packet width.
const NetworkTradeStartPacketSize51BAD0 = networkTradeStartPacketSize51BAD0

// NetworkTradeStartRuntime51BAD0 supplies the outer game gate and the
// native-width shop-session implementation.
type NetworkTradeStartRuntime51BAD0 struct {
	GameBlocked func() bool
	StartShop   func(*Object, *Object)
}

// NetworkTradeStart51BAD0 binds the original shop-start packet contract to
// native Object, PlayerUpdateData, and Player pointers on every pointer width.
func (s *Server) NetworkTradeStart51BAD0(
	unit *Object,
	update *PlayerUpdateData,
	packet *[NetworkTradeStartPacketSize51BAD0]byte,
	runtime NetworkTradeStartRuntime51BAD0,
) int32 {
	return networkTradeStart51BAD0(unit, update, networkTradeStartHooks51BAD0[
		*Object,
		*PlayerUpdateData,
		*Player,
	]{
		gameBlocked: runtime.GameBlocked,
		loadPlayer: func(update *PlayerUpdateData) *Player {
			return update.Player
		},
		loadPlayerStatus: func(player *Player) uint32 {
			return player.Field3680
		},
		loadWireCode: func() uint16 {
			return binary.LittleEndian.Uint16(packet[2:4])
		},
		dynamicUnitCode:   s.packetDynamicUnitCode578B40,
		objectFromNetCode: s.ObjectFromNetCode4ECCB0,
		loadMonsterSubclassLow: func(merchant *Object) uint8 {
			return uint8(merchant.ObjSubClass)
		},
		startShop: runtime.StartShop,
	})
}

// NetworkTradeExitPacketSize51BAD0 is the exact MSG_TRADE/0x12 packet width.
const NetworkTradeExitPacketSize51BAD0 = networkTradeExitPacketSize51BAD0

// NetworkTradeExit51BAD0 binds the original shop-exit packet contract to the
// native-width PlayerUpdateData and TradeSession layouts.
func NetworkTradeExit51BAD0(update *PlayerUpdateData, exitSession func(*TradeSession)) int32 {
	return networkTradeExit51BAD0(update, networkTradeExitHooks51BAD0[*PlayerUpdateData, TradeSession]{
		loadSession: func(update *PlayerUpdateData) *TradeSession {
			return update.Trade70
		},
		exitSession: exitSession,
	})
}
