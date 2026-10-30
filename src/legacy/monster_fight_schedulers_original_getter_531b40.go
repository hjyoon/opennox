package legacy

/*
#include "defs.h"
#include "GAME4_3.h"
*/
import "C"
import "unsafe"

func fightSchedulerPublicCBefore531B40(entry uint32) unsafe.Pointer {
	switch entry {
	case 0x531b40:
		return C.nox_xxx_TODOsomeCallingMeleeAttack_531B40
	case 0x531c10:
		return C.sub_531C10
	case 0x531c60:
		return C.nox_xxx_monsterAction_531C60
	case 0x531d50:
		return C.sub_531D50
	default:
		return nil
	}
}
