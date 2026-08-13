package legacy

import "github.com/opennox/opennox/v1/server"

var countInventoryClassSubclassNativeHooks4E7DA0 = countInventoryClassSubclassHooks4E7DA0[*server.Object]{
	loadFirst: func(owner *server.Object) *server.Object {
		return owner.InvFirstItem
	},
	loadClass: func(obj *server.Object) uint32 {
		return uint32(obj.ObjClass)
	},
	loadSubclass: func(obj *server.Object) uint32 {
		return uint32(obj.ObjSubClass)
	},
	loadNext: func(obj *server.Object) *server.Object {
		return obj.InvNextItem
	},
}

// Sub_4E7DA0 restores the unreferenced original entry point without adding a
// CGo export or reintroducing an ABI32 object pointer.
func Sub_4E7DA0(owner *server.Object, classMask, subclassMask uint32) int32 {
	return countInventoryClassSubclass4E7DA0(
		owner, classMask, subclassMask, countInventoryClassSubclassNativeHooks4E7DA0,
	)
}
