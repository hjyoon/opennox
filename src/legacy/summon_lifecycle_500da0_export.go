package legacy

/*
#include "GAME1_1.h"
#include "summon_lifecycle_500da0.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/libs/types"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/server"
)

const summonEffectIDOffset500DA0 = uintptr(1570276)

func summonRuntime500DA0() server.SummonRuntime500DA0 {
	return server.SummonRuntime500DA0{
		GameFlags: func(mask uint32) int32 {
			if noxflags.HasGame(noxflags.GameFlag(mask)) {
				return 1
			}
			return 0
		},
		GuideSize: func(guide int32) int32 {
			return int32(C.nox_xxx_guideGetUnitSize_427460(C.int(guide)))
		},
		CheckLimit:           Nox_xxx_checkSummonedCreaturesLimit_500D70,
		MapTileAllowTeleport: mapTileAllowTeleport411A90,
		SummonAt: func(typeID int, position types.Pointf, owner *server.Object, direction server.Dir16) *server.Object {
			return Nox_xxx_unitDoSummonAt_5016C0(int32(typeID), &position, owner, uint8(direction))
		},
		LoadEffectID: func() uint16 {
			return memmap.Uint16(0x5D4594, summonEffectIDOffset500DA0)
		},
		StoreEffectID: func(value uint16) {
			*memmap.PtrUint16(0x5D4594, summonEffectIDOffset500DA0) = value
		},
		FloatToInt: func(value float32) int32 {
			return int32(C.nox_float2int(C.float(value)))
		},
	}
}

var (
	summonStartCall500DA0 = func(record *server.DurSpell) int32 {
		return GetServer().S().SummonStart500DA0(record, summonRuntime500DA0())
	}
	summonFinishCall5010D0 = func(record *server.DurSpell) int32 {
		return GetServer().S().SummonFinish5010D0(record, summonRuntime500DA0())
	}
	summonCancelCall5011C0 = func(record *server.DurSpell) {
		GetServer().S().SummonCancel5011C0(record)
	}
)

func summonStartExportCall500DA0(record *server.DurSpell) int32 {
	return int32(C.nox_xxx_summonStart_500DA0(
		(*C.nox_dur_spell_t)(unsafe.Pointer(record)),
	))
}

func summonFinishExportCall5010D0(record *server.DurSpell) int32 {
	return int32(C.nox_xxx_summonFinish_5010D0(
		(*C.nox_dur_spell_t)(unsafe.Pointer(record)),
	))
}

func summonCancelExportCall5011C0(record *server.DurSpell) {
	C.nox_xxx_summonCancel_5011C0((*C.nox_dur_spell_t)(unsafe.Pointer(record)))
}

//export nox_xxx_summonStart_500DA0
func nox_xxx_summonStart_500DA0(record *C.nox_dur_spell_t) C.int32_t {
	return C.int32_t(summonStartCall500DA0((*server.DurSpell)(unsafe.Pointer(record))))
}

//export nox_xxx_summonFinish_5010D0
func nox_xxx_summonFinish_5010D0(record *C.nox_dur_spell_t) C.int32_t {
	return C.int32_t(summonFinishCall5010D0((*server.DurSpell)(unsafe.Pointer(record))))
}

//export nox_xxx_summonCancel_5011C0
func nox_xxx_summonCancel_5011C0(record *C.nox_dur_spell_t) {
	summonCancelCall5011C0((*server.DurSpell)(unsafe.Pointer(record)))
}
