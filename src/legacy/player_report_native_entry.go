package legacy

/*
#include "GAME3_2.h"
*/
import "C"

//export nox_xxx_playerReportAnything_4D9900
func nox_xxx_playerReportAnything_4D9900(unit *nox_object_t) {
	obj := asObjectS(unit)
	if obj == nil || uint8(obj.ObjClass)&0x04 == 0 {
		return
	}
	playerObjectSendRuntimeNative518C30().Self(obj)
}

//export sub_4D9DF0
func sub_4D9DF0(recipient C.int, unit *nox_object_t, presence C.char) C.int {
	return C.int(GetServer().S().QuestKeyReport4D9DF0(int32(recipient), asObjectS(unit), 0, uint8(presence)))
}

//export sub_4D9E30
func sub_4D9E30(recipient C.int, unit *nox_object_t, presence C.char) C.int {
	return C.int(GetServer().S().QuestKeyReport4D9DF0(int32(recipient), asObjectS(unit), 1, uint8(presence)))
}

//export nox_xxx_netFxShield_0_4D9200
func nox_xxx_netFxShield_0_4D9200(recipient C.int, unit *nox_object_t) C.int {
	return C.int(GetServer().S().Nox_xxx_netFxShield_0_4D9200(int(recipient), asObjectS(unit)))
}
