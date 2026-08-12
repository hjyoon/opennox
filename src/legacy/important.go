package legacy

/*
#include "GAME3_3.h"

extern uint32_t dword_5d4594_1565512;
extern uint32_t dword_5d4594_1565516;
*/
import "C"

import "unsafe"

func importantAllocClassC() unsafe.Pointer {
	return unsafe.Pointer(C.nox_server_getImportantAllocClass_4E4DE0())
}

func setImportantAllocClassC(p unsafe.Pointer) {
	C.nox_server_setImportantAllocClass_4E4DE0((*C.nox_alloc_class)(p))
}

func importantListHeadsC() (first, last uint32) {
	return uint32(C.dword_5d4594_1565512), uint32(C.dword_5d4594_1565516)
}

func setImportantListHeadsC(first, last uint32) {
	C.dword_5d4594_1565512 = C.uint32_t(first)
	C.dword_5d4594_1565516 = C.uint32_t(last)
}

func importantPacketSizeC() uintptr {
	return uintptr(C.sizeof_nox_important_packet_t)
}

func importantNativeListC() (first, last unsafe.Pointer) {
	return unsafe.Pointer(C.nox_server_getImportantFirst_4E4F80()),
		unsafe.Pointer(C.nox_server_getImportantLast_4E4F80())
}

func setImportantNativeListC(first, last unsafe.Pointer) {
	C.nox_server_setImportantFirst_4E4F80((*C.nox_important_packet_t)(first))
	C.nox_server_setImportantLast_4E4F80((*C.nox_important_packet_t)(last))
}

func importantPacketNextC(packet unsafe.Pointer) unsafe.Pointer {
	return unsafe.Pointer(C.nox_server_getImportantNext_4E4F80((*C.nox_important_packet_t)(packet)))
}

func importantPacketPrevC(packet unsafe.Pointer) unsafe.Pointer {
	return unsafe.Pointer(C.nox_server_getImportantPrev_4E4F80((*C.nox_important_packet_t)(packet)))
}

func setImportantPacketNextC(packet, next unsafe.Pointer) {
	C.nox_server_setImportantNext_4E4F80((*C.nox_important_packet_t)(packet), (*C.nox_important_packet_t)(next))
}

func setImportantPacketPrevC(packet, prev unsafe.Pointer) {
	C.nox_server_setImportantPrev_4E4F80((*C.nox_important_packet_t)(packet), (*C.nox_important_packet_t)(prev))
}

func cleanupImportantPacketsC() int {
	return int(C.sub_4E4F80())
}

func removeImportantPacketC(packet unsafe.Pointer) {
	C.sub_4E4FC0((*C.nox_important_packet_t)(packet))
}

func updateImportantRateControlC(ind int) int32 {
	return int32(C.sub_4E4E50(C.int(ind)))
}

func resetImportantPlayerCounterC(ind int) int {
	return int(C.sub_4E4F30(C.int(ind)))
}
