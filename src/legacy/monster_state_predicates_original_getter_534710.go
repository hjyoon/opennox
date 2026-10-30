package legacy

/*
#include "defs.h"
#include "GAME3_3.h"
#include "GAME4_3.h"
static unsigned nox_state_original_swap_cache(unsigned value) {
    unsigned* cache = getMemU32Ptr(0x5D4594, 2488524);
    unsigned old = *cache; *cache = value; return old;
}
static void* nox_state_original_cache_ptr(void) {return getMemU32Ptr(0x5D4594, 2488524);}
static int nox_state_original_lookup_control(void) {return nox_xxx_getNameId_4E3AA0("Mimic");}
*/
import "C"
import "unsafe"

func statePredicateOriginalPublicC534710(entry uint32) unsafe.Pointer {
	switch entry {
	case 0x534710:
		return C.sub_534710
	case 0x5347a0:
		return C.sub_5347A0
	case 0x5347c0:
		return C.sub_5347C0
	case 0x5347f0:
		return C.nox_xxx_isNotPoisoned_5347F0
	case 0x534810:
		return C.nox_xxx_mobGetMoveAttemptTime_534810
	case 0x534840:
		return C.nox_xxx_unitIsMimic_534840
	default:
		return nil
	}
}
func statePredicateOriginalSwapMimicCache534840(value uint32) uint32 {
	return uint32(C.nox_state_original_swap_cache(C.uint(value)))
}
func statePredicateOriginalMimicCachePointer534840() unsafe.Pointer {
	return C.nox_state_original_cache_ptr()
}
func statePredicateOriginalNameIDControl534840() uint32 {
	return uint32(C.nox_state_original_lookup_control())
}
