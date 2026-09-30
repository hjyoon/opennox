package legacy

/*
#include "quest_briefing_450a30.h"
#include "GAME1_2.h"
#include "GAME2.h"
#include "common__strman.h"
extern uint32_t dword_5d4594_832480;
*/
import "C"

import (
	"encoding/binary"
	"unsafe"

	"github.com/opennox/opennox/v1/common/memmap"
)

func questStartBriefingStatePtr450A30() *uint32 {
	return (*uint32)(unsafe.Pointer(&C.dword_5d4594_832480))
}

func questStartBriefingNativeHooks450A30() questStartBriefingHooks450A30[*byte, unsafe.Pointer, unsafe.Pointer] {
	return questStartBriefingHooks450A30[*byte, unsafe.Pointer, unsafe.Pointer]{
		storeState:     func(v uint32) { *questStartBriefingStatePtr450A30() = v },
		resetParticles: func() { Nox_client_resetScreenParticles_431510() },
		hideBook:       func(v int32) int32 { return int32(Nox_xxx_bookHideMB_45ACA0(int(v))) },
		prepare:        Sub_446780,
		offset:         func(p *byte, off uintptr) *byte { return (*byte)(unsafe.Add(unsafe.Pointer(p), off)) },
		loadImage: func(name *byte) unsafe.Pointer {
			return unsafe.Pointer(C.nox_xxx_gLoadImg_42F970((*C.char)(unsafe.Pointer(name))))
		},
		setImage: func(image unsafe.Pointer) { C.sub_450AD0((*C.nox_video_bag_image_t)(image)) },
		// Preserve the original unbounded NUL scan, not a capped 32-byte field.
		stringLength: func(key *byte) uintptr { return uintptr(C.strlen((*C.char)(unsafe.Pointer(key)))) },
		loadText: func(key *byte, source string, line int32) unsafe.Pointer {
			return unsafe.Pointer(C.nox_strman_loadString_40F1D0(
				(*C.char)(unsafe.Pointer(key)), nil, internCStr(source), C.int(line),
			))
		},
		// This is the address of an inline empty UTF-16 string, not a pointer slot.
		emptyText: func() unsafe.Pointer { return memmap.Ptr(0x5D4594 + 832548) },
		setText:   func(text unsafe.Pointer) { C.sub_450AF0((*C.wchar2_t)(text)) },
		stage: func(packet *byte) uint16 {
			return binary.LittleEndian.Uint16(unsafe.Slice((*byte)(unsafe.Add(unsafe.Pointer(packet), 2)), 2))
		},
		setStage: func(stage uint32) { C.nox_gui_setQuestStage_450B00(C.int(stage)) },
		lock: func(screen, mode int32, flags int8) int32 {
			return int32(C.nox_client_lockScreenBriefing_450160(C.int(screen), C.int(mode), C.char(flags)))
		},
	}
}

var questStartBriefingCall450A30 = func(packet *byte, show int32) int32 {
	return questStartBriefing450A30(packet, show, questStartBriefingNativeHooks450A30())
}

//export nox_client_showQuestBriefing_native_450A30
func nox_client_showQuestBriefing_native_450A30(packet *C.uchar, show C.int) C.int {
	return C.int(questStartBriefingCall450A30((*byte)(unsafe.Pointer(packet)), int32(show)))
}

func questStartBriefingCEntry450A30(packet *byte, show int32) int32 {
	return int32(C.nox_client_showQuestBriefing_450A30((*C.uchar)(unsafe.Pointer(packet)), C.int(show)))
}
