package legacy

import "github.com/opennox/opennox/v1/server"

var countSlavesNativeHooks4E7CF0 = countSlavesHooks4E7CF0[*server.Object]{
	loadFirst: func(owner *server.Object) *server.Object {
		return owner.Field129
	},
	loadClass: func(obj *server.Object) uint32 {
		return uint32(obj.ObjClass)
	},
	loadSubclass: func(obj *server.Object) uint32 {
		return uint32(obj.ObjSubClass)
	},
	loadNext: func(obj *server.Object) *server.Object {
		return obj.Field128
	},
}

// Nox_xxx_unitCountSlaves_4E7CF0 counts owned units matching both masks.
func Nox_xxx_unitCountSlaves_4E7CF0(owner *server.Object, classMask, subclassMask uint32) int32 {
	return countSlaves4E7CF0(owner, classMask, subclassMask, countSlavesNativeHooks4E7CF0)
}
