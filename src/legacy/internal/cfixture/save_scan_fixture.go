package cfixture

/*
#include "GAME3_2.h"
*/
import "C"

func CountSaveFilesC() int {
	return int(C.nox_client_countSaveFiles_4DC550())
}

func CountPlayerFilesC() int {
	return int(C.nox_client_countPlayerFiles02_4DC630())
}
