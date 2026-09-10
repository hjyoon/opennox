package legacy

/*
#include "GAME1_1.h"
#include "charm_lifecycle_5011f0.h"
*/
import "C"

import (
	"unsafe"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/server"
)

func charmRuntime5011F0() server.CharmRuntime5011F0 {
	return server.CharmRuntime5011F0{
		GameFlag: func(mask uint32) bool {
			return noxflags.HasGame(noxflags.GameFlag(mask))
		},
		GuideSize: func(guide int32) int32 {
			return int32(C.nox_xxx_guideGetUnitSize_427460(C.int(guide)))
		},
		Charmable: func(typeIndex uint16) int32 {
			return int32(C.nox_xxx_creatureIsCharmableByTT_4272B0(C.int(typeIndex)))
		},
		FloatToInt: func(value float32) int32 {
			return int32(C.nox_float2int(C.float(value)))
		},
		Distance:          objectDistance_4E6C00,
		CheckLimit:        Nox_xxx_checkSummonedCreaturesLimit_500D70,
		BuffApplyRuntime:  buffApplyRuntime4FF380(),
		BuffOffRuntime:    spellBuffOffRuntime4FF5B0(),
		Attribution:       recordPlayerAttributionRuntime4E7540,
		UnitOrderRuntime:  unitOrderRuntime533900(),
		ChangeTeam:        Nox_xxx_netChangeTeamMb_419570,
		SetHP:             Nox_xxx_unitSetHP_4E4560,
		QuestSpawnCleanup: charmQuestSpawnCleanupRuntime5013E0,
	}
}

var (
	charmStartCall5011F0 = func(record *server.DurSpell) int32 {
		return GetServer().S().CharmStart5011F0(record, charmRuntime5011F0())
	}
	charmFinishCall5013E0 = func(record *server.DurSpell) int32 {
		return GetServer().S().CharmFinish5013E0(record, charmRuntime5011F0())
	}
	charmCancelCall501690 = func(record *server.DurSpell) int32 {
		return GetServer().S().CharmCancel501690(record, charmRuntime5011F0())
	}
)

func charmStartExportCall5011F0(record *server.DurSpell) int32 {
	return int32(C.nox_xxx_charmCreature1_5011F0(
		(*C.nox_dur_spell_t)(unsafe.Pointer(record)),
	))
}

func charmFinishExportCall5013E0(record *server.DurSpell) int32 {
	return int32(C.nox_xxx_charmCreatureFinish_5013E0(
		(*C.nox_dur_spell_t)(unsafe.Pointer(record)),
	))
}

func charmCancelExportCall501690(record *server.DurSpell) int32 {
	return int32(C.nox_xxx_charmCreature2_501690(
		(*C.nox_dur_spell_t)(unsafe.Pointer(record)),
	))
}

//export nox_xxx_charmCreature1_5011F0
func nox_xxx_charmCreature1_5011F0(record *C.nox_dur_spell_t) C.int32_t {
	return C.int32_t(charmStartCall5011F0((*server.DurSpell)(unsafe.Pointer(record))))
}

//export nox_xxx_charmCreatureFinish_5013E0
func nox_xxx_charmCreatureFinish_5013E0(record *C.nox_dur_spell_t) C.int32_t {
	return C.int32_t(charmFinishCall5013E0((*server.DurSpell)(unsafe.Pointer(record))))
}

//export nox_xxx_charmCreature2_501690
func nox_xxx_charmCreature2_501690(record *C.nox_dur_spell_t) C.int32_t {
	return C.int32_t(charmCancelCall501690((*server.DurSpell)(unsafe.Pointer(record))))
}
