package legacy

/*
#include "GAME2_1.h"
#include "GAME2_3.h"
#include "GAME3_1.h"
#include "client__gui__window.h"
#include "noxstring.h"
extern int nox_win_width, nox_win_height;

_Static_assert(sizeof(nox_gui_item_amount_callback_t) == sizeof(uintptr_t),
	"item-amount callback must occupy one native pointer slot");

static inline void item_amount_call(nox_gui_item_amount_callback_t callback, int2* pos, uint32_t id,
									uint32_t thing, uint32_t amount, uint32_t extra) {
	callback(pos, id, thing, amount, extra);
}
static inline void item_amount_format_tag(wchar2_t* text, const wchar2_t* format) {
	nox_swprintf(text, format);
}
*/
import "C"

import (
	"image"
	"strconv"
	"unsafe"

	"github.com/opennox/opennox/v1/client/noxrender"
)

const (
	itemAmountImageBase = iota
	itemAmountImageUp
	itemAmountImageDown
	itemAmountImageYes
	itemAmountImageNo
	itemAmountImageBaseWithoutTag
	itemAmountImageYesWithoutTag
	itemAmountImageNoWithoutTag
)

type itemAmountCallback = C.nox_gui_item_amount_callback_t

var itemAmountDialog struct {
	root, amountInput, priceLabel *nox_window
	drawable                      *nox_drawable
	images                        [8]*nox_video_bag_image_t
	accept, cancel                itemAmountCallback
	maxAmount, unitPrice          uint32
	priceMode                     uint32
	active                        bool
}

func itemAmountRoot() *nox_window { return itemAmountDialog.root }

func itemAmountPosition(win *nox_window) (x, y C.int) {
	var px, py C.uint
	C.nox_gui_getWindowOffs_46AA20(win, &px, &py)
	return C.int(px), C.int(py)
}

func mirrorItemAmountPointer(offset uintptr, value unsafe.Pointer) {
	if unsafe.Sizeof(value) == 4 {
		*g2ServerWord(offset) = uint32(uintptr(value))
	}
}

func readItemAmount() uint32 {
	return uint32(parseLegacyWideDecimal(C.sub_46AF00(itemAmountDialog.amountInput)))
}

//export nox_gui_item_amount_state
func nox_gui_item_amount_state(amount, maxAmount *C.uint32_t) C.int {
	if amount != nil {
		*amount = 0
	}
	if maxAmount != nil {
		*maxAmount = 0
	}
	if !itemAmountDialog.active || itemAmountDialog.amountInput == nil {
		return 0
	}
	if amount != nil {
		*amount = C.uint32_t(readItemAmount())
	}
	if maxAmount != nil {
		*maxAmount = C.uint32_t(itemAmountDialog.maxAmount)
	}
	return 1
}

func setItemAmountNumber(win *nox_window, offset uintptr, value uint32) {
	text := g2ServerMem(offset)
	WStrCopy((*wchar2_t)(text), 16, strconv.FormatInt(int64(int32(value)), 10))
	g2ServerText(win, 16385, text, 0)
}

func writeItemAmount(amount uint32) {
	setItemAmountNumber(itemAmountDialog.amountInput, 1319164, amount)
	if itemAmountDialog.priceMode != 0 {
		setItemAmountNumber(itemAmountDialog.priceLabel, 1319068, itemAmountDialog.unitPrice*amount)
	}
}

//export nox_gui_item_amount_copy_attrs
func nox_gui_item_amount_copy_attrs(item *nox_drawable, attrs *C.nox_item_amount_const_attrs_t) {
	if item == nil || attrs == nil {
		return
	}
	fields := drawableItemData(item)
	fields.Field_108, fields.Field_109 = attrs.modifiers[0], attrs.modifiers[1]
	fields.Field_110, fields.Field_111 = attrs.modifiers[2], attrs.modifiers[3]
	fields.Field_112_0 = int16(attrs.field_16)
	fields.Field_112_2 = int16(attrs.field_16 >> 16)
}

//export nox_gui_item_amount_call_callback
func nox_gui_item_amount_call_callback(callback C.nox_gui_item_amount_callback_t, pos *C.int2, amount C.uint32_t) {
	if callback != nil {
		C.item_amount_call(callback, pos, C.uint32_t(*g2ServerWord(1319244)), C.uint32_t(*g2ServerWord(1319240)),
			amount, C.uint32_t(*g2ServerWord(1319252)))
	}
}

func invokeItemAmount(callback C.nox_gui_item_amount_callback_t, amount uint32) {
	x, y := itemAmountPosition(itemAmountRoot())
	pos := C.int2{field_0: x + C.int(legacyWords.dword_587000_183456), field_4: y + C.int(legacyWords.dword_587000_183460)}
	nox_gui_item_amount_call_callback(callback, &pos, C.uint32_t(amount))
}

