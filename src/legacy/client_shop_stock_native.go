package legacy

/*
#include "GAME2_2.h"
#include "client__gui__window.h"

typedef const uint8_t nox_shop_modifier_id_t;
typedef const uint32_t nox_shop_position_word_t;

extern nox_window* dword_5d4594_1098576;
extern uint32_t dword_5d4594_1098624, dword_5d4594_1098628, dword_5d4594_1107036;
extern wchar2_t *dword_5d4594_1098596, *dword_5d4594_1098600, *dword_5d4594_1098604;
extern char* nox_shop_dialog_file_1098608;
extern uint32_t dword_5d4594_1098616;
*/
import "C"

import "unsafe"

const (
	shopRows     = C.NOX_SHOP_INVENTORY_ROW_COUNT
	shopCols     = C.NOX_SHOP_INVENTORY_COLUMN_COUNT
	shopStackMax = C.NOX_SHOP_INVENTORY_STACK_MAX
)

type shopInventoryCell = C.nox_shop_inventory_cell_t

func shopRoot() *nox_window { return C.dword_5d4594_1098576 }

func shopCell(row, col int) *shopInventoryCell {
	return &C.nox_client_shop_inventory[row+shopRows*col]
}

//export sub_478080
func sub_478080(code C.uint32_t) *nox_drawable {
	if C.dword_5d4594_1098624 != 0 {
		if cell := sub_4780A0(code); cell != nil {
			return cell.drawable
		}
	}
	return nil
}

//export sub_4780A0
func sub_4780A0(code C.uint32_t) *shopInventoryCell {
	// GAME.EXE searches by row even though the cells are stored by column.
	for row := range shopRows {
		for col := range shopCols {
			cell := shopCell(row, col)
			if cell.drawable == nil {
				continue
			}
			for _, stored := range &cell.net_codes {
				if stored == code {
					return cell
				}
			}
		}
	}
	return nil
}

//export sub_478F10
func sub_478F10() {
	for row := range shopRows {
		for col := range shopCols {
			cell := shopCell(row, col)
			if cell.drawable != nil {
				nox_xxx_spriteDelete_45A4B0(cell.drawable)
			}
			*cell = shopInventoryCell{}
		}
	}
	C.dword_5d4594_1107036 = 0
}

//export nox_gui_shop_state
func nox_gui_shop_state(mode, count *C.uint32_t) C.int {
	if mode != nil {
		*mode = C.dword_5d4594_1098628
	}
	if count != nil {
		var total C.uint32_t
		for i := range C.nox_client_shop_inventory {
			total += C.nox_client_shop_inventory[i].count
		}
		*count = total
	}
	return C.int(bool2int(shopRoot() != nil && C.dword_5d4594_1098624 != 0))
}

//export sub_478F80
func sub_478F80() C.int {
	sub_478F10()
	sub_44D8F0()
	nox_xxx_windowDestroyMB_46C4E0(shopRoot())
	C.dword_5d4594_1098576 = nil
	C.dword_5d4594_1098624, C.dword_5d4594_1098628 = 0, 1
	C.dword_5d4594_1098596, C.dword_5d4594_1098600, C.dword_5d4594_1098604 = nil, nil, nil
	C.nox_shop_dialog_file_1098608 = nil
	C.dword_5d4594_1098616 = 0
	return 0
}

var shopPortraits = [...]struct {
	typeOffset uintptr
	merchant   string
	image      string
}{
	{1098396, "Shopkeeper", "ShopKeeperPic"},
	{1098560, "ShopkeeperWarriorsRealm", "ShopKeeperWarriorPic"},
	{1098556, "ShopkeeperConjurerRealm", "ShopKeeperConjurerPic"},
	{1098564, "ShopkeeperWizardRealm", "ShopKeeperWizardPic"},
	{1098572, "ShopkeeperLandOfTheDead", "ShopKeeperLandOfTheDeadPic"},
	{1098568, "ShopkeeperMagicShop", "ShopKeeperMagicShopPic"},
	{1098484, "ShopkeeperPurple", "ShopKeeperPurplePic"},
	{1097292, "ShopkeeperYellow", "ShopKeeperBrownPic"},
}

//export nox_xxx_getShopPic_4790F0
func nox_xxx_getShopPic_4790F0(typeID C.int) *nox_video_bag_image_t {
	portrait := nox_xxx_wndGetChildByID_46B0C0(shopRoot(), 3805)
	if *g2ServerWord(1107040) == 0 {
		for _, entry := range shopPortraits {
			*g2ServerWord(entry.typeOffset) = uint32(GetClient().Cli().Things.IndByID(entry.merchant))
		}
		*g2ServerWord(1107040) = 1
	}
	name := shopPortraits[0].image
	for _, entry := range shopPortraits {
		if *g2ServerWord(entry.typeOffset) == uint32(typeID) {
			name = entry.image
			break
		}
	}
	img := Nox_xxx_gLoadImg(name)
	if portrait != nil {
		asWindow(portrait).DrawData().SetBackgroundImage(img)
	}
	return (*nox_video_bag_image_t)(img.C())
}

