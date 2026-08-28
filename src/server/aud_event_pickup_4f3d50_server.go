package server

import "github.com/opennox/opennox/v1/common/sound"

type audEventPickupSoundTable5367B0 struct {
	initialized uint32
	rows        [audEventPickupRowStorage4F3D50]audEventPickupSoundRow4F3D50
}

func (tab *audEventPickupSoundTable5367B0) parse(objType *ObjectType, args []string) {
	var token *string
	if len(args) != 0 {
		token = &args[0]
	}
	audEventPickupParse5367B0(audEventPickupParseHooks5367B0[*string]{
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

func (tab *audEventPickupSoundTable5367B0) first(typeInd uint16) (uint16, bool) {
	for row := 0; row < len(tab.rows); row++ {
		if tab.rows[row].typeInd == audEventPickupSentinel4F3D50 {
			return 0, false
		}
		if tab.rows[row].typeInd == typeInd {
			return tab.rows[row].sound, true
		}
	}
	return 0, false
}

// AudEventPickupRuntime4F3D50 supplies the already-restored DefaultPickup
// dependency. Ordered sound lookup and audio are assembled from native Server
// state.
type AudEventPickupRuntime4F3D50 struct {
	DefaultPickup func(*Object, *Object, int32, int32) int32
}

type audEventPickupNativeDeps4F3D50 struct {
	defaultPickup func(*Object, *Object, int32, int32) int32
	loadRowType   func(int) uint16
	loadRowSound  func(int) uint16
	audio         func(uint32, *Object, int32, uint32)
}

func audEventPickupNative4F3D50(
	owner, item *Object,
	arg3, arg4 int32,
	deps audEventPickupNativeDeps4F3D50,
) int32 {
	return audEventPickup4F3D50(audEventPickupHooks4F3D50[*Object]{
		loadOwnerArg: func() *Object {
			return owner
		},
		loadItemArg: func() *Object {
			return item
		},
		loadArg4: func() int32 {
			return arg4
		},
		loadArg3: func() int32 {
			return arg3
		},
		defaultPickup: deps.defaultPickup,
		loadRowType:   deps.loadRowType,
		loadTypeInd: func(item *Object) uint16 {
			return item.TypeInd
		},
		loadRowSound: deps.loadRowSound,
		audio:        deps.audio,
	})
}

func audEventPickupServerDeps4F3D50(
	s *Server,
	runtime AudEventPickupRuntime4F3D50,
) audEventPickupNativeDeps4F3D50 {
	return audEventPickupNativeDeps4F3D50{
		defaultPickup: runtime.DefaultPickup,
		loadRowType: func(row int) uint16 {
			return s.Types.pickupSoundTable.rows[row].typeInd
		},
		loadRowSound: func(row int) uint16 {
			return s.Types.pickupSoundTable.rows[row].sound
		},
		audio: func(id uint32, obj *Object, kind int32, code uint32) {
			s.Audio.EventObj(sound.ID(id), obj, int(kind), code)
		},
	}
}

// AudEventPickup4F3D50 binds GAME.EXE 004F3D50 to native-width Object
// pointers while preserving the original callback, scalar, and table order.
func (s *Server) AudEventPickup4F3D50(
	owner, item *Object,
	arg3, arg4 int32,
	runtime AudEventPickupRuntime4F3D50,
) int32 {
	return audEventPickupNative4F3D50(owner, item, arg3, arg4, audEventPickupServerDeps4F3D50(s, runtime))
}
