package legacy

/*
#include "GAME3_3.h"
*/
import "C"

//export sub_4E54D0
func sub_4E54D0(clientMask C.uint32_t, packet *C.nox_important_packet_t, playerIndex C.int) {
	if related := C.nox_server_getImportantRelatedObject_4E5030(packet); related != nil {
		kind := packet.payload[0]
		if kind >= 0x31 && kind <= 0x33 {
			asObjectS(related).Field37 &^= uint32(clientMask)
		}
	}
	active := C.uint32_t(importantRecipientMaskC()) & packet.recipient_mask
	recipient := uint8(packet.recipient)
	if recipient&0x80 == 0 {
		if C.int(recipient) == playerIndex {
			sub_4E4FC0(packet)
		}
		return
	}
	if recipient != 0xff {
		active &^= 1 << (recipient & 31)
	}
	packet.acknowledged_mask |= clientMask
	if active&packet.acknowledged_mask == active {
		sub_4E4FC0(packet)
	}
}

//export nox_net_importantACK_4E55A0
func nox_net_importantACK_4E55A0(playerIndex C.uint8_t, frame C.uint32_t) C.int {
	if int(playerIndex) >= len(importantPacketC{}.last_send_frame) {
		return 0
	}
	mask := C.uint32_t(1) << (playerIndex & 31)
	for packet := C.nox_server_getImportantFirst_4E4F80(); packet != nil; {
		matches := packet.last_send_frame[playerIndex] == frame
		next := C.nox_server_getImportantNext_4E4F80(packet)
		if matches {
			sub_4E54D0(mask, packet, C.int(playerIndex))
		}
		packet = next
	}
	return 0
}

//export sub_4E55F0
func sub_4E55F0(playerIndex C.uint8_t) C.int {
	mask := C.uint32_t(1) << (playerIndex & 31)
	for packet := C.nox_server_getImportantFirst_4E4F80(); packet != nil; {
		next := C.nox_server_getImportantNext_4E4F80(packet)
		sub_4E54D0(mask, packet, C.int(playerIndex))
		packet = next
	}
	return 0
}

//export sub_4E5630
func sub_4E5630(playerIndex C.uint8_t, threshold, interval, resends *C.uint32_t) C.uint32_t {
	rate := &importantRates4E4E50()[playerIndex]
	*threshold = rate.threshold
	*interval = C.uint32_t(rate.resend_interval)
	*resends = C.uint32_t(rate.resends_per_update)
	return 12 * C.uint32_t(playerIndex)
}

//export nox_xxx_importantFreeSlots_4E5A90
func nox_xxx_importantFreeSlots_4E5A90() C.uint32_t {
	var count uint32
	for packet := C.nox_server_getImportantFirst_4E4F80(); packet != nil; packet = C.nox_server_getImportantNext_4E4F80(packet) {
		count++
	}
	return C.uint32_t(importantCapacityC() - count)
}
