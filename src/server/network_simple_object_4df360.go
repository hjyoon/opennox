package server

import "encoding/binary"

// NetSendSimpleObject4DF360 preserves GAME.EXE's nine-byte packet and read
// order. The object remains native-width; only its wire fields are narrowed.
func (s *Server) NetSendSimpleObject4DF360(recipient int32, obj *Object, floatToInt func(float32) int32) int32 {
	var packet [9]byte
	packet[0] = 0x2f
	binary.LittleEndian.PutUint16(packet[3:5], obj.TypeInd)
	binary.LittleEndian.PutUint16(packet[1:3], uint16(s.GetUnitNetCode(obj)))
	x := obj.PosVec.X
	binary.LittleEndian.PutUint16(packet[5:7], uint16(floatToInt(x)))
	y := obj.PosVec.Y
	binary.LittleEndian.PutUint16(packet[7:9], uint16(floatToInt(y)))
	return int32(s.NetSendPacketXxx1(int(recipient), packet[:], nil, 1))
}
