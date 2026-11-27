package legacy

/*
#include "GAME2_1.h"
#include "GAME2_2.h"
#include "client__gui__guishop.h"
#include "client__gui__guimsg.h"

*/
import "C"

import (
	"encoding/binary"

	"github.com/opennox/libs/noxnet/netmsg"
)

const shopSource = "C:\\NoxPost\\src\\client\\Gui\\GUIShop.c"

func shopItemPacket(command byte, item uint16, amount ...byte) C.int {
	var packet [5]byte
	packet[0], packet[1] = byte(netmsg.MSG_TRADE), command
	binary.LittleEndian.PutUint16(packet[2:4], item)
	n := 4 + copy(packet[4:], amount)
	return C.int(nox_netlist_addToMsgListCli_40EBC0(31, 0, (*C.uchar)(&packet[0]), int32(n)))
}

//export sub_478040
func sub_478040() C.int {
	if clientUIWords.shop_active == 0 {
		return 0
	}
	tradeActionPacket(18)
	sub_467680()
	return 1
}

//export sub_478850
func sub_478850(position *C.int2, itemID, typeID, amount, extra C.uint32_t) {
	switch amount {
	case 0:
	case 1:
		nox_client_tradeXxxBuyAccept_478880(C.int(typeID), C.short(itemID))
	default:
		sub_4788F0(C.int(typeID), C.int(amount))
	}
}

func shopBuyTooHeavy(line int32) {
	nox_xxx_clientPlaySoundSpecial_452D80(925, 100)
	nox_xxx_printCentered_445490(nox_strman_loadString_40F1D0(
		internCStr("pickup.c:CarryingTooMuch"), nil, internCStr(shopSource), line))
}

//export nox_client_tradeXxxBuyAccept_478880
func nox_client_tradeXxxBuyAccept_478880(typeID C.int, code C.short) {
	if C.sub_467B00(typeID, 1) != 0 {
		shopItemPacket(22, uint16(code))
	} else {
		shopBuyTooHeavy(207)
	}
}

//export sub_4788F0
func sub_4788F0(typeID, amount C.int) {
	if C.sub_467B00(typeID, amount) != 0 {
		shopItemPacket(23, uint16(typeID), byte(amount))
	} else {
		shopBuyTooHeavy(233)
	}
}

//export sub_479680
func sub_479680(position *C.int2, itemID, typeID, amount, extra C.uint32_t) {
	legacyWords.dword_5d4594_1098616 = 0
}

//export sub_479690
func sub_479690(position *C.int2, itemID, typeID, amount, extra C.uint32_t) {
	legacyWords.dword_5d4594_1098616 = 0
	switch amount {
	case 0:
	case 1:
		nox_client_tradeXxxSellAccept_4796D0(C.short(itemID))
	default:
		sub_479700(C.short(typeID), C.char(amount))
	}
}

//export nox_client_tradeXxxSellAccept_4796D0
func nox_client_tradeXxxSellAccept_4796D0(code C.short) C.int {
	return shopItemPacket(24, uint16(code))
}

//export sub_479700
func sub_479700(typeID C.short, amount C.char) C.int {
	return shopItemPacket(25, uint16(typeID), byte(amount))
}

//export sub_479810
func sub_479810(position *C.int2, itemID, typeID, amount, extra C.uint32_t) {
	legacyWords.dword_5d4594_1098620 = 0
}

//export sub_479820
func sub_479820(position *C.int2, itemID, typeID, amount, extra C.uint32_t) {
	sub_479840(C.short(itemID))
	legacyWords.dword_5d4594_1098620 = 0
}

//export sub_479840
func sub_479840(code C.short) C.int {
	return shopItemPacket(26, uint16(code))
}
