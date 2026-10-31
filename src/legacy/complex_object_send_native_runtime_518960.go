package legacy

/*
#include "defs.h"
#include "GAME1_1.h"
#include "GAME3_2.h"
#include "GAME4.h"
#include "GAME4_1.h"
#include "GAME4_3.h"
#include "GAME5_2.h"
#include "common__random.h"
#include "common__net_list.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

var complexObjectSendRuntimeNative518960 = func() server.ComplexObjectSendNativeRuntime518960 {
	return server.ComplexObjectSendNativeRuntime518960{
		Frame: func() uint32 { return GetServer().S().Frame() },
		ReportHealth: func(player int32, code uint16, delta int16) {
			C.nox_xxx_netReportHealthDelta_4D8760(C.int(player), C.short(code), C.short(delta))
		},
		UnitCode: func(unit *server.Object) uint32 {
			return uint32(C.nox_xxx_netGetUnitCodeServ_578AC0(asObjectC(unit)))
		},
		FloatToInt: func(value float32) int32 { return int32(C.nox_float2int(C.float(value))) },
		Animation: func(unit *server.Object) int32 {
			return int32(C.nox_xxx_mobActionToAnimation_533790(asObjectC(unit)))
		},
		Random: func(min, max int32) int32 {
			return int32(C.nox_common_randomInt_415FA0(C.int(min), C.int(max)))
		},
		Frames: func(index int32) (int32, int32) {
			var frames, duration C.int
			C.nox_xxx_animPlayerGetFrameRange_4F9F90(C.int(index), &frames, &duration)
			return int32(frames), int32(duration)
		},
		Direction: func(index int32, out *server.IndexedDirectionVector509E20) {
			var value C.int2
			C.nox_xxx_xferIndexedDirection_509E20(C.int(index), &value)
			out.X, out.Y = int32(value.field_0), int32(value.field_4)
		},
		Enqueue: func(player int32, packet [11]byte) int32 {
			if C.nox_netlist_addToMsgListSrv_40EF40(C.int(player), (*C.uchar)(unsafe.Pointer(&packet[0])), C.int(len(packet))) {
				return 1
			}
			return 0
		},
	}
}

// NetSendComplexObject518960 is the ordinary sender and the public C entry's
// shared native implementation. Dependencies remain lazy until reached.
func NetSendComplexObject518960(player int32, unit *server.Object, report int32) int32 {
	return server.ComplexObjectSendNative518960(player, unit, report, complexObjectSendRuntimeNative518960())
}
