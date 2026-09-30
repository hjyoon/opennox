package legacy

/*
#include "map_mode_setup_417ea0.h"
*/
import "C"

type mapModeSetupServer417EA0 interface {
	MapInfoSetCapflag417EA0() int
	MapInfoSetFlags417EC0() bool
	MapInfoSetFlagball417F30() int
	MapInfoSetKotr4180D0() int
}

//export nox_xxx_mapInfoSetCapflag_417EA0
func nox_xxx_mapInfoSetCapflag_417EA0() C.int {
	return C.int(GetServer().(mapModeSetupServer417EA0).MapInfoSetCapflag417EA0())
}

func mapInfoSetFlagsExportCall417EC0() bool {
	return bool(C.sub_417EC0())
}

//export sub_417EC0
func sub_417EC0() C.bool {
	return C.bool(GetServer().(mapModeSetupServer417EA0).MapInfoSetFlags417EC0())
}

//export nox_xxx_mapInfoSetFlagball_417F30
func nox_xxx_mapInfoSetFlagball_417F30() C.char {
	return C.char(GetServer().(mapModeSetupServer417EA0).MapInfoSetFlagball417F30())
}

//export nox_xxx_mapInfoSetKotr_4180D0
func nox_xxx_mapInfoSetKotr_4180D0() C.int {
	return C.int(GetServer().(mapModeSetupServer417EA0).MapInfoSetKotr4180D0())
}
