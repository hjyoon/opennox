//go:build amd64 || arm64

package legacy

/*
#include "common__system__team.h"
*/
import "C"

import (
	"encoding/binary"
	"unsafe"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/server"
)

//export nox_xxx_createAtImpl_native_4191D0
func nox_xxx_createAtImpl_native_4191D0(
	teamIDC C.uchar,
	valueC *C.nox_object_team_t,
	activeC C.int,
	netCodeC C.int,
	flagsC C.int,
) {
	// Pass nil through: the native server entry preserves GAME.EXE's GameBall
	// cache read before its null guard.
	srv := GetServer().S()
	srv.TeamCreateAtImpl4191D0(
		server.TeamID(uint8(teamIDC)),
		(*server.ObjectTeam)(unsafe.Pointer(valueC)),
		int32(activeC),
		uint32(netCodeC),
		int32(flagsC),
		teamCreateAtRuntimeNative4191D0(srv),
	)
}

func teamCreateAtRuntimeNative4191D0(srv *server.Server) server.TeamCreateAtRuntime4191D0 {
	return server.TeamCreateAtRuntime4191D0{
		ClientNetCode: uint32(ClientPlayerNetCode()),
		AfterAttach: func(
			team *server.Team,
			_ *server.ObjectTeam,
			active int32,
			netCode uint32,
			_ int32,
			gameBallType uint16,
		) {
			// Offline solo reaches this callback but has no team-network effects.
			// For an online host, preserve the safe, wire-visible attachment packet;
			// the legacy roster-window and join-message rendering remain UI work.
			if !noxflags.HasGame(noxflags.GameHost) || !noxflags.HasGame(noxflags.GameOnline) || active == 0 {
				return
			}
			var unit *server.Object
			if srv.ObjectByNetCode != nil {
				unit = srv.ObjectByNetCode(int(netCode))
			}
			player := srv.Players.ByID(int(netCode))
			if unit == nil || player == nil && unit.TypeInd != gameBallType {
				return
			}
			var packet [10]byte
			packet[0] = 0xc4
			packet[1] = 1
			binary.LittleEndian.PutUint32(packet[2:6], uint32(team.ID()))
			binary.LittleEndian.PutUint16(packet[6:8], uint16(netCode))
			binary.LittleEndian.PutUint16(packet[8:10], unit.TypeInd)
			srv.NetSendPacketXxx1(159, packet[:], nil, 1)
		},
	}
}
