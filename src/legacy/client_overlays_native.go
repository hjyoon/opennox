package legacy

/*
#include "defs.h"
#include "noxstring.h"


static inline void nox_format_overlay_map(wchar2_t* dst, const char* map) {
	nox_swprintf(dst, L"%S", map);
}
static inline void nox_format_overlay_coordinates(wchar2_t* dst, int x, int y) {
	nox_swprintf(dst, L"X:%d\tY:%d", x, y);
}
static inline void nox_format_overlay_player(wchar2_t* dst, const wchar2_t* format, int value, const wchar2_t* class_name) {
	nox_swprintf(dst, format, value, class_name);
}
*/
import "C"

import (
	"unsafe"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

const clientOverlayStrings = "C:\\NoxPost\\src\\client\\System\\client.c"

//export nox_xxx_clientDrawAll_436100_draw_A
func nox_xxx_clientDrawAll_436100_draw_A() {
	if sub_436550() != 0 {
		noxflags.SetEngine(noxflags.EngineFlag9)
	} else {
		noxflags.UnsetEngine(noxflags.EngineFlag9)
	}
	image := memmap.PtrPtr(0x5D4594, 814540)
	if *image == nil {
		*image = unsafe.Pointer(nox_xxx_gLoadImg_42F970(internCStr("MenuSystemBG")))
	}
	if !noxflags.HasEngine(noxflags.EngineFlag9) {
		return
	}
	font := nox_xxx_guiFontPtrByName_43F360(internCStr("large"))
	nox_client_drawImageAt_47D2C0((*nox_video_bag_image_t)(*image), 0, 0)
	title := nox_strman_loadString_40F1D0(internCStr("InProgress"), nil, internCStr(clientOverlayStrings), 1550)
	var width C.int
	nox_xxx_drawGetStringSize_43F840(font, title, &width, nil, 0)
	nox_xxx_drawSetTextColor_434390(int32(Get_nox_color_white_2523948()))
	x := int32(clientDisplayStorage.width-width) / 2
	y := int32(clientDisplayStorage.height) / 2
	title = nox_strman_loadString_40F1D0(internCStr("InProgress"), nil, internCStr(clientOverlayStrings), 1553)
	nox_xxx_drawString_43F6E0(font, title, x, y)
}

func overlayLineBreak(word C.wchar2_t) bool {
	return word == '\n' || word == '\r'
}

//export nox_xxx_clientDrawAll_436100_draw_B
func nox_xxx_clientDrawAll_436100_draw_B() {
	style := memmap.Uint32(0x5D4594, 811060)
	if style >= 2 {
		style = 0
	}
	offset := 4 * uintptr(style)
	image := memmap.PtrPtr(0x5D4594, 811888+offset)
	if *image == nil {
		name := (*C.char)(*memmap.PtrPtr(0x587000, 85712+offset))
		*image = unsafe.Pointer(nox_xxx_gLoadImg_42F970(name))
	}
	x, y := (int32(clientDisplayStorage.width)-310)/2, (int32(clientDisplayStorage.height)-200)/2
	nox_client_drawImageAt_47D2C0((*nox_video_bag_image_t)(*image), x, y)

	var text [264]C.wchar2_t
	stored := memmap.PtrT[[264]C.wchar2_t](0x5D4594, 811376)
	copy(text[:len(text)-1], stored[:])
	length := alloc.StrLenN(&text[0], len(text))
	var width, height C.int
	nox_xxx_drawGetStringSize_43F840(nil, &text[0], &width, &height, 220)
	x += 45
	y += (49-int32(height))/2 + 143
	nox_client_drawSetColor_434460(int32(Get_nox_color_white_2523948()))
	for next := 0; next < length; {
		for next < length && overlayLineBreak(text[next]) {
			next++
		}
		start := next
		for next < length && !overlayLineBreak(text[next]) {
			next++
		}
		if next == start {
			break
		}
		text[next] = 0
		nox_xxx_drawSetTextColor_434390(int32(Get_nox_color_white_2523948()))
		nox_xxx_drawGetStringSize_43F840(nil, &text[start], &width, nil, 0)
		nox_xxx_drawStringWrap_43FAF0(nil, &text[start], x+(220-int32(width))/2, y, 220, 0)
		y += nox_xxx_guiFontHeightMB_43F320(nil)
		next++
	}
}

//export sub_436F50
func sub_436F50() C.int {
	fontHeight := nox_xxx_guiFontHeightMB_43F320(nil)
	view := nox_draw_getViewport_437250()
	x, y := int32(view.x1)+10, int32(view.y1)+90
	nox_xxx_drawSetTextColor_434390(int32(Get_nox_color_white_2523948()))
	text := (*C.wchar2_t)(memmap.PtrOff(0x5D4594, 811120))
	C.nox_format_overlay_map(text, nox_server_currentMapGetFilename_409B30())
	nox_xxx_drawString_43F6E0(nil, text, x, y)
	local := (*nox_drawable)(*memmap.PtrPtr(0x852978, 8))
	if local == nil {
		return 0
	}
	player := g2qPlayer()
	if player == nil {
		return 1
	}
	y += fontHeight
	C.nox_format_overlay_coordinates(text, C.int(local.pos.x), C.int(local.pos.y))
	nox_xxx_drawString_43F6E0(nil, text, x, y)
	classID := uintptr(game2PlayerClass(player))
	name := (*C.char)(*memmap.PtrPtr(0x587000, 29456+4*classID))
	source := internCStr(clientOverlayStrings)
	class := nox_strman_loadString_40F1D0(name, nil, source, 1357)
	format := nox_strman_loadString_40F1D0(internCStr("PlayerInfo"), nil, source, 1355)
	C.nox_format_overlay_player(text, format, C.int(player.field_3684), class)
	return C.int(nox_xxx_drawString_43F6E0(nil, text, x, y+fontHeight))
}

//export nox_xxx_getRandomName_4358A0
func nox_xxx_getRandomName_4358A0() *C.char {
	names := [...]string{
		"Dweezle", "Glork", "Floogle", "Goombah", "Kraun", "Kloog", "Zurg", "Darg",
		"Arfingle", "Buurl", "Gurgin", "Grok", "Hurlong", "Luric", "Lupis", "Mallik",
		"Thrall", "Norwood", "Nulik", "Orin", "Olaf", "Orguk", "Pervis", "Paavik",
		"Qix", "Xevin", "Xurcon", "Markoan", "Yuric", "Yoovis", "Yalek", "Zug", "Zivik",
	}
	count := memmap.PtrUint32(0x5D4594, 814516)
	if *count == 0 || *count > uint32(len(names)) {
		*count = uint32(len(names))
	}
	index := nox_common_randomIntMinMax_415FF0(0, int32(*count)-1, internCStr(clientOverlayStrings), 559)
	return internCStr(names[index])
}
