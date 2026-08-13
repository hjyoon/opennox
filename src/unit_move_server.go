package opennox

import "github.com/opennox/opennox/v1/server"

// unitMoveOnlinePlayer4E7010 binds the independently tested load/call contract
// to the native-width server and player records used by the production path.
func unitMoveOnlinePlayer4E7010(s *server.Server, ud *server.PlayerUpdateData) {
	unitMoveOnline4E7010(ud, unitMoveOnline4E7010Hooks[*server.PlayerUpdateData, *server.Player]{
		frame:        s.Frame,
		setMoveFrame: func(ud *server.PlayerUpdateData, frame uint32) { ud.Field68 = frame },
		player:       func(ud *server.PlayerUpdateData) *server.Player { return ud.Player },
		playerIndex:  func(pl *server.Player) uint8 { return pl.PlayerInd },
		markPlayer:   func(ind uint8) { s.Sub4DE4D0(int(ind)) },
		sendPacket: func(ind uint8, buf [5]byte) {
			s.NetSendPacketXxx1(int(ind), buf[:], nil, 0)
		},
	})
}
