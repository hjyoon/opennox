package legacy

/*
#include "game2_packet_bridge.h"
*/
import "C"

import (
	"github.com/opennox/opennox/v1/common/ntype"
	"unsafe"
)

// The old C switch's three object cases carried DWORD-truncated drawables.
// Reuse the existing typed Go receiver, bounded to one message. Its full-loop
// API returns success, whereas the C switch returns consumed bytes.
//
//export nox_game2_decode_native_object
func nox_game2_decode_native_object(player, op C.int, data *C.uchar, size C.int) C.int {
	length := 0
	switch op {
	case 47:
		length = 9
	case 48:
		length = 11
	case 195:
		length = 12
	default:
		return -1
	}
	if data == nil || int(size) < length || byte(*data) != byte(op) || Nox_xxx_netOnPacketRecvCli_48EA70 == nil {
		return -1
	}
	if Nox_xxx_netOnPacketRecvCli_48EA70(ntype.PlayerInd(player), (*byte)(unsafe.Pointer(data)), length) != 1 {
		return -1
	}
	return C.int(length)
}
