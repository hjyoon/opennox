package legacy

/*
#include "GAME3_3.h"
*/
import "C"

import (
	"unsafe"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/ntype"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

//export nox_xxx_netSendPacket_4E5030
func nox_xxx_netSendPacket_4E5030(
	recipient C.int, payload C.nox_important_payload_t, size C.int,
	related *C.nox_object_t, removeIfDisconnected C.int, sequenceEnabled C.char,
) C.int {
	excludesPlayer := recipient != 0xff && recipient&0x80 != 0
	if excludesPlayer && importantRecipientMaskC()&^(1<<(uint32(recipient)&31)) == 0 {
		return 1
	}
	if size < 0 || size > C.int(len(importantPacketC{}.payload)) || (size != 0 && payload == nil) {
		return 0
	}
	if sequenceEnabled != 0 && recipient != 0xff && !excludesPlayer &&
		(recipient < 0 || int(recipient) >= len(importantPacketC{}.sequence)) {
		return 0
	}

	pool := alloc.AsClassT[importantPacketC](importantAllocClassC())
	if pool.Class == nil {
		capacity := 256
		if noxflags.HasGame(noxflags.GameModeCoop) {
			capacity = 512
		} else if noxflags.HasGame(noxflags.GameHost) {
			capacity = 3072
		}
		setImportantCapacityC(uint32(capacity))
		if noxflags.HasGame(noxflags.GameModeCoop) {
			pool = alloc.NewDynamicClassT("importantClass", importantPacketC{}, capacity)
		} else {
			pool = alloc.NewClassT("importantClass", importantPacketC{}, capacity)
		}
		setImportantAllocClassC(pool.UPtr())
	}
	packet := pool.NewObject()
	if packet == nil {
		if nox_xxx_importantCheckRate_4E52B0() != 1 {
			return 0
		}
		pool = alloc.AsClassT[importantPacketC](importantAllocClassC())
		packet = pool.NewObject()
		if packet == nil {
			return 0
		}
	}

	destination := unsafe.Slice((*byte)(unsafe.Pointer(&packet.payload[0])), int(size))
	copy(destination, unsafe.Slice((*byte)(unsafe.Pointer(payload)), int(size)))
	packet.payload_size = C.uint8_t(size)
	C.nox_server_setImportantRelatedObject_4E5030(packet, related)
	packet.recipient = C.int8_t(recipient)
	packet.remove_if_disconnected = C.uint32_t(removeIfDisconnected)
	packet.sequence_enabled = C.uint8_t(sequenceEnabled)
	packet.sent_mask = 0
	packet.acknowledged_mask = 0
	packet.send_count = 0
	packet.recipient_mask = C.uint32_t(importantRecipientMaskC())
	clear(packet.last_send_frame[:])
	packet.created_frame = C.uint32_t(gameFrame())
	clear(packet.sequence[:])

	if packet.sequence_enabled != 0 {
		counters := importantCounters4E4ED0()
		if recipient != 0xff && !excludesPlayer {
			packet.sequence[recipient] = counters[recipient]
			counters[recipient]++
		} else {
			mask := packet.recipient_mask
			if excludesPlayer && recipient&0x7f < C.int(len(packet.sequence)) {
				mask &^= 1 << (recipient & 0x7f)
			}
			for index := range packet.sequence {
				if mask&(1<<index) != 0 {
					packet.sequence[index] = counters[index]
					counters[index]++
				}
			}
		}
	}

	first := C.nox_server_getImportantFirst_4E4F80()
	C.nox_server_setImportantPrev_4E4F80(packet, nil)
	C.nox_server_setImportantNext_4E4F80(packet, first)
	if first == nil {
		C.nox_server_setImportantLast_4E4F80(packet)
	} else {
		C.nox_server_setImportantPrev_4E4F80(first, packet)
	}
	C.nox_server_setImportantFirst_4E4F80(packet)
	return 1
}

//export nox_xxx_importantCheckRate_4E52B0
func nox_xxx_importantCheckRate_4E52B0() C.int {
	var oldest *importantPacketC
	oldestFrame := C.uint32_t(999999999)
	var playerCounts [len(importantPacketC{}.sequence)]uint16
	kickIndex := -1
	var mostPackets uint16
	for packet := C.nox_server_getImportantFirst_4E4F80(); packet != nil; packet = C.nox_server_getImportantNext_4E4F80(packet) {
		if index := uint8(packet.recipient); index < uint8(len(playerCounts)-1) {
			playerCounts[index]++
			if playerCounts[index] > mostPackets {
				kickIndex, mostPackets = int(index), playerCounts[index]
			}
		}
		if packet.created_frame < oldestFrame {
			oldest, oldestFrame = packet, packet.created_frame
		}
	}
	if kickIndex != -1 {
		nox_xxx_playerKickDueToRate_4E5360(C.int(kickIndex))
	}
	if oldest == nil {
		return 0
	}
	// The kick can acknowledge and free the same packet selected for eviction.
	if pool := alloc.AsClass(importantAllocClassC()); pool.HasActiveObject(unsafe.Pointer(oldest)) {
		sub_4E4FC0(oldest)
	}
	return 1
}

//export nox_xxx_netSendPacket1_4E5390
func nox_xxx_netSendPacket1_4E5390(
	recipient C.int, payload C.nox_important_payload_t, size C.int,
	related *C.nox_object_t, removeIfDisconnected C.int,
) C.int {
	return nox_xxx_netSendPacket_4E5030(recipient, payload, size, related, removeIfDisconnected, 1)
}

//export nox_xxx_netClientSend2_4E53C0
func nox_xxx_netClientSend2_4E53C0(
	recipient C.int, payload C.nox_important_payload_t, size C.int,
	related *C.nox_object_t, removeIfDisconnected C.int,
) C.int {
	if noxflags.HasGame(noxflags.GameHost) {
		nox_netlist_addToMsgListCli_40EBC0(int32(recipient), 0, (*C.uchar)(unsafe.Pointer(payload)), int32(size))
		return 1
	}
	if recipient == 0xff || recipient&0x80 != 0 {
		return 0
	}
	return nox_xxx_netSendPacket0_4E5420(recipient, payload, size, related, removeIfDisconnected)
}

//export nox_xxx_netSendPacket0_4E5420
func nox_xxx_netSendPacket0_4E5420(
	recipient C.int, payload C.nox_important_payload_t, size C.int,
	related *C.nox_object_t, removeIfDisconnected C.int,
) C.int {
	return nox_xxx_netSendPacket_4E5030(recipient, payload, size, related, removeIfDisconnected, 0)
}

//export sub_4E5450
func sub_4E5450(
	recipient C.int, payload C.nox_important_payload_t, size C.int,
	related *C.nox_object_t, removeIfDisconnected C.int,
) C.int {
	if payload == nil {
		return 0
	}
	kind := *(*C.uint8_t)(unsafe.Pointer(payload))
	for packet := C.nox_server_getImportantFirst_4E4F80(); packet != nil; {
		next := C.nox_server_getImportantNext_4E4F80(packet)
		if kind == packet.payload[0] {
			if recipient == 0xff || recipient&0x80 != 0 {
				sub_4E4FC0(packet)
			} else {
				mask := C.uint32_t(1) << (uint32(recipient) & 31)
				sub_4E54D0(mask, packet, recipient)
			}
		}
		packet = next
	}
	return nox_xxx_netSendPacket0_4E5420(recipient, payload, size, related, removeIfDisconnected)
}

//export nox_xxx_importantCheckRate2_4E5670
func nox_xxx_importantCheckRate2_4E5670(playerIndex C.uint8_t) C.uint32_t {
	rates := importantRates4E4E50()
	if int(playerIndex) >= len(rates) {
		return 0
	}
	clientMask := C.uint32_t(1) << playerIndex
	var pending int32
	importantPlayerByIndHook(ntype.PlayerInd(playerIndex))
	for packet := C.nox_server_getImportantLast_4E4F80(); packet != nil; packet = C.nox_server_getImportantPrev_4E4F80(packet) {
		if packet.acknowledged_mask&clientMask != 0 {
			continue
		}
		recipient := uint8(packet.recipient)
		if recipient == 0xff || (recipient&0x80 == 0 && recipient == uint8(playerIndex)) ||
			(recipient&0x80 != 0 && recipient&0x7f != uint8(playerIndex)) {
			pending++
		}
	}

	rate := &rates[playerIndex]
	if importantRateGetHook() != uint32(rate.update_rate) {
		rate.update_rate = C.uint8_t(importantRateGetHook())
	}
	if pending > int32(rate.threshold) {
		rate.resend_interval++
		if rate.resend_interval > 5 {
			if rate.resends_per_update == 2 {
				nox_xxx_playerKickDueToRate_4E5360(C.int(playerIndex))
			}
			rate.resend_interval = 5
			rate.resends_per_update = 2
		}
		rate.update_rate = C.uint8_t(importantRateGetHook())
	} else if lower := int32(rate.lower_threshold); lower > 0 && pending < lower {
		if rate.resends_per_update == 2 {
			rate.resends_per_update = 1
		} else {
			rate.resend_interval--
			if rate.resend_interval < 2 {
				rate.resend_interval = 2
			}
		}
	}
	return C.uint32_t(sub_4E4E50(C.int(playerIndex)))
}
