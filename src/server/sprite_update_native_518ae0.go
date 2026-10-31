package server

import (
	"encoding/binary"
	"unsafe"
)

type SpriteUpdateNativeRuntime518AE0 struct {
	UnitCode     func(*Object) uint32
	ReadType     func(int) uint32
	ElevatorLink func(*Object, unsafe.Pointer) *Object
	Important    func(int32, [4]byte, int32, int32) int32
	Queue        func(int32, int32, [4]byte) int32
	Special      func(*Object, *Object) int32
}

// SpriteUpdateNative518AE0 retains the raw original entry's selection order
// and byte reads. Both object arguments and the exact cached shaft update
// remain native pointers; the obsolete PE32 link word is never dereferenced.
func SpriteUpdateNative518AE0(recipient *Object, player int32, unit *Object, runtime SpriteUpdateNativeRuntime518AE0) int32 {
	code := runtime.UnitCode(unit)
	class := uint32(unit.ObjClass)
	var packet [4]byte
	binary.LittleEndian.PutUint16(packet[1:], uint16(code))
	if class&0x400000 != 0 && uint32(unit.ObjSubClass)&0x18 != 0 {
		packet[0] = 0xb3
		packet[3] = *(*byte)(unsafe.Pointer(&unit.UpdateDataObelisk().Mana))
		return runtime.Important(player, packet, 0, 1)
	}
	firstType := runtime.ReadType(0)
	typ := uint32(unit.TypeInd)
	switch {
	case typ == firstType:
		packet[0] = 0x39
		packet[3] = unit.UpdateDataPentagram().AnimationStep
	case typ == runtime.ReadType(2):
		packet[0] = 0x39
		packet[3] = ^byte(uint32(unit.ObjFlags)>>24) & 1
	case typ == runtime.ReadType(1):
		packet[0] = 0xb4
		packet[3] = *(*byte)(unit.UpdateData) & 1
	case class&0x4000 != 0:
		packet[0] = 0x39
		packet[3] = *(*byte)(unsafe.Pointer(&unit.UpdateDataElevator().Field_4)) >> 2
	case class&0x8000 != 0:
		packet[0] = 0x39
		if data := unit.UpdateData; data != nil {
			if linked := runtime.ElevatorLink(unit, data); linked != nil {
				packet[3] = *(*byte)(unsafe.Pointer(&linked.UpdateDataElevator().Field_4)) >> 2
			}
		}
	case class&0x80 != 0:
		packet[0] = 0xb2
		packet[3] = *(*byte)(unsafe.Pointer(&unit.UpdateDataDoor().CurrentDirection))
	default:
		runtime.Special(recipient, unit)
		return 0
	}
	return runtime.Queue(player, 1, packet)
}
