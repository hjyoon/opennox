package legacy

/*
#include "monster_fight_run_state_native_531e20.h"
_Static_assert(sizeof(int) == 4, "fight/run status remains a signed DWORD");
#define NOX_FIGHT_RUN_ABI(fn) _Static_assert(_Generic(&(fn), int (*)(nox_object_t*): 1, default: 0), #fn " native unit/status ABI")
NOX_FIGHT_RUN_ABI(nox_xxx_mobActionFightStart_531E20);
NOX_FIGHT_RUN_ABI(sub_531E90);
NOX_FIGHT_RUN_ABI(sub_534750);
NOX_FIGHT_RUN_ABI(sub_534780);
NOX_FIGHT_RUN_ABI(sub_544740);
NOX_FIGHT_RUN_ABI(sub_544750);
NOX_FIGHT_RUN_ABI(nox_xxx_mobActionReturnToHome_544920);
NOX_FIGHT_RUN_ABI(sub_544930);
NOX_FIGHT_RUN_ABI(sub_544940);
NOX_FIGHT_RUN_ABI(sub_5456B0);
NOX_FIGHT_RUN_ABI(sub_5456C0);
NOX_FIGHT_RUN_ABI(sub_546410);
NOX_FIGHT_RUN_ABI(sub_546420);
*/
import "C"
import "unsafe"

//export nox_xxx_mobActionFightStart_531E20
func nox_xxx_mobActionFightStart_531E20(unit *C.nox_object_t) C.int {
	return C.int(int32(GetServer().S().MonsterActionFightStart531E20(asObjectS((*nox_object_t)(unit)), monsterFightStartRuntime531E20())))
}

//export sub_531E90
func sub_531E90(unit *C.nox_object_t) C.int {
	return C.int(int32(GetServer().S().MonsterActionFightEnd531E90(asObjectS((*nox_object_t)(unit)))))
}

//export sub_534750
func sub_534750(unit *C.nox_object_t) C.int {
	return C.int(int32(GetServer().S().MonsterActionRunStart534750(asObjectS((*nox_object_t)(unit)))))
}

//export sub_534780
func sub_534780(unit *C.nox_object_t) C.int {
	return C.int(int32(GetServer().S().MonsterActionRunEnd534780(asObjectS((*nox_object_t)(unit)))))
}

//export sub_544740
func sub_544740(unit *C.nox_object_t) C.int { return sub_534750(unit) }

//export sub_544750
func sub_544750(unit *C.nox_object_t) C.int { return sub_534780(unit) }

//export nox_xxx_mobActionReturnToHome_544920
func nox_xxx_mobActionReturnToHome_544920(unit *C.nox_object_t) C.int {
	return sub_534750(unit)
}

//export sub_544930
func sub_544930(unit *C.nox_object_t) C.int { return sub_534780(unit) }

//export sub_544940
func sub_544940(unit *C.nox_object_t) C.int { return sub_534780(unit) }

//export sub_5456B0
func sub_5456B0(unit *C.nox_object_t) C.int { return sub_534750(unit) }

//export sub_5456C0
func sub_5456C0(unit *C.nox_object_t) C.int { return sub_534780(unit) }

//export sub_546410
func sub_546410(unit *C.nox_object_t) C.int { return sub_534780(unit) }

//export sub_546420
func sub_546420(unit *C.nox_object_t) C.int { return sub_534780(unit) }

func monsterFightRunStatePublicC531E20(entry uint32) unsafe.Pointer {
	switch entry {
	case 0x531e20:
		return C.nox_xxx_mobActionFightStart_531E20
	case 0x531e90:
		return C.sub_531E90
	case 0x534750:
		return C.sub_534750
	case 0x534780:
		return C.sub_534780
	case 0x544740:
		return C.sub_544740
	case 0x544750:
		return C.sub_544750
	case 0x544920:
		return C.nox_xxx_mobActionReturnToHome_544920
	case 0x544930:
		return C.sub_544930
	case 0x544940:
		return C.sub_544940
	case 0x5456b0:
		return C.sub_5456B0
	case 0x5456c0:
		return C.sub_5456C0
	case 0x546410:
		return C.sub_546410
	case 0x546420:
		return C.sub_546420
	default:
		return nil
	}
}
