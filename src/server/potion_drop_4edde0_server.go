package server

import (
	"github.com/opennox/libs/types"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/sound"
)

// PotionDropRuntime4EDDE0 supplies the DefaultDrop dependency whose remaining
// services are still assembled at the legacy boundary. Audio, game flags,
// TickRate, and decay scheduling are native Server services.
type PotionDropRuntime4EDDE0 struct {
	DefaultDrop func(*Object, *Object, *types.Pointf) int32
}

type potionDropNativeDeps4EDDE0 struct {
	defaultDrop func(*Object, *Object, *types.Pointf) int32
	audio       func(uint32, *Object, int32, uint32)
	gameFlag    func(uint32) int32
	loadGameFPS func() uint32
	setDecay    func(*Object, uint32)
}

func potionDropNative4EDDE0(
	owner, item *Object,
	point *types.Pointf,
	deps potionDropNativeDeps4EDDE0,
) int32 {
	return potionDrop4EDDE0(potionDropHooks4EDDE0[*Object, *types.Pointf]{
		loadPointArg: func() *types.Pointf {
			return point
		},
		loadOwnerArg: func() *Object {
			return owner
		},
		loadItemArg: func() *Object {
			return item
		},
		defaultDrop: deps.defaultDrop,
		audio:       deps.audio,
		gameFlag:    deps.gameFlag,
		loadGameFPS: deps.loadGameFPS,
		setDecay:    deps.setDecay,
	})
}

func potionDropServerDeps4EDDE0(
	s *Server,
	runtime PotionDropRuntime4EDDE0,
) potionDropNativeDeps4EDDE0 {
	return potionDropNativeDeps4EDDE0{
		defaultDrop: runtime.DefaultDrop,
		audio: func(id uint32, obj *Object, kind int32, code uint32) {
			s.Audio.EventObj(sound.ID(id), obj, int(kind), code)
		},
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
	}
}

// PotionDrop4EDDE0 binds GAME.EXE 004EDDE0 to native-width Object and Pointf
// pointers while preserving the original callback order.
func (s *Server) PotionDrop4EDDE0(
	owner, item *Object,
	point *types.Pointf,
	runtime PotionDropRuntime4EDDE0,
) int32 {
	return potionDropNative4EDDE0(owner, item, point, potionDropServerDeps4EDDE0(s, runtime))
}
