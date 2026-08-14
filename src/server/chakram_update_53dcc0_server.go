package server

type ChakramInMotionUpdateRuntime53DCC0 struct {
	DelayedDelete func(*Object)
}

type chakramUpdateNativeDeps53DCC0 struct {
	mapCheck      func(*Object, *Object) bool
	frame         func() uint32
	frameRate     func() uint32
	delayedDelete func(*Object)
}

func chakramUpdateNative53DCC0(source *Object, deps chakramUpdateNativeDeps53DCC0) {
	chakramUpdate53DCC0(source, chakramUpdateHooks53DCC0[*Object, *ChakramUpdateData]{
		loadUpdateData: func(obj *Object) *ChakramUpdateData {
			return (*ChakramUpdateData)(obj.UpdateData)
		},
		inventoryFirst: func(obj *Object) *Object {
			return obj.InvFirstItem
		},
		loadFlags: func(obj *Object) uint32 {
			return uint32(obj.ObjFlags)
		},
		loadLastHit: func(update *ChakramUpdateData) *Object {
			return update.LastHit
		},
		storeLastHit: func(update *ChakramUpdateData, obj *Object) {
			update.LastHit = obj
		},
		loadOwner: func(obj *Object) *Object {
			return obj.ObjOwner
		},
		loadPosX: func(obj *Object) float32 {
			return obj.PosVec.X
		},
		loadPosY: func(obj *Object) float32 {
			return obj.PosVec.Y
		},
		storeOwnerPosX: func(update *ChakramUpdateData, value float32) {
			update.OwnerPos.X = value
		},
		storeOwnerPosY: func(update *ChakramUpdateData, value float32) {
			update.OwnerPos.Y = value
		},
		loadOwnerPosX: func(update *ChakramUpdateData) float32 {
			return update.OwnerPos.X
		},
		loadOwnerPosY: func(update *ChakramUpdateData) float32 {
			return update.OwnerPos.Y
		},
		mapCheck: deps.mapCheck,
		loadReturnState: func(update *ChakramUpdateData) uint8 {
			return update.ReturnState
		},
		storeReturnState: func(update *ChakramUpdateData, value uint8) {
			update.ReturnState = value
		},
		loadReturnTarget: func(update *ChakramUpdateData) *Object {
			return update.ReturnTarget
		},
		storeReturnTarget: func(update *ChakramUpdateData, obj *Object) {
			update.ReturnTarget = obj
		},
		loadSpeed: func(obj *Object) float32 {
			return obj.SpeedCur
		},
		storeVelocityX: func(obj *Object, value float32) {
			obj.VelVec.X = value
		},
		storeVelocityY: func(obj *Object, value float32) {
			obj.VelVec.Y = value
		},
		frame: deps.frame,
		loadCreateFrame: func(obj *Object) uint32 {
			return obj.Field32
		},
		frameRate:     deps.frameRate,
		delayedDelete: deps.delayedDelete,
	})
}

// ChakramInMotionUpdate53DCC0 binds the registered update callback to the
// native-width ChakramUpdateData record and server map/frame services.
func (s *Server) ChakramInMotionUpdate53DCC0(source *Object, runtime ChakramInMotionUpdateRuntime53DCC0) {
	chakramUpdateNative53DCC0(source, chakramUpdateNativeDeps53DCC0{
		mapCheck:      s.MapTraceVision,
		frame:         s.Frame,
		frameRate:     func() uint32 { return uint32(s.TickRate()) },
		delayedDelete: runtime.DelayedDelete,
	})
}
