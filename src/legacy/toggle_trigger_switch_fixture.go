package legacy

/*
#include "defs.h"
#include "GAME4_3.h"
static uint32_t nox_tts_cache_probe(int n) { return *getMemU32Ptr(0x5D4594, 2488680-4*n); }
static unsigned char nox_tts_probe(int kind,nox_object_t* source) {
 switch(kind) {
 case 0:return (unsigned char)((char(*)(nox_object_t*))nox_xxx_updateToggle_53B060)(source);
 case 1:return (unsigned char)((char(*)(nox_object_t*))nox_xxx_updateTrigger_53B1B0)(source);
 default:return (unsigned char)((char(*)(nox_object_t*))nox_xxx_updateSwitch_53B320)(source);
 }
}
*/
import "C"
import "github.com/opennox/opennox/v1/server"

func toggleTriggerSwitchProbe(kind int, source *server.Object) uint8 {
	return uint8(C.nox_tts_probe(C.int(kind), asObjectC(source)))
}

func toggleTriggerCacheProbe(n int) uint32 { return uint32(C.nox_tts_cache_probe(C.int(n))) }
