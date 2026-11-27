package legacy

/*
#include "GAME4_1.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

const scriptObjectCacheCapacity = 16

type scriptObjectCacheNode struct {
	object *nox_object_t
	prev   *scriptObjectCacheNode
	next   *scriptObjectCacheNode
}

type scriptObjectCacheList struct {
	first *scriptObjectCacheNode
	last  *scriptObjectCacheNode
}

func (list *scriptObjectCacheList) prepend(node *scriptObjectCacheNode) {
	node.prev, node.next = nil, list.first
	if list.first != nil {
		list.first.prev = node
	} else {
		list.last = node
	}
	list.first = node
}

func (list *scriptObjectCacheList) detach(node *scriptObjectCacheNode) {
	if node.next != nil {
		node.next.prev = node.prev
	} else {
		list.last = node.prev
	}
	if node.prev != nil {
		node.prev.next = node.next
	} else {
		list.first = node.next
	}
}

var scriptObjectCache struct {
	nodes   *[scriptObjectCacheCapacity]scriptObjectCacheNode
	objects [scriptObjectCacheCapacity]*server.Object
	free    scriptObjectCacheList
	used    scriptObjectCacheList
}

func scriptObjectCacheSlot(node *scriptObjectCacheNode) **server.Object {
	base := uintptr(unsafe.Pointer(scriptObjectCache.nodes))
	index := (uintptr(unsafe.Pointer(node)) - base) / unsafe.Sizeof(*node)
	return &scriptObjectCache.objects[index]
}

func initializeScriptObjectCache() {
	if scriptObjectCache.nodes == nil {
		// Returned entry addresses remain valid for the lifetime of the legacy runtime.
		scriptObjectCache.nodes, _ = alloc.New([scriptObjectCacheCapacity]scriptObjectCacheNode{})
	}
	clear(scriptObjectCache.objects[:])
	scriptObjectCache.free = scriptObjectCacheList{}
	scriptObjectCache.used = scriptObjectCacheList{}
	for i := range scriptObjectCache.nodes {
		scriptObjectCache.free.prepend(&scriptObjectCache.nodes[i])
	}
	legacyWords.dword_587000_237036 = 0
}

func scriptObjectCachePopFree() *scriptObjectCacheNode {
	first := scriptObjectCache.free.first
	if first != nil {
		scriptObjectCache.free.first = first.next
	}
	return first
}

func findScriptObjectCacheNode(id int32, nativeOnly bool) *scriptObjectCacheNode {
	if legacyWords.dword_587000_237036 != 0 {
		initializeScriptObjectCache()
	}
	for node := scriptObjectCache.used.first; node != nil; node = node.next {
		if nativeOnly && node.object == nil {
			continue
		}
		object := *scriptObjectCacheSlot(node)
		if object != nil && object.ObjFlags&0x20 == 0 && object.ScriptIDVal == id {
			scriptObjectCache.used.detach(node)
			scriptObjectCache.used.prepend(node)
			return node
		}
	}
	return nil
}

func rememberScriptObject(object *server.Object, native *nox_object_t) *scriptObjectCacheNode {
	if object == nil {
		return nil
	}
	if legacyWords.dword_587000_237036 != 0 || scriptObjectCache.nodes == nil {
		initializeScriptObjectCache()
	}
	node := scriptObjectCachePopFree()
	if node == nil {
		node = scriptObjectCache.used.last
		scriptObjectCache.used.detach(node)
	}
	// Go callers can own objects on the Go heap. The pointer exported to C
	// belongs to the native ABI, while the Go slot retains the object lifetime.
	*scriptObjectCacheSlot(node) = object
	storeForeignPointer(unsafe.Pointer(&node.object), unsafe.Pointer(native))
	scriptObjectCache.used.prepend(node)
	return node
}

func forgetCachedScriptObject(object *server.Object) uintptr {
	if flag := legacyWords.dword_587000_237036; flag != 0 {
		return uintptr(flag)
	}
	if scriptObjectCache.used.first == nil {
		return 0
	}
	for node := scriptObjectCache.used.first; node != nil; node = node.next {
		if slot := scriptObjectCacheSlot(node); *slot == object {
			scriptObjectCache.used.detach(node)
			*slot = nil
			scriptObjectCache.free.prepend(node)
			return uintptr(unsafe.Pointer(node))
		}
	}
	return uintptr(unsafe.Pointer(object))
}

func clearScriptObjectCache() uintptr {
	if flag := legacyWords.dword_587000_237036; flag != 0 {
		return uintptr(flag)
	}
	var last *scriptObjectCacheNode
	for node := scriptObjectCache.used.first; node != nil; node = scriptObjectCache.used.first {
		scriptObjectCache.used.detach(node)
		*scriptObjectCacheSlot(node) = nil
		scriptObjectCache.free.prepend(node)
		last = node
	}
	return uintptr(unsafe.Pointer(last))
}

//export nox_xxx_script_511C50
func nox_xxx_script_511C50(id C.int) *C.nox_object_t {
	if node := findScriptObjectCacheNode(int32(id), true); node != nil {
		return node.object
	}
	return nil
}

//export nox_xxx_scriptPrepareFoundUnit_511D70
func nox_xxx_scriptPrepareFoundUnit_511D70(object *C.nox_object_t) C.intptr_t {
	node := rememberScriptObject(asObjectS(object), object)
	return C.intptr_t(uintptr(unsafe.Pointer(node)))
}

//export sub_511DE0
func sub_511DE0(object *C.nox_object_t) C.intptr_t {
	return C.intptr_t(forgetCachedScriptObject(asObjectS(object)))
}

//export sub_511E20
func sub_511E20() C.intptr_t {
	return C.intptr_t(clearScriptObjectCache())
}
