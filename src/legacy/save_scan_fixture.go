package legacy

/*
#include "GAME3_2.h"
*/
import "C"

func countSaveFilesC() int {
	return int(C.nox_client_countSaveFiles_4DC550())
}

func countPlayerFilesC() int {
	return int(C.nox_client_countPlayerFiles02_4DC630())
}
