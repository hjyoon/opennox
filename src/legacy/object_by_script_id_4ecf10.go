package legacy

/*
#include <stdint.h>
typedef struct nox_object_t nox_object_t;
*/
import "C"

//export sub_4ECF10
func sub_4ECF10(scriptID C.int32_t) *C.nox_object_t {
	obj := GetServer().S().ObjectByScriptID4ECF10(int32(scriptID))
	return (*C.nox_object_t)(asObjectC(obj))
}
