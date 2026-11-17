package legacy

/*
#include "GAME3_3.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

func cacheListNativeLoad(slot unsafe.Pointer) *C.nox_dblLinkedList {
	return (*C.nox_dblLinkedList)(*(*unsafe.Pointer)(slot))
}

func cacheListNativeStore(slot unsafe.Pointer, value *C.nox_dblLinkedList) {
	// These are C-owned pointer cells; old C bytes must not enter Go's barrier.
	*(*uintptr)(slot) = uintptr(unsafe.Pointer(value))
}

func cacheListNativePrev(entry *C.nox_dblLinkedList) unsafe.Pointer {
	// The C field called next is the original logical previous link.
	return unsafe.Pointer(&entry.next)
}

func cacheListNativeNext(entry *C.nox_dblLinkedList) unsafe.Pointer {
	return unsafe.Pointer(&entry.prev)
}

//export sub_4ECDE0
func sub_4ECDE0(list **C.nox_dblLinkedList, entry *C.nox_dblLinkedList) *C.nox_dblLinkedList {
	return server.NetCodeCachePrependNative4ECDE0(entry,
		func(node, prev *C.nox_dblLinkedList) { cacheListNativeStore(cacheListNativePrev(node), prev) },
		func() *C.nox_dblLinkedList { return cacheListNativeLoad(unsafe.Pointer(list)) },
		func(node, next *C.nox_dblLinkedList) { cacheListNativeStore(cacheListNativeNext(node), next) },
		func(last *C.nox_dblLinkedList) {
			cacheListNativeStore(unsafe.Add(unsafe.Pointer(list), unsafe.Sizeof(uintptr(0))), last)
		},
		func(first *C.nox_dblLinkedList) { cacheListNativeStore(unsafe.Pointer(list), first) },
	)
}

//export sub_4ECE10
func sub_4ECE10(list **C.nox_dblLinkedList, entry *C.nox_dblLinkedList) *C.nox_dblLinkedList {
	return server.NetCodeCacheRemoveNative4ECE10(entry,
		func(node *C.nox_dblLinkedList) *C.nox_dblLinkedList {
			return cacheListNativeLoad(cacheListNativeNext(node))
		},
		func(node *C.nox_dblLinkedList) *C.nox_dblLinkedList {
			return cacheListNativeLoad(cacheListNativePrev(node))
		},
		func(node, prev *C.nox_dblLinkedList) { cacheListNativeStore(cacheListNativePrev(node), prev) },
		func(last *C.nox_dblLinkedList) {
			cacheListNativeStore(unsafe.Add(unsafe.Pointer(list), unsafe.Sizeof(uintptr(0))), last)
		},
		func(node, next *C.nox_dblLinkedList) { cacheListNativeStore(cacheListNativeNext(node), next) },
		func(first *C.nox_dblLinkedList) { cacheListNativeStore(unsafe.Pointer(list), first) },
	)
}
