package server

import (
	"unsafe"

	"github.com/opennox/opennox/v1/common/ntype"
	"github.com/opennox/opennox/v1/common/sound"
)

// SpellGestureCancelAllocator4FE680 snapshots the root-owned allocator at the
// exact point where GAME.EXE 004FE680 loads its allocator global.
type SpellGestureCancelAllocator4FE680 func(*MagicEntityClass)

// SpellGestureCancelRuntime4FE680 supplies the player-state operation and the
// intrusive queue globals still owned by the root runtime.
type SpellGestureCancelRuntime4FE680 struct {
	SetPlayerState func(*Object, PlayerState)
	LoadHead       func() *MagicEntityClass
	StoreHead      func(*MagicEntityClass)
	LoadAllocator  func() SpellGestureCancelAllocator4FE680
}

type spellGestureCancelNativeDeps4FE680 struct {
	compareTeams   func(*ObjectTeam, *ObjectTeam) int32
	mapCheck       func(*Object, *Object) int32
	informResult   func(uint8, uint8, int32)
	audioEvent     func(int32, *Object, int32, uint32)
	setPlayerState func(*Object, int32)
	loadHead       func() *MagicEntityClass
	storeHead      func(*MagicEntityClass)
	loadAllocator  func() SpellGestureCancelAllocator4FE680
}

// spellGestureCancelObjectTeam4FE680 models the original LEA/ADD of the
// inline ObjectTeam field. For valid native objects the result is an ordinary
// field pointer; the helper deliberately adds no synthetic nil guard.
func spellGestureCancelObjectTeam4FE680(object *Object) *ObjectTeam {
	return (*ObjectTeam)(unsafe.Add(unsafe.Pointer(object), unsafe.Offsetof(Object{}.TeamVal)))
}

func spellGestureCancelNative4FE680(
	source *Object,
	radius float32,
	deps spellGestureCancelNativeDeps4FE680,
) {
	spellGestureCancel4FE680(spellGestureCancelHooks4FE680[
		*MagicEntityClass,
		*Object,
		*ObjectTeam,
		*PlayerUpdateData,
		*Player,
		SpellGestureCancelAllocator4FE680,
	]{
		loadHead: deps.loadHead,
		loadSourceArg: func() *Object {
			return source
		},
		loadObject: func(entity *MagicEntityClass) *Object {
			return entity.Obj4
		},
		loadClass: func(object *Object) uint32 {
			return uint32(object.ObjClass)
		},
		loadTeam:     spellGestureCancelObjectTeam4FE680,
		compareTeams: deps.compareTeams,
		loadPosX: func(object *Object) float32 {
			return object.PosVec.X
		},
		loadPosY: func(object *Object) float32 {
			return object.PosVec.Y
		},
		loadRadiusArg: func() float32 {
			return radius
		},
		mapCheck: deps.mapCheck,
		loadUpdate: func(object *Object) *PlayerUpdateData {
			return (*PlayerUpdateData)(object.UpdateData)
		},
		storeSpellCastStart: func(update *PlayerUpdateData, value uint32) {
			update.SpellCastStart = value
		},
		storeCasting: func(update *PlayerUpdateData, value uint8) {
			update.Field47_0 = value
		},
		loadPlayer: func(update *PlayerUpdateData) *Player {
			return update.Player
		},
		loadPlayerIndex: func(player *Player) uint8 {
			return player.PlayerInd
		},
		informResult:   deps.informResult,
		audioEvent:     deps.audioEvent,
		setPlayerState: deps.setPlayerState,
		loadNext: func(entity *MagicEntityClass) *MagicEntityClass {
			return entity.Next52
		},
		loadPrev: func(entity *MagicEntityClass) *MagicEntityClass {
			return entity.Prev56
		},
		storePrev: func(entity, prev *MagicEntityClass) {
			entity.Prev56 = prev
		},
		storeNext: func(entity, next *MagicEntityClass) {
			entity.Next52 = next
		},
		storeHead:     deps.storeHead,
		loadAllocator: deps.loadAllocator,
		free: func(allocator SpellGestureCancelAllocator4FE680, entity *MagicEntityClass) {
			allocator(entity)
		},
	})
}

func spellGestureCancelServerDeps4FE680(
	s *Server,
	runtime SpellGestureCancelRuntime4FE680,
) spellGestureCancelNativeDeps4FE680 {
	return spellGestureCancelNativeDeps4FE680{
		compareTeams: func(first, second *ObjectTeam) int32 {
			if first.SameAs(second) {
				return 1
			}
			return 0
		},
		mapCheck: func(source, target *Object) int32 {
			if s.MapTraceVision(source, target) {
				return 1
			}
			return 0
		},
		informResult: func(playerIndex, code uint8, result int32) {
			_ = s.NetInformTextMsg(ntype.PlayerInd(playerIndex), byte(code), int(result))
		},
		audioEvent: func(id int32, object *Object, kind int32, code uint32) {
			s.Audio.EventObj(sound.ID(id), object, int(kind), code)
		},
		setPlayerState: func(object *Object, state int32) {
			runtime.SetPlayerState(object, PlayerState(state))
		},
		loadHead:      runtime.LoadHead,
		storeHead:     runtime.StoreHead,
		loadAllocator: runtime.LoadAllocator,
	}
}

// SpellGestureCancel4FE680 cancels hostile queued spell gestures within the
// caller-supplied Warcry radius using native-width Object and list pointers.
//
//go:noinline
func (s *Server) SpellGestureCancel4FE680(
	source *Object,
	radius float32,
	runtime SpellGestureCancelRuntime4FE680,
) {
	spellGestureCancelNative4FE680(source, radius, spellGestureCancelServerDeps4FE680(s, runtime))
}

var (
	_ = [1]struct{}{}[4-unsafe.Sizeof(Object{}.ObjClass)]
	_ = [1]struct{}{}[4-unsafe.Sizeof(Object{}.PosVec.X)]
	_ = [1]struct{}{}[4-unsafe.Sizeof(PlayerUpdateData{}.SpellCastStart)]
	_ = [1]struct{}{}[1-unsafe.Sizeof(PlayerUpdateData{}.Field47_0)]
	_ = [1]struct{}{}[1-unsafe.Sizeof(Player{}.PlayerInd)]
)
