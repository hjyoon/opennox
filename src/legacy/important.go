package legacy

/*
#include "GAME3_3.h"

extern uint32_t dword_5d4594_1565512;
extern uint32_t dword_5d4594_1565516;
extern uint32_t dword_5d4594_1565520;
extern uint32_t dword_5d4594_2649712;
*/
import "C"

import (
	"unsafe"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/common/ntype"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

type importantPacketC = C.nox_important_packet_t

var (
	importantPlayerByIndHook = func(ind ntype.PlayerInd) *server.Player {
		return GetServer().S().Players.ByInd(ind)
	}
	importantRateGetHook = func() uint32 {
		return memmap.Uint32(0x587000, 4728)
	}
	importantPlayerPacketCleanupHook = func(ind uint8) {
		C.sub_4E55F0(C.uint8_t(ind))
	}
	importantGameHostHook = func() bool {
		return noxflags.HasGame(noxflags.GameHost)
	}
	importantReplayReadHook = func() bool {
		return noxflags.HasEngine(noxflags.EngineReplayRead)
	}
	importantRateAdjustHook = func(ind uint8) {
		adjustImportantRateC(ind)
	}
)

const importantRateKickStatus = uint32(0x80)

//export nox_server_playerKickDueToRate_4E5360
func nox_server_playerKickDueToRate_4E5360(playerIndex C.int) C.int {
	ind := ntype.PlayerInd(playerIndex)
	player := importantPlayerByIndHook(ind)
	if player == nil {
		return 0
	}
	importantPlayerPacketCleanupHook(uint8(playerIndex))
	player.Field3680 |= importantRateKickStatus
	// The original calls 0x004174F0 with 0x80. That mask cannot enter its
	// 0x423 reporting branch, so its exact return is the GameHost check.
	return C.int(bool2int(importantGameHostHook()))
}

//export nox_server_importantPlayerLookup_4E5670
func nox_server_importantPlayerLookup_4E5670(playerIndex C.uint8_t) {
	importantPlayerByIndHook(ntype.PlayerInd(playerIndex))
}

//export nox_server_importantRateGet_4E5670
func nox_server_importantRateGet_4E5670() C.uint32_t {
	return C.uint32_t(importantRateGetHook())
}

//export nox_server_importantShouldProcess_4E5770
func nox_server_importantShouldProcess_4E5770(playerIndex C.uint8_t) C.int {
	player := importantPlayerByIndHook(ntype.PlayerInd(playerIndex))
	if player == nil || !importantGameHostHook() {
		return 1
	}
	return C.int(bool2int(player.Field3680&0x10 != 0))
}

//export nox_server_importantReplayRead_4E5770
func nox_server_importantReplayRead_4E5770() C.int {
	return C.int(bool2int(importantReplayReadHook()))
}

//export nox_server_importantGameHost_4E5770
func nox_server_importantGameHost_4E5770() C.int {
	return C.int(bool2int(importantGameHostHook()))
}

//export nox_server_importantRateAdjust_4E5770
func nox_server_importantRateAdjust_4E5770(playerIndex C.uint8_t) {
	importantRateAdjustHook(uint8(playerIndex))
}

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

func acknowledgeImportantPacketC(clientMask uint32, packet *importantPacketC, playerIndex int) {
	C.sub_4E54D0(C.uint32_t(clientMask), packet, C.int(playerIndex))
}

func acknowledgeImportantFrameC(playerIndex uint8, frame uint32) int {
	return int(C.nox_net_importantACK_4E55A0(C.uint8_t(playerIndex), C.uint32_t(frame)))
}

func cleanupImportantPlayerPacketsC(playerIndex uint8) int {
	return int(C.sub_4E55F0(C.uint8_t(playerIndex)))
}

func Sub_4E5630(playerIndex ntype.PlayerInd) (offset, threshold, resendInterval, resendsPerUpdate uint32) {
	var cThreshold C.uint32_t
	var cResendInterval C.uint32_t
	var cResendsPerUpdate C.uint32_t
	offset = uint32(C.sub_4E5630(
		C.uint8_t(playerIndex), &cThreshold, &cResendInterval, &cResendsPerUpdate,
	))
	return offset, uint32(cThreshold), uint32(cResendInterval), uint32(cResendsPerUpdate)
}

func adjustImportantRateC(playerIndex uint8) uint32 {
	return uint32(C.nox_xxx_importantCheckRate2_4E5670(C.uint8_t(playerIndex)))
}

func sendQueuedImportantC(playerIndex uint8, messageKind int) {
	C.nox_xxx_netImportant_4E5770(C.uint8_t(playerIndex), C.int(messageKind))
}

func importantFreeSlotsC() uint32 {
	return uint32(C.nox_xxx_importantFreeSlots_4E5A90())
}

func importantNoopC(p unsafe.Pointer) {
	C.nox_xxx_noop_4E5AB0(p)
}

func checkImportantRateC() int {
	return int(C.nox_xxx_importantCheckRate_4E52B0())
}

func kickPlayerDueToRateC(playerIndex int) int {
	return int(C.nox_xxx_playerKickDueToRate_4E5360(C.int(playerIndex)))
}

func cloneImportantPayload(payload []byte) (unsafe.Pointer, func()) {
	if len(payload) == 0 {
		return nil, nil
	}
	buf, free := alloc.CloneSlice(payload)
	return unsafe.Pointer(&buf[0]), free
}

func sendImportantPacketC(recipient int, payload []byte, relatedObject unsafe.Pointer, removeIfDisconnected int, sequenceEnabled bool) int {
	data, free := cloneImportantPayload(payload)
	if free != nil {
		defer free()
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

type importantPacketWrapperKind uint8

const (
	importantPacketSequenceDisabled importantPacketWrapperKind = iota
	importantPacketSequenceEnabled
	importantPacketReplaceExisting
)

func sendImportantPacketWrapperC(recipient int, payload []byte, relatedObject unsafe.Pointer, removeIfDisconnected int, kind importantPacketWrapperKind) int {
	data, free := cloneImportantPayload(payload)
	if free != nil {
		defer free()
	}
	switch kind {
	case importantPacketSequenceEnabled:
		return int(C.nox_xxx_netSendPacket1_4E5390(
			C.int(recipient), data, C.int(len(payload)), (*C.nox_object_t)(relatedObject), C.int(removeIfDisconnected),
		))
	case importantPacketReplaceExisting:
		return int(C.sub_4E5450(
			C.int(recipient), data, C.int(len(payload)), (*C.nox_object_t)(relatedObject), C.int(removeIfDisconnected),
		))
	case importantPacketSequenceDisabled:
		return int(C.nox_xxx_netSendPacket0_4E5420(
			C.int(recipient), data, C.int(len(payload)), (*C.nox_object_t)(relatedObject), C.int(removeIfDisconnected),
		))
	default:
		panic("invalid important-packet wrapper kind")
	}
}

func updateImportantRateControlC(ind int) int32 {
	return int32(C.sub_4E4E50(C.int(ind)))
}

func resetImportantPlayerCounterC(ind int) int {
	return int(C.sub_4E4F30(C.int(ind)))
}
