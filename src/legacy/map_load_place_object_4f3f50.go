package legacy

/*
#include "map_load_place_object_4f3f50.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/libs/types"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/common/ntype"
	"github.com/opennox/opennox/v1/server"
)

type mapLoadPlaceDeps4F3F50[T comparable, W any] struct {
	gameFlag23       func() int32
	loadTypeInd      func(T) uint16
	typeAllowed      func(uint16) int32
	loadFirstItem    func(T) T
	loadNextItem     func(T) T
	freeObject       func(T)
	wallSize         func() W
	loadWallWidth    func(W) uint32
	loadWallHeight   func(W) uint32
	loadX            func(T) float32
	loadY            func(T) float32
	loadTranslationX func() int32
	loadTranslationY func() int32
	storeX           func(T, float32)
	storeY           func(T, float32)
	stageObject      func(T)
	gameFlag22       func() int32
	placeAllowed     func(uint16) int32
	storeFirstItem   func(T, T)
	createAt         func(T, T, float32, float32, int32)
}

func mapLoadPlaceTranslatedCoordinate4F3F50(position float32, wallCount uint32, translation int32) float32 {
	// LEA/SHL/SUB computes 23*wallCount modulo 2^32, then FILD treats
	// that word as signed. Binary64 models the configured x87 53-bit
	// intermediate before GAME.EXE's single binary32 FSTP spill.
	wallOffset := int32(uint32(23) * wallCount)
	value := float64(position) - float64(wallOffset)
	value += float64(translation)
	value -= 11.0
	return float32(value)
}

func mapLoadPlaceFree4F3F50[T comparable, W any](
	object T,
	clearFirst bool,
	deps mapLoadPlaceDeps4F3F50[T, W],
) {
	var zero T
	for item := deps.loadFirstItem(object); item != zero; {
		next := deps.loadNextItem(item)
		deps.freeObject(item)
		item = next
	}
	if clearFirst {
		deps.storeFirstItem(object, zero)
	}
	deps.freeObject(object)
}

// mapLoadPlaceObject4F3F50 preserves GAME.EXE's two independent GameFlag23
// reads, the exact-one GameFlag22 gate, live field reloads, translation spill
// boundaries, and the distinct early/late inventory cleanup paths.
func mapLoadPlaceObject4F3F50[T comparable, W any](
	object, owner T,
	hasTranslation bool,
	deps mapLoadPlaceDeps4F3F50[T, W],
) int32 {
	if deps.gameFlag23() == 0 {
		if deps.typeAllowed(deps.loadTypeInd(object)) == 0 {
			mapLoadPlaceFree4F3F50(object, false, deps)
			return 0
		}
	}

	if hasTranslation {
		wall := deps.wallSize()
		width := deps.loadWallWidth(wall)
		x := deps.loadX(object)
		translationX := deps.loadTranslationX()
		deps.storeX(object, mapLoadPlaceTranslatedCoordinate4F3F50(x, width, translationX))

		height := deps.loadWallHeight(wall)
		y := deps.loadY(object)
		translationY := deps.loadTranslationY()
		deps.storeY(object, mapLoadPlaceTranslatedCoordinate4F3F50(y, height, translationY))
	}

	if deps.gameFlag23() != 0 {
		deps.stageObject(object)
		return 1
	}
	if deps.gameFlag22() != 1 {
		if deps.placeAllowed(deps.loadTypeInd(object)) == 0 {
			mapLoadPlaceFree4F3F50(object, true, deps)
			return 0
		}
	}

	y := deps.loadY(object)
	x := deps.loadX(object)
	deps.createAt(object, owner, x, y, 0)
	return 1
}

func mapLoadPlaceObjectNative4F3F50(
	object, owner *server.Object,
	translation *ntype.Point32,
) int32 {
	return mapLoadPlaceObject4F3F50(object, owner, translation != nil, mapLoadPlaceDeps4F3F50[*server.Object, *[2]uint32]{
		gameFlag23: func() int32 {
			return int32(bool2int(noxflags.HasGame(noxflags.GameFlag23)))
		},
		loadTypeInd: func(object *server.Object) uint16 {
			return object.TypeInd
		},
		typeAllowed: func(typeInd uint16) int32 {
			return int32(bool2int(GetServer().S().Types.ByInd(int(typeInd)).Allowed()))
		},
		loadFirstItem: func(object *server.Object) *server.Object {
			return object.InvFirstItem
		},
		loadNextItem: func(item *server.Object) *server.Object {
			return item.InvNextItem
		},
		freeObject: func(object *server.Object) {
			GetServer().S().Objs.FreeObject(object)
		},
		wallSize: func() *[2]uint32 {
			return (*[2]uint32)(unsafe.Pointer(memmap.PtrUint32(0x5D4594, 739980)))
		},
		loadWallWidth: func(wall *[2]uint32) uint32 {
			return wall[0]
		},
		loadWallHeight: func(wall *[2]uint32) uint32 {
			return wall[1]
		},
		loadX: func(object *server.Object) float32 {
			return object.PosVec.X
		},
		loadY: func(object *server.Object) float32 {
			return object.PosVec.Y
		},
		loadTranslationX: func() int32 {
			return translation.X
		},
		loadTranslationY: func() int32 {
			return translation.Y
		},
		storeX: func(object *server.Object, value float32) {
			object.PosVec.X = value
		},
		storeY: func(object *server.Object, value float32) {
			object.PosVec.Y = value
		},
		stageObject: func(object *server.Object) {
			mapObjectListAdd5048A0(object)
		},
		gameFlag22: func() int32 {
			return int32(bool2int(noxflags.HasGame(noxflags.GameFlag22)))
		},
		placeAllowed: func(typeInd uint16) int32 {
			return int32(Sub_4E3AD0(int(typeInd)))
		},
		storeFirstItem: func(object, first *server.Object) {
			object.InvFirstItem = first
		},
		createAt: func(object, owner *server.Object, x, y float32, _ int32) {
			var ownerObject server.Obj
			if owner != nil {
				ownerObject = owner
			}
			GetServer().CreateObjectAt(object, ownerObject, types.Pointf{X: x, Y: y})
		},
	})
}

//export nox_xxx_servMapLoadPlaceObj_4F3F50
func nox_xxx_servMapLoadPlaceObj_4F3F50(
	object, owner *C.nox_object_t,
	translation *C.nox_map_translation_4F3F50,
) C.int32_t {
	return C.int32_t(mapLoadPlaceObjectNative4F3F50(
		asObjectS((*nox_object_t)(object)),
		asObjectS((*nox_object_t)(owner)),
		(*ntype.Point32)(unsafe.Pointer(translation)),
	))
}

func Nox_xxx_servMapLoadPlaceObj_4F3F50(
	object, owner *server.Object,
	translation *ntype.Point32,
) int32 {
	return mapLoadPlaceObjectNative4F3F50(object, owner, translation)
}
