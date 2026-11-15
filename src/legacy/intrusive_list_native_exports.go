package legacy

/*
#include "defs.h"
#include <stdlib.h>
*/
import "C"
import "unsafe"

//export nox_common_list_clear_425760
func nox_common_list_clear_425760(node *C.nox_list_item_t) {
	intrusiveListClear425760((*intrusiveListNode425760)(unsafe.Pointer(node)))
}

//export sub_425770
func sub_425770(node unsafe.Pointer) unsafe.Pointer {
	return unsafe.Pointer(intrusiveListInit425770((*intrusiveListNode425760)(node)))
}

//export sub_425790
func sub_425790(head *C.int, node *C.uint32_t) C.int {
	return C.int(intrusiveListAscending425790((*intrusiveListNode425760)(unsafe.Pointer(head)), (*intrusiveListNode425760)(unsafe.Pointer(node))))
}

//export sub_4257F0
func sub_4257F0(head *C.int, node *C.uint32_t) {
	intrusiveListDescending4257F0((*intrusiveListNode425760)(unsafe.Pointer(head)), (*intrusiveListNode425760)(unsafe.Pointer(node)))
}

//export nox_common_list_getFirstSafe_425890
func nox_common_list_getFirstSafe_425890(node *C.nox_list_item_t) *C.nox_list_item_t {
	return (*C.nox_list_item_t)(unsafe.Pointer(intrusiveListFirstSafe425890((*intrusiveListNode425760)(unsafe.Pointer(node)))))
}

//export nox_common_list_getNextSafe_4258A0
func nox_common_list_getNextSafe_4258A0(node *C.nox_list_item_t) *C.nox_list_item_t {
	return (*C.nox_list_item_t)(unsafe.Pointer(intrusiveListNextSafe4258A0((*intrusiveListNode425760)(unsafe.Pointer(node)))))
}

//export sub_4258C0
func sub_4258C0(head **C.uint32_t, index C.int) *C.uint32_t {
	return (*C.uint32_t)(unsafe.Pointer(intrusiveListAt4258C0((*intrusiveListNode425760)(unsafe.Pointer(head)), int32(index))))
}

//export nox_common_list_append_4258E0
func nox_common_list_append_4258E0(head, node *C.nox_list_item_t) {
	intrusiveListBefore4258E0((*intrusiveListNode425760)(unsafe.Pointer(head)), (*intrusiveListNode425760)(unsafe.Pointer(node)))
}

//export sub_425900
func sub_425900(head, node *C.uint32_t) *C.uint32_t {
	return (*C.uint32_t)(unsafe.Pointer(intrusiveListAfter425900((*intrusiveListNode425760)(unsafe.Pointer(head)), (*intrusiveListNode425760)(unsafe.Pointer(node)))))
}

//export nox_common_list_remove_425920
func nox_common_list_remove_425920(node unsafe.Pointer) {
	intrusiveListRemove425920((*intrusiveListNode425760)(node))
}

//export nox_common_list_getNext_425940
func nox_common_list_getNext_425940(node *C.nox_list_item_t) *C.nox_list_item_t {
	return (*C.nox_list_item_t)(unsafe.Pointer(intrusiveListNext425940((*intrusiveListNode425760)(unsafe.Pointer(node)))))
}

//export sub_425960
func sub_425960(node *C.nox_list_item_t) *C.nox_list_item_t {
	return (*C.nox_list_item_t)(unsafe.Pointer(intrusiveListPrevious425960((*intrusiveListNode425760)(unsafe.Pointer(node)))))
}

func intrusiveListAbort4258E0() {
	C.abort()
}
