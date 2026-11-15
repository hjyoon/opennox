package server

import (
	"encoding/binary"

	"github.com/opennox/libs/noxnet/netmsg"
)

// PredictLinearNativeRuntime523530 preserves native object owners across the
// net-code and live player traversal services. Send consumes the packet now.
type PredictLinearNativeRuntime523530 struct {
	NetCode func(*Object) uint32
	First   func() *Object
	Next    func(*Object) *Object
	Send    func(uint8, []byte)
}

// PredictLinearNative523530 snapshots the fourteen-byte packet before First,
// then reloads each recipient's native update/player links for each Send.
// The original terminating null player-unit pointer supplies the int result.
func PredictLinearNative523530(obj *Object, rt PredictLinearNativeRuntime523530) int32 {
	var packet [14]byte
	packet[0] = byte(netmsg.MSG_CLIENT_PREDICT_LINEAR)
	code := rt.NetCode(obj)
	x := obj.PosVec.X
	binary.LittleEndian.PutUint16(packet[1:3], uint16(code))
	binary.LittleEndian.PutUint16(packet[3:5], obj.TypeInd)
	binary.LittleEndian.PutUint16(packet[5:7], uint16(x87TruncSignedQwordLow566DCC(float64(x))))
	binary.LittleEndian.PutUint16(packet[7:9], uint16(x87TruncSignedQwordLow566DCC(float64(obj.PosVec.Y))))
	damping := float64(obj.Float28) * 16
	binary.LittleEndian.PutUint16(packet[9:11], uint16(obj.Direction1))
	packet[11] = byte(x87TruncSignedQwordLow566DCC(damping))
	packet[12] = byte(x87TruncSignedQwordLow566DCC(float64(obj.VelVec.X) * 16))
	packet[13] = byte(x87TruncSignedQwordLow566DCC(float64(obj.VelVec.Y) * 16))
	for unit := rt.First(); unit != nil; unit = rt.Next(unit) {
		player := (*PlayerUpdateData)(unit.UpdateData).Player
		rt.Send(player.PlayerInd, packet[:])
	}
	return 0
}
