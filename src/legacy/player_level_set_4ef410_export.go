package legacy

/*
#include "player_level_set_4ef410.h"
*/
import "C"

//export sub_4EF410
func sub_4EF410(unit *C.nox_object_t, level C.uint8_t) {
	playerLevelSetCall4EF410(
		asObjectS((*nox_object_t)(unit)),
		uint8(level),
	)
}
