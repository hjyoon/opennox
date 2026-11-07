package legacy

/*
#include "GAME1.h"
*/
import "C"

//export sub_409B80
func sub_409B80() *C.char {
	return (*C.char)(sub409B80Native())
}
