package legacy

/*
#include "damage_predicates_raw_4e1400.h"
*/
import "C"

import "unsafe"

func init() {
	damagePredicateRawAttackEngine4E1400 = damagePredicateRawAttackNative4E1400
	damagePredicateRawFriendlyEngine4E1470 = damagePredicateRawFriendlyNative4E1470
}

//export sub_4E1400
func sub_4E1400(source, weapon *C.nox_object_t) C.int32_t {
	return C.int32_t(damagePredicateRawAttackNative4E1400(unsafe.Pointer(source), unsafe.Pointer(weapon)))
}

//export sub_4E1470
func sub_4E1470(weapon *C.nox_object_t) C.int32_t {
	return C.int32_t(damagePredicateRawFriendlyNative4E1470(unsafe.Pointer(weapon)))
}

func damagePredicateRawAttackNative4E1400(source, weapon unsafe.Pointer) int32 {
	if weapon != nil {
		// The original caches weapon class, including the final bit-1 result.
		class := damagePredicateRawRead4E1400(weapon, 2).Word
		if class&0x1000 != 0 {
			if damagePredicateRawRead4E1400(weapon, 3).Word&0x047f0000 == 0 {
				return 1
			}
			use := damagePredicateRawRead4E1400(weapon, 4).Pointer
			if damagePredicateRawRead4E1400(use, 5).Word&2 != 0 {
				return 1
			}
		} else if class&0x1000000 != 0 && damagePredicateRawRead4E1400(weapon, 3).Word&0x047f00fe == 0 {
			return 1
		}
		return int32((class >> 1) & 1)
	}
	// A nil source still reaches the original class read and its fault domain.
	class := damagePredicateRawRead4E1400(source, 0).Word
	if class&4 != 0 {
		return 1
	}
	if class&2 != 0 && damagePredicateRawRead4E1400(source, 1).Word&0x10 != 0 {
		return 1
	}
	return 0
}

func damagePredicateRawFriendlyNative4E1470(weapon unsafe.Pointer) int32 {
	if weapon != nil && damagePredicateRawRead4E1400(weapon, 2).Word&0x1000000 != 0 &&
		damagePredicateRawRead4E1400(weapon, 3).Word&0x4000 != 0 {
		return 1
	}
	return 0
}
