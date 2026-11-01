package server

import "encoding/binary"

type BasicObjectSendRuntime5187E0 struct {
	Frame        func() uint32
	HealthDelta  func(int32, uint16, int16)
	UnitCode     func(*Object) uint32
	FloatToInt   func(float32) int32
	Direction    func(int32) IndexedDirectionVector509E20
	ServerPacket func(int32, []byte) int32
	// The ordinary dispatcher previously accepted missing generator caches and
	// out-of-range sample slots. The raw C entries require valid health inputs.
	GuardedInputs bool
}

func basicObjectPacketNative5187E0(unit *Object, opcode byte, runtime BasicObjectSendRuntime5187E0) [11]byte {
	var packet [11]byte
	packet[0] = opcode
	// Type is captured before UnitCode; each coordinate is loaded immediately
	// before its conversion. The original helpers may expose live record changes.
	binary.LittleEndian.PutUint16(packet[3:], unit.TypeInd)
	binary.LittleEndian.PutUint16(packet[1:], uint16(runtime.UnitCode(unit)))
	x := unit.PosVec.X
	binary.LittleEndian.PutUint16(packet[5:], uint16(runtime.FloatToInt(x)))
	y := unit.PosVec.Y
	binary.LittleEndian.PutUint16(packet[7:], uint16(runtime.FloatToInt(y)))
	return packet
}

func PhantomObjectSendNative5187E0(index int32, unit *Object, runtime BasicObjectSendRuntime5187E0) int32 {
	packet := basicObjectPacketNative5187E0(unit, 48, runtime)
	packet[10] = 0xff
	if uint32(unit.ObjClass)&1 != 0 && uint32(unit.ObjSubClass)&0x30 != 0 {
		packet[10] = byte(uint16(unit.Direction1) >> 3)
	}
	vector := runtime.Direction(int32(int16(unit.Direction1)))
	direction := byte(vector.X) + 3*byte(vector.Y) + 4
	if direction > 3 {
		direction--
	}
	packet[9] = direction << 4
	return runtime.ServerPacket(index, packet[:])
}

func SimpleObjectSendNative5188A0(index int32, unit *Object, runtime BasicObjectSendRuntime5187E0) int32 {
	if uint32(unit.ObjClass)&0x20000 != 0 {
		health := unit.HealthData
		if health != nil {
			// Original5188C1 reads Frame before5188C7's damage timestamp and
			//5188CD's update pointer. Only the health link is cached before Frame.
			frame := runtime.Frame()
			elapsed := frame - unit.Frame134
			update := (*MonsterGenUpdateData)(unit.UpdateData)
			if elapsed > 2 && (!runtime.GuardedInputs || update != nil && index >= 0 && index < 32) {
				current, previous := health.Cur, update.HealthSamples[index]
				if current != previous {
					runtime.HealthDelta(index, uint16(unit.NetCode), int16(current-previous))
					update.HealthSamples[index] = unit.HealthData.Cur
				}
			}
		}
	}
	packet := basicObjectPacketNative5187E0(unit, 47, runtime)
	return runtime.ServerPacket(index, packet[:9])
}
