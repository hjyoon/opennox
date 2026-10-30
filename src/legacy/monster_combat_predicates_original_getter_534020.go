package legacy

/*
#include "defs.h"
#include "GAME4_3.h"
static double nox_combat_predicate_call_double(void* fnc, void* unit) { return ((double (*)(void*))(fnc))(unit); }
*/
import "C"
import (
	"github.com/opennox/opennox/v1/server"
	"unsafe"
)

func combatPredicateOriginalPublicC534020(entry uint32) unsafe.Pointer {
	switch entry {
	case 0x534020:
		return C.sub_534020
	case 0x534120:
		return C.sub_534120
	case 0x534220:
		return C.nox_xxx_monsterCanMelee_534220
	case 0x534280:
		return C.nox_xxx_monsterCanShoot_534280
	case 0x5342c0:
		return C.nox_xxx_monsterHasShield_5342C0
	case 0x534300:
		return C.nox_xxx_monsterCanCast_534300
	case 0x534320:
		return C.nox_xxx_monsterIsMoveing_534320
	case 0x534340:
		return C.sub_534340
	case 0x534390:
		return C.nox_xxx_monsterCanAttackAtWill_534390
	case 0x5343c0:
		return C.sub_5343C0
	case 0x534400:
		return C.sub_534400
	case 0x534440:
		return C.sub_534440
	case 0x534470:
		return C.sub_534470
	default:
		return nil
	}
}

func combatPredicateOriginalDouble534470(fnc unsafe.Pointer, unit *server.Object) float64 {
	return float64(C.nox_combat_predicate_call_double(fnc, unsafe.Pointer(unit)))
}
