package server

import (
	"encoding/binary"
	"unsafe"

	"github.com/opennox/libs/noxnet/netmsg"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/ntype"
)

func (s *Server) NetSendPlayerHealth4D86E0(recipient int, health uint16) {
	packet := [3]byte{byte(netmsg.MSG_REPORT_PLAYER_HEALTH)}
	binary.LittleEndian.PutUint16(packet[1:], health)
	s.NetSendPacketXxx1(recipient, packet[:], nil, 1)
}

func (s *Server) NetSendTeamHealth4D8670(recipient int, unit *Object) int {
	health := unit.HealthData
	if health == nil {
		return 0
	}
	packet := [5]byte{0xc4, 12}
	binary.LittleEndian.PutUint16(packet[2:], uint16(s.GetUnitNetCode(unit)))
	packet[4] = byte(100 * uint32(health.Cur) / uint32(health.Max))
	return s.NetSendPacketXxx1(recipient, packet[:], nil, 1)
}

// ReportPlayerHealthToTeam4D86E0 uses the player slot before sending health.
// When a unit has no HealthData, the original returns its address. Other
// paths return a transport result or zero in the same machine word.
func (s *Server) ReportPlayerHealthToTeam4D86E0(recipient int) uintptr {
	player := s.Players.ByInd(ntype.PlayerInd(recipient))
	if player == nil || player.PlayerUnit == nil {
		return 0
	}
	unit := player.PlayerUnit
	if unit.HealthData == nil {
		return uintptr(unsafe.Pointer(unit))
	}
	s.NetSendPlayerHealth4D86E0(recipient, unit.HealthData.Cur)
	if !noxflags.HasGame(noxflags.GameModeQuest) {
		return 0
	}
	return uintptr(uint32(s.NetSendTeamHealth4D8670(recipient|0x80, player.PlayerUnit)))
}

func (s *Server) ReportPlayerVitalsThrottled528030(unit *Object, reportHealth func(byte), reportMana func(byte, *Object)) int16 {
	update := unit.UpdateDataPlayer()
	health := unit.HealthData
	player := update.Player
	recipient := player.PlayerInd
	frame, fps := s.Frame(), s.TickRate()
	if health.Cur != update.Field2_1 &&
		(vitalDelta(health.Cur, update.Field2_1) >= health.Max/10 || frame-player.HealthReportFrame > fps>>2) {
		reportHealth(recipient)
		player := update.Player
		update.Field2_1 = unit.HealthData.Cur
		frame, fps = s.Frame(), s.TickRate()
		player.HealthReportFrame = frame
	}
	if update.ManaCur == update.ManaPrev {
		return int16(update.ManaCur)
	}
	if vitalDelta(update.ManaCur, update.ManaPrev) >= update.ManaMax/10 || frame-update.Player.ManaReportFrame > fps>>2 {
		reportMana(recipient, unit)
		update.ManaPrev = update.ManaCur
		frame := s.Frame()
		update.Player.ManaReportFrame = frame
		return int16(frame)
	}
	return 0
}

func vitalDelta(current, previous uint16) uint16 {
	return max(current, previous) - min(current, previous)
}
