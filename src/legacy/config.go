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
extern void* dword_587000_122852;
extern void* dword_587000_127004;
extern void* dword_587000_93164;
extern unsigned int nox_client_renderBubbles_80844;
extern uint32_t nox_server_kickQuestPlayerMinVotes_229992;
extern uint32_t nox_server_resetQuestMinVotes_229988;
*/
import "C"

var (
	WriteConfigLegacy func(name string)
)

//export nox_common_writecfgfile
func nox_common_writecfgfile(str *C.char) {
	WriteConfigLegacy(GoString(str))
}
