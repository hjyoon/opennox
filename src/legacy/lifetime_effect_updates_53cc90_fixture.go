package legacy

/*
#include "defs.h"
#include "GAME4_3.h"
static void nox_lifetime_effect_probe(int kind,nox_object_t* obj) {
 switch(kind) {
 case 0: ((void(*)(nox_object_t*))nox_xxx_updateSelfDestruct_53CC90)(obj);break;
 case 1: ((void(*)(nox_object_t*))nox_xxx_updateDeathBallFragment_53D220)(obj);break;
 case 2: ((void(*)(nox_object_t*))nox_xxx_updateMoonglow_53D270)(obj);break;
 case 3: ((void(*)(nox_object_t*))nox_xxx_updateTelekinesis_53D330)(obj);break;
 case 4: ((void(*)(nox_object_t*))nox_xxx_updateFist_53D400)(obj);break;
 case 5: ((void(*)(nox_object_t*))nox_xxx_updateFlameCleanse_53D510)(obj);break;
 }
}
*/
import "C"

import "github.com/opennox/opennox/v1/server"

func lifetimeEffectCProbe53CC90(kind server.LifetimeEffectKind53CC90, source *server.Object) {
	C.nox_lifetime_effect_probe(C.int(kind), asObjectC(source))
}
