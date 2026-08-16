package server

import (
	"unsafe"

	"github.com/opennox/libs/types"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/sound"
)

// TreasureDropRuntime4ED710 supplies the services that remain legacy-owned.
// Object, PlayerUpdateData, Player, and Pointf access stays native-width.
type TreasureDropRuntime4ED710 struct {
	DefaultDrop func(*Object, *Object, *types.Pointf) int32
	TreasureMax func() uint32
	Report      func(*Object)
}

type treasureDropNativeDeps4ED710 struct {
	defaultDrop func(*Object, *Object, *types.Pointf) int32
	gameFlag    func(uint32) int32
	treasureMax func() uint32
	report      func(*Object)
	audio       func(uint32, *Object, int32, uint32)
}

func treasureDropNative4ED710(
	owner, treasure *Object,
	point *types.Pointf,
	deps treasureDropNativeDeps4ED710,
) int32 {
	return treasureDrop4ED710(treasureDropHooks4ED710[
		*Object,
		*PlayerUpdateData,
		*Player,
		*types.Pointf,
	]{
		loadPointArg: func() *types.Pointf {
			return point
		},
		loadTreasureArg: func() *Object {
			return treasure
		},
		loadOwnerArg: func() *Object {
			return owner
		},
		defaultDrop: deps.defaultDrop,
		loadClassLow: func(obj *Object) uint8 {
			return uint8(obj.ObjClass)
		},
		gameFlag: deps.gameFlag,
		loadUpdate: func(obj *Object) *PlayerUpdateData {
			return (*PlayerUpdateData)(obj.UpdateData)
		},
		loadPlayer: func(update *PlayerUpdateData) *Player {
			return update.Player
		},
		loadCount: func(player *Player) uint32 {
			return player.Field2152
		},
		storeCount: func(player *Player, value uint32) {
			player.Field2152 = value
		},
		treasureMax: deps.treasureMax,
		storeMax: func(player *Player, value uint32) {
			player.Field2156 = value
		},
		report: deps.report,
		audio:  deps.audio,
	})
}

func treasureDropServerDeps4ED710(
	s *Server,
	runtime TreasureDropRuntime4ED710,
) treasureDropNativeDeps4ED710 {
	return treasureDropNativeDeps4ED710{
		defaultDrop: runtime.DefaultDrop,
		gameFlag: func(flag uint32) int32 {
			if noxflags.HasGame(noxflags.GameFlag(flag)) {
				return 1
			}
			return 0
		},
		treasureMax: runtime.TreasureMax,
		report:      runtime.Report,
		audio: func(id uint32, obj *Object, kind int32, code uint32) {
			s.Audio.EventObj(sound.ID(id), obj, int(kind), code)
		},
	}
}

// TreasureDrop4ED710 binds GAME.EXE 004ED710 to native-width Object,
// PlayerUpdateData, Player, and Pointf pointers.
func (s *Server) TreasureDrop4ED710(
	owner, treasure *Object,
	point *types.Pointf,
	runtime TreasureDropRuntime4ED710,
) int32 {
	return treasureDropNative4ED710(owner, treasure, point, treasureDropServerDeps4ED710(s, runtime))
}

var (
	_ = [1]struct{}{}[8-unsafe.Sizeof(types.Pointf{})]
)
