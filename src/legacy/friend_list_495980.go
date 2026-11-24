package legacy

/*
#include "GAME2_3.h"
#include "common/alloc/classes/alloc_class.h"

extern void *nox_alloc_friendList_1203860, *dword_5d4594_1203864;
*/
import "C"

import "unsafe"

type clientFriendNode struct {
	code uint32
	next *clientFriendNode
}

var (
	_ = [1]struct{}{}[unsafe.Offsetof(clientFriendNode{}.next)-unsafe.Sizeof(uintptr(0))]
	_ = [1]struct{}{}[unsafe.Sizeof(clientFriendNode{})-2*unsafe.Sizeof(uintptr(0))]
)

func friendListHeadSlot() **clientFriendNode {
	return (**clientFriendNode)(unsafe.Pointer(&C.dword_5d4594_1203864))
}

func friendListAllocator() *C.nox_alloc_class {
	return (*C.nox_alloc_class)(C.nox_alloc_friendList_1203860)
}

//export nox_xxx_allocClassListFriends_495980
func nox_xxx_allocClassListFriends_495980() C.int {
	pool := nox_new_alloc_class(internCStr("FriendListClass"), C.int(unsafe.Sizeof(clientFriendNode{})), 128)
	C.nox_alloc_friendList_1203860 = unsafe.Pointer(pool)
	return C.int(bool2int(pool != nil))
}

//export sub_4959B0
func sub_4959B0() {
	nox_alloc_class_free_all(friendListAllocator())
	*friendListHeadSlot() = nil
}

//export sub_4959D0
func sub_4959D0() C.int {
	nox_free_alloc_class(friendListAllocator())
	C.nox_alloc_friendList_1203860 = nil
	*friendListHeadSlot() = nil
	return 0
}

//export nox_xxx_cliAddObjFriend_4959F0
func nox_xxx_cliAddObjFriend_4959F0(code C.int) *C.uint32_t {
	node := (*clientFriendNode)(nox_alloc_class_new_obj_zero(friendListAllocator()))
	if node == nil {
		return nil
	}
	head := friendListHeadSlot()
	node.code, node.next = uint32(code), *head
	*head = node
	return (*C.uint32_t)(&node.code)
}

//export sub_495A20
func sub_495A20(code C.int) {
	for link := friendListHeadSlot(); *link != nil; link = &(*link).next {
		node := *link
		if node.code == uint32(code) {
			*link = node.next
			nox_alloc_class_free_obj_first(friendListAllocator(), unsafe.Pointer(node))
			return
		}
	}
}

//export sub_495A80
func sub_495A80(code C.int) C.int {
	for node := *friendListHeadSlot(); node != nil; node = node.next {
		if node.code == uint32(code) {
			return 1
		}
	}
	return 0
}

func Nox_xxx_cliAddObjFriend_4959F0(netCode uint32) unsafe.Pointer {
	return unsafe.Pointer(nox_xxx_cliAddObjFriend_4959F0(C.int(netCode)))
}

func Sub_4959B0() {
	sub_4959B0()
}

func Sub_495A20(netCode uint32) {
	sub_495A20(C.int(netCode))
}

func friendListHead495980() unsafe.Pointer {
	return unsafe.Pointer(*friendListHeadSlot())
}