//export sub_479280
func sub_479280() {
	if C.dword_5d4594_1098624 == 0 {
		return
	}
	sub_467680()
	C.dword_5d4594_1098624, C.dword_5d4594_1098628 = 0, 0
	sub_478F10()
	sub_44D8F0()
	nox_window_set_hidden(shopRoot(), 1)
	nox_xxx_wnd_46ABB0_go(shopRoot(), 0)
	Sub_467C10()
	nox_client_setCursorType_477610(0)
	if nox_client_getRenderGUI() == 0 && *g2ServerWord(1098612) == 1 {
		nox_client_setRenderGUI(1)
	}
}

//export sub_479300
func sub_479300(typeID, code, price C.uint32_t, durability C.uint16_t, modifiers *C.nox_shop_modifier_id_t) C.uint32_t {
	cell := sub_4793C0(typeID)
	if cell == nil {
		return 0
	}
	if cell.drawable == nil {
		dr := nox_new_drawable_for_thing(int32(typeID))
		if dr == nil {
			return 0
		}
		dr.flags30 |= 0x40000000
		dr.field_73_2, dr.field_73_1 = durability, durability
		if dr.flags28&0x13001000 != 0 && modifiers != nil {
			var attrs [4]unsafe.Pointer
			for i, id := range (*[4]C.nox_shop_modifier_id_t)(unsafe.Pointer(modifiers)) {
				if byte(id) != 0xff {
					attrs[i] = nox_xxx_modifGetDescById_413330(int32(byte(id)))
				}
			}
			item := drawableItemData(dr)
			item.Field_108, item.Field_109, item.Field_110, item.Field_111 = attrs[0], attrs[1], attrs[2], attrs[3]
		}
		cell.drawable, cell.count = dr, 0
	}
	cell.net_codes[cell.count] = code
	cell.count++
	cell.price = price
	return cell.count
}

//export sub_4793C0
func sub_4793C0(typeID C.uint32_t) *shopInventoryCell {
	for row := range shopRows {
		for col := range shopCols {
			cell := shopCell(row, col)
			if cell.count > 0 && cell.count < shopStackMax && cell.drawable != nil && cell.drawable.field_27 == typeID &&
				cell.drawable.flags28&0x4000000 == 0 {
				return cell
			}
		}
	}
	return sub_479430()
}

//export sub_479430
func sub_479430() *shopInventoryCell {
	for row := range shopRows {
		for col := range shopCols {
			cell := shopCell(row, col)
			if cell.count == 0 {
				return cell
			}
		}
	}
	return nil
}

//export sub_479480
func sub_479480(code C.uint32_t) C.uint32_t {
	cell := sub_4780A0(code)
	if cell == nil {
		return 0
	}
	sub_4794D0(cell, code)
	cell.count--
	if cell.count == 0 {
		nox_xxx_spriteDelete_45A4B0(cell.drawable)
		cell.drawable, cell.price = nil, 0
	}
	return cell.count
}

//export sub_4794D0
func sub_4794D0(cell *shopInventoryCell, code C.uint32_t) C.int {
	for i, stored := range &cell.net_codes {
		if stored == code {
			copy(cell.net_codes[i:], cell.net_codes[i+1:])
			cell.net_codes[shopStackMax-1] = 0
			return C.int(i)
		}
	}
	return shopStackMax
}

//export sub_4795A0
func sub_4795A0(mode C.int) {
	if C.dword_5d4594_1098628 == 4 && mode != 4 {
		sub_467680()
	}
	C.dword_5d4594_1098628 = C.uint32_t(mode)
}

//export sub_479880
func sub_479880(point *C.uint32_t) C.bool {
	return C.bool(nox_xxx_pointInRect_4281F0((*C.int2)(unsafe.Pointer(point)), (*C.int4)(g2ServerMem(1098380))) != 0)
}

func shopCellAt(point *C.int2) *shopInventoryCell {
	col := (uint32(point.field_0) - *g2ServerWord(1098380)) / 50
	row := (uint32(point.field_4) - *g2ServerWord(1098384) + uint32(C.dword_5d4594_1107036)) / 50
	return shopCell(int(min(row, shopRows-1)), int(min(col, shopCols-1)))
}

//export sub_4798A0
func sub_4798A0(point *C.nox_shop_position_word_t) *nox_drawable {
	if !sub_479880((*C.uint32_t)(unsafe.Pointer(point))) {
		return nil
	}
	cell := shopCellAt((*C.int2)(unsafe.Pointer(point)))
	if cell.count == 0 {
		return nil
	}
	cell.drawable.field_32 = cell.net_codes[0]
	return cell.drawable
}
