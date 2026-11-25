package legacy

/*
#include "GAME2_1.h"
#include "GAME3_2.h"

typedef const nox_drawable inventory_const_drawable_t;
extern uint32_t dword_5d4594_1049864, dword_5d4594_1062496, nox_color_white_2523948;
extern uintptr_t dword_5d4594_1062492;
extern void* nox_inventory_font;
*/
import "C"

import (
	"encoding/binary"
	"unsafe"

	"github.com/opennox/libs/noxnet/netmsg"
	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/legacy/common/ccall"
)

//export sub_461F90
func sub_461F90(code C.int) *C.nox_drawable {
	heads := inventoryStateNativeHeads()
	for slot, head := range heads {
		for item := head; item != nil; item = item.NextPtr {
			if item.NetCode32 != uint32(code) {
				continue
			}
			prev, next := item.Field_93, item.NextPtr
			if prev != nil {
				prev.NextPtr = next
			} else {
				heads[slot] = next
			}
			if next != nil {
				next.Field_93 = prev
			}
			if uint32(item.ObjClass)&0x1000 != 0 || nox_xxx_ammoCheck_415880(int32(item.TypeIDVal)) == 2 ||
				nox_xxx_ammoCheck_415880(int32(item.TypeIDVal)) == 128 {
				sub_470D70()
			}
			item.NextPtr, item.Field_93 = nil, nil
			return (*C.nox_drawable)(item.C())
		}
	}
	return nil
}

//export sub_4622E0
func sub_4622E0(drawable *C.nox_drawable) C.int {
	primary, secondary := drawable.flags28, drawable.flags29
	if primary&0x1000000 != 0 && secondary&2 != 0 {
		return 0
	}
	if primary&0x2000000 != 0 {
		switch {
		case secondary&1 != 0:
			return 1
		case secondary&0x144 != 0:
			return 2
		case secondary&0x90 != 0:
			return 3
		case secondary&0x20 != 0:
			return 4
		case secondary&2 != 0:
			return 8
		case secondary&8 != 0:
			return 5
		}
	}
	if primary&0x1000000 != 0 {
		if secondary&4 != 0 {
			return 8
		}
		return 7
	}
	if primary&0x1000 != 0 {
		return 7
	}
	return 9
}

func clientInventorySendPacket(packet []byte) C.int {
	return C.int(nox_netlist_addToMsgListCli_40EBC0(31, 0, (*C.uchar)(&packet[0]), int32(len(packet))))
}

func clientInventorySendItem(op netmsg.Op, drawable *nox_drawable) C.int {
	packet := [3]byte{byte(op)}
	binary.LittleEndian.PutUint16(packet[1:], uint16(nox_xxx_netGetUnitCodeCli_578B00(drawable)))
	return clientInventorySendPacket(packet[:])
}

//export nox_xxx_clientEquip_4623B0
func nox_xxx_clientEquip_4623B0(drawable *C.nox_drawable) C.int {
	return clientInventorySendItem(netmsg.MSG_TRY_EQUIP, drawable)
}

//export nox_xxx_clientDequip_464B70
func nox_xxx_clientDequip_464B70(drawable *C.nox_drawable) C.int {
	return clientInventorySendItem(netmsg.MSG_TRY_DEQUIP, drawable)
}

//export sub_4623E0
func sub_4623E0(drawable *C.nox_drawable, slot C.int) *C.nox_drawable {
	heads := inventoryStateNativeHeads()
	if slot < 0 || int(slot) >= len(heads) {
		return nil
	}
	item := asDrawable(drawable)
	head := heads[slot]
	var previous *client.Drawable
	if drawable.flags28&0x2000000 != 0 && drawable.flags29&0x150 != 0 {
		for tail := head; tail != nil; tail = tail.NextPtr {
			previous = tail
		}
		if previous != nil && drawable.flags29&0x40 != 0 &&
			uint32(previous.ObjClass)&0x2000000 != 0 && uint32(previous.SubClass())&0x100 != 0 {
			previous = previous.Field_93
		}
	}
	if previous != nil {
		item.NextPtr = previous.NextPtr
		if item.NextPtr != nil {
			item.NextPtr.Field_93 = item
		}
		previous.NextPtr = item
		item.Field_93 = previous
		return (*C.nox_drawable)(previous.C())
	}
	item.Field_93 = nil
	item.NextPtr = head
	if head != nil {
		head.Field_93 = item
	}
	heads[slot] = item
	return (*C.nox_drawable)(head.C())
}

func clientInventoryAlternateCell() *C.nox_inventory_cell_t {
	return (*C.nox_inventory_cell_t)(globalStateNativePointer(2))
}

