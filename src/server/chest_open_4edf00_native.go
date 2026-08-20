package server

import (
	"image"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
)

type chestOpenNativeDeps4EDF00 struct {
	normalize func(*types.Pointf)
	mapTrace  func(*chestOpenRay4EDF00, *types.Pointf, *image.Point, uint8) int32
	refresh   func(*Object)
	drop      func(*Object, *Object, *types.Pointf) int32
}

func chestShapeExtentNative4EE2A0(obj *Object) float64 {
	return chestShapeExtent4EE2A0(obj, chestShapeExtentHooks4EE2A0[*Object]{
		loadShapeKind: func(obj *Object) uint32 {
			return uint32(obj.Shape.Kind)
		},
		loadCircleR: func(obj *Object) float32 {
			return obj.Shape.Circle.R
		},
		loadBoxExtentW: func(obj *Object) float32 {
			return obj.Shape.Box.W
		},
		loadBoxExtentH: func(obj *Object) float32 {
			return obj.Shape.Box.H
		},
	})
}

// chestOpenNative4EDF00 binds the restored control flow to native-width
// Object pointers and named fields. The Object layout may widen on a 64-bit
// target without changing the original game-field meanings.
func chestOpenNative4EDF00(
	chest, unit *Object,
	deps chestOpenNativeDeps4EDF00,
) {
	chestOpen4EDF00(chestOpenHooks4EDF00[*Object]{
		loadChestArg: func() *Object {
			return chest
		},
		loadUnitArg: func() *Object {
			return unit
		},
		countInventory: func(chest *Object, typeInd int32) int32 {
			return chest.CountInventoryWithType(typeInd)
		},
		loadSubClass: func(obj *Object) uint32 {
			return uint32(obj.ObjSubClass)
		},
		loadPosX: func(obj *Object) float32 {
			return obj.PosVec.X
		},
		loadPosY: func(obj *Object) float32 {
			return obj.PosVec.Y
		},
		normalize:   deps.normalize,
		shapeExtent: chestShapeExtentNative4EE2A0,
		mapTrace:    deps.mapTrace,
		firstItem: func(chest *Object) *Object {
			return chest.InvFirstItem
		},
		nextItem: func(item *Object) *Object {
			return item.InvNextItem
		},
		loadWeight: func(item *Object) uint8 {
			return item.Weight
		},
		loadClassLow: func(item *Object) uint8 {
			return uint8(item.ObjClass)
		},
		loadFlags: func(item *Object) uint32 {
			return uint32(item.ObjFlags)
		},
		storeFlags: func(item *Object, flags uint32) {
			item.ObjFlags = object.Flags(flags)
		},
		refresh: deps.refresh,
		drop:    deps.drop,
	})
}
