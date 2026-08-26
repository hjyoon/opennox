package legacy

/*
#include "GAME5.h"
*/
import "C"

//export sub_551C40_go
func sub_551C40_go(shaft, candidate *nox_object_t) {
	GetServer().S().ElevatorShaftCollision551C40(asObjectS(shaft), asObjectS(candidate))
}
