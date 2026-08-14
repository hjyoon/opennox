package server

import (
	"unsafe"

	"github.com/opennox/libs/noxnet/netmsg"
	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/sound"
)

// GameBallUpdateData4EA800 is the native-width prefix used by GameBall. On
// ABI32 it exactly occupies the original four words; ABI64 expands the Object
// pointer and aligns the trailing tick counter without converting it to an
// integer address.
type GameBallUpdateData4EA800 struct {
	Carrier *Object
	Field4  uint32
	Ticks   uint64
}

// FlagPickupBallRuntime4EA800 supplies effects that are still owned by the
// legacy/root runtime. Object, update-data, Player, Team, list and vector fields
// stay native-width inside the server adapter.
type FlagPickupBallRuntime4EA800 struct {
	GameData         func(uint32) uint16
	ObserverMode     func() uint32
	ObserverUpdate   func(*Player, *Player)
	InformScore      func(uint32, uint32, uint32)
	SetGameFlags     func(uint32)
	FlagBallWinner   func(*Team)
	DropBall         func(*Object, *Object)
	ChangeObjectTeam func(*ObjectTeam, uint32)
	SetHPMax         func(*Object)
	Ticks            func() uint64
	MoveTo           func(*Object, types.Pointf)
	BallStatus       func(uint8, uint16) int32
}

type flagPickupBallNativeDeps4EA800 struct {
	loadBallCache    func() uint32
	lookupType       func(string) uint32
	storeBallCache   func(uint32)
	unitIsGameBall   func(*Object) int32
	gameData         func(uint32) uint16
	teamByID         func(uint8) *Team
	nextTeam         func(*Team) *Team
	firstTeam        func() *Team
	reportLesson     func(*Object)
	changeTeamScore  func(*Team, int32)
	observerMode     func() uint32
	observerUpdate   func(*Player, *Player)
	audio            func(uint32, *Object)
	informScore      func(uint32, uint32, uint32)
	pointFX          func(uint32, types.Pointf)
	setGameFlags     func(uint32)
	flagBallWinner   func(*Team)
	loadStartCache   func() uint32
	storeStartCache  func(uint32)
	firstObject      func() *Object
	nextObject       func(*Object) *Object
	randomInt        func(int32, int32) int32
	clearOwner       func(*Object)
	dropBall         func(*Object, *Object)
	changeObjectTeam func(*ObjectTeam, uint32)
	setHPMax         func(*Object)
	ticks            func() uint64
	moveTo           func(*Object, types.Pointf)
	ballStatus       func(uint8, uint16) int32
}

func flagPickupBallNative4EA800(
	source, target *Object,
	collision *types.Pointf,
	deps flagPickupBallNativeDeps4EA800,
) {
	flagPickupBall4EA800(source, target, collision,
		flagPickupBallHooks4EA800[*Object, unsafe.Pointer, *Team, *Player]{
			loadBallCache:  deps.loadBallCache,
			lookupType:     deps.lookupType,
			storeBallCache: deps.storeBallCache,
			loadClassLow: func(obj *Object) uint8 {
				return uint8(obj.ObjClass)
			},
			unitIsGameBall: deps.unitIsGameBall,
			firstOwned: func(obj *Object) *Object {
				return obj.Field129
			},
			nextOwned: func(obj *Object) *Object {
				return obj.Field128
			},
			loadTypeInd: func(obj *Object) uint16 {
				return obj.TypeInd
			},
			loadUpdate: func(obj *Object) unsafe.Pointer {
				return obj.UpdateData
			},
			loadCarrier: func(update unsafe.Pointer) *Object {
				return (*GameBallUpdateData4EA800)(update).Carrier
			},
			loadFlagsLow: func(obj *Object) uint8 {
				return uint8(obj.ObjFlags)
			},
			storeCarrier: func(update unsafe.Pointer, carrier *Object) {
				(*GameBallUpdateData4EA800)(update).Carrier = carrier
			},
			loadTeamID: func(obj *Object) uint8 {
				return uint8(obj.TeamVal.ID)
			},
			teamByID:  deps.teamByID,
			nextTeam:  deps.nextTeam,
			firstTeam: deps.firstTeam,
			loadTeamIDValue: func(team *Team) uint8 {
				return uint8(team.IDVal)
			},
			gameData: deps.gameData,
			changeScore: func(obj *Object, delta int32) {
				if uint8(obj.ObjClass)&uint8(object.ClassPlayer) != 0 {
					update := (*PlayerUpdateData)(obj.UpdateData)
					update.Player.Lessons += delta
				}
			},
			reportLesson:    deps.reportLesson,
			loadTeamScore:   func(team *Team) int32 { return int32(team.Lessons) },
			changeTeamScore: deps.changeTeamScore,
			observerMode:    deps.observerMode,
			playerFromUpdate: func(update unsafe.Pointer) *Player {
				return (*PlayerUpdateData)(update).Player
			},
			observerUpdate: deps.observerUpdate,
			audio:          deps.audio,
			loadNetCode: func(obj *Object) uint32 {
				return obj.NetCode
			},
			informScore: deps.informScore,
			pointFX: func(code uint32, obj *Object) {
				deps.pointFX(code, obj.PosVec)
			},
			setGameFlags:    deps.setGameFlags,
			flagBallWinner:  deps.flagBallWinner,
			loadStartCache:  deps.loadStartCache,
			storeStartCache: deps.storeStartCache,
			firstObject:     deps.firstObject,
			nextObject:      deps.nextObject,
			randomInt:       deps.randomInt,
			clearOwner:      deps.clearOwner,
			dropBall:        deps.dropBall,
			changeObjectTeam: func(obj *Object, netCode uint32) {
				deps.changeObjectTeam(&obj.TeamVal, netCode)
			},
			setHPMax: deps.setHPMax,
			ticks:    deps.ticks,
			storeTicks: func(update unsafe.Pointer, ticks uint64) {
				(*GameBallUpdateData4EA800)(update).Ticks = ticks
			},
			moveToMarker: func(ball, marker *Object) {
				deps.moveTo(ball, marker.PosVec)
			},
			ballStatus: func(state, netCode uint32) {
				deps.ballStatus(uint8(state), uint16(netCode))
			},
			clearMotion: func(obj *Object) {
				obj.VelVec.X = 0
				obj.VelVec.Y = 0
				obj.ForceVec.X = 0
				obj.Pos24.Y = 0
			},
		},
	)
}

