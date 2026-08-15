package legacy

/*
typedef struct nox_object_t nox_object_t;
*/
import "C"

//export nox_server_getObjectFromNetCode_4ECCB0
func nox_server_getObjectFromNetCode_4ECCB0(code C.int) *C.nox_object_t {
	obj := GetServer().S().ObjectFromNetCode4ECCB0(uint32(code))
	return (*C.nox_object_t)(asObjectC(obj))
}
