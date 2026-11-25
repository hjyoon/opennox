package legacy

/*
#include "GAME2_3.h"

extern nox_render_data_t* nox_draw_curDrawData_3799572;
#include "client_ui_words.h"
*/
import "C"

//export sub_49F6D0
func sub_49F6D0(flag C.int) C.int {
	data := C.nox_draw_curDrawData_3799572
	previous := data.flag_0
	data.flag_0 = flag
	return C.int(previous)
}

func currentDrawClipRect(x, y, width, height C.int) C.int {
	return nox_client_copyRect_native_49F6F0(C.nox_draw_curDrawData_3799572, x, y, width, height)
}

//export nox_client_clipHorizontal_49F780
func nox_client_clipHorizontal_49F780(left, right C.int) C.int {
	clip := C.nox_draw_curDrawData_3799572.clip
	left = max(left, C.int(clip.min_x))
	right = min(right, C.int(clip.max_x))
	top, bottom := C.int(clip.min_y), C.int(clip.max_y)
	return currentDrawClipRect(left, top, right-left, bottom-top)
}

func saveDrawClipBounds(offset uintptr, bounds C.nox_rect) {
	words := (*[4]uint32)(g2ServerMem(offset))
	*words = [4]uint32{uint32(bounds.min_x), uint32(bounds.min_y), uint32(bounds.max_x), uint32(bounds.max_y)}
}

func loadDrawClipBounds(offset uintptr) C.nox_rect {
	words := (*[4]int32)(g2ServerMem(offset))
	return C.nox_rect{min_x: C.intptr_t(words[0]), min_y: C.intptr_t(words[1]), max_x: C.intptr_t(words[2]), max_y: C.intptr_t(words[3])}
}

//export nox_xxx_wndDraw_49F7F0
func nox_xxx_wndDraw_49F7F0() {
	if C.dword_5d4594_1305748 != 0 {
		return
	}
	data := C.nox_draw_curDrawData_3799572
	*g2ServerWord(1305772) = uint32(data.flag_0)
	saveDrawClipBounds(1305756, data.clip)
	saveDrawClipBounds(1305732, data.rect2)
	C.dword_5d4594_1305748 = 1
}

//export sub_49F860
func sub_49F860() C.int {
	if C.dword_5d4594_1305748 == 0 {
		return 0
	}
	data := C.nox_draw_curDrawData_3799572
	data.flag_0 = C.int(*g2ServerWord(1305772))
	data.clip = loadDrawClipBounds(1305756)
	data.rect2 = loadDrawClipBounds(1305732)
	C.dword_5d4594_1305748 = 0
	return C.int(data.rect2.max_x)
}

//export nox_xxx_utilRect_49F930
func nox_xxx_utilRect_49F930(out, a, b *C.int4) *C.int4 {
	left, right := max(a.field_0, b.field_0), min(a.field_8, b.field_8)
	top, bottom := max(a.field_4, b.field_4), min(a.field_C, b.field_C)
	if left >= right || top >= bottom {
		return nil
	}
	*out = C.int4{field_0: left, field_4: top, field_8: right, field_C: bottom}
	return out
}
