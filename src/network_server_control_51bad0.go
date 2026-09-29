package opennox

import (
	"encoding/binary"

	"github.com/opennox/libs/noxnet/netmsg"
	"github.com/opennox/libs/types"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/server"
)

const (
	networkPlayerSetWaypointPacketSize51BAD0 = 7
	networkVoteNamedPacketSize51BAD0         = 52
	networkVoteControlPacketSize51BAD0       = 2
	networkVoteNameUnits51BAD0               = (networkVoteNamedPacketSize51BAD0 - 2) / 2
)

type networkControlRuntime51BAD0 struct {
	questMode   bool
	setWaypoint func(*server.Object, types.Pointf)
	voteStart   func(int, *server.Object, []uint16)
	voteCancel  func(int, *server.Object, []uint16)
	gauntlet    func(*[server.NetworkGauntletPacketSize51BAD0]byte, *server.Object, *server.PlayerUpdateData) int32
}

// dispatchNetworkControlNative51BAD0 decodes the remaining control packets
// before the legacy server C switch can narrow Object and PlayerUpdateData
// pointers into IA-32 int temporaries. handled distinguishes a malformed known
// packet from an unrelated opcode so malformed input never falls back to C.
func dispatchNetworkControlNative51BAD0(
	op netmsg.Op,
	data []byte,
	unit *server.Object,
	update *server.PlayerUpdateData,
	runtime networkControlRuntime51BAD0,
) (n int, handled, valid bool) {
	switch op {
	case netmsg.MSG_PLAYER_SET_WAYPOINT:
		if len(data) < networkPlayerSetWaypointPacketSize51BAD0 {
			return 0, true, false
		}
		position := types.Pointf{
			X: float32(binary.LittleEndian.Uint16(data[3:5])),
			Y: float32(binary.LittleEndian.Uint16(data[5:7])),
		}
		runtime.setWaypoint(unit, position)
		return networkPlayerSetWaypointPacketSize51BAD0, true, true

	case netmsg.MSG_VOTE:
		if len(data) < networkVoteControlPacketSize51BAD0 {
			return 0, true, false
		}
		subtype := data[1]
		if subtype <= 3 {
			if len(data) < networkVoteNamedPacketSize51BAD0 {
				return 0, true, false
			}
			// The wire field has no spare terminator when all 25 code units are
			// populated. Copy it and append one before calling the legacy lookup.
			name := make([]uint16, networkVoteNameUnits51BAD0+1)
			for i := 0; i < networkVoteNameUnits51BAD0; i++ {
				name[i] = binary.LittleEndian.Uint16(data[2+i*2 : 4+i*2])
			}
			kind := 1
			if subtype&1 == 0 {
				kind = 0
				if runtime.questMode {
					kind = 3
				}
			}
			if subtype < 2 {
				runtime.voteStart(kind, unit, name)
			} else {
				runtime.voteCancel(kind, unit, name)
			}
			return networkVoteNamedPacketSize51BAD0, true, true
		}
		switch subtype {
		case 4:
			runtime.voteStart(2, unit, nil)
		case 5:
			runtime.voteCancel(2, unit, nil)
		default:
			return 0, true, false
		}
		return networkVoteControlPacketSize51BAD0, true, true

	case netmsg.MSG_GAUNTLET:
		if len(data) < server.NetworkGauntletPacketSize51BAD0 {
			return 0, true, false
		}
		var packet [server.NetworkGauntletPacketSize51BAD0]byte
		copy(packet[:], data[:server.NetworkGauntletPacketSize51BAD0])
		n = int(runtime.gauntlet(&packet, unit, update))
		if n <= 0 || n > len(data) {
			return 0, true, false
		}
		return n, true, true

	default:
		return 0, false, false
	}
}

func (s *Server) onPacketControlNative51BAD0(
	op netmsg.Op,
	data []byte,
	unit *server.Object,
	update *server.PlayerUpdateData,
) (n int, handled, valid bool) {
	return dispatchNetworkControlNative51BAD0(op, data, unit, update, networkControlRuntime51BAD0{
		questMode:   noxflags.HasGame(noxflags.GameModeQuest),
		setWaypoint: s.playerSetCustomWaypointNative4F79A0,
		voteStart: func(kind int, unit *server.Object, name []uint16) {
			legacy.VoteStart506870(kind, unit, name)
		},
		voteCancel: legacy.VoteCancel506C90,
		gauntlet:   legacy.NetworkGauntletCall51BAD0,
	})
}

// playerSetCustomWaypointNative4F79A0 preserves the three-slot waypoint
// behavior while keeping every live object reference at native pointer width.
func (s *Server) playerSetCustomWaypointNative4F79A0(unit *server.Object, position types.Pointf) {
	if unit == nil || unit.UpdateData == nil {
		return
	}
	update := unit.UpdateDataPlayer()
	if update.Player == nil || update.Player.Field3680&3 != 0 {
		return
	}
	index := int(update.CustomWaypointWrite)
	if index < 0 || index >= len(update.CustomWaypoints) {
		return
	}
	if waypoint := update.CustomWaypoints[index]; waypoint != nil {
		asObjectS(waypoint).SetPos(position)
		return
	}
	waypoint := s.NewObjectByTypeID("PlayerWaypoint")
	if waypoint == nil {
		return
	}
	update.CustomWaypoints[index] = waypoint
	s.CreateObjectAt(waypoint, unit, position)
}
