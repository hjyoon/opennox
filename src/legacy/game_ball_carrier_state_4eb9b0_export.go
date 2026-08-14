package legacy

/*
#include "game_ball_carrier_state_4eb9b0.h"
*/
import "C"

//export sub_4EB9B0
func sub_4EB9B0(ball, target *C.nox_object_t) *C.nox_object_t {
	return (*C.nox_object_t)(asObjectC(GetServer().S().GameBallCarrierState4EB9B0(
		asObjectS((*nox_object_t)(ball)),
		asObjectS((*nox_object_t)(target)),
	)))
}
