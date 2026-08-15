package server

import (
	"github.com/opennox/libs/object"
	"github.com/opennox/libs/strman"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/sound"
)

// BallCollideRuntime4EBA00 supplies team/network effects whose shared service
// bodies remain separate restoration units. Object, Team, owned-list and
// GameBall update-data pointers stay native-width across this boundary.
type BallCollideRuntime4EBA00 struct {
	TeamMemberCount   func(*Team) int32
	LoadFeedbackFrame func() uint32
	StoreFeedback     func(uint32)
	ChangeTeam        func(*ObjectTeam, *Team, uint32, int32) int32
	CreateTeam        func(TeamID, *ObjectTeam, int32, uint32, int32)
	BallStatus        func(uint8, uint16) int32
	BuffPurge         FlagPickupBuffPurgeRuntime4EA7A0
}

type ballCollideNativeDeps4EBA00 struct {
	teamByID        func(uint8) *Team
	teamMemberCount func(*Team) int32
	loadFrame       func() uint32
	loadFeedback    func() uint32
	priorityMessage func(*Object, string, int32)
	storeFeedback   func(uint32)
	audio           func(uint32, *Object)
	setOwner        func(*Object, *Object)
	changeTeam      func(*ObjectTeam, *Team, uint32, int32) int32
	createTeam      func(TeamID, *ObjectTeam, int32, uint32, int32)
	ballStatus      func(uint8, uint16) int32
	carrierState    func(*Object, *Object) *Object
	purgeBuffs      func(*Object)
}

func ballCollideNative4EBA00(
	ball, target *Object,
	collision *types.Pointf,
	deps ballCollideNativeDeps4EBA00,
) {
	ballCollide4EBA00(ball, target, collision, ballCollideHooks4EBA00[
		*Object,
		*Team,
		*GameBallUpdateData4EA800,
	]{
		loadUpdateData: func(obj *Object) *GameBallUpdateData4EA800 {
			return (*GameBallUpdateData4EA800)(obj.UpdateData)
		},
		loadTeamID: func(obj *Object) uint8 {
			return uint8(obj.TeamVal.ID)
		},
		findTeamByID: deps.teamByID,
		loadCarrier: func(data *GameBallUpdateData4EA800) *Object {
			return data.Carrier
		},
		teamMemberCount:   deps.teamMemberCount,
		loadFrame:         deps.loadFrame,
		loadFeedbackFrame: deps.loadFeedback,
		priorityMessage:   deps.priorityMessage,
		storeFeedback:     deps.storeFeedback,
		audio:             deps.audio,
		loadOwner: func(obj *Object) *Object {
			return obj.ObjOwner
		},
		loadClassLow: func(obj *Object) uint8 {
			return uint8(obj.ObjClass)
		},
		loadOwnedFirst: func(obj *Object) *Object {
			return obj.Field129
		},
		loadTypeInd: func(obj *Object) uint16 {
			return obj.TypeInd
		},
		loadOwnedNext: func(obj *Object) *Object {
			return obj.Field128
		},
		setOwner: deps.setOwner,
		hasTeam: func(obj *Object) int32 {
			if obj.TeamVal.Has() {
				return 1
			}
			return 0
		},
		loadNetCode: func(obj *Object) uint32 {
			return obj.NetCode
		},
		changeTeam: func(obj *Object, team *Team, netCode uint32, flags int32) {
			_ = deps.changeTeam(&obj.TeamVal, team, netCode, flags)
		},
		createTeam: func(id uint8, obj *Object, active int32, netCode uint32, flags int32) {
			deps.createTeam(TeamID(id), &obj.TeamVal, active, netCode, flags)
		},
		loadTeamKind: func(team *Team) uint8 {
			return uint8(team.ColorInd)
		},
		loadNetCode16: func(obj *Object) uint16 {
			return uint16(obj.NetCode)
		},
		ballStatus: func(state uint8, netCode uint16) {
			_ = deps.ballStatus(state, netCode)
		},
		carrierState: deps.carrierState,
		loadFlags: func(obj *Object) uint32 {
			return uint32(obj.ObjFlags)
		},
		storeFlags: func(obj *Object, flags uint32) {
			obj.ObjFlags = object.Flags(flags)
		},
		purgeBuffs: deps.purgeBuffs,
	})
}

func ballCollideServerDeps4EBA00(
	s *Server,
	runtime BallCollideRuntime4EBA00,
) ballCollideNativeDeps4EBA00 {
	return ballCollideNativeDeps4EBA00{
		teamByID: func(id uint8) *Team {
			return s.Teams.ByID(TeamID(id))
		},
		teamMemberCount: runtime.TeamMemberCount,
		loadFrame:       s.Frame,
		loadFeedback:    runtime.LoadFeedbackFrame,
		priorityMessage: func(obj *Object, message string, arg int32) {
			s.NetPriMsgToPlayer(obj, strman.ID(message), byte(arg))
		},
		storeFeedback: runtime.StoreFeedback,
		audio: func(id uint32, obj *Object) {
			s.Audio.EventObj(sound.ID(id), obj, 0, 0)
		},
		setOwner:   s.ObjSetOwner,
		changeTeam: runtime.ChangeTeam,
		createTeam: runtime.CreateTeam,
		ballStatus: runtime.BallStatus,
		carrierState: func(ball, target *Object) *Object {
			return s.GameBallCarrierState4EB9B0(ball, target)
		},
		purgeBuffs: func(obj *Object) {
			s.FlagPickupBuffPurge4EA7A0(obj, runtime.BuffPurge)
		},
	}
}

// BallCollide4EBA00 binds the zero-byte BallCollide registration to native
// Object, Team and GameBall update-data layouts. The collision pointer remains
// in the callback ABI but is intentionally unread.
func (s *Server) BallCollide4EBA00(
	ball, target *Object,
	collision *types.Pointf,
	runtime BallCollideRuntime4EBA00,
) {
	ballCollideNative4EBA00(ball, target, collision, ballCollideServerDeps4EBA00(s, runtime))
}
