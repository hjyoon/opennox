package legacy

/*
#include "GAME3_3.h"
*/
import "C"

import (
	"encoding/binary"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/common/ntype"
)

//export nox_xxx_netImportant_4E5770
func nox_xxx_netImportant_4E5770(index C.uint8_t, messageKind C.int) {
	rates := importantRates4E4E50()
	if int(index) >= len(rates) {
		return
	}
	player := importantPlayerByIndHook(ntype.PlayerInd(index))
	if player != nil && importantGameHostHook() && player.Field3680&0x10 == 0 {
		return
	}

	rate := &rates[index]
	clientMask := C.uint32_t(1) << index
	var resent int
	headerPending := true
	var prev *importantPacketC
	for packet := C.nox_server_getImportantLast_4E4F80(); packet != nil; packet = prev {
		prev = C.nox_server_getImportantPrev_4E4F80(packet)
		related := asObjectS(C.nox_server_getImportantRelatedObject_4E5030(packet))
		if related != nil && related.Flags().Has(object.FlagDestroyed) {
			C.nox_server_setImportantRelatedObject_4E5030(packet, nil)
			related = nil
		}

		if packet.acknowledged_mask&clientMask != 0 {
			continue
		}
		recipient := uint8(packet.recipient)
		if (recipient&0x80 == 0 && recipient != uint8(index)) ||
			(recipient != 0xff && recipient&0x80 != 0 && recipient&0x7f == uint8(index)) {
			continue
		}
		if (related != nil && related.Field37&uint32(clientMask) == 0) ||
			(packet.remove_if_disconnected != 0 && importantRecipientMaskC()&uint32(clientMask) == 0) {
			sub_4E54D0(clientMask, packet, C.int(index))
			return
		}

		if packet.sent_mask&clientMask != 0 {
			if packet.retry_delay[index] != 0 {
				packet.retry_delay[index]--
				continue
			}
			if resent >= int(rate.resends_per_update) {
				continue
			}
			packet.sent_mask &^= clientMask
			packet.send_count++
			resent++
		}

		if headerPending {
			header := [5]byte{0xaa}
			length := 1
			if messageKind == 0 || index == 31 {
				binary.LittleEndian.PutUint32(header[1:], gameFrame())
				length = len(header)
			}
			if !importantNetSendHook(uint8(index), int(messageKind), header[:length]) {
				return
			}
			headerPending = false
		}

		size := int(packet.payload_size)
		payload := (*[len(importantPacketC{}.payload)]byte)(unsafe.Pointer(&packet.payload))[:size]
		if packet.sequence_enabled != 0 {
			scratch := memmap.PtrT[[4 + len(importantPacketC{}.payload)]byte](0x5D4594, 1564964)
			scratch[0] = 0xcc
			binary.LittleEndian.PutUint16(scratch[1:], uint16(packet.sequence[index]))
			scratch[3] = byte(size)
			copy(scratch[4:], payload)
			payload = scratch[:4+size]
		}
		if importantNetSendHook(uint8(index), int(messageKind), payload) {
			packet.sent_mask |= clientMask
			updateRate := importantRateGetHook()
			interval := uint32(rate.resend_interval)
			packet.retry_delay[index] = C.uint8_t((gameFPS() * interval) / updateRate)
			packet.last_send_frame[index] = C.uint32_t(gameFrame())
			if importantReplayReadHook() {
				sub_4E54D0(clientMask, packet, C.int(index))
			}
		}
	}

	if importantGameHostHook() {
		frame := gameFrame()
		interval := uint32(rate.resend_interval)
		if frame%(gameFPS()*interval) == 0 {
			importantRateAdjustHook(uint8(index))
		}
	}
}
