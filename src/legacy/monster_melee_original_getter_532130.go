package legacy

/*
#include "defs.h"
#include "GAME4_3.h"
*/
import "C"
import "unsafe"

func meleePublicCBefore532130(entry uint32) unsafe.Pointer {
	switch entry {
	case 0x532130:
		return C.nox_xxx_mobActionMelee1_532130
	case 0x532390:
		return C.sub_532390
	case 0x532440:
		return C.nox_xxx_mobActionMeleeAtt_532440
	default:
		return nil
	}
}
