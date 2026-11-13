package legacy

/*
#include "defs.h"
*/
import "C"

// Public ABI names share the restored Go handlers. No legacy body executes here.

//export sub_5485B0
func sub_5485B0(a1 *nox_object_t, a2 *nox_object_t) C.int { return sub_5485B0_go(a1, a2) }

//export sub_5488B0
func sub_5488B0(a1 *nox_object_t, a2 *nox_object_t, a3 C.int) { sub_5488B0_go(a1, a2, a3) }

//export sub_54A7D0
func sub_54A7D0(unit *nox_object_t) C.int { return nox_monster_die_callback_native(1, unit) }

//export sub_54A850
func sub_54A850(unit *nox_object_t) C.int { return nox_monster_die_callback_native(2, unit) }

//export sub_54A890
func sub_54A890(unit *nox_object_t) C.int { return nox_monster_die_callback_native(3, unit) }

//export sub_54A900
func sub_54A900(unit *nox_object_t) C.int { return nox_monster_die_callback_native(4, unit) }

//export sub_54A950
func sub_54A950(unit *nox_object_t) C.int { return nox_monster_die_callback_native(5, unit) }

//export sub_54A990
func sub_54A990(center *C.float2, radius C.float, box *nox_object_t, normal *C.float2) C.double {
	return sub_54A990_go(center, radius, box, normal)
}

//export sub_54AD50
func sub_54AD50(circle *nox_object_t, box *nox_object_t, move_box C.int) {
	sub_54AD50_go(circle, box, move_box)
}

//export sub_54CBD0
func sub_54CBD0(unit *nox_object_t) { nox_server_questPenalty_native_54CBD0(unit) }

//export sub_54CC40
func sub_54CC40(unit *nox_object_t) { nox_server_questLoseWeapon_native_54CC40(unit) }

//export sub_54CD30
func sub_54CD30(unit *nox_object_t) { nox_server_questLoseArmor_native_54CD30(unit) }

//export sub_54CE00
func sub_54CE00(unit *nox_object_t) { nox_server_questLoseSpell_native_54CE00(unit) }

//export sub_54CEE0
func sub_54CEE0(unit *nox_object_t) { nox_server_questLoseGuide_native_54CEE0(unit) }

//export sub_54CFB0
func sub_54CFB0(unit *nox_object_t) C.char { return nox_server_questLoseAbility_native_54CFB0(unit) }

//export sub_54D080
func sub_54D080(unit *nox_object_t) { nox_server_questLoseGems_native_54D080(unit) }

//export nox_xxx_collideTrigger_54FCD0
func nox_xxx_collideTrigger_54FCD0(trigger *nox_object_t, candidate *nox_object_t) {
	nox_xxx_collideTrigger_54FCD0_go(trigger, candidate)
}

//export sub_54FEF0
func sub_54FEF0(obj *nox_object_t) { sub_54FEF0_go(obj) }

//export sub_54FFC0
func sub_54FFC0(a1 *C.int2, obj *nox_object_t) C.int { return sub_54FFC0_go(a1, obj) }

//export sub_550280
func sub_550280(a1 *C.float2, a2 C.float, a3 C.float, a4 C.int, a5 C.int, a6 *C.float2, a7 *C.float2) C.int {
	return sub_550280_go(a1, a2, a3, a4, a5, a6, a7)
}

//export sub_5502F0
func sub_5502F0(a1 *C.float2, a2 C.float, a3 C.float, a4 C.int, a5 C.int, a6 *C.float2, a7 *C.float2) C.int {
	return sub_5502F0_go(a1, a2, a3, a4, a5, a6, a7)
}

//export sub_550380
func sub_550380(a1 C.int, obj *nox_object_t, a3 *C.float2) C.int { return sub_550380_go(a1, obj, a3) }

//export sub_5504B0
func sub_5504B0(obj *nox_object_t) { sub_5504B0_go(obj) }

//export nox_xxx_collisionCheckCircleCircle_550D00
func nox_xxx_collisionCheckCircleCircle_550D00(moving *nox_object_t, candidate *nox_object_t) {
	nox_xxx_collisionCheckCircleCircle_550D00_go(moving, candidate)
}

//export sub_550F80
func sub_550F80(first *nox_object_t, second *nox_object_t) { sub_550F80_go(first, second) }

//export sub_551250
func sub_551250(door *nox_object_t, box *nox_object_t, move_box C.int) {
	sub_551250_go(door, box, move_box)
}

//export sub_551AE0
func sub_551AE0(elevator *nox_object_t, candidate *nox_object_t, candidate_moves C.int) {
	sub_551AE0_go(elevator, candidate, candidate_moves)
}

//export sub_551C40
func sub_551C40(shaft *nox_object_t, candidate *nox_object_t) { sub_551C40_go(shaft, candidate) }
