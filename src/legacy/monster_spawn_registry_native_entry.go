package legacy

/*
#include "monster_spawn_registry_50d780.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/server"
)

func monsterSpawnRuntimeNative50D890() server.MonsterSpawnDeleteRuntime50E210 {
	return server.MonsterSpawnDeleteRuntime50E210{DelayedDelete: GetServer().DelayedDelete}
}

//export nox_xxx_allocMonsterRelatedArrays_50D780
func nox_xxx_allocMonsterRelatedArrays_50D780() C.int {
	return C.int(bool2int(GetServer().S().MonsterSpawnInit50D780()))
}

//export sub_50D7E0
func sub_50D7E0() { GetServer().S().MonsterSpawnReset50D7E0() }

//export sub_50D820
func sub_50D820() { GetServer().S().MonsterSpawnFree50D820() }

//export sub_50D890
func sub_50D890() C.uint {
	return C.uint(GetServer().S().MonsterSpawnTick50D890(monsterSpawnRuntimeNative50D890()))
}

//export sub_50D8D0
func sub_50D8D0() { GetServer().S().MonsterSpawnCullDistant50D8D0(monsterSpawnRuntimeNative50D890()) }

//export sub_50D960
func sub_50D960() C.int {
	return C.int(GetServer().S().MonsterSpawnCullVisible50D960(monsterSpawnRuntimeNative50D890()))
}

//export sub_50DE10
func sub_50DE10() C.int { return C.int(GetServer().S().MonsterSpawnScratchClear50DE10()) }

// Keep native pointer values in integer slots only in private Cgo return
// frames. Public C entries retain typed pointers to pinned owned records.

//export nox_server_spawn_find_native_50DE60
func nox_server_spawn_find_native_50DE60(obj *C.nox_object_t) C.uintptr_t {
	record := GetServer().S().MonsterSpawnScratchFind50DE60(asObjectS(obj))
	return C.uintptr_t(uintptr(unsafe.Pointer(record)))
}

//export sub_50DE80
func sub_50DE80(_ *C.nox_object_t, point *C.float2) C.int {
	return C.int(bool2int(GetServer().S().MonsterSpawnCanPlace50DE80(*(*types.Pointf)(unsafe.Pointer(point)))))
}

//export sub_50DFB0
func sub_50DFB0(obj, player *C.nox_object_t) {
	GetServer().S().MonsterSpawnVisible50DFB0(asObjectS(obj), asObjectS(player), monsterSpawnRuntimeNative50D890())
}

//export sub_50E030
func sub_50E030(generator, spawned *C.nox_object_t) C.int {
	s := GetServer().S()
	return C.int(bool2int(s.MonsterSpawnRegister50E030(asObjectS(generator), asObjectS(spawned), server.MonsterSpawnRegisterRuntime50E030{
		CreateGlyph:  func() *server.Object { return s.NewObjectByTypeInd(s.Types.GlyphID()) },
		InventoryPut: func(owner, item *server.Object, report int32) { inventoryPutImpl4F3070(owner, item, report != 0) },
	})))
}

//export nox_server_spawn_prepend_native_50E110
func nox_server_spawn_prepend_native_50E110(link *C.nox_monster_spawn_link) C.uintptr_t {
	value := GetServer().S().MonsterSpawnPrepend50E110((*server.MonsterSpawnLink)(unsafe.Pointer(link)))
	return C.uintptr_t(uintptr(unsafe.Pointer(value)))
}

//export sub_50E140
func sub_50E140(obj *C.nox_object_t) { GetServer().S().MonsterSpawnCleanup50E140(asObjectS(obj)) }

//export nox_server_spawn_unlink_native_50E1B0
func nox_server_spawn_unlink_native_50E1B0(link *C.nox_monster_spawn_link) C.uintptr_t {
	value := GetServer().S().MonsterSpawnUnlink50E1B0((*server.MonsterSpawnLink)(unsafe.Pointer(link)))
	return C.uintptr_t(uintptr(unsafe.Pointer(value)))
}

//export sub_50E1E0
func sub_50E1E0(obj *C.nox_object_t) {
	GetServer().S().MonsterSpawnCleanupUnlessZombie50E1E0(asObjectS(obj))
}

//export sub_50E210
func sub_50E210(obj *C.nox_object_t) { Sub_50E210(asObjectS(obj)) }
