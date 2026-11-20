package legacy

/*
#include "defs.h"
char nox_xxx_unitHasCollideOrUpdateFn_537610(nox_object_t* a1p);
void nox_xxx_unitNeedSync_4E44F0(nox_object_t* a1);
int* sub_4E4500(nox_object_t* a1, int a2, int a3, int a4);
nox_object_t* nox_xxx_findObjectAtCursor_54AF40(nox_object_t* a1);
*/
import "C"
import "github.com/opennox/opennox/v1/server"

func objectNeedSyncC(obj *server.Object) {
	C.nox_xxx_unitNeedSync_4E44F0(asObjectC(obj))
}

func objectStatusMaskC(obj *server.Object, val1, val2 uint32, set bool) {
	defer pinObjectForC(obj)()
	var cset C.int
	if set {
		cset = 1
	}
	C.sub_4E4500(asObjectC(obj), C.int(val1), C.int(val2), cset)
}

func Nox_xxx_findObjectAtCursor_54AF40(a1 *server.Object) *server.Object {
	return asObjectS(C.nox_xxx_findObjectAtCursor_54AF40(asObjectC(a1)))
}
