package legacy

/*
#include "defs.h"
#include "GAME4_3.h"
static inline void nox_zombie_probe_warm_534A10(void) {
 *getMemU32Ptr(0x5D4594,2488528)=100;
 *getMemU32Ptr(0x5D4594,2488532)=100;
 *getMemU32Ptr(0x5D4594,2488536)=200;
}
static inline uint32_t nox_zombie_probe_call_534A10(int kind,nox_object_t* unit) {
 switch(kind) {
 case 0:return nox_xxx_unitIsPlant_534A10(unit);
 case 1:return nox_xxx_unitIsZombie_534A40(unit);
 case 2:return (uint8_t)nox_xxx_mobActionGetUp_534A90(unit);
 default:return nox_xxx_mobRaiseZombie_534AB0(unit);
 }
}
*/
import "C"

import "github.com/opennox/opennox/v1/server"

func zombieEntryCProbe534A10(kind int, unit *server.Object) uint32 {
	C.nox_zombie_probe_warm_534A10()
	return uint32(C.nox_zombie_probe_call_534A10(C.int(kind), asObjectC(unit)))
}
