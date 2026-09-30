package legacy

/*
#include "GAME4_1.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/server"
)

var questThemeCall51A1F0 = func(group int32) {
	outer := GetServer()
	outer.S().QuestTheme51A1F0(group, server.QuestThemeRuntime51A1F0{
		QuestStage:       func() uint32 { return uint32(Nox_game_getQuestStage_4E3CC0()) },
		HecubahType:      func() uint32 { return memmap.Uint32(0x5D4594, 2388668) },
		NecroType:        func() uint32 { return memmap.Uint32(0x5D4594, 2388672) },
		StoreHecubahType: func(id uint32) { *memmap.PtrUint32(0x5D4594, 2388668) = id },
		StoreNecroType:   func(id uint32) { *memmap.PtrUint32(0x5D4594, 2388672) = id },
		GeneratorType:    questGeneratorTypeCEntry51A500,
		DelayedDelete:    outer.DelayedDelete,
		SetMinions:       func(value int32) { C.sub_51A940(C.int(value)) },
		SpawnHecubah: func(position *types.Pointf) {
			C.nox_xxx_spawnHecubahQuest_51A5A0((*C.int)(unsafe.Pointer(position)))
		},
		SpawnNecro: func(position *types.Pointf) {
			C.nox_xxx_spawnNecroQuest_51A7A0((*C.int)(unsafe.Pointer(position)))
		},
	})
}

//export nox_xxx_questTheme_native_51A1F0
func nox_xxx_questTheme_native_51A1F0(group C.int) {
	questThemeCall51A1F0(int32(group))
}
