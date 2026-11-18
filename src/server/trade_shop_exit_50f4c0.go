package server

import (
	"github.com/opennox/libs/noxnet/netmsg"
	"github.com/opennox/libs/object"
	noxflags "github.com/opennox/opennox/v1/common/flags"
)

// NotifyShopExit50F4C0 clears participant links, unfreezes the player and
// sends the close packet. The caller retains responsibility for the session.
func (s *Server) NotifyShopExit50F4C0(session *TradeSession, unfreeze func(*Object, int)) (int, bool) {
	session.Field0 = 0
	var player *Object
	for _, unit := range [...]*Object{session.Field8, session.Field12} {
		if !unit.Class().Has(object.ClassPlayer) {
			continue
		}
		update := unit.UpdateDataPlayer()
		if update != nil && update.Trade70 == session {
			update.Trade70 = nil
		}
		if player == nil {
			player = unit
		}
	}
	if player == nil {
		return 0, false
	}
	if unfreeze != nil {
		unfreeze(player, 0)
	}
	update := player.UpdateDataPlayer()
	if update == nil || update.Player == nil {
		return 0, false
	}
	index := update.Player.Index()
	packet := [...]byte{byte(netmsg.MSG_TRADE), 0x02}
	s.NetSendPacketXxx1(index, packet[:], nil, 1)
	return index, true
}

// CloseShopNative50F4C0 handles an owned shop session. A synchronous send
// can close the original allocation before the notification returns.
func (s *Server) CloseShopNative50F4C0(session *TradeSession, unfreeze func(*Object, int)) bool {
	owner := s.tradeNative.sessions[session]
	if owner == nil {
		return false
	}
	keep := false
	defer func() {
		if !keep && s.tradeNative.sessions[session] == owner {
			s.ReleaseTradeSessionNative510000(session)
		}
	}()
	index, notified := s.NotifyShopExit50F4C0(session, unfreeze)
	if notified && s.tradeNative.sessions[session] == owner && noxflags.HasGame(noxflags.GameModeQuest) {
		keep = s.CacheQuestShopSessionNative50F4C0(index, session)
	}
	return true
}
