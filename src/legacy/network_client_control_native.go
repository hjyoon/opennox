package legacy

/*
#include "client__network__cdecode.h"
*/
import "C"

import (
	"encoding/binary"
	"unsafe"

	"github.com/opennox/libs/noxnet/netmsg"

	"github.com/opennox/opennox/v1/common/ntype"
)

var clientControlPacketSizes = [256]uint8{
	netmsg.MSG_IMPORTANT_ACK:        5,
	netmsg.MSG_OUTGOING_CLIENT:      3,
	netmsg.MSG_KICK_NOTIFICATION:    1,
	netmsg.MSG_TIMEOUT_NOTIFICATION: 1,
	netmsg.MSG_CHAT_KILL:            3,
	netmsg.MSG_MESSAGES_KILL:        1,
	netmsg.MSG_TIMER_STATUS:         13,
	netmsg.MSG_REPORT_ALL_LATENCY:   5,
	netmsg.MSG_MUSIC_EVENT:          3,
	netmsg.MSG_MUSIC_PUSH_EVENT:     3,
	netmsg.MSG_MUSIC_POP_EVENT:      3,
}

func clientControlPacket(ind ntype.PlayerInd, op netmsg.Op, data []byte) (int, bool) {
	switch op {
	case netmsg.MSG_SEQ_IMPORTANT:
		return ClientSequencePacket48D5A0(data, nox_client_isConnected_43C700() != 0), true
	case netmsg.MSG_VOTE:
		if len(data) >= 2 {
			switch data[1] {
			case 6:
				if len(data) >= 3 {
					sub_48D4B0(C.int(data[2]))
					return 3, true
				}
			case 7:
				sub_48D4A0()
				return 2, true
			}
		}
		return -1, true
	}
	size := int(clientControlPacketSizes[op])
	if size == 0 {
		return 0, false
	}
	if len(data) < size {
		return -1, true
	}
	switch op {
	case netmsg.MSG_OUTGOING_CLIENT:
		return size, true
	case netmsg.MSG_KICK_NOTIFICATION:
		sub_43B6E0()
		return size, true
	case netmsg.MSG_TIMEOUT_NOTIFICATION:
		sub_43B750()
		return size, true
	case netmsg.MSG_REPORT_ALL_LATENCY:
		if player := nox_common_playerInfoGetByID_417040(int32(binary.LittleEndian.Uint16(data[1:3]))); player != nil {
			*(*C.uint16_t)(unsafe.Pointer(&player.field_2148)) = C.uint16_t(binary.LittleEndian.Uint16(data[3:5]))
		}
		return size, true
	}
	if nox_client_isConnected_43C700() == 0 {
		return size, true
	}
	switch op {
	case netmsg.MSG_IMPORTANT_ACK:
		nox_net_importantACK_4E55A0(C.uint8_t(ind), C.uint32_t(binary.LittleEndian.Uint32(data[1:5])))
	case netmsg.MSG_CHAT_KILL:
		code := binary.LittleEndian.Uint16(data[1:3])
		if int16(code) == -8531 {
			sub_48E940()
		} else {
			sub_48E8E0(C.int(code))
		}
	case netmsg.MSG_MESSAGES_KILL:
		sub_445450()
	case netmsg.MSG_TIMER_STATUS:
		frame := binary.LittleEndian.Uint32(data[9:13])
		if gameFrame()-frame >= 30 {
			Nox_xxx_netClientSend2_4E53C0(31, []byte{byte(netmsg.MSG_REQUEST_TIMER_STATUS)}, nil, 1)
		} else if binary.LittleEndian.Uint32(data[1:5]) != 0 {
			sub_40A1F0(1)
			sub_40A310(C.int(binary.LittleEndian.Uint32(data[5:9])))
		} else {
			sub_40A1F0(0)
		}
	case netmsg.MSG_MUSIC_EVENT:
		sub_43D9B0(int32(data[1]), int32(data[2]))
	case netmsg.MSG_MUSIC_PUSH_EVENT:
		sub_43DA80()
	case netmsg.MSG_MUSIC_POP_EVENT:
		sub_43DAD0()
	}
	return size, true
}

//export nox_client_decodeControlPacket
func nox_client_decodeControlPacket(ind, op C.int, data *C.uchar, size C.int) C.int {
	if data == nil || size < 0 {
		return -1
	}
	n, _ := clientControlPacket(ntype.PlayerInd(ind), netmsg.Op(op), unsafe.Slice((*byte)(unsafe.Pointer(data)), int(size)))
	return C.int(n)
}
