package legacy

/*
#include <stdlib.h>
#include <stdint.h>
#include "GAME1.h"
#include "GAME1_1.h"
#include "GAME3_2.h"
#include "GAME3_3.h"
#include "noxstring.h"
#include "world_native_storage.h"
static void world_defend_call(void* callback,void* effect,nox_object_t* item,float* value){
 ((void(*)(void*,nox_object_t*,uintptr_t,nox_object_t*,uintptr_t,float*))callback)(effect,item,0,item,0,value);
}
_Static_assert(sizeof(int)==4 && sizeof(short)==2 && sizeof(float)==4,"world/player scalar ABI");
*/
import "C"

import (
	"runtime"
	"unsafe"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/server"
)

var worldPlayerDefaults40A830 = worldPlayerRuntime40A830{
	PlayerFirst:       func() *server.Player { return GetServer().S().Players.First() },
	PlayerNext:        func(p *server.Player) *server.Player { return GetServer().S().Players.Next(p) },
	UnitFirst:         func() *server.Object { return GetServer().S().Players.FirstUnit() },
	UnitNext:          func(p *server.Object) *server.Object { return GetServer().S().Players.NextUnit(p) },
	TeamFirst:         func() *server.Team { return GetServer().S().Teams.First() },
	TeamNext:          func(p *server.Team) *server.Team { return GetServer().S().Teams.Next(p) },
	TeamContains:      func(p *server.ObjectTeam, id server.TeamID) bool { return GetServer().S().Teams.ContainsObject(p, id) },
	EngineNoRendering: func() bool { return noxflags.HasEngine(noxflags.EngineNoRendering) },
	GameplayTeams:     func() bool { return noxflags.HasGamePlay(noxflags.GameplayFlag(4)) },
	GameFlag: func(mask uint32) uint32 {
		if noxflags.HasGame(noxflags.GameFlag(mask)) {
			return 1
		}
		return 0
	},
	GameFlags:   func() uint32 { return uint32(noxflags.GetGame()) },
	TimerStatus: sub40A220Native,
	TimerMode:   sub40A180Native,
	TimerStart:  func() { sub40A250Native() },
	TimerSet:    sub40A1F0Native,
	Send: func(recipient int32, packet []byte) int32 {
		result := int32(nox_xxx_netSendPacket1_4E5390(C.int(recipient), C.nox_important_payload_t(unsafe.Pointer(&packet[0])), C.int(len(packet)), nil, 0))
		runtime.KeepAlive(packet)
		return result
	},
	Lesson: func(unit *server.Object) {
		nox_xxx_netReportLesson_4D8EF0((*C.nox_object_t)(unsafe.Pointer(unit)))
		runtime.KeepAlive(unit)
	},
	Minimap: func(mark bool, index byte, object *server.Object, flags uint32) {
		if mark {
			nox_xxx_netMarkMinimapObject_417190(int32(index), asObjectC(object), flags)
		} else {
			nox_xxx_netUnmarkMinimapObj_417300(int32(index), asObjectC(object), flags)
		}
		runtime.KeepAlive(object)
	},
	WallCreate: func(x, y int32) *server.Wall {
		return (*server.Wall)(nox_xxx_wallCreateAt_410250(x, y))
	},
	WallAttach: func(w *server.Wall, o *server.Object) unsafe.Pointer {
		return nox_server_wallAttachDoor(unsafe.Pointer(w), asObjectC(o))
	},
	ArmorDefinition: func(kind uint16) *server.Modifier {
		return (*server.Modifier)(nox_xxx_equipClothFindDefByTT_413270(int32(kind)))
	},
	CompareNames: func(a, b *uint16) int32 {
		return int32(_nox_wcsicmp((*wchar2_t)(unsafe.Pointer(a)), (*wchar2_t)(unsafe.Pointer(b))))
	},
}

func worldPlayerCallDefend415C00(callback unsafe.Pointer, effect *server.ModifierEff, item *server.Object, value *float32) {
	if applyModifierDefendScalar(callback, effect, value) {
		return
	}
	C.world_defend_call(callback, unsafe.Pointer(effect), (*C.nox_object_t)(unsafe.Pointer(item)), (*C.float)(unsafe.Pointer(value)))
	runtime.KeepAlive(effect)
	runtime.KeepAlive(item)
	runtime.KeepAlive(value)
}
func worldTileHeadCell411160() unsafe.Pointer {
	return unsafe.Pointer(&worldOwnerStorage.tile_coordinates)
}
func worldTileCapacity411160() int32 { return C.NOX_WORLD_TILE_CAPACITY }
func worldTileCalloc411160(count, size uintptr) unsafe.Pointer {
	return game4SupportAllocate(count, size)
}
func worldTileFree411160(p unsafe.Pointer)          { game4SupportReleaseMemory(p) }
func worldTileFloatToInt411160(value float32) int32 { return int32(C.nox_float2int(C.float(value))) }
func worldTileDefinitions411160() []server.TileDef {
	return unsafe.Slice((*server.TileDef)(unsafe.Pointer(&floorTileStorage.definitions[0])), len(floorTileStorage.definitions))
}
