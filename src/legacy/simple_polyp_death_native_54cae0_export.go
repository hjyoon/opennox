package legacy

/*
#include <stdint.h>
#include "defs.h"
#include "GAME1_1.h"
#include "GAME3_2.h"
#include "GAME3_3.h"
#include "GAME4.h"
#include "GAME5.h"
static uint32_t polyp_raw_resolve_54cb10(void) {return (uint32_t)nox_xxx_getNameId_4E3AA0("ToxicCloud");}
static double polyp_raw_balance_54cb10(void) {return nox_xxx_gamedataGetFloat_419D40("ToxicCloudLifetime");}
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/server"
)

func simplePolypRawRuntime54CAE0() server.SimpleObjectDeathRuntime54CAE0 {
	if simplePolypOriginalSimpleRuntimeOverride54CAE0 != nil {
		return *simplePolypOriginalSimpleRuntimeOverride54CAE0
	}
	return server.SimpleObjectDeathRuntime54CAE0{
		Audio: func(id uint32, source *server.Object) {
			C.nox_xxx_aud_501960(C.int(int32(id)), asObjectC(source), 0, 0)
		},
		DelayedDelete: func(source *server.Object) {
			C.nox_xxx_delayedDeleteObject_4E5CC0(asObjectC(source))
		},
	}
}

func polypRawState54CB10() *server.PolypDeathState54CB10 {
	if simplePolypOriginalStateOverride54CB10 != nil {
		return simplePolypOriginalStateOverride54CB10
	}
	return (*server.PolypDeathState54CB10)(unsafe.Pointer(memmap.PtrUint32(0x5D4594, 2491672)))
}

func polypRawRuntime54CB10() server.PolypDeathRawRuntime54CB10 {
	if simplePolypOriginalRuntimeOverride54CB10 != nil {
		return *simplePolypOriginalRuntimeOverride54CB10
	}
	return server.PolypDeathRawRuntime54CB10{
		SimpleObjectDeathRuntime54CAE0: simplePolypRawRuntime54CAE0(),
		Resolve:                        func() uint32 { return uint32(C.polyp_raw_resolve_54cb10()) },
		New: func(id uint32) *server.Object {
			return asObjectS(C.nox_xxx_newObjectWithTypeInd_4E3450(C.int(int32(id))))
		},
		Create: func(cloud, owner *server.Object, position types.Pointf, extra uint32) {
			C.nox_xxx_createAt_4DAA50(asObjectC(cloud), asObjectC(owner), C.float(position.X), C.float(position.Y))
		},
		Balance: func() float64 { return float64(C.polyp_raw_balance_54cb10()) },
		FPS:     func() uint32 { return uint32(C.gameFPS()) },
	}
}

//export nox_xxx_dieImpEgg_54CAE0
func nox_xxx_dieImpEgg_54CAE0(source *nox_object_t) C.int {
	return C.int(server.ImpEggDieNative54CAE0(asObjectS(source), simplePolypRawRuntime54CAE0()))
}

//export nox_xxx_diePolyp_54CB10
func nox_xxx_diePolyp_54CB10(source *nox_object_t) {
	asObjectS(source).PolypDeathNative54CB10(polypRawState54CB10(), polypRawRuntime54CB10())
}

//export nox_xxx_diePotion_54CBB0
func nox_xxx_diePotion_54CBB0(source *nox_object_t) {
	server.PotionDieNative54CBB0(asObjectS(source), simplePolypRawRuntime54CAE0())
}
