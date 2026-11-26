package legacy

/*
#include "defs.h"
#include "GAME2_3.h"
#include "client_ui_words.h"
static uintptr_t draw_clip_rect_test_49F6F0(int x, int y, int w, int h) {
	return (uintptr_t)nox_client_copyRect_49F6F0(x, y, w, h);
}
*/
import "C"
import (
	"github.com/opennox/opennox/v1/client/noxrender"
)

func drawClipRectCEntry49F6F0(data *noxrender.RenderData, x, y, w, h int32) uintptr {
	saved := currentDrawData
	Set_nox_draw_curDrawData_3799572(data)
	defer func() { currentDrawData = saved }()
	return uintptr(C.draw_clip_rect_test_49F6F0(C.int(x), C.int(y), C.int(w), C.int(h)))
}

func useDrawClipTestData(data *noxrender.RenderData) func() {
	savedData, savedFlag := currentDrawData, drawClipSaved
	words := (*[11]uint32)(g2ServerMem(1305732))
	savedWords := *words
	Set_nox_draw_curDrawData_3799572(data)
	drawClipSaved = 0
	return func() {
		currentDrawData, drawClipSaved = savedData, savedFlag
		*words = savedWords
	}
}
