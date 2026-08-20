package server

import (
	"github.com/opennox/libs/types"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/sound"
)

// FoodDropRuntime4EDE50 supplies the DefaultDrop dependency whose remaining
// services are assembled from native Server state.
type FoodDropRuntime4EDE50 struct {
	DefaultDrop func(*Object, *Object, *types.Pointf) int32
}

type foodDropNativeDeps4EDE50 struct {
	defaultDrop func(*Object, *Object, *types.Pointf) int32
	gameFlag    func(uint32) int32
	loadGameFPS func() uint32
	setDecay    func(*Object, uint32)
	audio       func(uint32, *Object, int32, uint32)
}

func foodDropNative4EDE50(
	owner, food *Object,
	point *types.Pointf,
	deps foodDropNativeDeps4EDE50,
) int32 {
	return foodDrop4EDE50(foodDropHooks4EDE50[*Object, *types.Pointf]{
		loadOwnerArg: func() *Object {
			return owner
		},
		loadFoodArg: func() *Object {
			return food
		},
		loadPointArg: func() *types.Pointf {
			return point
		},
		defaultDrop: deps.defaultDrop,
		gameFlag:    deps.gameFlag,
		loadGameFPS: deps.loadGameFPS,
		setDecay:    deps.setDecay,
		loadRuleSound: func(row int) uint16 {
			return foodDropSoundRules4EDE50[row].sound
		},
		loadSubClass: func(food *Object) uint32 {
			return uint32(food.ObjSubClass)
		},
		loadRuleSubClassMask: func(row int) uint32 {
			return foodDropSoundRules4EDE50[row].subClassMask
		},
		loadRuleFlagsLowMask: func(row int) uint16 {
			return foodDropSoundRules4EDE50[row].flagsLowMask
		},
		loadFlagsLow: func(food *Object) uint16 {
			return uint16(food.ObjFlags)
		},
		audio: deps.audio,
	})
}

func foodDropServerDeps4EDE50(
	s *Server,
	runtime FoodDropRuntime4EDE50,
) foodDropNativeDeps4EDE50 {
	return foodDropNativeDeps4EDE50{
		defaultDrop: runtime.DefaultDrop,
		gameFlag: func(flag uint32) int32 {
			if noxflags.HasGame(noxflags.GameFlag(flag)) {
				return 1
			}
			return 0
		},
		loadGameFPS: s.TickRate,
		setDecay: func(obj *Object, delay uint32) {
			s.DecaySetTime511660(obj, delay)
		},
		audio: func(id uint32, obj *Object, kind int32, code uint32) {
			s.Audio.EventObj(sound.ID(id), obj, int(kind), code)
		},
	}
}

// FoodDrop4EDE50 binds GAME.EXE 004EDE50 to native-width Object and Pointf
// pointers while preserving the original field-read and callback order.
func (s *Server) FoodDrop4EDE50(
	owner, food *Object,
	point *types.Pointf,
	runtime FoodDropRuntime4EDE50,
) int32 {
	return foodDropNative4EDE50(owner, food, point, foodDropServerDeps4EDE50(s, runtime))
}
