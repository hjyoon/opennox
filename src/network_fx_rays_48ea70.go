package opennox

import (
	"encoding/binary"
	"image"

	"github.com/opennox/libs/noxnet/netmsg"
)

const rayFXPacketSize48EA70 = 9

type rayFXState48EA70 struct {
	Op       netmsg.Op
	From, To image.Point
	Packet   [rayFXPacketSize48EA70]byte
}

type rayFXHooks48EA70 struct {
	connected       func() bool
	drawRay         func([rayFXPacketSize48EA70]byte)
	lightningSparks func(from, to image.Point)
	plasmaEndSparks func(to image.Point)
}

func isRayFXOp48EA70(op netmsg.Op) bool {
	switch op {
	case netmsg.MSG_FX_PLASMA,
		netmsg.MSG_FX_LIGHTNING,
		netmsg.MSG_FX_ENERGY_BOLT,
		netmsg.MSG_FX_CHAIN_LIGHTNING_BOLT,
		netmsg.MSG_FX_DRAIN_MANA,
		netmsg.MSG_FX_CHARM,
		netmsg.MSG_FX_GREATER_HEAL:
		return true
	default:
		return false
	}
}

func decodeRayFXState48EA70(op netmsg.Op, data []byte) (rayFXState48EA70, bool) {
	if !isRayFXOp48EA70(op) || len(data) < rayFXPacketSize48EA70 || netmsg.Op(data[0]) != op {
		return rayFXState48EA70{}, false
	}
	state := rayFXState48EA70{
		Op: op,
		From: image.Pt(
			int(binary.LittleEndian.Uint16(data[1:3])),
			int(binary.LittleEndian.Uint16(data[3:5])),
		),
		To: image.Pt(
			int(binary.LittleEndian.Uint16(data[5:7])),
			int(binary.LittleEndian.Uint16(data[7:9])),
		),
	}
	copy(state.Packet[:], data[:rayFXPacketSize48EA70])
	return state, true
}

func handleRayFXNative48EA70(op netmsg.Op, data []byte, hooks rayFXHooks48EA70) int {
	state, ok := decodeRayFXState48EA70(op, data)
	if !ok {
		return -1
	}
	if !hooks.connected() {
		return rayFXPacketSize48EA70
	}

	hooks.drawRay(state.Packet)
	switch op {
	case netmsg.MSG_FX_LIGHTNING, netmsg.MSG_FX_CHAIN_LIGHTNING_BOLT:
		hooks.lightningSparks(state.From, state.To)
	case netmsg.MSG_FX_PLASMA:
		hooks.plasmaEndSparks(state.To)
	}
	return rayFXPacketSize48EA70
}
