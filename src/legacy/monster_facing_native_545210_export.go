package legacy

/*
#include "monster_facing_native_545210.h"

_Static_assert(_Generic(&sub_545210, char (*)(nox_object_t*): 1, default: 0), "native face location ABI");
_Static_assert(_Generic(&sub_545300, char (*)(nox_object_t*): 1, default: 0), "native face object ABI");
_Static_assert(_Generic(&sub_545340, char (*)(nox_object_t*): 1, default: 0), "native face angle ABI");
_Static_assert(_Generic(&sub_5453E0, char (*)(nox_object_t*): 1, default: 0), "native set angle ABI");
_Static_assert(_Generic(&nox_ai_action_pop_532100, char (*)(nox_object_t*): 1, default: 0), "native facing cancel ABI");
*/
import "C"

//export sub_545210
func sub_545210(unit *C.nox_object_t) C.char {
	return C.char(GetServer().S().MonsterActionFaceLocation545210(asObjectS((*nox_object_t)(unit))))
}

//export sub_545300
func sub_545300(unit *C.nox_object_t) C.char {
	return C.char(GetServer().S().MonsterActionFaceObject545300(asObjectS((*nox_object_t)(unit))))
}

//export sub_545340
func sub_545340(unit *C.nox_object_t) C.char {
	return C.char(GetServer().S().MonsterActionFaceAngle545340(asObjectS((*nox_object_t)(unit))))
}

//export sub_5453E0
func sub_5453E0(unit *C.nox_object_t) C.char {
	return C.char(GetServer().S().MonsterActionSetAngle5453E0(asObjectS((*nox_object_t)(unit))))
}

//export nox_ai_action_pop_532100
func nox_ai_action_pop_532100(unit *C.nox_object_t) C.char {
	return C.char(asObjectS((*nox_object_t)(unit)).MonsterPopAction())
}
