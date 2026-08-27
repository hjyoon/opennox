package server

import (
	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/sound"
)

// PickupTreasureRuntime4F3580 supplies the scalar treasure limit and the
// elimination-death service owned by the root runtime. All object, update,
// player, team, iteration, audio, score, and report access remains native.
type PickupTreasureRuntime4F3580 struct {
	DefaultPickup      PickupDefaultRuntime4F31E0
	TreasureMax        func() uint32
	IncrementElimDeath func(*Object)
}

type pickupTreasureNativeDeps4F3580 struct {
	defaultPickup      func(*Object, *Object, int32, int32) int32
	gameFlag           func(uint32) int32
	audio              func(uint32, *Object, int32, uint32)
	treasureMax        func() uint32
	report             func(*Object)
	findTeam           func(uint8) *Team
	firstPlayer        func() *Object
	nextPlayer         func(*Object) *Object
	setGameFlags       func(uint32)
	changeScore        func(*Object, int32)
	reportLesson       func(*Object)
	incrementElimDeath func(*Object)
}

func pickupTreasureNative4F3580(
	owner, item *Object,
	arg3, arg4 int32,
	deps pickupTreasureNativeDeps4F3580,
) int32 {
	return pickupTreasure4F3580(
		owner,
		item,
		pickupTreasureHooks4F3580[*Object, *PlayerUpdateData, *Player, *Team]{
			loadArg4: func() int32 {
				return arg4
			},
			loadArg3: func() int32 {
				return arg3
			},
			defaultPickup: deps.defaultPickup,
			loadClassLow: func(obj *Object) uint8 {
				return uint8(obj.ObjClass)
			},
			loadUpdate: func(obj *Object) *PlayerUpdateData {
				return (*PlayerUpdateData)(obj.UpdateData)
			},
			gameFlag: deps.gameFlag,
			audio:    deps.audio,
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
			hasTeam: func(obj *Object) int32 {
				if obj.TeamVal.Has() {
					return 1
				}
				return 0
			},
			loadObjectTeam: func(obj *Object) uint8 {
				return uint8(obj.TeamVal.ID)
			},
			findTeam: deps.findTeam,
			loadTeamID: func(team *Team) uint8 {
				return uint8(team.IDVal)
			},
			teamContains: func(obj *Object, id uint8) int32 {
				if obj.TeamVal.ID != 0 && id != 0 && obj.TeamVal.ID == TeamID(id) {
					return 1
				}
				return 0
			},
			firstPlayer:     deps.firstPlayer,
			nextPlayer:      deps.nextPlayer,
			setGameFlags:    deps.setGameFlags,
			changeScore:     deps.changeScore,
			reportLesson:    deps.reportLesson,
			incrementDeaths: deps.incrementElimDeath,
		},
	)
}

func pickupTreasureServerDeps4F3580(
	s *Server,
	runtime PickupTreasureRuntime4F3580,
) pickupTreasureNativeDeps4F3580 {
	return pickupTreasureNativeDeps4F3580{
		defaultPickup: func(owner, item *Object, arg3, arg4 int32) int32 {
			return s.PickupDefault4F31E0(owner, item, arg3, arg4, runtime.DefaultPickup)
		},
		gameFlag: func(flag uint32) int32 {
			if noxflags.HasGame(noxflags.GameFlag(flag)) {
				return 1
			}
			return 0
		},
		audio: func(id uint32, owner *Object, kind int32, code uint32) {
			s.Audio.EventObj(sound.ID(id), owner, int(kind), code)
		},
		treasureMax: runtime.TreasureMax,
		report:      s.ScavengerHuntReport4D8CD0,
		findTeam: func(id uint8) *Team {
			return s.Teams.ByID(TeamID(id))
		},
		firstPlayer: s.Players.FirstUnit,
		nextPlayer:  s.questNextPlayerUnit4DA7F0,
		setGameFlags: func(flags uint32) {
			noxflags.SetGame(noxflags.GameFlag(flags))
		},
		changeScore: func(owner *Object, value int32) {
			owner.changeScore(int(value))
		},
		reportLesson:       s.Nox_xxx_netReportLesson_4D8EF0,
		incrementElimDeath: runtime.IncrementElimDeath,
	}
}

// PickupTreasure4F3580 binds GAME.EXE's registered four-argument
// TreasurePickup callback to native-width Object, PlayerUpdateData, Player,
// ObjectTeam, and Team layouts.
func (s *Server) PickupTreasure4F3580(
	owner, item *Object,
	arg3, arg4 int32,
	runtime PickupTreasureRuntime4F3580,
) int32 {
	return pickupTreasureNative4F3580(
		owner,
		item,
		arg3,
		arg4,
		pickupTreasureServerDeps4F3580(s, runtime),
	)
}
