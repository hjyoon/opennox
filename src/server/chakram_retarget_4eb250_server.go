package server

import "github.com/opennox/libs/types"

type chakramRetargetNativeDeps4EB250 struct {
	eachInRect func(types.Rectf, func(*Object))
	mapCheck   func(*Object, *Object) bool
}

// chakramRetargetNative4EB250 binds the recovered nearest-target search to
// native-width Object and ChakramUpdateData pointers.
func chakramRetargetNative4EB250(source *Object, deps chakramRetargetNativeDeps4EB250) *Object {
	return chakramRetarget4EB250(source, chakramRetargetHooks4EB250[*Object, *ChakramUpdateData]{
		loadUpdateData: func(obj *Object) *ChakramUpdateData {
			return (*ChakramUpdateData)(obj.UpdateData)
		},
		loadLastHit: func(update *ChakramUpdateData) *Object {
			return update.LastHit
		},
		loadOwner: func(obj *Object) *Object {
			return obj.ObjOwner
		},
		loadClass: func(obj *Object) uint32 {
			return uint32(obj.ObjClass)
		},
		loadFlags: func(obj *Object) uint32 {
			return uint32(obj.ObjFlags)
		},
		hasEnchant: func(obj *Object, enchant uint32) bool {
			return obj.HasEnchant(EnchantID(enchant))
		},
		mapCheck: deps.mapCheck,
		loadPosX: func(obj *Object) float32 {
			return obj.PosVec.X
		},
		loadPosY: func(obj *Object) float32 {
			return obj.PosVec.Y
		},
		loadSpeed: func(obj *Object) float32 {
			return obj.SpeedCur
		},
		eachInRect: func(rect chakramRetargetRect4EB250, callback func(*Object)) {
			deps.eachInRect(types.Rectf{
				Min: types.Pointf{X: rect.MinX, Y: rect.MinY},
				Max: types.Pointf{X: rect.MaxX, Y: rect.MaxY},
			}, callback)
		},
		storeState: func(update *ChakramUpdateData, state uint8) {
			update.ReturnState = state
		},
		storeVelocityX: func(obj *Object, value float32) {
			obj.VelVec.X = value
		},
		storeVelocityY: func(obj *Object, value float32) {
			obj.VelVec.Y = value
		},
	})
}
