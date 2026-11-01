package server

import "encoding/binary"

type PlayerObjectSendRuntime518C30 struct {
	Frame        func() uint32
	HealthDelta  func(int32, uint16, int16)
	Self         func(*Object)
	UnitCode     func(*Object) uint32
	FloatToInt   func(float32) int32
	Action       func(*Object) uint32
	Random       func(int32, int32) int32
	Frames       func(int32) (int32, int32)
	Direction    func(int32) IndexedDirectionVector509E20
	ServerPacket func(int32, [12]byte) int32
	ClientPacket func(int32, int32, [12]byte) int32
	// The convenience sender previously accepted absent peer links and
	// out-of-range health-cache indices. Raw C requires valid player slots.
	GuardedInputs bool
}

// PlayerObjectSendNative518C30 preserves the cached update records and live
// Player/health reloads of the original sender. stream and health are
// independent integer flags; publication returns its complete 32-bit word.
func PlayerObjectSendNative518C30(recipient, unit *Object, stream, health int32, runtime PlayerObjectSendRuntime518C30) int32 {
	receiverUpdate := (*PlayerUpdateData)(recipient.UpdateData)
	initialPlayer := receiverUpdate.Player
	unitUpdate := (*PlayerUpdateData)(unit.UpdateData)
	index := int32(initialPlayer.PlayerInd)
	if health != 0 {
		data := unit.HealthData
		if data != nil && (!runtime.GuardedInputs || index < 32) {
			frame := runtime.Frame()
			if frame-unit.Frame134 > 2 {
				current := data.Cur
				previous := unitUpdate.HealthSamples[index]
				if current != previous {
					runtime.HealthDelta(index, uint16(unit.NetCode), int16(current-previous))
					unitUpdate.HealthSamples[index] = unit.HealthData.Cur
				}
			}
		}
	}
	if recipient == unit {
		runtime.Self(recipient)
	}
	// Load the target slot before the recipient Player link, then compute the
	// destination before the frame accessor. Both update records stay cached.
	targetPlayer := unitUpdate.Player
	if !runtime.GuardedInputs || targetPlayer != nil && receiverUpdate.Player != nil && targetPlayer.PlayerInd < 32 {
		targetIndex := targetPlayer.PlayerInd
		player := receiverUpdate.Player
		destination := &player.PeerUpdateFrames[targetIndex]
		*destination = runtime.Frame()
	}
	var packet [12]byte
	packet[0] = 0xc3
	binary.LittleEndian.PutUint16(packet[3:], unit.TypeInd)
	binary.LittleEndian.PutUint16(packet[1:], uint16(runtime.UnitCode(unit)))
	x := unit.PosVec.X
	binary.LittleEndian.PutUint16(packet[5:], uint16(runtime.FloatToInt(x)))
	y := unit.PosVec.Y
	binary.LittleEndian.PutUint16(packet[7:], uint16(runtime.FloatToInt(y)))
	packet[11] = byte(runtime.Action(unit))
	packet[10] = 0xff
	switch byte(unitUpdate.State) {
	case 1, 10, 2, 15, 16, 17, 14, 20, 18, 19, 21, 22, 24, 25, 27, 28, 29, 26, 30, 32:
		packet[10] = unitUpdate.Field59_0
	}
	if unitUpdate.Field40_0 != 0 && runtime.Random(0, 10) >= 8 {
		packet[11], packet[10] = 50, 0xff
	}
	if (unitUpdate.State == 3 || unitUpdate.State == 4) && unit.Field131 == 16 {
		packet[11], packet[10] = 51, 0xff
	}
	if unitUpdate.State != 30 && unitUpdate.Field41 != 0 {
		count, delay := runtime.Frames(52)
		frame := runtime.Frame()
		elapsed := frame - unitUpdate.Field41
		position := elapsed / (uint32(delay) + 1)
		if int32(position) >= count || elapsed >= 4 {
			unitUpdate.Field41 = 0
		} else {
			packet[11], packet[10] = 52, byte(position)
		}
	}
	vector := runtime.Direction(int32(int16(unit.Direction1)))
	direction := byte(vector.X) + 3*byte(vector.Y) + 4
	if direction > 3 {
		direction--
	}
	packet[9] |= direction << 4
	destination := int32(receiverUpdate.Player.PlayerInd)
	if stream != 0 {
		return runtime.ServerPacket(destination, packet)
	}
	return runtime.ClientPacket(destination, 1, packet)
}
