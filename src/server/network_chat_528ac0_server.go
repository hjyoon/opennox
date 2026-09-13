package server

import (
	"encoding/binary"
	"unicode/utf16"

	"github.com/opennox/libs/noxnet/netmsg"

	"github.com/opennox/opennox/v1/internal/netlist"
)

// chatPacket528AC0 preserves the original MSG_TEXT_MESSAGE wire format. The
// original PE32 routine used a 520-byte stack buffer and an 8-bit code-unit
// count, so clamp long strings before building the packet.
func (s *Server) chatPacket528AC0(obj *Object, message string, duration uint16) []byte {
	units := utf16.Encode([]rune(message))
	for i, ch := range units {
		if ch == 0 {
			units = units[:i]
			break
		}
	}
	narrow := true
	for _, ch := range units {
		if ch > 0xff {
			narrow = false
			break
		}
	}
	maxUnits := 254
	width := 1
	flags := byte(2)
	if !narrow {
		maxUnits = 253 // 11-byte header plus 2*(253+1) bytes fits 520.
		width = 2
		flags = 4
	}
	if len(units) > maxUnits {
		units = units[:maxUnits]
		if !narrow && len(units) != 0 && units[len(units)-1] >= 0xd800 && units[len(units)-1] <= 0xdbff {
			units = units[:len(units)-1]
		}
	}
	packet := make([]byte, 11+width*(len(units)+1))
	packet[0] = byte(netmsg.MSG_TEXT_MESSAGE)
	binary.LittleEndian.PutUint16(packet[1:3], uint16(s.GetUnitNetCode(obj)))
	packet[3] = flags
	binary.LittleEndian.PutUint16(packet[4:6], uint16(int64(obj.PosVec.X)))
	binary.LittleEndian.PutUint16(packet[6:8], uint16(int64(obj.PosVec.Y)))
	packet[8] = byte(len(units) + 1)
	binary.LittleEndian.PutUint16(packet[9:11], duration)
	for i, ch := range units {
		if narrow {
			packet[11+i] = byte(ch)
		} else {
			binary.LittleEndian.PutUint16(packet[11+2*i:], ch)
		}
	}
	return packet
}

// NetSendChat528AC0 broadcasts a scripted object's chat message without
// crossing the PE32 C object's int-sized pointer boundary.
func (s *Server) NetSendChat528AC0(obj *Object, message string, duration uint16) {
	packet := s.chatPacket528AC0(obj, message, duration)
	for playerUnit := s.Players.FirstUnit(); playerUnit != nil; playerUnit = s.Players.NextUnit(playerUnit) {
		s.NetList.AddToMsgListCli(playerUnit.UpdateDataPlayer().Player.PlayerIndex(), netlist.Kind1, packet)
	}
}
