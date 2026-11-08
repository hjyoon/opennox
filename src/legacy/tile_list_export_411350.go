package legacy

/*
#include "GAME1.h"
*/
import "C"
import "unsafe"

//export sub_411350
func sub_411350(head *C.nox_tile_list_node_t, xy *C.int, initial C.int) C.int {
	return C.int(sub411350Native(unsafe.Pointer(head), unsafe.Pointer(xy), int32(initial)))
}

//export sub_4113A0
func sub_4113A0(xy *C.int, selector C.int) C.int {
	return C.int(sub4113A0Native(unsafe.Pointer(xy), int32(selector)))
}

//export sub_411490
func sub_411490(edge, variation C.int) C.int {
	return C.int(sub411490Native(int32(edge), int32(variation)))
}
