package legacy

/*
#include "GAME3_2.h"
extern uint32_t dword_5d4594_1548476;
extern uint32_t dword_5d4594_1548480;
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/common/memmap"
)

func questMapFileGlobals4D0F60() (count, clock *uint32) {
	return (*uint32)(unsafe.Pointer(&C.dword_5d4594_1548476)),
		(*uint32)(unsafe.Pointer(&C.dword_5d4594_1548480))
}

func questMapFileCEntry4D0F60() unsafe.Pointer {
	return unsafe.Pointer(C.nox_xxx_getQuestMapFile_4D0F60())
}

func questMapEntryNative4D0F60(index int32) *questMapEntry4D0F60 {
	offset := uint32(1525132) + (uint32(index) << 5)
	return (*questMapEntry4D0F60)(memmap.PtrOff(0x5D4594, uintptr(offset)))
}

func questMapNameNative4D0F60(index int32) unsafe.Pointer {
	offset := uint32(1525136) + (uint32(index) << 5)
	return memmap.PtrOff(0x5D4594, uintptr(offset))
}

var questMapFileCall4D0F60 = func() unsafe.Pointer {
	count, clock := questMapFileGlobals4D0F60()
	return questMapFile4D0F60(questMapFileHooks4D0F60{
		Count:     func() uint32 { return *count },
		LastIndex: func() uint32 { return memmap.Uint32(0x587000, 191880) },
		Clock:     func() uint32 { return *clock },
		Entry:     questMapEntryNative4D0F60,
		Name:      questMapNameNative4D0F60,
		Random: func(min, max int32) int32 {
			return nox_common_randomInt_415FA0(min, max)
		},
	})
}

//export nox_server_questMapFile_native_4D0F60
func nox_server_questMapFile_native_4D0F60() *C.char {
	return (*C.char)(questMapFileCall4D0F60())
}
