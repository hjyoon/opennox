package legacy

/*
#include "defs.h"
void nox_xxx_updateFallLogic_51B870(nox_object_t* a1);
char nox_xxx_unitHasCollideOrUpdateFn_537610(nox_object_t* a1p);
void nox_xxx_unitNeedSync_4E44F0(nox_object_t* a1);
int* sub_4E4500(nox_object_t* a1, int a2, int a3, int a4);
void sub_51B810(nox_object_t* a1);
void sub_537770(nox_object_t* a1);
nox_object_t* nox_xxx_findObjectAtCursor_54AF40(nox_object_t* a1);
*/
import "C"
import "github.com/opennox/opennox/v1/server"

func objectNeedSyncC(obj *server.Object) {
	C.nox_xxx_unitNeedSync_4E44F0(asObjectC(obj))
}

func objectStatusMaskC(obj *server.Object, val1, val2 uint32, set bool) {
	var cset C.int
	if set {
		cset = 1
	}
	C.sub_4E4500(asObjectC(obj), C.int(val1), C.int(val2), cset)
}

func Nox_xxx_findObjectAtCursor_54AF40(a1 *server.Object) *server.Object {
	return asObjectS(C.nox_xxx_findObjectAtCursor_54AF40(asObjectC(a1)))
}
func Nox_xxx_updateFallLogic_51B870(a1 *server.Object) {
	C.nox_xxx_updateFallLogic_51B870(asObjectC(a1))
}
func Sub_51B810(a1 *server.Object) {
	C.sub_51B810(asObjectC(a1))
}
func Sub_537770(a1 *server.Object) {
	C.sub_537770(asObjectC(a1))
}