//export sub_4BFD40
func sub_4BFD40() {
	root := itemAmountRoot()
	if itemAmountDialog.active {
		g2ServerHidden(root, true)
		g2ServerEnabled(root, false)
		nox_xxx_wnd_46C6E0(root)
		deleteItemAmountDrawable()
		itemAmountDialog.active = false
		return
	}
	g2ServerEnabled(root, true)
	sub_46C690(root)
	nox_xxx_wndShowModalMB_46A8C0(root)
	itemAmountDialog.active = true
}

//export sub_4BFDD0
func sub_4BFDD0(win *nox_window, event C.int, packed, unused C.uintptr_t) C.int {
	switch event {
	case 5, 9, 13:
		if !C.nox_xxx_wndPointInWnd_46AAB0(win, C.int(uint16(packed)), C.int(uint32(packed)>>16)) {
			sub_4BFE40()
		}
		return 1
	case 6, 7, 10, 11, 14, 15:
		return 1
	default:
		return 0
	}
}

//export sub_4BFE40
func sub_4BFE40() C.int {
	if !itemAmountDialog.active {
		return 0
	}
	invokeItemAmount(itemAmountDialog.cancel, min(readItemAmount(), itemAmountDialog.maxAmount))
	sub_4BFD40()
	return 1
}

// ItemAmountCancel4BFE40 closes an active quantity dialog through its cancel callback.
func ItemAmountCancel4BFE40() int {
	return int(sub_4BFE40())
}

//export nox_gui_itemAmount_init_4BFEF0
func nox_gui_itemAmount_init_4BFEF0() C.int {
	itemAmountDialog.priceMode = 0
	root := nox_new_window_from_file(internCStr("MultMove.wnd"), C.sub_4C01C0)
	itemAmountDialog.root = root
	if root == nil {
		return 0
	}
	nox_window_set_all_funcs(root, C.sub_4BFDD0, C.sub_4C0030, nil)
	itemAmountDialog.amountInput = g2ServerChild(root, 3601)
	itemAmountDialog.priceLabel = g2ServerChild(root, 3607)
	g2ServerHidden(root, true)
	g2ServerEnabled(root, false)

	*g2ServerWord(1319108), *g2ServerWord(1319112) = 0, 0
	*g2ServerWord(1319116), *g2ServerWord(1319120) = uint32(C.nox_win_width), uint32(C.nox_win_height)
	*g2ServerWord(1319140), *g2ServerWord(1319144) = uint32(C.nox_win_width), uint32(C.nox_win_height)
	*g2ServerWord(1319124), *g2ServerWord(1319128) = 0, 0

	names := [...]string{
		"MultiMoveBase", "MultiMoveUpLit", "MultiMoveDownLit", "MultiMoveYesPressed",
		"MultiMoveNoPressed", "MultiMoveBaseNoTag", "MultiMoveYesPressedNoTag", "MultiMoveNoPressedNoTag",
	}
	for i, name := range names {
		img := Nox_xxx_gLoadImg(name).C()
		itemAmountDialog.images[i] = (*nox_video_bag_image_t)(img)
		mirrorItemAmountPointer(1319196+4*uintptr(i), unsafe.Pointer(img))
	}
	return 1
}

func packedGUIViewport(offset uintptr) noxrender.Viewport {
	words := (*[13]uint32)(g2ServerMem(offset))
	pair := func(i int) image.Point {
		return image.Pt(int(int32(words[i])), int(int32(words[i+1])))
	}
	return noxrender.Viewport{
		Screen:   image.Rectangle{Min: pair(0), Max: pair(2)},
		World:    image.Rectangle{Min: pair(4), Max: pair(6)},
		Size:     pair(8),
		Field10:  uint(words[10]),
		Field11:  uint(words[11]),
		Jiggle12: int(int32(words[12])),
	}
}

//export sub_4C0030
func sub_4C0030(win *nox_window, draw unsafe.Pointer) C.int {
	C.nox_client_drawRectFilledAlpha_49CF10(0, 0, C.nox_win_width, C.nox_win_height)
	x, y := itemAmountPosition(win)
	drawImage := func(index int) {
		nox_client_drawImageAt_47D2C0(itemAmountDialog.images[index], int32(x), int32(y))
	}
	base, yes, no := itemAmountImageBaseWithoutTag, itemAmountImageYesWithoutTag, itemAmountImageNoWithoutTag
	if itemAmountDialog.priceMode != 0 {
		base, yes, no = itemAmountImageBase, itemAmountImageYes, itemAmountImageNo
	}
	drawImage(base)
	item := asDrawable(itemAmountDialog.drawable)
	item.SetPos(image.Pt(int(int32(x+C.int(legacyWords.dword_587000_183456))), int(int32(y+C.int(legacyWords.dword_587000_183460)))))
	viewport := packedGUIViewport(1319108)
	item.CallDraw(&viewport)

	root := itemAmountRoot()
	overlays := [...]struct{ id, image int }{{3603, itemAmountImageDown}, {3602, itemAmountImageUp}, {3604, yes}, {3605, no}}
	for _, part := range overlays {
		if g2ServerChild(root, part.id).draw_data.field_0&4 != 0 {
			drawImage(part.image)
		}
	}
	return 1
}

