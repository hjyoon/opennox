package server

import "github.com/opennox/libs/types"

type mapFindPlayerStartNativeDeps4F7AB0 struct {
	loadCachedType  func() uint32
	lookupType      func(string) uint32
	storeCachedType func(uint32)
	touchTeam       func(uint8)
	firstObject     func() *Object
	firstPlayer     func() *Object
	nextPlayer      func(*Object) *Object
	isEnemyTo       func(*Object, *Object) bool
	teamContains    func(*Object, uint8) bool
	randomInt       func(int32, int32, string, int32) int32
}

func mapFindPlayerStartNative4F7AB0(
	player *Object,
	output *types.Pointf,
	deps mapFindPlayerStartNativeDeps4F7AB0,
) {
	mapFindPlayerStart4F7AB0(player, mapFindPlayerStartHooks4F7AB0[*Object]{
		loadCachedType:  deps.loadCachedType,
		lookupType:      deps.lookupType,
		storeCachedType: deps.storeCachedType,
		hasTeam: func(object *Object) bool {
			return object.TeamVal.Has()
		},
		loadTeamID: func(object *Object) uint8 {
			return uint8(object.TeamVal.ID)
		},
		touchTeam:   deps.touchTeam,
		firstObject: deps.firstObject,
		nextObject: func(object *Object) *Object {
			return object.Next()
		},
		loadTypeIndex: func(object *Object) uint16 {
			return object.TypeInd
		},
		loadObjectFlags: func(object *Object) uint32 {
			return uint32(object.ObjFlags)
		},
		teamContains: deps.teamContains,
		firstPlayer:  deps.firstPlayer,
		nextPlayer:   deps.nextPlayer,
		isEnemyTo:    deps.isEnemyTo,
		loadPosX: func(object *Object) float32 {
			return object.PosVec.X
		},
		loadPosY: func(object *Object) float32 {
			return object.PosVec.Y
		},
		randomInt: deps.randomInt,
		storeOutputX: func(value float32) {
			output.X = value
		},
		storeOutputY: func(value float32) {
			output.Y = value
		},
	})
}

func mapFindPlayerStartServerDeps4F7AB0(s *Server) mapFindPlayerStartNativeDeps4F7AB0 {
	return mapFindPlayerStartNativeDeps4F7AB0{
		loadCachedType: func() uint32 {
			return s.Types.fast.playerStart4F7AB0
		},
		lookupType: func(name string) uint32 {
			return uint32(s.Types.IndByID(name))
		},
		storeCachedType: func(value uint32) {
			s.Types.fast.playerStart4F7AB0 = value
		},
		touchTeam: func(id uint8) {
			_ = s.Teams.ByID(TeamID(id))
		},
		firstObject: s.Objs.First,
		firstPlayer: s.Players.FirstUnit,
		nextPlayer:  s.Players.NextUnit,
		isEnemyTo:   s.IsEnemyTo,
		teamContains: func(value *Object, id uint8) bool {
			return s.Teams.ContainsObject(&value.TeamVal, TeamID(id))
		},
		randomInt: func(minimum, maximum int32, _ string, _ int32) int32 {
			return int32(s.Rand.Logic.IntClamp(int(minimum), int(maximum)))
		},
	}
}

// MapFindPlayerStartInto4F7AB0 binds GAME.EXE 004F7AB0 to native-width
// objects. A nil player leaves output untouched; all other nil behavior is the
// original function's behavior and is deliberately not guarded here.
func (s *Server) MapFindPlayerStartInto4F7AB0(output *types.Pointf, player *Object) {
	mapFindPlayerStartNative4F7AB0(player, output, mapFindPlayerStartServerDeps4F7AB0(s))
}

// MapFindPlayerStart4F7AB0 is the value-oriented Go API over the pointer-
// accurate implementation.
func (s *Server) MapFindPlayerStart4F7AB0(player *Object) types.Pointf {
	var output types.Pointf
	s.MapFindPlayerStartInto4F7AB0(&output, player)
	return output
}
