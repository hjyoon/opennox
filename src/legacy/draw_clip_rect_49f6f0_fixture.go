package legacy

/*
#include "defs.h"
#include "GAME2_3.h"
extern nox_render_data_t* nox_draw_curDrawData_3799572;
extern unsigned int dword_5d4594_1305748;
static uintptr_t draw_clip_rect_test_49F6F0(nox_render_data_t* data, int x, int y, int w, int h) {
	nox_render_data_t* saved = nox_draw_curDrawData_3799572;
	nox_draw_curDrawData_3799572 = data;
	uintptr_t result = (uintptr_t)nox_client_copyRect_49F6F0(x, y, w, h);
	nox_draw_curDrawData_3799572 = saved;
	return result;
}
*/
import "C"
import (
	"github.com/opennox/opennox/v1/client/noxrender"
)

func drawClipRectCEntry49F6F0(data *noxrender.RenderData, x, y, w, h int32) uintptr {
	return uintptr(C.draw_clip_rect_test_49F6F0((*C.nox_render_data_t)(data.C()), C.int(x), C.int(y), C.int(w), C.int(h)))
}

func useDrawClipTestData(data *noxrender.RenderData) func() {
	savedData, savedFlag := C.nox_draw_curDrawData_3799572, C.dword_5d4594_1305748
	words := (*[11]uint32)(g2ServerMem(1305732))
	savedWords := *words
	Set_nox_draw_curDrawData_3799572(data)
	C.dword_5d4594_1305748 = 0
	return func() {
		C.nox_draw_curDrawData_3799572, C.dword_5d4594_1305748 = savedData, savedFlag
		*words = savedWords
	}
}
