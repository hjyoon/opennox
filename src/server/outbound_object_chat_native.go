package server

import (
	"encoding/binary"
	"strings"
	"unsafe"

	"github.com/opennox/libs/noxnet/netmsg"
	"github.com/opennox/libs/object"
	"github.com/opennox/opennox/v1/common/ntype"
	"github.com/opennox/opennox/v1/internal/netlist"
)

// OutboundObjectChatRuntime binds the packet builders to the live object and
// player owners. Return values use the widths of their original services.
type OutboundObjectChatRuntime struct {
	UnitCode  func(*Object) uint32
	Send      func(int, []byte, *Object, int) int
	FirstUnit func() *Object
	NextUnit  func(*Object) *Object
	Queue     func(uint8, netlist.Kind, []byte) int32
}

func (s *Server) OutboundObjectChatRuntime() OutboundObjectChatRuntime {
	return OutboundObjectChatRuntime{
		UnitCode:  func(unit *Object) uint32 { return uint32(s.GetUnitNetCode(unit)) },
		Send:      s.NetSendPacketXxx0,
		FirstUnit: s.Players.FirstUnit,
		NextUnit:  s.questNextPlayerUnit4DA7F0,
		Queue: func(index uint8, kind netlist.Kind, packet []byte) int32 {
			if s.NetList.AddToMsgListCli(ntype.PlayerInd(index), kind, packet) {
				return 1
			}
			return 0
		},
	}
}

func visibilityPacket528A60(op netmsg.Op, code uint16) [3]byte {
	return [3]byte{byte(op), byte(code), byte(code >> 8)}
}

func NetObjectVisibility528A60(recipient int, unit *Object, op netmsg.Op, rt OutboundObjectChatRuntime) int {
	packet := visibilityPacket528A60(op, uint16(rt.UnitCode(unit)))
	return rt.Send(recipient, packet[:], nil, 1)
}

func chatPlayerIndex(unit *Object) uint8 {
	return (*PlayerUpdateData)(unit.UpdateData).Player.PlayerInd
}

func queueChatForPlayerUnits(packet []byte, rt OutboundObjectChatRuntime) {
	for unit := rt.FirstUnit(); unit != nil; unit = rt.NextUnit(unit) {
		rt.Queue(chatPlayerIndex(unit), netlist.Kind1, packet)
	}
}

// MonsterCommandSendRaw528BD0 returns the full source pointer for a
// non-player source and the queue result for a direct player recipient.
func MonsterCommandSendRaw528BD0(unit, source *Object, command string, duration int16, rt OutboundObjectChatRuntime) uintptr {
	code := uint16(rt.UnitCode(unit))
	if end := strings.IndexByte(command, 0); end >= 0 {
		command = command[:end]
	}
	count := uint8(len(command) + 1)
	packet := make([]byte, 11+int(count))
	packet[0] = byte(netmsg.MSG_TEXT_MESSAGE)
	binary.LittleEndian.PutUint16(packet[1:3], code)
	packet[3] = 8
	binary.LittleEndian.PutUint16(packet[4:6], uint16(x87TruncSignedQwordLow566DCC(float64(unit.PosVec.X))))
	binary.LittleEndian.PutUint16(packet[6:8], uint16(x87TruncSignedQwordLow566DCC(float64(unit.PosVec.Y))))
	packet[8] = count
	binary.LittleEndian.PutUint16(packet[9:11], uint16(duration))
	copy(packet[11:], command)
	if source == nil {
		queueChatForPlayerUnits(packet, rt)
		return 0
	}
	if uint8(source.ObjClass)&uint8(object.ClassPlayer) != 0 {
		return uintptr(uint32(rt.Queue(chatPlayerIndex(source), netlist.Kind1, packet)))
	}
	return uintptr(unsafe.Pointer(source))
}

func clearChatPacket528D60() [3]byte {
	return [3]byte{byte(netmsg.MSG_CHAT_KILL), 0xad, 0xde}
}

func DestroyEveryChatRaw528D60(rt OutboundObjectChatRuntime) int32 {
	packet := clearChatPacket528D60()
	queueChatForPlayerUnits(packet[:], rt)
	return 0
}
