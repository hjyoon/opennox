package legacy

/*
#include <stdint.h>
typedef struct nox_object_t nox_object_t;
*/
import "C"

//export nox_xxx_netGetUnitByExtent_4ED020
func nox_xxx_netGetUnitByExtent_4ED020(extent C.uint32_t) *C.nox_object_t {
	obj := GetServer().S().ObjectByExtent4ED020(uint32(extent))
	return (*C.nox_object_t)(asObjectC(obj))
}
