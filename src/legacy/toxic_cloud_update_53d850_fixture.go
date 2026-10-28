package legacy

/*
#include "defs.h"
#include "GAME4_3.h"
static void nox_toxic_cloud_native_probe(int kind, nox_object_t* cloud, nox_object_t* target) {
 switch (kind) {
 case 0: ((void(*)(nox_object_t*))nox_xxx_updateToxicCloud_53D850)(cloud);break;
 case 1: ((void(*)(nox_object_t*,nox_object_t*))sub_53D8C0)(target,cloud);break;
 case 2: ((void(*)(nox_object_t*))nox_xxx_updateSmallToxicCloud_53D960)(cloud);break;
 case 3: ((void(*)(nox_object_t*,nox_object_t*))nox_xxx_toxicCloudPoison_53D9D0)(target,cloud);break;
 }
}
static int nox_toxic_cloud_damage_probe0(nox_object_t* t,nox_object_t* s,nox_object_t* w,int32_t d,int32_t typ) {
 (void)t;(void)s;(void)w;(void)d;(void)typ;return 0;
}
static int nox_toxic_cloud_damage_probe1(nox_object_t* t,nox_object_t* s,nox_object_t* w,int32_t d,int32_t typ) {
 (void)t;(void)s;(void)w;(void)d;(void)typ;return 1;
}
static void* nox_toxic_cloud_damage_probe_ptr(int index) {
 return index==0 ? nox_toxic_cloud_damage_probe0 : nox_toxic_cloud_damage_probe1;
}
*/
import "C"

import "unsafe"

import "github.com/opennox/opennox/v1/server"

func toxicCloudNativeProbe53D850(kind int, cloud, target *server.Object) {
	C.nox_toxic_cloud_native_probe(C.int(kind), asObjectC(cloud), asObjectC(target))
}

func toxicCloudDamageProbePtr53D850(index int) unsafe.Pointer {
	return C.nox_toxic_cloud_damage_probe_ptr(C.int(index))
}
