package server

import "github.com/opennox/opennox/v1/common/sound"

// PickupFoodRuntime4F3350 supplies the two root-owned object-list operations
// needed by FoodPickup's nested DefaultPickup call.
type PickupFoodRuntime4F3350 struct {
	DefaultPickup PickupDefaultRuntime4F31E0
}

type pickupFoodNativeDeps4F3350 struct {
	playerState   func(*Object) int32
	defaultPickup func(*Object, *Object, int32, int32) int32
	audio         func(uint32, *Object, int32, uint32)
}

func pickupFoodNative4F3350(
	owner, item *Object,
	arg3, arg4 int32,
	deps pickupFoodNativeDeps4F3350,
) int32 {
	return pickupFood4F3350(
		owner,
		item,
		arg3,
		arg4,
		pickupFoodHooks4F3350[*Object, UseFunc]{
			playerState: deps.playerState,
			loadSubClassLow: func(item *Object) uint8 {
				return uint8(item.ObjSubClass)
			},
			loadUse: func(item *Object) UseFunc {
				return item.Use.Get()
			},
			callUse: func(use UseFunc, owner, item *Object) int32 {
				if use(owner, item) {
					return 1
				}
				return 0
			},
			loadFlagsLow: func(item *Object) uint8 {
				return uint8(item.ObjFlags)
			},
			defaultPickup: deps.defaultPickup,
			loadRuleSound: func(row int) uint16 {
				return pickupFoodSoundRules4F3350[row].sound
			},
			loadSubClass: func(item *Object) uint32 {
				return uint32(item.ObjSubClass)
			},
			loadRuleSubClassMask: func(row int) uint32 {
				return pickupFoodSoundRules4F3350[row].subClassMask
			},
			loadRuleMaterialMask: func(row int) uint16 {
				return pickupFoodSoundRules4F3350[row].materialMask
			},
			loadMaterialLow: func(item *Object) uint16 {
				return item.Material
			},
			audio: deps.audio,
		},
	)
}

func pickupFoodServerDeps4F3350(
	s *Server,
	runtime PickupFoodRuntime4F3350,
) pickupFoodNativeDeps4F3350 {
	return pickupFoodNativeDeps4F3350{
		playerState: func(owner *Object) int32 {
			if s.Players.CheckXxx(owner) {
				return 1
			}
			return 0
		},
		defaultPickup: func(owner, item *Object, arg3, arg4 int32) int32 {
			return s.PickupDefault4F31E0(owner, item, arg3, arg4, runtime.DefaultPickup)
		},
		audio: func(id uint32, owner *Object, kind int32, code uint32) {
			s.Audio.EventObj(sound.ID(id), owner, int(kind), code)
		},
	}
}

// PickupFood4F3350 binds GAME.EXE's registered four-argument FoodPickup
// callback to native-width Object and Use-function fields.
func (s *Server) PickupFood4F3350(
	owner, item *Object,
	arg3, arg4 int32,
	runtime PickupFoodRuntime4F3350,
) int32 {
	return pickupFoodNative4F3350(
		owner,
		item,
		arg3,
		arg4,
		pickupFoodServerDeps4F3350(s, runtime),
	)
}
