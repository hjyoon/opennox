package legacy

import "github.com/opennox/opennox/v1/server"

var countInventoryClassNativeHooks4E7D70 = countInventoryClassHooks4E7D70[*server.Object]{
	loadFirst: func(owner *server.Object) *server.Object {
		return owner.InvFirstItem
	},
	loadClass: func(obj *server.Object) uint32 {
		return uint32(obj.ObjClass)
	},
	loadNext: func(obj *server.Object) *server.Object {
		return obj.InvNextItem
	},
}

// Sub_4E7D70 restores the unreferenced original entry point without adding a
// CGo export or reintroducing an ABI32 object pointer.
func Sub_4E7D70(owner *server.Object, classMask uint32) int32 {
	return countInventoryClass4E7D70(owner, classMask, countInventoryClassNativeHooks4E7D70)
}
