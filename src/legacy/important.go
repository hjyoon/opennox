package legacy

/*
#include "GAME3_3.h"

extern uint32_t dword_5d4594_1565512;
extern uint32_t dword_5d4594_1565516;
extern uint32_t dword_5d4594_1565520;
extern uint32_t dword_5d4594_2649712;

static int nox_important_test_kick_count;
static int nox_important_test_kick_player;

static void nox_important_test_capture_kick(int player_index) {
	nox_important_test_kick_count++;
	nox_important_test_kick_player = player_index;
}

static void nox_important_test_begin_kick_capture(void) {
	nox_important_test_kick_count = 0;
	nox_important_test_kick_player = -1;
	nox_server_setImportantKickHandler_4E52B0(nox_important_test_capture_kick);
}

static void nox_important_test_end_kick_capture(void) {
	nox_server_setImportantKickHandler_4E52B0(NULL);
}

static int nox_important_test_get_kick_count(void) {
	return nox_important_test_kick_count;
}

static int nox_important_test_get_kick_player(void) {
	return nox_important_test_kick_player;
}
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

func importantAllocClassC() unsafe.Pointer {
	return unsafe.Pointer(C.nox_server_getImportantAllocClass_4E4DE0())
}

func setImportantAllocClassC(p unsafe.Pointer) {
	C.nox_server_setImportantAllocClass_4E4DE0((*C.nox_alloc_class)(p))
}

func importantListHeadsC() (first, last uint32) {
	return uint32(C.dword_5d4594_1565512), uint32(C.dword_5d4594_1565516)
}

func importantRecipientMaskC() uint32 {
	return uint32(C.dword_5d4594_2649712)
}

func importantCapacityC() uint32 {
	return uint32(C.dword_5d4594_1565520)
}

func setImportantCapacityC(capacity uint32) {
	C.dword_5d4594_1565520 = C.uint32_t(capacity)
}

func setImportantRecipientMaskC(mask uint32) {
	C.dword_5d4594_2649712 = C.uint32_t(mask)
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

func importantPacketRelatedObjectC(packet unsafe.Pointer) unsafe.Pointer {
	return unsafe.Pointer(C.nox_server_getImportantRelatedObject_4E5030((*C.nox_important_packet_t)(packet)))
}

func setImportantPacketRelatedObjectC(packet, object unsafe.Pointer) {
	C.nox_server_setImportantRelatedObject_4E5030(
		(*C.nox_important_packet_t)(packet), (*C.nox_object_t)(object),
	)
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

func checkImportantRateC() int {
	return int(C.nox_xxx_importantCheckRate_4E52B0())
}

func beginImportantKickCaptureC() {
	C.nox_important_test_begin_kick_capture()
}

func importantKickCaptureC() (count, player int) {
	return int(C.nox_important_test_get_kick_count()), int(C.nox_important_test_get_kick_player())
}

func endImportantKickCaptureC() {
	C.nox_important_test_end_kick_capture()
}

func sendImportantPacketC(recipient int, payload []byte, relatedObject unsafe.Pointer, removeIfDisconnected int, sequenceEnabled bool) int {
	var data unsafe.Pointer
	if len(payload) != 0 {
		buf, free := alloc.CloneSlice(payload)
		defer free()
		data = unsafe.Pointer(&buf[0])
	}
	var sequence C.char
	if sequenceEnabled {
		sequence = 1
	}
	return int(C.nox_xxx_netSendPacket_4E5030(
		C.int(recipient), data, C.int(len(payload)), (*C.nox_object_t)(relatedObject),
		C.int(removeIfDisconnected), sequence,
	))
}

func updateImportantRateControlC(ind int) int32 {
	return int32(C.sub_4E4E50(C.int(ind)))
}

func resetImportantPlayerCounterC(ind int) int {
	return int(C.sub_4E4F30(C.int(ind)))
}
