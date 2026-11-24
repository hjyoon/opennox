package legacy

/*
#include "GAME2_3.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/client"
)

//export nox_xxx_clientAddRayEffect_49C160
func nox_xxx_clientAddRayEffect_49C160(packet *C.uint8_t) *C.nox_drawable {
	if packet == nil || GetClient == nil {
		return nil
	}
	if cli := GetClient(); cli != nil {
		dr := cli.AddLegacyDurationRay(*(*[7]byte)(unsafe.Pointer(packet)))
		return (*C.nox_drawable)(dr.C())
	}
	return nil
}

//export nox_xxx_clientRemoveRayEffect_49C450
func nox_xxx_clientRemoveRayEffect_49C450(packet *C.uint8_t) {
	if packet == nil || GetClient == nil {
		return
	}
	if cli := GetClient(); cli != nil {
		cli.RemoveLegacyDurationRay(*(*[7]byte)(unsafe.Pointer(packet)))
	}
}

//export nox_xxx_spriteDeleteSomeList_49C4B0
func nox_xxx_spriteDeleteSomeList_49C4B0() {
	Nox_xxx_spriteDeleteSomeList_49C4B0()
}

func Nox_xxx_spriteDeleteSomeList_49C4B0() {
	count := clientTransientRayCount49BDD0()
	for i := range count {
		if addr := clientTransientRayAddress49BDD0(i); addr != 0 {
			GetClient().Nox_xxx_spriteDeleteStatic_45A4E0_drawable(AsDrawableP(unsafe.Pointer(addr)))
		}
	}
	clientTransientRayClear49BDD0()
	sub_4C5050()
}

//export nox_xxx_sprite_49C4F0
func nox_xxx_sprite_49C4F0() {
	if GetClient == nil {
		return
	}
	if cli := GetClient(); cli != nil {
		cli.ClearDurationRays()
	}
}

func ClientTransientRayContains(dr *client.Drawable) bool {
	return dr != nil && clientTransientRayContainsAddress49BDD0(uintptr(dr.C()))
}

//export sub_49C520
func sub_49C520(dr *C.nox_drawable) C.int {
	if dr == nil {
		return 0
	}
	drawable := asDrawable(dr)
	if GetClient != nil {
		if cli := GetClient(); cli != nil && cli.IsDurationRayDrawable(drawable) {
			return 1
		}
	}
	if ClientTransientRayContains(drawable) {
		return 1
	}
	return 0
}
