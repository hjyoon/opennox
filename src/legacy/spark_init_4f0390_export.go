package legacy

/*
#include "spark_init_4f0390.h"
*/
import "C"
import "unsafe"

//export nox_xxx_unitSparkInit_4F0390
func nox_xxx_unitSparkInit_4F0390(unit *C.nox_object_t) *C.nox_spark_update_data_t {
	update := sparkInitCall4F0390(asObjectS((*nox_object_t)(unit)))
	return (*C.nox_spark_update_data_t)(unsafe.Pointer(update))
}
