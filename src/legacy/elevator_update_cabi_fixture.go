package legacy

/*
#include "GAME4_3.h"
*/
import "C"

import "github.com/opennox/opennox/v1/server"

func elevatorUpdateCABI53B5D0(unit *server.Object) { C.nox_xxx_updateElevator_53B5D0(asObjectC(unit)) }
func elevatorShaftUpdateCABI53B380(unit *server.Object) uint8 {
	return uint8(C.nox_xxx_updateElevatorShaft_53B380(asObjectC(unit)))
}
func elevatorPassengerUpCABI53B750(candidate, owner *server.Object) {
	C.nox_xxx_elevatorFn_53B750(asObjectC(candidate), asObjectC(owner))
}
func elevatorPassengerDownCABI53B410(candidate, owner *server.Object) {
	C.nox_xxx_fnElevatorShaft_53B410(asObjectC(candidate), asObjectC(owner))
}
func elevatorAudioCABI53B490(owner *server.Object, direction int32) {
	C.nox_xxx_elevatorAud_53B490(asObjectC(owner), C.int(direction))
}
