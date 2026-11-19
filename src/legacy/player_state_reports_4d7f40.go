package legacy

/*
#include "GAME3_2.h"
*/
import "C"

//export nox_xxx_netReportObjectPoison_4D7F40
func nox_xxx_netReportObjectPoison_4D7F40(receiver, object *nox_object_t, poison C.char) C.int {
	s := GetServer().S()
	return C.int(s.Nox_xxx_netReportObjectPoison_4D7F40(asObjectS(receiver), asObjectS(object), int8(poison)))
}

//export sub_4D81A0
func sub_4D81A0(player *nox_object_t) {
	GetServer().S().NetReportExperience(asObjectS(player))
}

//export nox_xxx_netReportAnimFrame_4D81F0
func nox_xxx_netReportAnimFrame_4D81F0(recipient C.int, object *nox_object_t) C.int {
	return C.int(GetServer().S().Nox_xxx_netReportAnimFrame_4D81F0(int(recipient), asObjectS(object)))
}

//export nox_xxx_netReportXStatus_4D8230
func nox_xxx_netReportXStatus_4D8230(recipient C.int, object *nox_object_t) C.int {
	return C.int(GetServer().S().Nox_xxx_netReportXStatus_4D8230(int(recipient), asObjectS(object)))
}