//export sub_4624D0
func sub_4624D0(code C.int) C.int {
	equipped := sub_461F90(code)
	if equipped == nil {
		return 0
	}
	source, _ := inventoryLookupNativeSearch(uint32(code))
	if source == nil {
		return C.int(nox_xxx_spriteDelete_45A4B0(equipped))
	}
	source.field_132 = 0
	if C.dword_5d4594_1062492 != C.uintptr_t(uintptr(unsafe.Pointer(equipped))) {
		alternate := clientInventoryAlternateCell()
		if nox_xxx_ammoCheck_415880(int32(equipped.field_27))&0xc != 0 && alternate != nil &&
			alternate.field_0 != nil && nox_xxx_ammoCheck_415880(int32(alternate.field_0.field_27)) == 2 {
			alternate.field_136 = 0
			nox_xxx_clientSetAltWeapon_461550(nil)
		}
		return C.int(nox_xxx_spriteDelete_45A4B0(equipped))
	}
	C.dword_5d4594_1062492 = 0
	if alternate := clientInventoryAlternateCell(); alternate != nil && alternate.field_0 != nil {
		C.dword_5d4594_1062496 = equipped.field_32
		alternate.field_0.field_32 = alternate.field_4
		nox_xxx_clientEquip_4623B0(alternate.field_0)
	} else {
		nox_xxx_clientSetAltWeapon_461550(source)
		source.field_136 = 1
	}
	return C.int(nox_xxx_spriteDelete_45A4B0(equipped))
}

//export sub_4625D0
func sub_4625D0(win *C.nox_window, draw *C.nox_window_data) C.int {
	if C.dword_5d4594_1049864 == 5 {
		return 1
	}
	x, y := meterWindowPosition(win)
	w, h := C.int(win.width), C.int(win.height)
	if y+h <= 0 {
		return 1
	}
	nox_xxx_drawSetTextColor_434390(int32(C.nox_color_white_2523948))
	if cell := clientInventoryAlternateCell(); cell != nil && cell.field_0 != nil {
		item := cell.field_0
		item.pos.x = C.intptr_t(x + w/2)
		item.pos.y = C.intptr_t(y + h/2)
		if item.draw_func != nil {
			ccall.CallIntPtr2(unsafe.Pointer(item.draw_func), unsafe.Pointer(nox_client_inventory_viewport_native()), unsafe.Pointer(item))
		}
	}
	if key := sub_42E8E0(35, 1); key != nil {
		nox_xxx_drawString_43F6E0(C.nox_inventory_font, key, int32(x+22), int32(y+41))
	}
	return 1
}

//export sub_4626C0
func sub_4626C0(drawable *C.inventory_const_drawable_t) C.double {
	return clientInventoryPreHitAmount((*nox_drawable)(unsafe.Pointer(drawable)), unsafe.Pointer(C.nox_xxx_lightngEffect_4E06F0))
}

//export sub_462700
func sub_462700(drawable *C.inventory_const_drawable_t) C.double {
	return clientInventoryPreHitAmount((*nox_drawable)(unsafe.Pointer(drawable)), unsafe.Pointer(C.nox_xxx_fireEffect_4E0550))
}

func clientInventoryPreHitAmount(drawable *nox_drawable, effect unsafe.Pointer) C.double {
	if drawable == nil || drawable.flags28&0x13001000 == 0 {
		return 0
	}
	item := drawableItemData(drawable)
	for _, modifier := range [2]unsafe.Pointer{item.Field_110, item.Field_111} {
		if modifier != nil && nox_modifier_effect_getPreHitFunc(modifier) == effect {
			return C.double(nox_modifier_effect_getPreHitFloat(modifier))
		}
	}
	return 0
}

//export sub_463370
func sub_463370(win *C.nox_window, pos *C.nox_point, output *C.uint32_t) C.int {
	x, y := meterWindowPosition(win)
	out := unsafe.Slice(output, 2)
	out[0] = C.uint32_t(pos.x) - C.uint32_t(x)
	out[1] = C.uint32_t(pos.y) - C.uint32_t(y)
	return C.int(out[1])
}

//export sub_4633B0
func sub_4633B0(drawable *C.inventory_const_drawable_t, current, maximum *C.float) {
	item := (*nox_drawable)(unsafe.Pointer(drawable))
	*current = C.float(item.field_73_1)
	*maximum = C.float(item.field_73_2)
	if item.flags28&0x13001000 == 0 {
		return
	}
	modifier := drawableItemData(item).Field_109
	if modifier != nil && nox_modifier_effect_getDefendFunc(modifier) == unsafe.Pointer(C.sub_4E0380) {
		multiplier := float64(nox_modifier_effect_getDefendFloat(modifier))
		*current = C.float(nox_float2int(C.float(float64(*current) * multiplier)))
		*maximum = C.float(nox_float2int(C.float(float64(*maximum) * multiplier)))
	}
}

//export sub_463420
func sub_463420(value C.int) C.int {
	*memmap.PtrUint32(0x5D4594, 1050012) = uint32(value)
	return value
}
