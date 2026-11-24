package legacy

/*
#include "GAME2_2.h"
#include "GAME3_2.h"
int nox_xxx_mapLoadRequired_4DCC80(void);
*/
import "C"

func callSaveGateC() int32 {
	return int32(C.nox_xxx_game_4DCCB0())
}

func setMapLoadFlagC(value int32) int32 {
	return int32(C.nox_xxx_mapLoadOrSaveMB_4DCC70(C.int(value)))
}

func mapLoadFlagC() int32 {
	return int32(C.nox_xxx_mapLoadRequired_4DCC80())
}

func saveGateCursorC() int32 {
	return int32(C.nox_xxx_guiCursor_477600())
}
