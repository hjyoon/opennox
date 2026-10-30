package legacy

/*
#include "defs.h"
#include "GAME4_3.h"
_Static_assert(sizeof(int)==4, "state predicate results retain DWORD width");
_Static_assert(sizeof(((nox_object_t*)0)->typ_ind)==2, "Mimic retains WORD type IDs");
#define NOX_STATE_PREDICATE_ABI(fn) _Static_assert(_Generic(&(fn), int (*)(nox_object_t*):1, default:0), #fn " native object/DWORD ABI")
NOX_STATE_PREDICATE_ABI(sub_534710);
NOX_STATE_PREDICATE_ABI(sub_5347A0);
NOX_STATE_PREDICATE_ABI(sub_5347C0);
NOX_STATE_PREDICATE_ABI(nox_xxx_isNotPoisoned_5347F0);
NOX_STATE_PREDICATE_ABI(nox_xxx_mobGetMoveAttemptTime_534810);
NOX_STATE_PREDICATE_ABI(nox_xxx_unitIsMimic_534840);
*/
import "C"
import "github.com/opennox/opennox/v1/server"

//export sub_534710
func sub_534710(unit *C.nox_object_t) C.int {
	return monsterPredicateResultC534020(server.MonsterHoldPosition534710(asObjectS((*nox_object_t)(unit))))
}

//export sub_5347A0
func sub_5347A0(unit *C.nox_object_t) C.int {
	return monsterPredicateResultC534020(server.MonsterStatusBit95347A0(asObjectS((*nox_object_t)(unit))))
}

//export sub_5347C0
func sub_5347C0(unit *C.nox_object_t) C.int {
	return monsterPredicateResultC534020(server.UnitInjured5347C0(asObjectS((*nox_object_t)(unit))))
}

//export nox_xxx_isNotPoisoned_5347F0
func nox_xxx_isNotPoisoned_5347F0(unit *C.nox_object_t) C.int {
	return monsterPredicateResultC534020(server.UnitPoisoned5347F0(asObjectS((*nox_object_t)(unit))))
}

//export nox_xxx_mobGetMoveAttemptTime_534810
func nox_xxx_mobGetMoveAttemptTime_534810(unit *C.nox_object_t) C.int {
	return monsterPredicateResultC534020(GetServer().S().MonsterMoveAttemptRecent534810(asObjectS((*nox_object_t)(unit))))
}
