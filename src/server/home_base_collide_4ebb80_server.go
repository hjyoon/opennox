package server

import (
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/sound"
)

// HomeBaseCollideRuntime4EBB80 supplies the effects that are still owned by
// the legacy/root runtime. Object, GameBall update-data, Player, Team, object
// list and motion fields stay native-width inside the server adapter.
type HomeBaseCollideRuntime4EBB80 struct {
	ObserverMode   func() uint32
	ObserverUpdate func(*Player, *Player)
	MoveTo         func(*Object, types.Pointf)
	PointFX        func(uint32, types.Pointf) uint32
}

type homeBaseCollideNativeDeps4EBB80 struct {
	lookupType        func(string) uint32
	teamByID          func(uint8) *Team
	changeScore       func(*Object, int32)
	reportLesson      func(*Object)
	changeTeamLessons func(*Team, int32)
	observerMode      func() uint32
	observerUpdate    func(*Player, *Player)
	audio             func(uint32, *Object)
	pointFX           func(uint32, types.Pointf) uint32
	firstObject       func() *Object
	nextObject        func(*Object) *Object
	randomInt         func(int32, int32) int32
	clearOwner        func(*Object)
	moveTo            func(*Object, types.Pointf)
}

func homeBaseCollideNative4EBB80(
	homeBase, other *Object,
	collision *types.Pointf,
	deps homeBaseCollideNativeDeps4EBB80,
) uint32 {
	return homeBaseCollide4EBB80(
		homeBase,
		other,
		collision,
		homeBaseCollideHooks4EBB80[
			*Object,
			*GameBallUpdateData4EA800,
			*Team,
			*PlayerUpdateData,
			*Player,
		]{
			lookupType: deps.lookupType,
			loadTypeIndex: func(obj *Object) uint16 {
				return obj.TypeInd
			},
			loadUpdate: func(obj *Object) *GameBallUpdateData4EA800 {
				return (*GameBallUpdateData4EA800)(obj.UpdateData)
			},
			loadCarrier: func(update *GameBallUpdateData4EA800) *Object {
				return update.Carrier
			},
			hasTeam: func(obj *Object) bool {
				return obj.TeamVal.Has()
			},
			loadTeamID: func(obj *Object) uint8 {
				return uint8(obj.TeamVal.ID)
			},
			teamByID:     deps.teamByID,
			changeScore:  deps.changeScore,
			reportLesson: deps.reportLesson,
			loadTeamLessons: func(team *Team) int32 {
				return team.Lessons
			},
			changeTeamLessons: deps.changeTeamLessons,
			observerMode:      deps.observerMode,
			loadPlayerUpdate: func(obj *Object) *PlayerUpdateData {
				return (*PlayerUpdateData)(obj.UpdateData)
			},
			loadPlayer: func(update *PlayerUpdateData) *Player {
				return update.Player
			},
			observerUpdate: deps.observerUpdate,
			audio:          deps.audio,
			pointFX: func(code uint32, obj *Object) uint32 {
				return deps.pointFX(code, obj.PosVec)
			},
			firstObject: deps.firstObject,
			nextObject:  deps.nextObject,
			randomInt: func(minimum, maximum int32, _ string, _ int32) int32 {
				return deps.randomInt(minimum, maximum)
			},
			clearOwner: deps.clearOwner,
			moveToMarker: func(obj, marker *Object) {
				deps.moveTo(obj, marker.PosVec)
			},
			storeVelocityX: func(obj *Object, value float32) {
				obj.VelVec.X = value
			},
			storeVelocityY: func(obj *Object, value float32) {
				obj.VelVec.Y = value
			},
			storeForceX: func(obj *Object, value float32) {
				obj.ForceVec.X = value
			},
			storePos24Y: func(obj *Object, value float32) {
				obj.Pos24.Y = value
			},
		},
	)
}

func homeBaseCollideServerDeps4EBB80(
	s *Server,
	runtime HomeBaseCollideRuntime4EBB80,
) homeBaseCollideNativeDeps4EBB80 {
	return homeBaseCollideNativeDeps4EBB80{
		lookupType: func(name string) uint32 {
			return uint32(s.Types.IndByID(name))
		},
		teamByID: func(id uint8) *Team {
			return s.Teams.ByID(TeamID(id))
		},
		changeScore: func(obj *Object, delta int32) {
			obj.changeScore(int(delta))
		},
		reportLesson: s.Nox_xxx_netReportLesson_4D8EF0,
		changeTeamLessons: func(team *Team, lessons int32) {
			s.TeamChangeLessons(team, int(lessons))
		},
		observerMode:   runtime.ObserverMode,
		observerUpdate: runtime.ObserverUpdate,
		audio: func(id uint32, obj *Object) {
			s.Audio.EventObj(sound.ID(id), obj, 0, 0)
		},
		pointFX:     runtime.PointFX,
		firstObject: s.Objs.First,
		nextObject: func(obj *Object) *Object {
			return obj.Next()
		},
		randomInt: func(minimum, maximum int32) int32 {
			return int32(s.Rand.Logic.IntClamp(int(minimum), int(maximum)))
		},
		clearOwner: s.ObjClearOwner,
		moveTo:     runtime.MoveTo,
	}
}

// HomeBaseCollide4EBB80 runs the native-width HomeBase scoring and GameBall
// respawn handler. The collision point is part of the registered callback ABI
// but, as in GAME.EXE, is never read.
func (s *Server) HomeBaseCollide4EBB80(
	homeBase, other *Object,
	collision *types.Pointf,
	runtime HomeBaseCollideRuntime4EBB80,
) uint32 {
	return homeBaseCollideNative4EBB80(
		homeBase,
		other,
		collision,
		homeBaseCollideServerDeps4EBB80(s, runtime),
	)
}
