package legacy

/*
#include <stdint.h>
#include "defs.h"
#include "GAME1_1.h"
#include "GAME3_2.h"
#include "GAME3_3.h"
#include "GAME4.h"
#include "GAME5.h"
static uint32_t troll_death_resolve_54a270(void) { return (uint32_t)nox_xxx_getNameId_4E3AA0("SmallToxicCloud"); }
static double troll_death_balance_54a270(void) { return nox_xxx_gamedataGetFloat_419D40("SmallToxicCloudLifetime"); }
*/
import "C"

import (
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/server"
	"unsafe"
)

func trollDeathState54A270() *server.TrollDeathState54A270 {
	if trollDeathOriginalStateOverride54A270 != nil {
		return trollDeathOriginalStateOverride54A270
	}
	return (*server.TrollDeathState54A270)(unsafe.Pointer(memmap.PtrUint32(0x5D4594, 2491584)))
}

func trollDeathRuntime54A270() server.TrollDeathRuntime54A270 {
	if trollDeathOriginalRuntimeOverride54A270 != nil {
		return *trollDeathOriginalRuntimeOverride54A270
	}
	return server.TrollDeathRuntime54A270{
		Resolve: func() uint32 { return uint32(C.troll_death_resolve_54a270()) },
		New: func(id uint32) *server.Object {
			return asObjectS(C.nox_xxx_newObjectWithTypeInd_4E3450(C.int(int32(id))))
		},
		Create: func(o, owner *server.Object, pos types.Pointf, extra uint32) uint32 {
			C.nox_xxx_createAt_4DAA50(asObjectC(o), asObjectC(owner), C.float(pos.X), C.float(pos.Y))
			return 0
		},
		Audio: func(id uint32, o *server.Object, kind, arg uint32) uint32 {
			C.nox_xxx_aud_501960(C.int(int32(id)), asObjectC(o), C.int(int32(kind)), C.int(int32(arg)))
			return 0
		},
		Balance: func() float64 { return float64(C.troll_death_balance_54a270()) },
		FPS:     func() uint32 { return uint32(C.gameFPS()) },
	}
}

//export nox_xxx_monsterDeadTroll_54A270
func nox_xxx_monsterDeadTroll_54A270(unit *nox_object_t) C.int {
	return C.int(asObjectS(unit).TrollDeathNative54A270(trollDeathState54A270(), trollDeathRuntime54A270()))
}
