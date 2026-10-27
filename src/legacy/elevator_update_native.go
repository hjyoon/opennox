package legacy

/*
#include "GAME4_3.h"
*/
import "C"

import "github.com/opennox/opennox/v1/server"

func elevatorRuntimeNative53B5D0() server.ElevatorUpdateRuntime53B5D0 {
	return server.ElevatorUpdateRuntime53B5D0{
		Move:           Nox_xxx_unitMove_4E7010,
		QueueCollision: Nox_xxx_unitHasCollideOrUpdateFn_537610,
	}
}

//export nox_xxx_updateElevator_53B5D0_go
func nox_xxx_updateElevator_53B5D0_go(unit *nox_object_t) {
	GetServer().S().ElevatorUpdate53B5D0(asObjectS(unit), elevatorRuntimeNative53B5D0())
}

//export nox_xxx_updateElevatorShaft_53B380_go
func nox_xxx_updateElevatorShaft_53B380_go(unit *nox_object_t) C.char {
	return C.char(GetServer().S().ElevatorShaftUpdate53B380(asObjectS(unit), elevatorRuntimeNative53B5D0()))
}

//export nox_xxx_elevatorFn_53B750_go
func nox_xxx_elevatorFn_53B750_go(candidate, elevator *nox_object_t) {
	GetServer().S().ElevatorPassengerUp53B750(asObjectS(candidate), asObjectS(elevator), elevatorRuntimeNative53B5D0())
}

//export nox_xxx_fnElevatorShaft_53B410_go
func nox_xxx_fnElevatorShaft_53B410_go(candidate, shaft *nox_object_t) {
	GetServer().S().ElevatorPassengerDown53B410(asObjectS(candidate), asObjectS(shaft), elevatorRuntimeNative53B5D0())
}

//export nox_xxx_elevatorAud_53B490_go
func nox_xxx_elevatorAud_53B490_go(unit *nox_object_t, up C.int) {
	GetServer().S().ElevatorAudio53B490(asObjectS(unit), up != 0)
}
