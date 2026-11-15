package server

import (
	"encoding/binary"

	"github.com/opennox/libs/types"
)

// PointFXNativeRuntime522FF0 keeps the live native point owner and the current
// C sender's signed-dword result. Send consumes the packet synchronously.
type PointFXNativeRuntime522FF0 struct {
	Send func(*types.Pointf, []byte) int32
}

func pointFXNativeCoordinates522FF0(packet []byte, point *types.Pointf) {
	binary.LittleEndian.PutUint16(packet[0:2], uint16(netFXTruncateSignedDWORD419A70(point.X)))
	binary.LittleEndian.PutUint16(packet[2:4], uint16(netFXTruncateSignedDWORD419A70(point.Y)))
}

func NetSendPointFXNative522FF0(op byte, point *types.Pointf, rt PointFXNativeRuntime522FF0) int32 {
	var packet [5]byte
	packet[0] = op
	pointFXNativeCoordinates522FF0(packet[1:], point)
	return rt.Send(point, packet[:])
}

func NetSendPointSubtypeFXNative523150(op, subtype byte, point *types.Pointf, rt PointFXNativeRuntime522FF0) int32 {
	var packet [6]byte
	packet[0], packet[1] = op, subtype
	pointFXNativeCoordinates522FF0(packet[2:], point)
	return rt.Send(point, packet[:])
}

func NetSparkExplosionFXNative5231B0(point *types.Pointf, subtype byte, rt PointFXNativeRuntime522FF0) int32 {
	var packet [6]byte
	packet[0], packet[5] = 0x93, subtype
	pointFXNativeCoordinates522FF0(packet[1:5], point)
	return rt.Send(point, packet[:])
}

// SendGeneratorBreakFXNative523200 retains the public C entry's void contract.
func SendGeneratorBreakFXNative523200(point *types.Pointf, subtype byte, rt PointFXNativeRuntime522FF0) {
	var packet [7]byte
	packet[0], packet[1], packet[6] = 0xf0, 25, subtype
	pointFXNativeCoordinates522FF0(packet[2:6], point)
	rt.Send(point, packet[:])
}
