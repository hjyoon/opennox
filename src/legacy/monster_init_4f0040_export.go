package legacy

/*
#include "monster_init_4f0040.h"
*/
import "C"

//export nox_xxx_unitMonsterInit_4F0040
func nox_xxx_unitMonsterInit_4F0040(unit *C.nox_object_t) {
	monsterInitCall4F0040(asObjectS((*nox_object_t)(unit)))
}
