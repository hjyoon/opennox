package legacy

/*
#include "defs.h"
extern nox_object_t* dword_5d4594_2491592;
int nox_float2int(float);
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/libs/types"
	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/server"
)

var game5Cursor struct {
	source, result *server.Object
}

//export nox_xxx_findObjectAtCursor_54AF40
func nox_xxx_findObjectAtCursor_54AF40(raw *nox_object_t) *nox_object_t {
	source := asObjectS(raw)
	cursor := source.UpdateDataPlayer().Player.CursorPos()
	game5Cursor.source, game5Cursor.result = source, nil
	C.dword_5d4594_2491592 = raw
	*memmap.PtrUint32(0x5D4594, 2491600) = 0
	GetServer().S().Map.EachObjInCircle(cursor, 100, func(obj *server.Object) bool {
		game5CursorScan(obj, &cursor)
		return true
	})
	return asObjectC(game5Cursor.result)
}

func game5CursorScan(obj *server.Object, cursor *types.Pointf) {
	polyp := memmap.PtrUint32(0x5D4594, 2491604)
	if *polyp == 0 {
		*polyp = uint32(GetServer().S().Types.IndByID("Polyp"))
	}
	source := asObjectS(C.dword_5d4594_2491592)
	if obj == source || obj.Flags()&0x8020 != 0 || obj.HasEnchant(0) && !source.HasEnchant(21) || obj.ObjClass&0x80000206 == 0 && uint32(obj.TypeInd) != *polyp {
		return
	}
	if !GetServer().S().MapTraceVision(obj, source) {
		return
	}
	if obj.ObjClass&4 != 0 {
		if obj.NetCode == uint32(ClientPlayerNetCode()) && noxflags.HasEngine(noxflags.EngineNoRendering) {
			return
		}
		player := GetServer().S().Players.ByID(int(obj.NetCode))
		if player == nil || player.Field3680&1 != 0 {
			return
		}
	}
	if obj.ObjClass&0x200 != 0 && obj.Flags()&0x4000 == 0 {
		return
	}
	if !game5CursorContains(obj, cursor) {
		return
	}
	height := server.PlayerCursorHeight54AFB0(obj)
	best := memmap.PtrFloat32(0x5D4594, 2491600)
	if height > float64(*best) {
		*best = float32(height)
		game5Cursor.result = obj
	}
}

// Keep the projected-circle and box tests in the same coordinate space as the
// original cursor scan. The shape is addressed by its native field location.
func game5CursorContains(obj *server.Object, cursor *types.Pointf) bool {
	return server.PlayerCursorGeometry54AFB0(obj, cursor, func(radius float32) int32 {
		return int32(C.nox_float2int(C.float(radius)))
	})
}

//export nox_xxx_playerCursorScanFn_54AFB0
func nox_xxx_playerCursorScanFn_54AFB0(raw *nox_object_t, cursor *C.float2) {
	game5CursorScan(asObjectS(raw), (*types.Pointf)(unsafe.Pointer(cursor)))
}
