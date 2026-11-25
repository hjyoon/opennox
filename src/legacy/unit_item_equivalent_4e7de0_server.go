package legacy

import (
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

var itemEquivalentNativeHooks4E7DE0 = itemEquivalentHooks4E7DE0[
	*server.Object, *server.ModifierInitData, unsafe.Pointer, *server.ModifierEff,
]{
	loadType: func(obj *server.Object) uint16 {
		return obj.TypeInd
	},
	loadClass: func(obj *server.Object) uint32 {
		return uint32(obj.ObjClass)
	},
	loadInitData: func(obj *server.Object) *server.ModifierInitData {
		return (*server.ModifierInitData)(obj.InitData)
	},
	loadModifier: func(attrs *server.ModifierInitData, index int) *server.ModifierEff {
		return attrs.Modifiers[index]
	},
	loadSubclass: func(obj *server.Object) uint32 {
		return uint32(obj.ObjSubClass)
	},
	loadUseData: func(obj *server.Object) unsafe.Pointer {
		return obj.UseData.Ptr
	},
	loadUseByte: func(data unsafe.Pointer, index int) byte {
		return *(*byte)(unsafe.Add(data, uintptr(index)))
	},
}

// Sub_4E7DE0 compares item types, modifiers, and book payloads.
func Sub_4E7DE0(candidate, item *server.Object) int32 {
	if itemEquivalent4E7DE0(candidate, item, itemEquivalentNativeHooks4E7DE0) {
		return 1
	}
	return 0
}
