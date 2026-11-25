package legacy

/*
#include "client__network__cdecode.h"
*/
import "C"

import (
	"encoding/binary"
	"unsafe"

	"github.com/opennox/libs/noxnet/netmsg"

	"github.com/opennox/opennox/v1/common/memmap"
)

func clientInventoryReportSize(op netmsg.Op) int {
	switch op {
	case netmsg.MSG_REPORT_STAMINA:
		return 2
	case netmsg.MSG_REPORT_PLAYER_HEALTH, netmsg.MSG_REPORT_EQUIP,
		netmsg.MSG_REPORT_DEQUIP, netmsg.MSG_REPORT_LAST_QUIVER:
		return 3
	case netmsg.MSG_REPORT_HEALTH, netmsg.MSG_REPORT_OBJECT_POISON,
		netmsg.MSG_REPORT_SECONDARY_WEAPON, netmsg.MSG_INFO_BOOK_DATA:
		return 4
	case netmsg.MSG_REPORT_HEALTH_DELTA, netmsg.MSG_REPORT_MANA,
		netmsg.MSG_REPORT_ARMOR_VALUE, netmsg.MSG_REPORT_GOLD,
		netmsg.MSG_REPORT_PICKUP, netmsg.MSG_REPORT_DROP, netmsg.MSG_REPORT_CHARGES:
		return 5
	case netmsg.MSG_REPORT_ITEM_HEALTH, netmsg.MSG_REPORT_TOTAL_HEALTH, netmsg.MSG_REPORT_TOTAL_MANA:
		return 7
	case netmsg.MSG_REPORT_MODIFIABLE_PICKUP:
		return 9
	default:
		return 0
	}
}

func clientPacketDrawable(rawCode uint16) *nox_drawable {
	code := int32(rawCode & 0x7fff)
	if rawCode&0x8000 != 0 {
		return nox_xxx_netSpriteByCodeStatic_45A720(code)
	}
	return nox_xxx_netSpriteByCodeDynamic_45A6F0(code)
}

func clientInventoryReportPickup(op netmsg.Op, code C.int, data []byte) {
	attrs := C.nox_modifier_attrs_t{field_16: ^C.uint32_t(0)}
	if op == netmsg.MSG_REPORT_MODIFIABLE_PICKUP {
		for i, id := range data[5:9] {
			attrs.modifiers[i] = nox_xxx_modifGetDescById_413330(int32(id))
		}
	}
	thing := binary.LittleEndian.Uint16(data[3:5])
	if nox_xxx_spritePickup_461660(code, C.int(thing), &attrs) == 0 {
		nox_xxx_send2ServInvenFail_461630(C.short(code))
	}
}

func clientInventoryReportBook(rawCode uint16, page byte) {
	code := C.uint32_t(rawCode & 0x7fff)
	drawable := sub_478080(code)
	if drawable == nil {
		drawable = sub_4676D0(C.int(code))
	}
	if drawable == nil {
		drawable = clientPacketDrawable(rawCode)
	}
	if drawable != nil {
		drawableEffectBytes(asDrawable(drawable))[0] = page
	}
}

func clientInventoryReportPacket(op netmsg.Op, data []byte) int {
	size := clientInventoryReportSize(op)
	if size == 0 {
		return 0
	}
	if len(data) < size {
		return -1
	}
	switch op {
	case netmsg.MSG_REPORT_HEALTH_DELTA:
		code := binary.LittleEndian.Uint16(data[1:3])
		delta := binary.LittleEndian.Uint16(data[3:5])
		nox_xxx_cliAddHealthChange_49A650(C.int(code), C.short(delta))
		return size
	case netmsg.MSG_REPORT_SECONDARY_WEAPON:
		code := binary.LittleEndian.Uint16(data[1:3]) & 0x7fff
		sub_467750(C.int(code), C.char(data[3]))
		return size
	case netmsg.MSG_REPORT_LAST_QUIVER:
		code := binary.LittleEndian.Uint16(data[1:3]) & 0x7fff
		sub_467740(C.int(code))
		return size
	case netmsg.MSG_INFO_BOOK_DATA:
		clientInventoryReportBook(binary.LittleEndian.Uint16(data[1:3]), data[3])
		return size
	}

	if nox_client_isConnected_43C700() == 0 {
		return size
	}
	if op == netmsg.MSG_REPORT_STAMINA {
		sub_470D20(C.int(data[1]), C.int(memmap.Int32(0x587000, 157092)))
		return size
	}

	rawCode := binary.LittleEndian.Uint16(data[1:3])
	code := C.int(rawCode & 0x7fff)
	word := func(offset int) uint16 { return binary.LittleEndian.Uint16(data[offset : offset+2]) }
	switch op {
	case netmsg.MSG_REPORT_HEALTH:
		if nox_xxx_unitSpriteCheckAlly_4951F0(C.int(rawCode)) != 0 {
			sub_495150(C.int(rawCode), 2*C.short(data[3]))
		}
	case netmsg.MSG_REPORT_PLAYER_HEALTH:
		sub_470CB0(C.int(int16(rawCode)))
	case netmsg.MSG_REPORT_ITEM_HEALTH:
		sub_4675E0(code, C.short(word(3)), C.short(word(5)))
	case netmsg.MSG_REPORT_MANA:
		if int(code) == ClientPlayerNetCode() {
			nox_xxx_cliSetMana_470D10(C.int(word(3)))
		}
	case netmsg.MSG_REPORT_ARMOR_VALUE:
		sub_467450(C.int(binary.LittleEndian.Uint32(data[1:5])))
	case netmsg.MSG_REPORT_GOLD:
		sub_467490(C.int(binary.LittleEndian.Uint32(data[1:5])))
	case netmsg.MSG_REPORT_PICKUP, netmsg.MSG_REPORT_MODIFIABLE_PICKUP:
		clientInventoryReportPickup(op, code, data)
	case netmsg.MSG_REPORT_DROP:
		sub_461A80(code)
	case netmsg.MSG_REPORT_EQUIP:
		sub_462040(code)
	case netmsg.MSG_REPORT_DEQUIP:
		sub_4624D0(code)
	case netmsg.MSG_REPORT_CHARGES:
		sub_467930(code, C.int(data[3]), C.int(data[4]))
	case netmsg.MSG_REPORT_OBJECT_POISON:
		if nox_xxx_unitSpriteCheckAlly_4951F0(code) != 0 {
			sub_4950F0(code, C.char(data[3]))
		}
		nox_npc_set_328(int32(code), int32(data[3]))
	case netmsg.MSG_REPORT_TOTAL_HEALTH:
		if int(code) == ClientPlayerNetCode() {
			nox_xxx_cliSetTotalHealth_470C80(C.int(word(3)), C.int(word(5)))
		} else if nox_xxx_unitSpriteCheckAlly_4951F0(C.int(rawCode)) != 0 {
			sub_495120(C.int(rawCode), C.short(word(3)), C.short(word(5)))
		}
	case netmsg.MSG_REPORT_TOTAL_MANA:
		if int(code) == ClientPlayerNetCode() {
			nox_xxx_cliSetManaAndMax_470CE0(C.int(word(3)), C.int(word(5)))
		}
	}
	return size
}

//export nox_client_decodeInventoryReport
func nox_client_decodeInventoryReport(op C.int, data *C.uchar, size C.int) C.int {
	if data == nil || size < 0 {
		return -1
	}
	return C.int(clientInventoryReportPacket(netmsg.Op(op), unsafe.Slice((*byte)(unsafe.Pointer(data)), int(size))))
}
