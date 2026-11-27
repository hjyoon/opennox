package legacy

/*
#include "GAME2_3.h"
*/
import "C"
import (
	"github.com/opennox/opennox/v1/client/noxrender"
)

func drawClipRectCEntry49F6F0(data *noxrender.RenderData, x, y, w, h int32) int32 {
	saved := currentDrawData
	Set_nox_draw_curDrawData_3799572(data)
	defer func() { currentDrawData = saved }()
	return int32(C.nox_client_copyRect_49F6F0(C.int(x), C.int(y), C.int(w), C.int(h)))
}

func drawClipHorizontalCEntry49F780(left, right int32) int32 {
	return int32(C.sub_49F780(C.int(left), C.int(right)))
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
