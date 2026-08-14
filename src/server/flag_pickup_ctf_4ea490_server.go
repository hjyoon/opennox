package server

import (
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/strman"
	"github.com/opennox/libs/types"
)

// flagPickupCTFFlagIndex4ECBD0 restores the flag-only material lookup used by
// GAME.EXE 004ECBD0/004ECC00. Flag objects use ModifierInitData and the second
// modifier slot is the team material. The original static table deliberately
// orders Green before Blue while mapping them to team IDs 3 and 2.
func flagPickupCTFFlagIndex4ECBD0(obj *Object) uint32 {
	if uint32(obj.ObjClass)&uint32(object.ClassFlag) == 0 {
		return 0
	}
	material := obj.InitDataModifier().Modifiers[1]
	if material == nil {
		return 0
	}
	switch material.Name() {
	case "MaterialTeamRed":
		return 1
	case "MaterialTeamGreen":
		return 3
	case "MaterialTeamBlue":
		return 2
	case "MaterialTeamYellow":
		return 5
	case "MaterialTeamCyan":
		return 4
	case "MaterialTeamViolet":
		return 6
	case "MaterialTeamBlack":
		return 7
	case "MaterialTeamWhite":
		return 8
	case "MaterialTeamOrange":
		return 9
	default:
		return 0
	}
}

// FlagUpdateData4EA490 is the pointer-free prefix used by team flags. Its
// three original ABI32 words retain the same offsets on every Go target.
type FlagUpdateData4EA490 struct {
	Home  types.Pointf
	State uint32
}

var (
	_ = [1]struct{}{}[12-unsafe.Sizeof(FlagUpdateData4EA490{})]
	_ = [1]struct{}{}[8-unsafe.Offsetof(FlagUpdateData4EA490{}.State)]
)

// FlagPickupCTFRuntime4EA490 supplies effects still owned by the legacy/root
// runtime. Native Object, update-data, Player, and Team fields are handled by
// the server adapter without pointer-to-int conversions.
type FlagPickupCTFRuntime4EA490 struct {
	GameData        func(uint32) uint16
	MoveHome        func(*Object, *FlagUpdateData4EA490)
	InformReturn    func(uint32)
	InformFlag      func(uint32, uint32, uint32)
	FlagStatus      func(uint8, uint8, uint8, uint16) int32
	ObserverMode    func() uint32
	ObserverUpdate  func(*Player, *Player)
	DetachInventory func(*Object, *Object)
	CreateAt        func(*Object, *Object, types.Pointf)
	SetGameFlags    func(uint32)
	FlagWinner      func(*Team, uint32)
	TeamEligible    func(*Team) int32
	ForceDrop       func(*Object, *Object)
	FinalizeDelete  func(*Object)
	InventoryPut    func(*Object, *Object, int32)
	ReportObject    func(uint32, *Object)
	BuffPurge       FlagPickupBuffPurgeRuntime4EA7A0
}

type flagPickupCTFNativeDeps4EA490 struct {
	flagIndex       func(*Object) uint32
	gameData        func(uint32) uint16
	moveHome        func(*Object, *FlagUpdateData4EA490)
	informReturn    func(uint32)
	informFlag      func(uint32, uint32, uint32)
	flagStatus      func(uint8, uint8, uint8, uint16) int32
	reportLesson    func(*Object)
	teamByID        func(uint8) *Team
	changeTeamScore func(*Team, int32)
	observerMode    func() uint32
	observerUpdate  func(*Player, *Player)
	detachInventory func(*Object, *Object)
	createAt        func(*Object, *Object, types.Pointf)
	raise           func(*Object, float32)
	markMinimap     func(*Object, uint32)
	firstTeam       func() *Team
	nextTeam        func(*Team) *Team
	setGameFlags    func(uint32)
	flagWinner      func(*Team, uint32)
	teamEligible    func(*Team) int32
	forceDrop       func(*Object, *Object)
	finalizeDelete  func(*Object)
	inventoryPut    func(*Object, *Object, int32)
	reportObject    func(uint32, *Object)
	unmarkMinimap   func(*Object, uint32)
	purgeBuffs      func(*Object)
	priorityMessage func(*Object, string, uint32)
}

