package legacy

/*
#include "GAME3_2.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/libs/noxnet/netmsg"
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/common/ntype"
)

func sendEarthquakeFromNative(pos *types.Pointf, strength int) {
	GetServer().S().EarthquakeFromNative4D9110(pos, int32(strength))
}

//export sub_4D90E0
func sub_4D90E0(recipient C.int, offset C.char) C.int {
	return C.int(bool2int(GetServer().S().NetQueueFxJiggle(ntype.PlayerInd(recipient), int8(offset))))
}

//export nox_xxx_earthquakeSend_4D9110
func nox_xxx_earthquakeSend_4D9110(pos *C.float, strength C.int) C.int {
	sendEarthquakeFromNative((*types.Pointf)(unsafe.Pointer(pos)), int(strength))
	return 0
}

//export sub_4D9CF0
func sub_4D9CF0(recipient C.int) C.int {
	packet := [2]byte{byte(netmsg.MSG_GAUNTLET), 0}
	return sendImportantPacketBytes(recipient, packet[:], true, true)
}

//export nox_xxx_netReportAllLatency_4D3050
func nox_xxx_netReportAllLatency_4D3050() *nox_playerInfo {
	GetServer().S().NetReportAllLatency4D3050(func(ind ntype.PlayerInd) int32 {
		return sub_554240(int32(ind))
	})
	return nil
}
