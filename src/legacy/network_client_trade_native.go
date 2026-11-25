package legacy

/*
#include "client__network__cdecode.h"
#include "GAME3_1.h"

typedef const wchar2_t nox_shop_const_wchar2_t;
typedef const uint8_t nox_shop_modifier_id_t;
*/
import "C"

import (
	"encoding/binary"
	"unsafe"
)

var clientTradePacketSizes = [...]uint8{
	1: 2, 2: 2, 3: 3, 4: 15, 5: 4, 6: 14, 7: 2, 8: 18, 9: 4,
	12: 52, 13: 86, 27: 4, 29: 8, 31: 8,
}

func clientTradeStartPeer(data []byte) {
	// The ABI reader accepts a terminated name after the two header bytes.
	var packet [27]wchar2_t
	for i := 0; i < len(data)/2; i++ {
		packet[i] = wchar2_t(binary.LittleEndian.Uint16(data[2*i : 2*i+2]))
	}
	nox_xxx_netP2PStartTrade_4C1320((*C.nox_trade_const_u8_t)(unsafe.Pointer(&packet[0])))
}

func clientTradeStartShop(data []byte) {
	var title [26]C.nox_shop_const_wchar2_t
	for i := 0; i < len(title)-1; i++ {
		title[i] = C.nox_shop_const_wchar2_t(binary.LittleEndian.Uint16(data[4+2*i : 6+2*i]))
	}
	var intro [33]byte
	copy(intro[:], data[54:86])
	kind := C.int(binary.LittleEndian.Uint16(data[2:4]))
	nox_xxx_cliStartShopDlg_478FD0(&title[0], (*C.char)(unsafe.Pointer(&intro[0])), kind)
}

func clientTradePacket(data []byte) int {
	if len(data) < 2 || int(data[1]) >= len(clientTradePacketSizes) {
		return -1
	}
	size := int(clientTradePacketSizes[data[1]])
	if size == 0 || len(data) < size {
		return -1
	}
	if nox_client_isConnected_43C700() == 0 {
		return size
	}
	raw := (*C.nox_trade_const_u8_t)(unsafe.Pointer(&data[0]))
	word := func(offset int) uint16 { return binary.LittleEndian.Uint16(data[offset : offset+2]) }
	switch data[1] {
	case 1:
		sub_4C1590()
	case 2:
		sub_479280()
	case 3:
		sub_4C1BC0(raw)
	case 4:
		nox_xxx_tradeClientAddItem_4C1790(raw)
	case 5:
		sub_4C15D0(raw)
	case 6:
		sub_4C1B50(raw)
	case 7:
		nox_xxx_prepareP2PTrade_4C1BF0()
	case 8:
		price := binary.LittleEndian.Uint32(data[6:10])
		sub_479300(C.uint32_t(word(2)), C.uint32_t(word(4)), C.uint32_t(price), C.uint16_t(word(10)), (*C.nox_shop_modifier_id_t)(unsafe.Pointer(&data[14])))
	case 9:
		sub_479480(C.uint32_t(word(2)))
	case 12:
		clientTradeStartPeer(data[:size])
	case 13:
		clientTradeStartShop(data[:size])
	case 27:
		sub_479520(C.int(word(2)))
	case 29:
		sub_4795E0(C.int(word(2)), C.int(binary.LittleEndian.Uint32(data[4:8])))
	case 31:
		sub_479740(C.int(word(2)), C.uint(binary.LittleEndian.Uint32(data[4:8])))
	}
	return size
}

//export nox_client_decodeTradePacket
func nox_client_decodeTradePacket(data *C.uchar, size C.int) C.int {
	if data == nil || size < 0 {
		return -1
	}
	return C.int(clientTradePacket(unsafe.Slice((*byte)(unsafe.Pointer(data)), int(size))))
}
