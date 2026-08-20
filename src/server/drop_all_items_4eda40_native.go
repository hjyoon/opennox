package server

import (
	"image"

	"github.com/opennox/libs/types"
)

type dropAllItemsNativeDeps4EDA40 struct {
	randomFloat func(float32, float32, string, int32) float64
	mapTrace    func(*dropAllItemsRay4EDA40, *types.Pointf, *image.Point, uint8) int32
	drop        func(*Object, *Object, *types.Pointf) int32
}

// dropAllItemsNative4EDA40 binds the restored control flow to native-width
// Object pointers. Named fields deliberately replace raw ABI32 offsets: on a
// 64-bit target their addresses expand while their game meaning stays fixed.
func dropAllItemsNative4EDA40(owner *Object, deps dropAllItemsNativeDeps4EDA40) int32 {
	return dropAllItems4EDA40(dropAllItemsHooks4EDA40[*Object, *Object]{
		loadOwnerArg: func() *Object {
			return owner
		},
		loadInventoryHead: func(owner *Object) *Object {
			return owner.InvFirstItem
		},
		loadInventoryNext: func(item *Object) *Object {
			return item.InvNextItem
		},
		dropEligible: dropEligibilityNative4EDCD0,
		loadItemRadius: func(item *Object) float32 {
			return item.Shape.Circle.R
		},
		loadOwnerX: func(owner *Object) float32 {
			return owner.PosVec.X
		},
		loadOwnerY: func(owner *Object) float32 {
			return owner.PosVec.Y
		},
		ownerPosition: func(owner *Object) *types.Pointf {
			return &owner.PosVec
		},
		randomFloat: deps.randomFloat,
		mapTrace:    deps.mapTrace,
		drop:        deps.drop,
	})
}
