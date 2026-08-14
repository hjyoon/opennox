package server

import (
	"github.com/opennox/libs/types"

	noxflags "github.com/opennox/opennox/v1/common/flags"
)

// FlagCollideRuntime4EA400 supplies the two downstream pickup callbacks. They
// remain separate restoration units at GAME.EXE 004EA490 and 004EA800.
type FlagCollideRuntime4EA400 struct {
	PickupCTF      func(*Object, *Object, *types.Pointf)
	PickupGameBall func(*Object, *Object, *types.Pointf)
}

type flagCollideNativeDeps4EA400 struct {
	hasGameFlag       func(uint32) int32
	loadGameBallCache func() uint32
	lookupGameBall    func(string) uint32
	storeGameBall     func(uint32)
	pickupCTF         func(*Object, *Object, *types.Pointf)
	pickupGameBall    func(*Object, *Object, *types.Pointf)
}

func flagCollideNative4EA400(
	source, target *Object,
	collision *types.Pointf,
	deps flagCollideNativeDeps4EA400,
) {
	flagCollide4EA400(source, target, collision, flagCollideHooks4EA400[*Object, *types.Pointf]{
		loadFlags: func(obj *Object) uint32 {
			return uint32(obj.ObjFlags)
		},
		hasGameFlag:       deps.hasGameFlag,
		loadGameBallCache: deps.loadGameBallCache,
		lookupGameBall:    deps.lookupGameBall,
		storeGameBall:     deps.storeGameBall,
		loadTypeInd: func(obj *Object) uint16 {
			return obj.TypeInd
		},
		loadClassLow: func(obj *Object) uint8 {
			return uint8(obj.ObjClass)
		},
		pickupCTF:      deps.pickupCTF,
		pickupGameBall: deps.pickupGameBall,
	})
}

func flagCollideServerDeps4EA400(
	s *Server,
	runtime FlagCollideRuntime4EA400,
) flagCollideNativeDeps4EA400 {
	return flagCollideNativeDeps4EA400{
		hasGameFlag: func(mask uint32) int32 {
			if noxflags.HasGame(noxflags.GameFlag(mask)) {
				return 1
			}
			return 0
		},
		loadGameBallCache: func() uint32 {
			return uint32(s.Types.fast.ball)
		},
		lookupGameBall: func(name string) uint32 {
			return uint32(s.Types.IndByID(name))
		},
		storeGameBall: func(ind uint32) {
			s.Types.fast.ball = int(ind)
		},
		pickupCTF:      runtime.PickupCTF,
		pickupGameBall: runtime.PickupGameBall,
	}
}

// FlagCollide4EA400 routes FlagCollide through native Object fields. The two
// pickup implementations are injected until their sequential restoration is
// complete; this function is not exported to legacy C before that point.
func (s *Server) FlagCollide4EA400(
	source, target *Object,
	collision *types.Pointf,
	runtime FlagCollideRuntime4EA400,
) {
	flagCollideNative4EA400(source, target, collision, flagCollideServerDeps4EA400(s, runtime))
}
