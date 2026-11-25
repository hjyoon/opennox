package legacy

/*
#include "game2_packet_bridge.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/libs/noxnet/netmsg"

	"github.com/opennox/opennox/v1/common/ntype"
)

// clientNativeObjectPacket gives the full Go receiver one object message.
// The full loop reports success while this entry reports consumed bytes.
func clientNativeObjectPacket(player ntype.PlayerInd, op netmsg.Op, data []byte) int {
	var size int
	switch op {
	case netmsg.MSG_SIMPLE_OBJ:
		size = 9
	case netmsg.MSG_COMPLEX_OBJ:
		size = 11
	case netmsg.MSG_PLAYER_OBJ:
		size = 12
	default:
		return -1
	}
	if len(data) < size || data[0] != byte(op) || Nox_xxx_netOnPacketRecvCli_48EA70 == nil {
		return -1
	}
	if Nox_xxx_netOnPacketRecvCli_48EA70(player, &data[0], size) != 1 {
		return -1
	}
	return size
}

//export nox_game2_decode_native_object
func nox_game2_decode_native_object(player, op C.int, data *C.uchar, size C.int) C.int {
	if op < 0 || op > 255 || data == nil || size < 0 {
		return -1
	}
	packet := unsafe.Slice((*byte)(unsafe.Pointer(data)), int(size))
	return C.int(clientNativeObjectPacket(ntype.PlayerInd(player), netmsg.Op(op), packet))
}

// DecodeNativeClientPacket enters the same typed handlers as the normal
// client receiver for callers of the legacy switch ABI.
var DecodeNativeClientPacket func(op netmsg.Op, data []byte) int
