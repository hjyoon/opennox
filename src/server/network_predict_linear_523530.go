package server

import (
	"encoding/binary"

	"github.com/opennox/libs/noxnet/netmsg"

	"github.com/opennox/opennox/v1/common/ntype"
	"github.com/opennox/opennox/v1/internal/netlist"
)

type predictLinearDeps523530 struct {
	netCode func(*Object) uint32
	first   func() *Object
	next    func(*Object) *Object
	send    func(uint8, []byte)
}

// predictLinear523530 restores GAME.EXE 00523530..005235EE. Object,
// PlayerUpdateData and Player pointers stay native-width; the fourteen-byte
// packet retains the original low WORD/BYTE values, including wrapped and
// invalid x87 signed-qword conversions (00566DCC).
func predictLinear523530(obj *Object, h predictLinearDeps523530) int32 {
	var packet [14]byte
	packet[0] = byte(netmsg.MSG_CLIENT_PREDICT_LINEAR)
	code := h.netCode(obj)
	x := obj.PosVec.X // Read the live object only after the net-code callback.
	binary.LittleEndian.PutUint16(packet[1:3], uint16(code))
	binary.LittleEndian.PutUint16(packet[3:5], obj.TypeInd)
	binary.LittleEndian.PutUint16(packet[5:7], uint16(x87TruncSignedQwordLow566DCC(float64(x))))
	binary.LittleEndian.PutUint16(packet[7:9], uint16(x87TruncSignedQwordLow566DCC(float64(obj.PosVec.Y))))
	damping := float64(obj.Float28) * 16
	binary.LittleEndian.PutUint16(packet[9:11], uint16(obj.Direction1))
	packet[11] = byte(x87TruncSignedQwordLow566DCC(damping))
	packet[12] = byte(x87TruncSignedQwordLow566DCC(float64(obj.VelVec.X) * 16))
	packet[13] = byte(x87TruncSignedQwordLow566DCC(float64(obj.VelVec.Y) * 16))
	for unit := h.first(); unit != nil; unit = h.next(unit) {
		player := (*PlayerUpdateData)(unit.UpdateData).Player
		h.send(player.PlayerInd, packet[:])
	}
	return 0 // The original returns the terminating null player-unit pointer.
}

func (s *Server) NetClientPredictLinear523530(obj *Object) int32 {
	return predictLinear523530(obj, predictLinearDeps523530{
		netCode: func(obj *Object) uint32 { return uint32(s.GetUnitNetCode(obj)) },
		first:   s.Players.FirstUnit,
		next:    s.questNextPlayerUnit4DA7F0,
		send: func(index uint8, packet []byte) {
			s.NetList.AddToMsgListCli(ntype.PlayerInd(index), netlist.Kind1, packet)
		},
	})
}
