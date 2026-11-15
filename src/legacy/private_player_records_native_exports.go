package legacy

/*
#include "defs.h"
*/
import "C"
import "unsafe"

//export sub_4259C0
func sub_4259C0() { privatePlayerRecords4259C0.init() }

//export sub_4259F0
func sub_4259F0() *C.int {
	return (*C.int)(unsafe.Pointer(privatePlayerRecords4259C0.freeAll()))
}

//export sub_425A50
func sub_425A50() *C.int {
	return (*C.int)(unsafe.Pointer(privatePlayerRecords4259C0.first()))
}

//export sub_425A60
func sub_425A60(p *C.int) *C.int {
	return (*C.int)(unsafe.Pointer(privatePlayerRecordNext425A60((*privatePlayerRecord425AD0)(unsafe.Pointer(p)))))
}

//export sub_425A70
func sub_425A70(id C.int) *C.int {
	return (*C.int)(unsafe.Pointer(privatePlayerRecords4259C0.find(int32(id))))
}

//export sub_425AA0
func sub_425AA0(id C.int) C.int {
	if privatePlayerRecords4259C0.find(int32(id)) != nil {
		return 1
	}
	return 0
}

//export sub_425AD0
func sub_425AD0(id C.int, name *C.ushort) *C.ushort {
	return (*C.ushort)(unsafe.Pointer(privatePlayerRecords4259C0.create(int32(id), (*uint16)(unsafe.Pointer(name)))))
}

//export sub_425B30
func sub_425B30(p unsafe.Pointer, peer C.int) {
	privatePlayerRecords4259C0.addChild((*privatePlayerRecord425AD0)(p), int32(peer))
}

//export sub_425B60
func sub_425B60(p unsafe.Pointer, peer C.int) *C.char {
	return (*C.char)(privatePlayerRecords4259C0.removeChild((*privatePlayerRecord425AD0)(p), int32(peer)))
}

//export sub_425BC0
func sub_425BC0(p unsafe.Pointer) *C.int {
	return (*C.int)(unsafe.Pointer(privatePlayerFirstChild425BC0((*privatePlayerRecord425AD0)(p))))
}

//export sub_425BE0
func sub_425BE0(p *C.int) *C.int {
	return (*C.int)(unsafe.Pointer(privatePlayerNextChild425BE0((*privatePlayerChild425B30)(unsafe.Pointer(p)))))
}
