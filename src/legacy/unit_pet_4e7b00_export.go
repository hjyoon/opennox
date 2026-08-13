package legacy

/*
#include "GAME3_3.h"
*/
import "C"

import "github.com/opennox/opennox/v1/server"

func unitBecomePetRuntime4E7B00(owner, pet *server.Object) {
	s := GetServer().S()
	unitBecomePetNative4E7B00(owner, pet, unitPetRuntimeDeps4E7B00(s))
}

func unitBecomeEnemyRuntime4E7B60(owner, pet *server.Object) {
	s := GetServer().S()
	unitBecomeEnemyNative4E7B60(owner, pet, unitPetRuntimeDeps4E7B00(s))
}

//export nox_xxx_unitBecomePet_4E7B00
func nox_xxx_unitBecomePet_4E7B00(owner, pet *nox_object_t) {
	unitBecomePetRuntime4E7B00(asObjectS(owner), asObjectS(pet))
}

//export nox_xxx_monsterRemoveMonitors_4E7B60
func nox_xxx_monsterRemoveMonitors_4E7B60(owner, pet *nox_object_t) {
	unitBecomeEnemyRuntime4E7B60(asObjectS(owner), asObjectS(pet))
}
