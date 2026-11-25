package legacy

/*
#include "GAME2_1.h"
#include "GAME3_3.h"

typedef const nox_draw_viewport_t nox_const_inventory_viewport;
typedef const nox_inventory_cell_t nox_const_inventory_cell;

extern uint32_t dword_587000_136184;
*/
import "C"

import (
	"encoding/binary"
	"image"
	"unsafe"

	"github.com/opennox/opennox/v1/client/noxrender"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

var clientInventoryViewport, _ = alloc.NewHigh(noxrender.Viewport{})
var clientInventoryDragged *nox_drawable

func legacyViewportFromWords32(words *[13]uint32) noxrender.Viewport {
	point := func(i int) image.Point {
		return image.Pt(int(int32(words[i])), int(int32(words[i+1])))
	}
	return noxrender.Viewport{
		Screen:   image.Rectangle{Min: point(0), Max: point(2)},
		World:    image.Rectangle{Min: point(4), Max: point(6)},
		Size:     point(8),
		Field10:  uint(words[10]),
		Field11:  uint(words[11]),
		Jiggle12: int(int32(words[12])),
	}
}

//export nox_client_inventory_viewport_native
func nox_client_inventory_viewport_native() *C.nox_const_inventory_viewport {
	words := memmap.PtrT[[13]uint32](0x5D4594, 1049732)
	*clientInventoryViewport = legacyViewportFromWords32(words)
	return (*C.nox_const_inventory_viewport)(clientInventoryViewport.C())
}

//export nox_client_inventory_get_dragged
func nox_client_inventory_get_dragged() *nox_drawable {
	return clientInventoryDragged
}

//export nox_client_inventory_set_dragged
func nox_client_inventory_set_dragged(drawable *nox_drawable) {
	clientInventoryDragged = drawable
	if unsafe.Sizeof(uintptr(0)) == 4 {
		*memmap.PtrUint32(0x5D4594, 1049848) = uint32(uintptr(unsafe.Pointer(drawable)))
	}
}

//export nox_client_inventory_animation_state
func nox_client_inventory_animation_state() C.int {
	return C.int(memmap.Uint8(0x5D4594, 1049868))
}

//export nox_client_inventory_animation_offset
func nox_client_inventory_animation_offset() C.int {
	return C.int(C.dword_587000_136184)
}

//export nox_client_inventory_has_dragged
func nox_client_inventory_has_dragged() C.int {
	return C.int(bool2int(clientInventoryDragged != nil))
}

//export nox_inventory_cell_last_net_code_4657E0
func nox_inventory_cell_last_net_code_4657E0(cell *C.nox_const_inventory_cell) C.uint32_t {
	if cell == nil {
		return 0
	}
	const codes = int(unsafe.Offsetof(C.nox_inventory_cell_t{}.field_140)-unsafe.Offsetof(C.nox_inventory_cell_t{}.field_4)) / 4
	count := int(cell.field_140)
	if count == 0 || count > codes {
		return 0
	}
	return unsafe.Slice(&cell.field_4, codes)[count-1]
}

//export sub_461520
func sub_461520() *C.char {
	for player := nox_common_playerInfoGetFirst_416EA0(); player != nil; player = nox_common_playerInfoGetNext_416EE0(player) {
		nox_xxx_playerInitColors_461460(player)
	}
	return nil
}

//export sub_4615C0
func sub_4615C0() *nox_drawable {
	bow := memmap.PtrUint32(0x5D4594, 1063640)
	if *bow == 0 {
		*bow = uint32(nox_xxx_getTTByNameSpriteMB_44CFC0(internCStr("Bow")))
	}
	equipped := inventoryStateNativeHeads()
	for item := equipped[8]; item != nil; item = item.NextPtr {
		if item.TypeIDVal == *bow {
			return (*nox_drawable)(equipped[8].C())
		}
	}
	return (*nox_drawable)(equipped[7].C())
}

//export nox_xxx_send2ServInvenFail_461630
func nox_xxx_send2ServInvenFail_461630(code C.short) C.int {
	packet := [3]byte{0xf1}
	binary.LittleEndian.PutUint16(packet[1:], uint16(code))
	return nox_xxx_netClientSend2_4E53C0(31, C.nox_important_payload_t(unsafe.Pointer(&packet[0])), C.int(len(packet)), nil, 0)
}
