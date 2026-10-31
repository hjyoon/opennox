package legacy

/*
#include "defs.h"
#include "GAME4_3.h"
static uintptr_t missileLeadPublicCNativeABI533080(nox_object_t* unit, nox_object_t* target, float speed, float2* out) {
 return ((uintptr_t (*)(nox_object_t*, nox_object_t*, float, float2*))(void*)nox_xxx_projAddVelocitySmth_533080)(unit, target, speed, out);
}
*/
import "C"
import "unsafe"

func missilePublicCBefore532540(entry uint32) unsafe.Pointer {
	switch entry {
	case 0x532540:
		return C.sub_532540
	case 0x532610:
		return C.nox_xxx_mobActionMissileAtt_532610
	default:
		return nil
	}
}

func missileLeadPublicCBefore533080(unit, target unsafe.Pointer, speed float32, out unsafe.Pointer) uintptr {
 return uintptr(C.missileLeadPublicCNativeABI533080((*C.nox_object_t)(unit), (*C.nox_object_t)(target), C.float(speed), (*C.float2)(out)))
}
