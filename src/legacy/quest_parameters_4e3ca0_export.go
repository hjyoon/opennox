package legacy

/*
#include "GAME3_3.h"
*/
import "C"

import (
	"math"

	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/server"
)

const (
	questParametersBase = 0x587000
	questDifficulty     = 202024
	questStage          = 202028
	questDamageScale    = 202032
	questHealthScale    = 202036

	questDifficultyCacheBase  = 0x5D4594
	questDifficultyDelta      = 1563912
	questDifficultyDeltaReady = 1563928
)

func questParameterFloat(offset uintptr) C.double {
	bits := memmap.Uint32(questParametersBase, offset)
	// The x87 load preserves the sign and payload, including the quiet bit.
	if bits&0x7f800000 == 0x7f800000 {
		wide := uint64(bits&0x80000000)<<32 | 0x7ff0000000000000 | uint64(bits&0x007fffff)<<29
		return C.double(math.Float64frombits(wide))
	}
	return C.double(math.Float32frombits(bits))
}

func questParameterStore(value float64) float32 {
	bits := math.Float64bits(value)
	payload := uint32(bits>>29) & 0x007fffff
	// A legacy balance value has already been read as a binary32.
	if bits&0x7ff0000000000000 == 0x7ff0000000000000 && payload != 0 {
		return math.Float32frombits(uint32(bits>>32)&0x80000000 | 0x7f800000 | payload)
	}
	return float32(value)
}

//export sub_4E3CA0
func sub_4E3CA0() C.double { return questParameterFloat(questDifficulty) }

//export sub_4E3CB0
func sub_4E3CB0(value C.float) C.int {
	bits := math.Float32bits(float32(value))
	*memmap.PtrUint32(questParametersBase, questDifficulty) = bits
	return C.int(int32(bits))
}

//export nox_game_getQuestStage_4E3CC0
func nox_game_getQuestStage_4E3CC0() C.int {
	return C.int(memmap.Int32(questParametersBase, questStage))
}

//export nox_game_setQuestStage_4E3CD0
func nox_game_setQuestStage_4E3CD0(stage C.int) {
	*memmap.PtrInt32(questParametersBase, questStage) = int32(stage)
}

//export sub_4E3D50
func sub_4E3D50() C.int {
	ready := memmap.PtrUint32(questDifficultyCacheBase, questDifficultyDeltaReady)
	delta := memmap.PtrFloat32(questDifficultyCacheBase, questDifficultyDelta)
	if *ready == 0 {
		*delta = questParameterStore(GetServer().S().Balance.Float("PlayerDifficultyDelta"))
		*ready = 1
	}
	players := int32(nox_xxx_player_4E3CE0())
	stage := uint32(nox_game_getQuestStage_4E3CC0())
	return sub_4E3CB0(C.float(server.QuestDifficulty4E3D50(stage, players, *delta)))
}

func setQuestScale(offset uintptr, value float32, key string) {
	cap := GetServer().S().Balance.Float(key)
	// FCOMP stores the input bits for less, equal and unordered operands.
	if float64(value) > cap {
		value = questParameterStore(cap)
	}
	*memmap.PtrFloat32(questParametersBase, offset) = value
}

//export sub_4E4080
func sub_4E4080(value C.float) { setQuestScale(questDamageScale, float32(value), "PlayerDamageCap") }

//export sub_4E40B0
func sub_4E40B0() C.double { return questParameterFloat(questDamageScale) }

//export sub_4E40C0
func sub_4E40C0(value C.float) { setQuestScale(questHealthScale, float32(value), "SystemHealthCap") }

//export sub_4E40F0
func sub_4E40F0() C.double { return questParameterFloat(questHealthScale) }

//export sub_4E4DC0
func sub_4E4DC0() { nox_platform_srand(0x1429) }

//export sub_4E4DD0
func sub_4E4DD0() { nox_platform_srand(0x490) }

//export nox_xxx_noop_4E5AB0
func nox_xxx_noop_4E5AB0(_ C.nox_important_payload_t) {}

//export sub_4E5AC0
func sub_4E5AC0() { nox_platform_srand(0x13d11) }
