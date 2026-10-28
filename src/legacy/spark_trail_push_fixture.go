package legacy

/*
#include "defs.h"
#include "GAME4_3.h"
static nox_object_t* nox_spark_trail_push_probe(int kind,nox_object_t* source) {
 switch(kind) {
 case 0: { void (*call)(nox_object_t*)=nox_xxx_updateSpark_53ADC0;call(source);return 0; }
 case 1: { nox_object_t* (*call)(nox_object_t*)=nox_xxx_updateProjTrail_53AEC0;return call(source); }
 default: { void (*call)(nox_object_t*)=nox_xxx_updatePush_53B030;call(source);return 0; }
 }
}
*/
import "C"
import (
	"github.com/opennox/opennox/v1/server"
	"unsafe"
)

func sparkTrailPushProbe(kind int, source *server.Object) *server.Object {
	return (*server.Object)(unsafe.Pointer(C.nox_spark_trail_push_probe(C.int(kind), asObjectC(source))))
}
