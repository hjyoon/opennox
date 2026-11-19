package legacy

/*
#include "GAME3_2.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

//export nox_xxx_scavengerHuntReport_4D8CD0
func nox_xxx_scavengerHuntReport_4D8CD0(raw *nox_object_t) C.intptr_t {
	owner, result := server.ScavengerHuntReportResult4D8CD0(asObjectS(raw),
		func(recipient int, packet []byte, related *server.Object, remove int) int {
			return GetServer().S().NetSendPacketXxx0(recipient, packet, related, remove)
		})
	if owner != nil {
		return C.intptr_t(uintptr(unsafe.Pointer(owner)))
	}
	return C.intptr_t(result)
}

//export nox_xxx_playerIncrementElimDeath_4D8D40
func nox_xxx_playerIncrementElimDeath_4D8D40(raw *nox_object_t) {
	GetServer().PlayerIncrementElimDeath4D8D40(asObjectS(raw))
}
