package legacy

/*
#include "defs.h"
#include "GAME4_1.h"
typedef struct { nox_object_t* values[16]; unsigned count; } chakram_rect_probe_4eb250;
static void chakram_rect_collect_4eb250(nox_object_t* object,void* context) {
 chakram_rect_probe_4eb250* out=context;
 if (out->count<16) out->values[out->count++]=object;
}
static void chakram_rect_probe_call_4eb250(float4* rect,chakram_rect_probe_4eb250* out) {
 nox_xxx_getUnitsInRect_517C10(rect,chakram_rect_collect_4eb250,out);
}
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/server"
)

func chakramRetargetProductionRectControl4EB250(rect types.Rectf) []*server.Object {
	var probe C.chakram_rect_probe_4eb250
	C.chakram_rect_probe_call_4eb250((*C.float4)(unsafe.Pointer(&rect)), &probe)
	out := make([]*server.Object, int(probe.count))
	for i := range out {
		out[i] = (*server.Object)(unsafe.Pointer(probe.values[i]))
	}
	return out
}
