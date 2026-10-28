package legacy

/*
#include "defs.h"
#include "memmap.h"
#include "GAME4_3.h"
static uint32_t nox_arachnaphobia_cache_probe(void) { return *getMemU32Ptr(0x5D4594, 2488708); }
static void nox_state_update_native_probe(int kind,nox_object_t* source) {
 switch(kind) {
 case 0:((void(*)(nox_object_t*))nox_xxx_updateArachnaphobia_53DA60)(source);break;
 case 1:((void(*)(nox_object_t*))nox_xxx_updateBreak_53DB30)(source);break;
 case 2:((void(*)(nox_object_t*))nox_xxx_updateOpen_53DBB0)(source);break;
 case 3:((void(*)(nox_object_t*))nox_xxx_updateBreakAndRemove_53DC30)(source);break;
 }
}
*/
import "C"

import "github.com/opennox/opennox/v1/server"

func stateUpdateNativeProbe53DA60(kind int, source *server.Object) {
	C.nox_state_update_native_probe(C.int(kind), asObjectC(source))
}

func arachnaphobiaCacheProbe53DA60() uint32 { return uint32(C.nox_arachnaphobia_cache_probe()) }
