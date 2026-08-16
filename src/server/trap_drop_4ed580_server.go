package server

import (
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/sound"
)

// TrapDropRuntime4ED580 contains the two predicates still supplied by the
// legacy-facing runtime. Audio and ownership are native Server services.
type TrapDropRuntime4ED580 struct {
	MapTileAllowTeleport func(*types.Pointf) int32
	DefaultDrop          func(*Object, *Object, *types.Pointf) int32
}

type trapDropNativeDeps4ED580 struct {
	mapTile     func(*types.Pointf) int32
	defaultDrop func(*Object, *Object, *types.Pointf) int32
	audio       func(uint32, *Object, int32, uint32)
	setOwner    func(*Object, *Object)
}

func trapDropNative4ED580(
	owner, glyph *Object,
	point *types.Pointf,
	deps trapDropNativeDeps4ED580,
) int32 {
	return trapDrop4ED580(trapDropHooks4ED580[*Object, *types.Pointf]{
		loadPointArg: func() *types.Pointf {
			return point
		},
		mapTile: deps.mapTile,
		loadOwnerArg: func() *Object {
			return owner
		},
		loadNetCode: func(owner *Object) uint32 {
			return owner.NetCode
		},
		audio: deps.audio,
		loadGlyphArg: func() *Object {
			return glyph
		},
		defaultDrop: deps.defaultDrop,
		setOwner:    deps.setOwner,
	})
}

// TrapDrop4ED580 binds GAME.EXE 004ED580 to native Object and Pointf pointers.
func (s *Server) TrapDrop4ED580(
	owner, glyph *Object,
	point *types.Pointf,
	runtime TrapDropRuntime4ED580,
) int32 {
	return trapDropNative4ED580(owner, glyph, point, trapDropNativeDeps4ED580{
		mapTile:     runtime.MapTileAllowTeleport,
		defaultDrop: runtime.DefaultDrop,
		audio: func(id uint32, obj *Object, kind int32, code uint32) {
			s.Audio.EventObj(sound.ID(id), obj, int(kind), code)
		},
		setOwner: s.ObjSetOwner,
	})
}