//export sub_4C01C0
func sub_4C01C0(win *nox_window, event C.int, control *nox_window, arg C.uintptr_t) C.int {
	if event != 16391 || !itemAmountDialog.active {
		return 0
	}
	id := C.nox_xxx_wndGetID_46B0A0(control)
	nox_xxx_clientPlaySoundSpecial_452D80(766, 100)
	switch id {
	case 3602:
		if amount := readItemAmount() + 1; amount <= itemAmountDialog.maxAmount {
			writeItemAmount(amount)
		}
	case 3603:
		if amount := readItemAmount(); amount > 1 {
			writeItemAmount(amount - 1)
		}
	case 3604, 3606:
		invokeItemAmount(itemAmountDialog.accept, min(readItemAmount(), itemAmountDialog.maxAmount))
		sub_4BFD40()
	case 3605:
		sub_4BFE40()
	}
	return 0
}

func deleteItemAmountDrawable() {
	if dr := itemAmountDialog.drawable; dr != nil {
		nox_xxx_spriteDelete_45A4B0(dr)
	}
	itemAmountDialog.drawable = nil
}

//export nox_gui_itemAmount_free_4C03E0
func nox_gui_itemAmount_free_4C03E0() {
	deleteItemAmountDrawable()
	nox_xxx_windowDestroyMB_46C4E0(itemAmountRoot())
	itemAmountDialog.root, itemAmountDialog.amountInput, itemAmountDialog.priceLabel = nil, nil, nil
	itemAmountDialog.priceMode = 0
	itemAmountDialog.active = false
}

//export nox_gui_itemAmountDialog_4C0430
func nox_gui_itemAmountDialog_4C0430(title *wchar2_t, x, y C.int, id, thing C.uint32_t, attrs *C.nox_item_amount_const_attrs_t, maxAmount, extra C.uint32_t, accept, cancel C.nox_gui_item_amount_callback_t) C.int {
	g2ServerText(g2ServerChild(itemAmountRoot(), 3606), 16385, unsafe.Pointer(title), 0)
	item := nox_new_drawable_for_thing(int32(thing))
	itemAmountDialog.drawable = item
	item.flags30 |= 0x40000000
	nox_gui_item_amount_copy_attrs(item, attrs)
	itemAmountDialog.accept = accept
	itemAmountDialog.cancel = cancel
	mirrorItemAmountPointer(1319160, unsafe.Pointer(accept))
	mirrorItemAmountPointer(1319100, unsafe.Pointer(cancel))
	*g2ServerWord(1319240), *g2ServerWord(1319244) = uint32(thing), uint32(id)
	itemAmountDialog.maxAmount = uint32(maxAmount)
	*g2ServerWord(1319252) = uint32(extra)
	sub_4BFD40()
	sub_4C0560(x, y)
	writeItemAmount(1)
	if itemAmountDialog.priceMode == 0 {
		text := g2ServerMem(1319068)
		C.item_amount_format_tag((*wchar2_t)(text), (*wchar2_t)(g2ServerMem(1319272)))
		g2ServerText(itemAmountDialog.priceLabel, 16385, text, 0)
	}
	return 0
}

//export sub_4C0560
func sub_4C0560(x, y C.int) C.int {
	root := itemAmountRoot()
	var width, height C.int
	C.nox_window_get_size(root, &width, &height)
	x = max(0, x-C.int(legacyWords.dword_587000_183456))
	y = max(0, y-C.int(legacyWords.dword_587000_183460))
	if x+width >= C.nox_win_width {
		x = C.nox_win_width - width
	}
	if y+height >= C.nox_win_height {
		y = C.nox_win_height - height
	}
	return C.int(nox_window_setPos_46A9B0(root, int32(x), int32(y)))
}

//export sub_4C05F0
func sub_4C05F0(priceMode, unitPrice C.int) C.int {
	itemAmountDialog.priceMode = uint32(priceMode)
	itemAmountDialog.unitPrice = uint32(unitPrice)
	return priceMode
}
