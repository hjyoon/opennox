package opennox

import (
	"github.com/opennox/libs/noxnet/netmsg"

	"github.com/opennox/opennox/v1/common/ntype"
	"github.com/opennox/opennox/v1/server"
)

// These boundaries reuse the same runtime factories as the ordinary typed
// network path. The legacy C entries do not maintain a second transaction.
func (s *Server) Game4TradeP2PRuntime() server.TradeP2PRuntime50F3A0 {
	return s.tradeP2PRuntime50F3A0()
}

func (s *Server) Game4ShopBuyRuntime() server.ShopBuyRuntime5100C0 {
	return s.shopBuyRuntime5100C0()
}

func (s *Server) Game4ShopSellRuntime() server.ShopSellRuntime5109C0 {
	return s.shopSellRuntime5109C0()
}

func (s *Server) Game4ShopRepairRuntime() server.ShopRepairRuntime5108D0 {
	return s.shopRepairRuntime5108D0()
}

// Game4TradePacketNative handles one MSG_TRADE packet. onPacketOp never calls
// the C decoder, so a C-to-Go handoff cannot recurse back into its old case.
func (s *Server) Game4TradePacketNative(unit *server.Object, packet []byte) int {
	if unit == nil || unit.UpdateData == nil || len(packet) < 2 || packet[0] != byte(netmsg.MSG_TRADE) {
		return -1
	}
	player := unit.UpdateDataPlayer().Player
	if player == nil {
		return -1
	}
	n, ok := s.onPacketOp(ntype.PlayerInd(player.PlayerInd), netmsg.MSG_TRADE, packet, player, unit)
	if !ok {
		return -1
	}
	return n
}
