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

//export nox_xxx_packetDynamicUnitCode_578B40
func nox_xxx_packetDynamicUnitCode_578B40(code C.int) C.int {
	if code&0x8000 == 0 {
		return code
	}
	obj := GetServer().S().ObjectByExtent4ED020(uint32(code) &^ 0x8000)
	if obj == nil {
		return 0
	}
	return C.int(obj.NetCode)
}
