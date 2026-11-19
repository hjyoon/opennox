package legacy

/*
#include "GAME3_2.h"
*/
import "C"

import "github.com/opennox/opennox/v1/server"

//export nox_xxx_projectileReflect_4E0A70
func nox_xxx_projectileReflect_4E0A70(projectile, other *nox_object_t) *nox_object_t {
	server.SpellProjectileReflect4E0A70(asObjectS(projectile), asObjectS(other))
	return projectile
}