func flagPickupCTFNative4EA490(
	source, target *Object,
	collision *types.Pointf,
	deps flagPickupCTFNativeDeps4EA490,
) {
	flagPickupCTF4EA490(source, target, collision,
		flagPickupCTFHooks4EA490[*Object, unsafe.Pointer, *Team, *Player]{
			loadUpdate: func(obj *Object) unsafe.Pointer {
				return obj.UpdateData
			},
			flagIndex: deps.flagIndex,
			loadTeamID: func(obj *Object) uint8 {
				return uint8(obj.TeamVal.ID)
			},
			teamsSame: func(target, source *Object) int32 {
				if target.TeamVal.SameAs(&source.TeamVal) {
					return 1
				}
				return 0
			},
			loadPosX: func(obj *Object) float32 {
				return obj.PosVec.X
			},
			loadPosY: func(obj *Object) float32 {
				return obj.PosVec.Y
			},
			loadHomeX: func(update unsafe.Pointer) float32 {
				return (*FlagUpdateData4EA490)(update).Home.X
			},
			loadHomeY: func(update unsafe.Pointer) float32 {
				return (*FlagUpdateData4EA490)(update).Home.Y
			},
			moveHome: func(obj *Object, update unsafe.Pointer) {
				deps.moveHome(obj, (*FlagUpdateData4EA490)(update))
			},
			loadNetCode: func(obj *Object) uint32 {
				return obj.NetCode
			},
			informReturn: deps.informReturn,
			informFlag:   deps.informFlag,
			storeFlagState: func(update unsafe.Pointer, state uint32) {
				(*FlagUpdateData4EA490)(update).State = state
			},
			flagStatus: deps.flagStatus,
			firstInventory: func(obj *Object) *Object {
				return obj.InvFirstItem
			},
			nextInventory: func(obj *Object) *Object {
				return obj.InvNextItem
			},
			loadClass: func(obj *Object) uint32 {
				return uint32(obj.ObjClass)
			},
			gameData: deps.gameData,
			changeScore: func(obj *Object, delta int32) {
				if uint8(obj.ObjClass)&uint8(object.ClassPlayer) != 0 {
					update := (*PlayerUpdateData)(obj.UpdateData)
					update.Player.Lessons += delta
				}
			},
			reportLesson: deps.reportLesson,
			hasTeam: func(obj *Object) int32 {
				if obj.TeamVal.Has() {
					return 1
				}
				return 0
			},
			teamByID: deps.teamByID,
			loadTeamScore: func(team *Team) int32 {
				return int32(team.Lessons)
			},
			changeTeamScore: deps.changeTeamScore,
			observerMode:    deps.observerMode,
			playerFromUpdate: func(update unsafe.Pointer) *Player {
				return (*PlayerUpdateData)(update).Player
			},
			observerUpdate:  deps.observerUpdate,
			detachInventory: deps.detachInventory,
			createAt: func(obj, owner *Object, x, y float32) {
				deps.createAt(obj, owner, types.Pointf{X: x, Y: y})
			},
			raise:        deps.raise,
			markMinimap:  deps.markMinimap,
			firstTeam:    deps.firstTeam,
			nextTeam:     deps.nextTeam,
			setGameFlags: deps.setGameFlags,
			flagWinner:   deps.flagWinner,
			inventoryHolder: func(obj *Object) *Object {
				return obj.InvHolder
			},
			teamEligible:   deps.teamEligible,
			forceDrop:      deps.forceDrop,
			finalizeDelete: deps.finalizeDelete,
			inventoryPut:   deps.inventoryPut,
			markPlayerPickup: func(update unsafe.Pointer, flags uint32) {
				(*PlayerUpdateData)(update).Player.WeaponEquip |= flags
			},
			reportObject:    deps.reportObject,
			unmarkMinimap:   deps.unmarkMinimap,
			purgeBuffs:      deps.purgeBuffs,
			priorityMessage: deps.priorityMessage,
		},
	)
}

func flagPickupCTFMarkMinimapAll4EA490(s *Server, obj *Object, flags uint32) {
	for player := s.Players.First(); player != nil; player = s.Players.Next(player) {
		s.Players.Nox_xxx_netMarkMinimapObject_417190(player.PlayerIndex(), obj, flags)
	}
}

func flagPickupCTFServerDeps4EA490(
	s *Server,
	runtime FlagPickupCTFRuntime4EA490,
) flagPickupCTFNativeDeps4EA490 {
	return flagPickupCTFNativeDeps4EA490{
		flagIndex:    flagPickupCTFFlagIndex4ECBD0,
		gameData:     runtime.GameData,
		moveHome:     runtime.MoveHome,
		informReturn: runtime.InformReturn,
		informFlag:   runtime.InformFlag,
		flagStatus:   runtime.FlagStatus,
		reportLesson: s.Nox_xxx_netReportLesson_4D8EF0,
		teamByID: func(id uint8) *Team {
			return s.Teams.ByID(TeamID(id))
		},
		changeTeamScore: func(team *Team, score int32) {
			s.TeamChangeLessons(team, int(score))
		},
		observerMode:    runtime.ObserverMode,
		observerUpdate:  runtime.ObserverUpdate,
		detachInventory: runtime.DetachInventory,
		createAt:        runtime.CreateAt,
		raise: func(obj *Object, z float32) {
			obj.Raise(z)
		},
		markMinimap: func(obj *Object, flags uint32) {
			flagPickupCTFMarkMinimapAll4EA490(s, obj, flags)
		},
		firstTeam:      s.Teams.First,
		nextTeam:       s.Teams.Next,
		setGameFlags:   runtime.SetGameFlags,
		flagWinner:     runtime.FlagWinner,
		teamEligible:   runtime.TeamEligible,
		forceDrop:      runtime.ForceDrop,
		finalizeDelete: runtime.FinalizeDelete,
		inventoryPut:   runtime.InventoryPut,
		reportObject:   runtime.ReportObject,
		unmarkMinimap: func(obj *Object, flags uint32) {
			s.Players.Nox_xxx_netUnmarkMinimapSpec_417470(obj, flags)
		},
		purgeBuffs: func(obj *Object) {
			s.FlagPickupBuffPurge4EA7A0(obj, runtime.BuffPurge)
		},
		priorityMessage: func(obj *Object, message string, arg uint32) {
			s.NetPriMsgToPlayer(obj, strman.ID(message), byte(arg))
		},
	}
}

// FlagPickupCTF4EA490 runs the native-width CTF flag handler. It is reached
// through the shared typed legacy router together with the sibling FlagBall
// handler, so no callback can fall back into the old int-address body.
func (s *Server) FlagPickupCTF4EA490(
	source, target *Object,
	collision *types.Pointf,
	runtime FlagPickupCTFRuntime4EA490,
) {
	flagPickupCTFNative4EA490(source, target, collision, flagPickupCTFServerDeps4EA490(s, runtime))
}
