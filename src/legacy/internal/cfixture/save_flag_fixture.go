package cfixture

/*
#include "GAME2_2.h"
#include "GAME3_2.h"
int nox_xxx_mapLoadRequired_4DCC80(void);
*/
import "C"

func CallSaveGateC() int32 {
	return int32(C.nox_xxx_game_4DCCB0())
}

func SetMapLoadFlagC(value int32) int32 {
	return int32(C.nox_xxx_mapLoadOrSaveMB_4DCC70(C.int(value)))
}

func MapLoadFlagC() int32 {
	return int32(C.nox_xxx_mapLoadRequired_4DCC80())
}

func SaveGateCursorC() int32 {
	return int32(C.nox_xxx_guiCursor_477600())
}
