package opennox

import (
	"encoding/binary"

	"github.com/opennox/libs/noxnet/netmsg"
	"github.com/opennox/libs/object"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/server"
)

const playerStatsPacketSize48EA70 = 14

type playerStatsState48EA70 struct {
	RawNetCode uint16
	NetCode    uint16
	HealthMax  uint16
	ManaMax    uint16
	Capacity   uint16
	Speed      uint16
	Strength   uint16
	Level      byte
}

type playerStatsHooks48EA70 struct {
	connected   func() bool
	localCode   func() uint16
	gameHost    func() bool
	localPlayer func() *server.Player
	refreshName func()
}

func decodePlayerStatsState48EA70(data []byte) (playerStatsState48EA70, bool) {
	if len(data) < playerStatsPacketSize48EA70 {
		return playerStatsState48EA70{}, false
	}
	rawCode := binary.LittleEndian.Uint16(data[1:3])
	return playerStatsState48EA70{
		RawNetCode: rawCode,
		NetCode:    nox_xxx_netClearHighBit_578B30(rawCode),
		HealthMax:  binary.LittleEndian.Uint16(data[3:5]),
		ManaMax:    binary.LittleEndian.Uint16(data[5:7]),
		Capacity:   binary.LittleEndian.Uint16(data[7:9]),
		Speed:      binary.LittleEndian.Uint16(data[9:11]),
		Strength:   binary.LittleEndian.Uint16(data[11:13]),
		Level:      data[13],
	}, true
}

func applyPlayerStatsState48EA70(player *server.Player, state playerStatsState48EA70) {
	if player == nil {
		return
	}
	info := player.Info()
	info.SetField2247(uint32(state.HealthMax))
	info.SetField2243(uint32(state.ManaMax))
	player.SetStatsCapacity(state.Capacity)
	info.SetField2235(uint32(state.Speed))
	info.SetField2239(uint32(state.Strength))
	player.Level = state.Level
}

func handlePlayerStatsNative48EA70(data []byte, hooks playerStatsHooks48EA70) int {
	state, ok := decodePlayerStatsState48EA70(data)
	if !ok {
		return -1
	}
	if !hooks.connected() || state.NetCode != hooks.localCode() {
		return playerStatsPacketSize48EA70
	}
	// The listen server shares this Player object with its client. Reapplying
	// the packet is unnecessary there, but the inventory name still refreshes.
	if !hooks.gameHost() {
		applyPlayerStatsState48EA70(hooks.localPlayer(), state)
	}
	hooks.refreshName()
	return playerStatsPacketSize48EA70
}

func (c *Client) handlePlayerStatsPacketNative48EA70(data []byte) int {
	return handlePlayerStatsNative48EA70(data, playerStatsHooks48EA70{
		connected: nox_client_isConnected,
		localCode: func() uint16 {
			return uint16(legacy.ClientPlayerNetCode())
		},
		gameHost: func() bool {
			return noxflags.HasGame(noxflags.GameHost)
		},
		localPlayer: getCurPlayer,
		refreshName: func() {
			legacy.Nox_xxx_j_inventoryNameSignInit_467460()
		},
	})
}

func playerStatsPacketNative4D8990(unit *server.Object, netCode func(*server.Object) int) ([playerStatsPacketSize48EA70]byte, bool) {
	var packet [playerStatsPacketSize48EA70]byte
	if unit == nil || !unit.ObjClass.Has(object.ClassPlayer) || unit.UpdateData == nil || unit.HealthData == nil {
		return packet, false
	}
	update := (*server.PlayerUpdateData)(unit.UpdateData)
	if update.Player == nil {
		return packet, false
	}
	player := update.Player
	info := player.Info()
	packet[0] = byte(netmsg.MSG_REPORT_STATS)
	binary.LittleEndian.PutUint16(packet[1:3], uint16(netCode(unit)))
	binary.LittleEndian.PutUint16(packet[3:5], unit.HealthData.Max)
	binary.LittleEndian.PutUint16(packet[5:7], update.ManaMax)
	binary.LittleEndian.PutUint16(packet[7:9], unit.CarryCapacity)
	binary.LittleEndian.PutUint16(packet[9:11], uint16(info.Field2235()))
	binary.LittleEndian.PutUint16(packet[11:13], uint16(info.Field2239()))
	packet[13] = player.Level
	return packet, true
}

func (s *Server) playerStatsReportNative4D8990(playerInd byte, unit *server.Object) int {
	packet, ok := playerStatsPacketNative4D8990(unit, func(unit *server.Object) int {
		return s.GetUnitNetCode(unit)
	})
	if !ok {
		return 0
	}
	return s.Server.NetSendPacketXxx0(int(playerInd), packet[:], nil, 1)
}

type playerStatsReportHooks4D9900 struct {
	totalHealth func(byte, *server.Object)
	totalMana   func(byte, *server.Object)
	stats       func(byte, *server.Object)
}

// playerReportStatsNative4D9900 restores the dirty-stat block from GAME.EXE
// 004D9900. PlayerReadValues marks StatsReportPending, and the next self
// update sends all maxima plus the derived speed, strength, capacity and level.
func playerReportStatsNative4D9900(unit *server.Object, hooks playerStatsReportHooks4D9900) {
	if unit == nil || !unit.ObjClass.Has(object.ClassPlayer) || unit.UpdateData == nil {
		return
	}
	update := (*server.PlayerUpdateData)(unit.UpdateData)
	player := update.Player
	if player == nil || player.StatsReportPending == 0 {
		return
	}
	playerInd := player.PlayerInd
	hooks.totalHealth(playerInd, unit)
	hooks.totalMana(playerInd, unit)
	hooks.stats(playerInd, unit)
	player.StatsReportPending = 0
}
