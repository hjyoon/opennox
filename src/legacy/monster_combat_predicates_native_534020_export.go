package legacy

/*
#include "defs.h"
#include "GAME4_3.h"
_Static_assert(sizeof(int) == 4, "combat predicates retain DWORD results");
#define NOX_COMBAT_PREDICATE_ABI(fn) _Static_assert(_Generic(&(fn), int (*)(nox_object_t*): 1, default: 0), #fn " native object/DWORD ABI")
NOX_COMBAT_PREDICATE_ABI(sub_534020);
NOX_COMBAT_PREDICATE_ABI(nox_xxx_monsterCanMelee_534220);
NOX_COMBAT_PREDICATE_ABI(nox_xxx_monsterCanShoot_534280);
NOX_COMBAT_PREDICATE_ABI(nox_xxx_monsterHasShield_5342C0);
NOX_COMBAT_PREDICATE_ABI(nox_xxx_monsterCanCast_534300);
NOX_COMBAT_PREDICATE_ABI(nox_xxx_monsterIsMoveing_534320);
NOX_COMBAT_PREDICATE_ABI(sub_534340);
NOX_COMBAT_PREDICATE_ABI(nox_xxx_monsterCanAttackAtWill_534390);
NOX_COMBAT_PREDICATE_ABI(sub_5343C0);
NOX_COMBAT_PREDICATE_ABI(sub_534400);
NOX_COMBAT_PREDICATE_ABI(sub_534440);
_Static_assert(_Generic(&sub_534120, int (*)(nox_object_t*, float2*): 1, default: 0), "native facing vector ABI");
_Static_assert(_Generic(&sub_534470, double (*)(nox_object_t*): 1, default: 0), "native melee range ABI");
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/server"
)

func monsterPredicateResultC534020(value bool) C.int {
	if value {
		return 1
	}
	return 0
}

//export sub_534020
func sub_534020(unit *C.nox_object_t) C.int {
	return monsterPredicateResultC534020(server.MonsterFlying534020(asObjectS((*nox_object_t)(unit))))
}

//export sub_534120
func sub_534120(unit *C.nox_object_t, vector *C.float2) C.int {
	if vector == nil {
		return 0
	}
	return monsterPredicateResultC534020(server.MonsterFacingDirection534120(
		asObjectS((*nox_object_t)(unit)), *(*types.Pointf)(unsafe.Pointer(vector))))
}

//export nox_xxx_monsterCanMelee_534220
func nox_xxx_monsterCanMelee_534220(unit *C.nox_object_t) C.int {
	return monsterPredicateResultC534020(server.MonsterCanMelee534220(asObjectS((*nox_object_t)(unit))))
}

//export nox_xxx_monsterCanShoot_534280
func nox_xxx_monsterCanShoot_534280(unit *C.nox_object_t) C.int {
	return monsterPredicateResultC534020(server.MonsterCanShoot534280(asObjectS((*nox_object_t)(unit))))
}

//export nox_xxx_monsterHasShield_5342C0
func nox_xxx_monsterHasShield_5342C0(unit *C.nox_object_t) C.int {
	return monsterPredicateResultC534020(server.MonsterHasShield5342C0(asObjectS((*nox_object_t)(unit))))
}

//export nox_xxx_monsterCanCast_534300
func nox_xxx_monsterCanCast_534300(unit *C.nox_object_t) C.int {
	return monsterPredicateResultC534020(server.MonsterCanCast534300(asObjectS((*nox_object_t)(unit))))
}

//export nox_xxx_monsterIsMoveing_534320
func nox_xxx_monsterIsMoveing_534320(unit *C.nox_object_t) C.int {
	return monsterPredicateResultC534020(server.MonsterIsMoving534320(asObjectS((*nox_object_t)(unit))))
}

//export sub_534340
func sub_534340(unit *C.nox_object_t) C.int {
	return monsterPredicateResultC534020(server.MonsterHasInterruptibleAction534340(asObjectS((*nox_object_t)(unit))))
}

//export nox_xxx_monsterCanAttackAtWill_534390
func nox_xxx_monsterCanAttackAtWill_534390(unit *C.nox_object_t) C.int {
	object := asObjectS((*nox_object_t)(unit))
	return monsterPredicateResultC534020(object != nil && object.UpdateData != nil && object.Nox_xxx_monsterCanAttackAtWill_534390())
}

//export sub_5343C0
func sub_5343C0(unit *C.nox_object_t) C.int {
	object := asObjectS((*nox_object_t)(unit))
	return monsterPredicateResultC534020(object != nil && object.UpdateData != nil && object.Sub_5343C0())
}

//export sub_534400
func sub_534400(unit *C.nox_object_t) C.int {
	return monsterPredicateResultC534020(server.MonsterLowAggression534400(asObjectS((*nox_object_t)(unit))))
}

//export sub_534440
func sub_534440(unit *C.nox_object_t) C.int {
	object := asObjectS((*nox_object_t)(unit))
	return monsterPredicateResultC534020(object != nil && object.UpdateData != nil && object.Sub_534440())
}

//export sub_534470
func sub_534470(unit *C.nox_object_t) C.double {
	return C.double(server.MonsterMeleeRange534470(asObjectS((*nox_object_t)(unit))))
}
