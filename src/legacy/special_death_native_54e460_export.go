package legacy

/*
#include <stdint.h>
#include "defs.h"
#include "GAME1.h"
#include "GAME1_1.h"
#include "GAME3_2.h"
#include "GAME3_3.h"
#include "GAME4.h"
#include "GAME4_2.h"
#include "GAME4_3.h"
#include "GAME5.h"
#include "server__script__script.h"
#include "native_death_calls.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/libs/types"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/server"
)

// BoulderDeathIndex54E4B0 is the actual extracted C cache also shared by the
// ordinary Boulder dispatcher. It is distinct from the Golem debris cache.
func BoulderDeathIndex54E4B0() *uint32 {
	return (*uint32)(unsafe.Pointer(&legacyAddressableWords.dword_5d4594_2491704))
}
func specialDeathState54E460() *server.GolemDeathState549E90 {
	if specialDeathOriginalStateOverride54E460 != nil {
		return specialDeathOriginalStateOverride54E460
	}
	return &server.GolemDeathState549E90{Index: BoulderDeathIndex54E4B0(), Divisor: memmap.PtrUint32(0x587000, 290340)}
}
func specialDeathRuntime54E460() server.SpecialDeathRuntime54E460 {
	if specialDeathOriginalRuntimeOverride54E460 != nil {
		return *specialDeathOriginalRuntimeOverride54E460
	}
	return server.SpecialDeathRuntime54E460{GolemDeathRuntime549E90: server.GolemDeathRuntime549E90{
		RandomInt: func(lo, hi int32, _ server.GolemRandomSite549E90) int32 {
			return nox_common_randomInt_415FA0(lo, hi)
		},
		RandomFloat: func(lo, hi float32) float64 {
			return GetServer().S().RandomFloat416030(lo, hi)
		},
		FPS: gameFPS,
		RockName: func(index int32) unsafe.Pointer {
			off := uint32(291512) + uint32(index)*4
			return *memmap.PtrPtr(0x587000, uintptr(off))
		},
		Mass: func(index int32) byte {
			off := uint32(290328) + uint32(index)
			return *memmap.PtrUint8(0x587000, uintptr(off))
		},
		New: func(name unsafe.Pointer) *server.Object {
			return asObjectS(nox_xxx_newObjectByTypeID_4E3810((*C.char)(name)))
		},
		Point: debrisPoint,
		Create: func(o, owner *server.Object, pos types.Pointf, _ uint32) uint32 {
			nox_xxx_createAt_4DAA50(asObjectC(o), asObjectC(owner), C.float(pos.X), C.float(pos.Y))
			return 0
		},
		Raise: func(o *server.Object, height float32) uint32 {
			o.Raise(height)
			return 0
		},
		Force: debrisForce,
		Audio: func(id uint32, o *server.Object, kind, arg uint32) uint32 {
			nox_xxx_aud_501960(int32(id), asObjectC(o), int32(kind), int32(arg))
			return 0
		},
		PointFX: debrisPointFX,
		Decay: func(o *server.Object, duration uint32) uint32 {
			return uint32(GetServer().S().DecaySetTime511660(o, duration))
		},
	},
		DestroyedName: unsafe.Pointer(internCStr("DestroyedGenerator")),
		Parent: func(unit *server.Object) *server.Object {
			return asObjectS(nox_xxx_findParentChainPlayer_4EC580(asObjectC(unit)))
		},
		Delete: func(unit *server.Object) uint32 { nox_xxx_delayedDeleteObject_4E5CC0(asObjectC(unit)); return 0 },
		Frame:  gameFrame,
		Timer:  func(frame uint32) uint32 { return uint32(sub_4D71E0(C.int(int32(frame)))) },
		Mode:   func(mode int32) uint32 { return uint32(sub_4D7520(C.int(mode))) },
		Script: func(data server.SpecialDeathCallback54E460, caller, trigger *server.Object, event uint32) uint32 {
			if callback := data.NativeCallback(); callback != nil {
				nox_xxx_scriptCallByEventBlock_502490(unsafe.Pointer(callback), unsafe.Pointer(caller), unsafe.Pointer(trigger), int32(event))
			} else {
				C.nox_legacy_death_script_at_address(C.uintptr_t(data.Address()), asObjectC(caller), asObjectC(trigger), C.int(int32(event)))
			}
			return 0
		},
		GeneratorFX: func(origin server.SkeletonOrigin54A310, intensity uint32) uint32 {
			if point := origin.NativePoint(); point != nil {
				server.SendGeneratorBreakFXNative523200(point, byte(intensity), pointFXRuntimeForCall522FF0())
			} else {
				C.nox_legacy_generator_fx_at_address(C.uintptr_t(origin.Address()), C.char(intensity))
			}
			return 0
		},
		Quest: func() int32 {
			if noxflags.HasGame(noxflags.GameModeQuest) {
				return 1
			}
			return 0
		},
		Credit: func(unit *server.Object) uint32 { sub_4D61B0(asObjectC(unit)); return 0 },
	}
}

//export nox_xxx_dieMarker_54E460
func nox_xxx_dieMarker_54E460(unit *nox_object_t) {
	asObjectS(unit).SpecialDeathNative54E460(0, specialDeathState54E460(), specialDeathRuntime54E460())
}

//export nox_xxx_dieBoulder_54E4B0
func nox_xxx_dieBoulder_54E4B0(unit *nox_object_t) {
	asObjectS(unit).SpecialDeathNative54E460(1, specialDeathState54E460(), specialDeathRuntime54E460())
}

//export nox_xxx_dieMonsterGen_54E630
func nox_xxx_dieMonsterGen_54E630(unit *nox_object_t) {
	asObjectS(unit).SpecialDeathNative54E460(2, specialDeathState54E460(), specialDeathRuntime54E460())
}
