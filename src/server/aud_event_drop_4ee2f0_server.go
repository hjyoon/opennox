package server

import (
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/sound"
)

type audEventDropSoundTable536AC0 struct {
	initialized uint32
	rows        [audEventDropRowStorage4EE2F0]audEventDropSoundRow4EE2F0
}

func (tab *audEventDropSoundTable536AC0) parse(objType *ObjectType, args []string) {
	var token *string
	if len(args) != 0 {
		token = &args[0]
	}
	audEventDropParse536AC0(audEventDropParseHooks536AC0[*string]{
		loadInit: func() uint32 {
			return tab.initialized
		},
		storeRowType: func(row int, value uint16) {
			tab.rows[row].typeInd = value
		},
		storeRowSound: func(row int, value uint16) {
			tab.rows[row].sound = value
		},
		storeInit: func(value uint32) {
			tab.initialized = value
		},
		loadRowType: func(row int) uint16 {
			return tab.rows[row].typeInd
		},
		nextToken: func() *string {
			return token
		},
		loadTokenByte: func(token *string) byte {
			if len(*token) == 0 {
				return 0
			}
			return (*token)[0]
		},
		resolveSound: func(token *string) uint16 {
			return uint16(sound.ByName(*token))
		},
		loadTypeInd: func() uint16 {
			return objType.ind
		},
	})
}

func (tab *audEventDropSoundTable536AC0) first(typeInd uint16) (uint16, bool) {
	for row := 0; row < len(tab.rows); row++ {
		if tab.rows[row].typeInd == audEventDropSentinel4EE2F0 {
			return 0, false
		}
		if tab.rows[row].typeInd == typeInd {
			return tab.rows[row].sound, true
		}
	}
	return 0, false
}

// AudEventDropRuntime4EE2F0 supplies the already-restored DefaultDrop
// dependency. Ordered sound lookup and audio are assembled from native Server
// state.
type AudEventDropRuntime4EE2F0 struct {
	DefaultDrop func(*Object, *Object, *types.Pointf) int32
}

type audEventDropNativeDeps4EE2F0 struct {
	defaultDrop  func(*Object, *Object, *types.Pointf) int32
	loadRowType  func(int) uint16
	loadRowSound func(int) uint16
	audio        func(uint32, *Object, int32, uint32)
}

func audEventDropNative4EE2F0(
	owner, item *Object,
	point *types.Pointf,
	deps audEventDropNativeDeps4EE2F0,
) int32 {
	return audEventDrop4EE2F0(audEventDropHooks4EE2F0[*Object, *types.Pointf]{
		loadOwnerArg: func() *Object {
			return owner
		},
		loadItemArg: func() *Object {
			return item
		},
		loadPointArg: func() *types.Pointf {
			return point
		},
		defaultDrop: deps.defaultDrop,
		loadRowType: deps.loadRowType,
		loadTypeInd: func(item *Object) uint16 {
			return item.TypeInd
		},
		loadRowSound: deps.loadRowSound,
		audio:        deps.audio,
	})
}

func audEventDropServerDeps4EE2F0(
	s *Server,
	runtime AudEventDropRuntime4EE2F0,
) audEventDropNativeDeps4EE2F0 {
	return audEventDropNativeDeps4EE2F0{
		defaultDrop: runtime.DefaultDrop,
		loadRowType: func(row int) uint16 {
			return s.Types.dropSoundTable.rows[row].typeInd
		},
		loadRowSound: func(row int) uint16 {
			return s.Types.dropSoundTable.rows[row].sound
		},
		audio: func(id uint32, obj *Object, kind int32, code uint32) {
			s.Audio.EventObj(sound.ID(id), obj, int(kind), code)
		},
	}
}

// AudEventDrop4EE2F0 binds GAME.EXE 004EE2F0 to native-width Object and
// Pointf pointers while preserving the original callback and table-read order.
func (s *Server) AudEventDrop4EE2F0(
	owner, item *Object,
	point *types.Pointf,
	runtime AudEventDropRuntime4EE2F0,
) int32 {
	return audEventDropNative4EE2F0(owner, item, point, audEventDropServerDeps4EE2F0(s, runtime))
}