func flagPickupBallUnitIsGameBall4EA800(s *Server, owner *Object) int32 {
	typeInd := uint32(s.Types.GameBallID())
	for obj := owner.Field129; obj != nil; obj = obj.Field128 {
		if uint32(obj.TypeInd) == typeInd {
			return 1
		}
	}
	return 0
}

func flagPickupBallServerDeps4EA800(
	s *Server,
	runtime FlagPickupBallRuntime4EA800,
) flagPickupBallNativeDeps4EA800 {
	return flagPickupBallNativeDeps4EA800{
		loadBallCache: func() uint32 {
			return s.Types.fast.flagPickupGameBall
		},
		lookupType: func(name string) uint32 {
			return uint32(s.Types.IndByID(name))
		},
		storeBallCache: func(ind uint32) {
			s.Types.fast.flagPickupGameBall = ind
		},
		unitIsGameBall: func(owner *Object) int32 {
			return flagPickupBallUnitIsGameBall4EA800(s, owner)
		},
		gameData: runtime.GameData,
		teamByID: func(id uint8) *Team {
			return s.Teams.ByID(TeamID(id))
		},
		nextTeam:     s.Teams.Next,
		firstTeam:    s.Teams.First,
		reportLesson: s.Nox_xxx_netReportLesson_4D8EF0,
		changeTeamScore: func(team *Team, score int32) {
			s.TeamChangeLessons(team, int(score))
		},
		observerMode:   runtime.ObserverMode,
		observerUpdate: runtime.ObserverUpdate,
		audio: func(id uint32, obj *Object) {
			s.Audio.EventObj(sound.ID(id), obj, 0, 0)
		},
		informScore: runtime.InformScore,
		pointFX: func(code uint32, pos types.Pointf) {
			s.Nox_xxx_netSendPointFx_522FF0(netmsg.Op(code), pos)
		},
		setGameFlags:   runtime.SetGameFlags,
		flagBallWinner: runtime.FlagBallWinner,
		loadStartCache: func() uint32 {
			return s.Types.fast.flagPickupBallStart
		},
		storeStartCache: func(ind uint32) {
			s.Types.fast.flagPickupBallStart = ind
		},
		firstObject: s.Objs.First,
		nextObject: func(obj *Object) *Object {
			return obj.Next()
		},
		randomInt: func(minimum, maximum int32) int32 {
			return int32(s.Rand.Logic.IntClamp(int(minimum), int(maximum)))
		},
		clearOwner:       s.ObjClearOwner,
		dropBall:         runtime.DropBall,
		changeObjectTeam: runtime.ChangeObjectTeam,
		setHPMax:         runtime.SetHPMax,
		ticks:            runtime.Ticks,
		moveTo:           runtime.MoveTo,
		ballStatus:       runtime.BallStatus,
	}
}

// FlagPickupBall4EA800 runs the native-width FlagBall scoring and respawn
// handler. Its legacy export is installed together with the completed sibling
// CTF handler and collision router so no architecture can fall back into the
// old integer-address implementation.
func (s *Server) FlagPickupBall4EA800(
	source, target *Object,
	collision *types.Pointf,
	runtime FlagPickupBallRuntime4EA800,
) {
	flagPickupBallNative4EA800(source, target, collision, flagPickupBallServerDeps4EA800(s, runtime))
}
