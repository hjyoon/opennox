package legacy

import "github.com/opennox/opennox/v1/server"

var countOwnedClassNativeHooks4E7CC0 = countOwnedClassHooks4E7CC0[*server.Object]{
	loadFirst: func(owner *server.Object) *server.Object {
		return owner.Field129
	},
	loadClass: func(obj *server.Object) uint32 {
		return uint32(obj.ObjClass)
	},
	loadNext: func(obj *server.Object) *server.Object {
		return obj.Field128
	},
}

// Sub_4E7CC0 restores the unreferenced original entry point without adding a
// CGo export or reintroducing an ABI32 object pointer.
func Sub_4E7CC0(owner *server.Object, classMask uint32) int32 {
	return countOwnedClass4E7CC0(owner, classMask, countOwnedClassNativeHooks4E7CC0)
}
