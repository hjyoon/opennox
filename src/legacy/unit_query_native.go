package legacy

/*
#include "GAME3_3.h"
typedef const nox_object_t unit_query_const_object_t;
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/libs/object"

	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/server"
)

//export sub_4E7BC0
func sub_4E7BC0(unit *C.unit_query_const_object_t) C.int {
	player := (*server.Object)(unsafe.Pointer(unit)).Class().Has(object.ClassPlayer)
	return C.int(bool2int(player))
}

//export nox_xxx_unitCountSlaves_4E7CF0
func nox_xxx_unitCountSlaves_4E7CF0(owner *C.unit_query_const_object_t, class, subclass C.uint32_t) C.int32_t {
	unit := (*server.Object)(unsafe.Pointer(owner))
	return C.int32_t(Nox_xxx_unitCountSlaves_4E7CF0(unit, uint32(class), uint32(subclass)))
}

//export sub_4E7DE0
func sub_4E7DE0(candidate, item *C.unit_query_const_object_t) C.int32_t {
	return C.int32_t(Sub_4E7DE0((*server.Object)(unsafe.Pointer(candidate)), (*server.Object)(unsafe.Pointer(item))))
}

//export sub_4E7EC0
func sub_4E7EC0(owner, item *C.unit_query_const_object_t) C.int32_t {
	return C.int32_t(Sub_4E7EC0((*server.Object)(unsafe.Pointer(owner)), (*server.Object)(unsafe.Pointer(item))))
}

var unitOwnedTypeNativeHooks = unitOwnedTypeHooks4E7BE0[*uint32, *server.Object]{
	loadCache:  func(cache *uint32) uint32 { return *cache },
	storeCache: func(cache *uint32, id uint32) { *cache = id },
	lookupType: func(name string) uint32 { return uint32(GetServer().S().Types.IndByID(name)) },
	loadFirst:  func(owner *server.Object) *server.Object { return owner.Field129 },
	loadType:   func(obj *server.Object) uint16 { return obj.TypeInd },
	loadNext:   func(obj *server.Object) *server.Object { return obj.Field128 },
}

func unitHasCrown(owner *server.Object) bool {
	cache := memmap.PtrUint32(0x5D4594, 1567716)
	return unitIsCrown4E7BE0(cache, owner, unitOwnedTypeNativeHooks) != 0
}

//export nox_xxx_unitIsCrown_4E7BE0
func nox_xxx_unitIsCrown_4E7BE0(owner *C.unit_query_const_object_t) C.int {
	return C.int(bool2int(unitHasCrown((*server.Object)(unsafe.Pointer(owner)))))
}

//export nox_xxx_unitIsGameball_4E7C30
func nox_xxx_unitIsGameball_4E7C30(owner *C.unit_query_const_object_t) C.int {
	cache := memmap.PtrUint32(0x5D4594, 1567720)
	return C.int(unitIsGameBall4E7C30(cache, (*server.Object)(unsafe.Pointer(owner)), unitOwnedTypeNativeHooks))
}

//export nox_xxx_unitFreezeGateRef_4E79B0
func nox_xxx_unitFreezeGateRef_4E79B0() *C.uint32_t {
	return (*C.uint32_t)(unsafe.Pointer(unitFreezeGateRuntimeRef4E79B0()))
}

//export sub_4E79B0
func sub_4E79B0(value C.uint32_t) C.uint32_t {
	return C.uint32_t(Sub_4E79B0(uint32(value)))
}
