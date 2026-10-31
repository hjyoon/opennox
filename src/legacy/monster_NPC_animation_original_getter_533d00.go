package legacy

/*
#include "defs.h"
#include "GAME4.h"
#include "GAME4_3.h"
*/
import "C"

import "unsafe"

func NPCAnimationPublicC533D00() unsafe.Pointer { return C.nox_xxx_unitNPCActionToAnim_533D00 }
func NPCWeaponAnimationActualC533D00(value uint32) int32 { return int32(C.sub_4FA280(C.int(value))) }
func NPCFramesActualC533D00(index int32) (int32, int32) {
	var frames, duration C.int
	C.nox_xxx_animPlayerGetFrameRange_4F9F90(C.int(index), &frames, &duration)
	return int32(frames), int32(duration)
}
