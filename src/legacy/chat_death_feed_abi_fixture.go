package legacy

/*
#include "GAME2_3.h"
*/
import "C"

import "unsafe"

type chatBubbleCandidateRecord = C.nox_chat_bubble

func chatBubbleCandidateCallC(direction byte, moving, earlier *chatBubbleCandidateRecord, x, y *int32) uintptr {
	return uintptr(C.sub_48E6A0(C.char(direction), (*C.uint32_t)(unsafe.Pointer(moving)), (*C.uint32_t)(unsafe.Pointer(earlier)),
		(*C.int)(unsafe.Pointer(x)), (*C.int)(unsafe.Pointer(y))))
}

func deathFeedEntryCallC(entry *deathFeedEntry) uintptr {
	return uintptr(C.sub_495500((*C.int)(unsafe.Pointer(entry))))
}
