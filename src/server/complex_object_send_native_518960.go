package server

import "encoding/binary"

// ComplexObjectSendNative518960 preserves GAME.EXE 00518960's helper order
// and entry update snapshots. The opaque C object pointer never becomes an int.
func ComplexObjectSendNative518960(player int32, unit *Object, report int32, runtime ComplexObjectSendNativeRuntime518960) int32 {
	if report != 0 {
		if health := unit.HealthData; health != nil {
			frame := runtime.Frame()
			last := unit.Frame134
			update := (*MonsterUpdateData)(unit.UpdateData)
			if frame-last > 2 {
				current := health.Cur
				previous := update.HealthGraph103[player]
				if current != previous {
					runtime.ReportHealth(player, uint16(unit.NetCode), int16(current-previous))
					update.HealthGraph103[player] = unit.HealthData.Cur
				}
			}
		}
	}
	var packet [11]byte
	packet[0] = 48
	binary.LittleEndian.PutUint16(packet[3:], unit.TypeInd)
	binary.LittleEndian.PutUint16(packet[1:], uint16(runtime.UnitCode(unit)))
	binary.LittleEndian.PutUint16(packet[5:], uint16(runtime.FloatToInt(unit.PosVec.X)))
	y := runtime.FloatToInt(unit.PosVec.Y)
	update := (*MonsterUpdateData)(unit.UpdateData)
	binary.LittleEndian.PutUint16(packet[7:], uint16(y))
	packet[9] = uint8(runtime.Animation(unit))
	subclass := uint8(unit.ObjSubClass)
	packet[10] = update.Field120_1
	if subclass&0x10 != 0 && update.Field523_2 != 0 && runtime.Random(0, 10) >= 8 {
		packet[9] = 14
		frames, _ := runtime.Frames(50)
		packet[10] = uint8(runtime.Random(0, frames))
	}
	var direction IndexedDirectionVector509E20
	runtime.Direction(int32(int16(unit.Direction1)), &direction)
	value := uint8(direction.X) + 3*uint8(direction.Y) + 4
	if value > 3 {
		value--
	}
	packet[9] |= value << 4
	return runtime.Enqueue(player, packet)
}

// ComplexObjectPacket518960 uses the same native sender without health
// reporting or network publication. Its random draws still advance game state.
func (s *Server) ComplexObjectPacket518960(unit *Object) (packet [11]byte) {
	ComplexObjectSendNative518960(0, unit, 0, ComplexObjectSendNativeRuntime518960{
		UnitCode:   func(unit *Object) uint32 { return uint32(s.GetUnitNetCode(unit)) },
		FloatToInt: netFXTruncateSignedDWORD419A70,
		Animation:  s.MonsterAnimationIndex533790,
		Random:     func(min, max int32) int32 { return int32(s.Rand.Logic.IntClamp(int(min), int(max))) },
		Frames: func(index int32) (int32, int32) {
			frames, duration := s.PlayerAnimFrames(int(index))
			return int32(frames), int32(duration)
		},
		Direction: func(index int32, out *IndexedDirectionVector509E20) { IndexedDirection509E20(index, out) },
		Enqueue: func(_ int32, value [11]byte) int32 {
			packet = value
			return 1
		},
	})
	return packet
}
