package legacy

/*
#include "defs.h"
#include "GAME5.h"
#include "map_push_units_52e040.h"
_Static_assert(_Generic(&sub_549D80,int (*)(nox_object_t*):1,default:0),"549D80 native public ABI");
_Static_assert(_Generic(&sub_549E00,int (*)(nox_object_t*):1,default:0),"549E00 native public ABI");
_Static_assert(_Generic(&sub_549E70,int (*)(nox_object_t*):1,default:0),"549E70 native public ABI");
_Static_assert(_Generic(&sub_54A250,int (*)(nox_object_t*):1,default:0),"54A250 native public ABI");
static void monster_death_typed_push_control_549d80(nox_object_t* unit,float outer,float inner,float force) {
 nox_xxx_mapPushUnitsAround_52E040(&unit->x,outer,inner,force,unit,0,0);
}
*/
import "C"

import "github.com/opennox/opennox/v1/server"

// MonsterDeathExplosionProductionC549D80 enters the real public C symbol with
// its default runtime. Integration tests use the ordinary root server wiring.
func MonsterDeathExplosionProductionC549D80(entry uint32, unit *server.Object) int32 {
	if monsterDeathOriginalRuntimeOverride549D80 != nil {
		panic("production death entry cannot use an overridden runtime")
	}
	return monsterDeathOriginalC549D80(entry, unit)
}

// MonsterDeathExplosionPushControl549D80 independently enters the preexisting
// typed C world service; it does not call the new death runtime adapter.
func MonsterDeathExplosionPushControl549D80(unit *server.Object, outer, inner, force float32) {
	C.monster_death_typed_push_control_549d80(asObjectC(unit), C.float(outer), C.float(inner), C.float(force))
}
