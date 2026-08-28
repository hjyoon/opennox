package server

import "github.com/opennox/opennox/v1/common/sound"

// PickupAnkhTradableRuntime4F3DD0 supplies the root-owned delayed object
// deletion operation. Audio and native player data remain server-owned.
type PickupAnkhTradableRuntime4F3DD0 struct {
	DelayedDelete func(*Object)
}

type pickupAnkhTradableNativeDeps4F3DD0 struct {
	delayedDelete func(*Object)
	audio         func(uint32, *Object, int32, uint32)
}

func pickupAnkhTradableNative4F3DD0(
	owner, item *Object,
	arg3, arg4 int32,
	deps pickupAnkhTradableNativeDeps4F3DD0,
) int32 {
	_ = arg3
	_ = arg4
	return pickupAnkhTradable4F3DD0(owner, pickupAnkhTradableHooks4F3DD0[
		*Object,
		*PlayerUpdateData,
	]{
		loadOwnerClassLow: func(owner *Object) uint8 {
			return uint8(owner.ObjClass)
		},
		loadOwnerUpdate: func(owner *Object) *PlayerUpdateData {
			return (*PlayerUpdateData)(owner.UpdateData)
		},
		loadExtraLives: func(update *PlayerUpdateData) uint32 {
			return update.ExtraLives
		},
		storeExtraLives: func(update *PlayerUpdateData, value uint32) {
			update.ExtraLives = value
		},
		loadItemArg: func() *Object {
			return item
		},
		delayedDelete: deps.delayedDelete,
		audio:         deps.audio,
	})
}

func pickupAnkhTradableServerDeps4F3DD0(
	s *Server,
	runtime PickupAnkhTradableRuntime4F3DD0,
) pickupAnkhTradableNativeDeps4F3DD0 {
	return pickupAnkhTradableNativeDeps4F3DD0{
		delayedDelete: runtime.DelayedDelete,
		audio: func(id uint32, owner *Object, kind int32, code uint32) {
			s.Audio.EventObj(sound.ID(id), owner, int(kind), code)
		},
	}
}

// PickupAnkhTradable4F3DD0 binds GAME.EXE's registered four-argument
// AnkhTradablePickup callback to native-width Object and PlayerUpdateData
// pointers. The two trailing callback arguments remain deliberately unread.
func (s *Server) PickupAnkhTradable4F3DD0(
	owner, item *Object,
	arg3, arg4 int32,
	runtime PickupAnkhTradableRuntime4F3DD0,
) int32 {
	return pickupAnkhTradableNative4F3DD0(
		owner,
		item,
		arg3,
		arg4,
		pickupAnkhTradableServerDeps4F3DD0(s, runtime),
	)
}
