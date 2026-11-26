package legacy

/*
#include <stdint.h>
#include "GAME1.h"
#include "GAME1_3.h"
#include "GAME2.h"
#include "GAME2_2.h"
#include "GAME3_2.h"
#include "GAME5_2.h"
#include "common__system__settings.h"
extern void* dword_587000_127004;
extern void* dword_587000_93164;
*/
import "C"

var (
	WriteConfigLegacy func(name string)
)

//export nox_common_writecfgfile
func nox_common_writecfgfile(str *C.char) {
	WriteConfigLegacy(GoString(str))
}
