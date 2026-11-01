package server

import (
	"encoding/binary"
	"math"
)

// UnitHeightRuntime4D9020 declares the original code and netlist boundaries.
type UnitHeightRuntime4D9020 struct {
	UnitCode func(*Object) uint32
	Submit   func(recipient, kind int32, packet []byte) int32
}

// ObjectHeightPacketNative4D9020 shares unchanged packet construction with
// the ordinary sender. Nil Object faults, as at the original public entry.
func ObjectHeightPacketNative4D9020(unit *Object, unitCode func(*Object) uint32) ([6]byte, int) {
	var packet [6]byte
	if byte(unit.Field5)&0x20 != 0 {
		packet[0] = 159
		binary.LittleEndian.PutUint16(packet[1:3], uint16(unit.NetCode))
		packet[3] = byte(x87TruncSignedQwordLow566DCC(float64(unit.ZVal)))
		packet[4] = byte(x87TruncSignedQwordLow566DCC(float64(unit.Field27)))
		packet[5] = byte(x87TruncSignedQwordLow566DCC(float64(math.Float32frombits(unit.Field29))))
		return packet, 6
	}
	// FCMP sets C0 for negative and unordered input. Cache that choice before
	// the code child, then reload Z after the child exactly as the root does.
	negative := !(unit.ZVal >= 0)
	packet[0] = 94
	if negative {
		packet[0] = 95
	}
	code := unitCode(unit)
	binary.LittleEndian.PutUint16(packet[1:3], uint16(code))
	z := float64(unit.ZVal)
	if negative {
		z = -z
	}
	packet[3] = byte(x87TruncSignedQwordLow566DCC(z))
	return packet, 4
}

// ObjectHeightReportNative4D9020 retains kind1 and the full transport result.
func ObjectHeightReportNative4D9020(recipient int32, unit *Object, runtime UnitHeightRuntime4D9020) int32 {
	packet, size := ObjectHeightPacketNative4D9020(unit, runtime.UnitCode)
	return runtime.Submit(recipient, 1, packet[:size])
}
