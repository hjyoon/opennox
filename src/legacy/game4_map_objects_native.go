package legacy

/*
#include "game4_list_vote_native.h"
*/
import "C"

import "unsafe"

func game4MapObjectHead() *C.nox_map_object_list_node_5048A0 {
	return (*C.nox_map_object_list_node_5048A0)(C.dword_5d4594_1599540)
}
func game4SetMapObjectHead(node *C.nox_map_object_list_node_5048A0) {
	C.dword_5d4594_1599540 = unsafe.Pointer(node)
}

//export nox_xxx_unitAddToList_5048A0
func nox_xxx_unitAddToList_5048A0(object *C.nox_object_t) *C.nox_map_object_list_node_5048A0 {
	node := (*C.nox_map_object_list_node_5048A0)(game4SupportAllocate(1, C.sizeof_nox_map_object_list_node_5048A0))
	if node == nil {
		return nil
	}
	node.previous, node.object, node.next = nil, object, game4MapObjectHead()
	if node.next != nil {
		node.next.previous = node
	}
	game4SetMapObjectHead(node)
	object.object_prev = nil
	if node.next != nil {
		object.object_next = node.next.object
		node.next.object.object_prev = object
	} else {
		object.object_next = nil
	}
	return node
}

//export sub_504910
func sub_504910(deltaX, deltaY C.int32_t) C.int32_t {
	x, y := C.float(deltaX), C.float(deltaY)
	for node := game4MapObjectHead(); node != nil; node = node.next {
		object := node.object
		object.x += x
		object = node.object
		object.y += y
		object = node.object
		posY, posX := object.y, object.x
		C.nox_xxx_createAt_4DAA50(object, nil, posX, posY)
		object = node.object
		object.obj_flags |= 0x80000000
	}
	return 1
}

//export sub_504980
func sub_504980() *C.nox_object_t {
	loaded := C.dword_5d4594_1599480
	if !(loaded == C.dword_5d4594_3835396 && int32(loaded) != -1 && C.dword_5d4594_1599476 != 1) && C.nox_xxx_mapgenSaveMap_503830(C.int(int32(C.dword_5d4594_3835396))) == 0 {
		return nil
	}
	node := game4MapObjectHead()
	if node == nil {
		return nil
	}
	return node.object
}

//export sub_5049C0
func sub_5049C0(object *C.nox_object_t) *C.nox_object_t {
	if object == nil {
		return nil
	}
	return object.object_next
}

//export sub_5049D0
func sub_5049D0() *C.nox_map_object_list_node_5048A0 { return game4MapObjectHead() }

//export sub_5049E0
func sub_5049E0(node *C.nox_map_object_list_node_5048A0) *C.nox_map_object_list_node_5048A0 {
	if node == nil {
		return nil
	}
	return node.next
}

//export sub_504A10
func sub_504A10(object *C.nox_object_t) C.int32_t {
	if object == nil {
		return 0
	}
	node := game4MapObjectHead()
	for node != nil && node.object != object {
		node = node.next
	}
	if node == nil {
		return 0
	}
	if node.next != nil {
		node.next.previous = node.previous
	}
	if node.previous != nil {
		node.previous.next = node.next
	}
	if node == game4MapObjectHead() {
		game4SetMapObjectHead(node.next)
	}
	if object.object_next != nil {
		object.object_next.object_prev = object.object_prev
	}
	if object.object_prev != nil {
		object.object_prev.object_next = object.object_next
	}
	C.nox_xxx_objectFreeMem_4E38A0(object)
	game4SupportFree(unsafe.Pointer(node))
	return 1
}

//export nox_map_object_list_node_object_5048A0
func nox_map_object_list_node_object_5048A0(node *C.nox_map_object_list_node_5048A0) *C.nox_object_t {
	return node.object
}

//export nox_map_object_list_node_next_5048A0
func nox_map_object_list_node_next_5048A0(node *C.nox_map_object_list_node_5048A0) *C.nox_map_object_list_node_5048A0 {
	return node.next
}

//export nox_map_object_list_node_free_5048A0
func nox_map_object_list_node_free_5048A0(node *C.nox_map_object_list_node_5048A0) {
	game4SupportFree(unsafe.Pointer(node))
}
