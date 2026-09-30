package legacy

/*
#include "quest_briefing_450980.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/common/memmap"
)

func questStageBriefingNativeHooks450980() questStageBriefingHooks450980[*byte, unsafe.Pointer, unsafe.Pointer] {
	h := questStageBriefingHooks450980[*byte, unsafe.Pointer, unsafe.Pointer]{
		questStartBriefingHooks450A30: questStartBriefingNativeHooks450A30(),
		flags: func(packet *byte) uint8 {
			return *(*byte)(unsafe.Add(unsafe.Pointer(packet), 4))
		},
	}
	// Inline empty UTF-16 data, not a pointer slot and not 450A30's +832548.
	h.emptyText = func() unsafe.Pointer { return memmap.Ptr(0x5D4594 + 832544) }
	return h
}

var questStageBriefingCall450980 = func(packet *byte, show int32) int32 {
	return questStageBriefing450980(packet, show, questStageBriefingNativeHooks450980())
}

//export nox_client_showQuestBriefing2_native_450980
func nox_client_showQuestBriefing2_native_450980(packet *C.uchar, show C.int) C.int {
	return C.int(questStageBriefingCall450980((*byte)(unsafe.Pointer(packet)), int32(show)))
}

func questStageBriefingCEntry450980(packet *byte, show int32) int32 {
	return int32(C.nox_client_showQuestBriefing2_450980((*C.uchar)(unsafe.Pointer(packet)), C.int(show)))
}
