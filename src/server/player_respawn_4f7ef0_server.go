package server

import (
	"encoding/binary"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/strman"
	"github.com/opennox/libs/types"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/sound"
)

// PlayerRespawnRuntime4F7EF0 supplies services that remain owned by the outer
// server or legacy runtime. Every Object, Player, Settings, Team, and point
// pointer stays native-width throughout this boundary.
type PlayerRespawnRuntime4F7EF0 struct {
	LoadSettings         func() *Settings
	MakeDefaultItems     func(*Object, int32, int32)
	NetworkMode          func() uint32
	RespawnCorpse        func(*types.Pointf, int16)
	MapTileAllowTeleport func(*types.Pointf) int32
	Move                 func(*Object, types.Pointf)
	CrownPickup          func(*Object, *Object, int32, int32)
	ApplyBuff            func(*Object, EnchantID, uint16, uint8)
}

type playerRespawnNativeDeps4F7EF0 struct {
	loadSettings     func() *Settings
	gameFlag         func(uint32) int32
	makeDefaultItems func(*Object, int32, int32)
	priorityMessage  func(*Object, string, uint8)
	audio            func(uint32, *Object, int32, int32)
	loadNetworkMode  func() uint32
	respawnCorpse    func(*types.Pointf, int16)
	soulGatePoint    func(*Object, *types.Pointf)
	mapPlayerStart   func(*types.Pointf, *Object)
	move             func(*Object, types.Pointf)
	gameplayFlag     func(uint32) int32
	teamByID         func(uint8) *Team
	loadTeamCrown    func(*Team) *Object
	crownPickup      func(*Object, *Object, int32, int32)
	loadTickRate     func() uint32
	applyBuff        func(*Object, EnchantID, uint16, uint8)
}

func playerRespawnNative4F7EF0(
	unit *Object,
	deps playerRespawnNativeDeps4F7EF0,
) playerRespawnResult4F7EF0[*Settings] {
	return playerRespawn4F7EF0(playerRespawnHooks4F7EF0[
		*Object,
		*PlayerUpdateData,
		*Player,
		*Settings,
		*Team,
	]{
		loadSettings: deps.loadSettings,
		loadUnitArg: func() *Object {
			return unit
		},
		loadUpdateData: func(unit *Object) *PlayerUpdateData {
			return (*PlayerUpdateData)(unit.UpdateData)
		},
		loadPlayer: func(update *PlayerUpdateData) *Player {
			return update.Player
		},
		gameFlag: deps.gameFlag,
		loadQuestBlock: func(update *PlayerUpdateData) uint32 {
			return update.Field137
		},
		storePlayerDone: func(player *Player, value uint32) {
			player.Field4700 = value
		},
		makeDefaultItems: deps.makeDefaultItems,
		loadPlayerIndex: func(player *Player) uint8 {
			return player.PlayerInd
		},
		storeRespawnMarker: func(update *PlayerUpdateData, index, value uint8) {
			update.RespawnMarkers[index] = value
		},
		priorityMessage: deps.priorityMessage,
		audio:           deps.audio,
		loadNetworkMode: deps.loadNetworkMode,
		loadSkeletonSetting: func(settings *Settings) uint32 {
			return binary.LittleEndian.Uint32(settings.PlayerSkeletons58[:])
		},
		respawnCorpse: func(unit *Object) {
			deps.respawnCorpse(&unit.PosVec, int16(unit.Direction1))
		},
		loadPositionX: func(unit *Object) float32 {
			return unit.PosVec.X
		},
		loadPositionY: func(unit *Object) float32 {
			return unit.PosVec.Y
		},
		loadSoulGate: func(update *PlayerUpdateData) *Object {
			return update.SoulGate
		},
		soulGatePoint:  deps.soulGatePoint,
		mapPlayerStart: deps.mapPlayerStart,
		move: func(unit *Object, destination *types.Pointf) {
			deps.move(unit, *destination)
		},
		gameplayFlag: deps.gameplayFlag,
		loadPlayerUnit: func(player *Player) *Object {
			return player.PlayerUnit
		},
		loadTeamID: func(unit *Object) uint8 {
			return uint8(unit.TeamVal.ID)
		},
		teamByID: deps.teamByID,
		loadTeamCrown: func(team *Team) *Object {
			return deps.loadTeamCrown(team)
		},
		loadInventoryHolder: func(crown *Object) *Object {
			return crown.InvHolder
		},
		crownPickup:  deps.crownPickup,
		loadTickRate: deps.loadTickRate,
		applyBuff: func(unit *Object, enchant uint32, duration uint16, power uint32) {
			deps.applyBuff(unit, EnchantID(enchant), duration, uint8(power))
		},
	})
}

func playerRespawnResultLow4F7EF0(result playerRespawnResult4F7EF0[*Settings]) int16 {
	if result.kind == playerRespawnResultSettings4F7EF0 {
		return int16(uint16(uintptr(unsafe.Pointer(result.settings))))
	}
	return int16(uint16(result.value))
}

func playerRespawnServerDeps4F7EF0(
	s *Server,
	runtime PlayerRespawnRuntime4F7EF0,
) playerRespawnNativeDeps4F7EF0 {
	return playerRespawnNativeDeps4F7EF0{
		loadSettings: runtime.LoadSettings,
		gameFlag: func(flag uint32) int32 {
			if noxflags.HasGame(noxflags.GameFlag(flag)) {
				return 1
			}
			return 0
		},
		makeDefaultItems: runtime.MakeDefaultItems,
		priorityMessage: func(unit *Object, message string, value uint8) {
			s.NetPriMsgToPlayer(unit, strman.ID(message), value)
		},
		audio: func(id uint32, unit *Object, first, second int32) {
			s.Audio.EventObj(sound.ID(id), unit, int(first), uint32(second))
		},
		loadNetworkMode: runtime.NetworkMode,
		respawnCorpse:   runtime.RespawnCorpse,
		soulGatePoint: func(gate *Object, output *types.Pointf) {
			s.SoulGateRespawnPointInto4F80C0(gate, output, runtime.MapTileAllowTeleport)
		},
		mapPlayerStart: s.MapFindPlayerStartInto4F7AB0,
		move:           runtime.Move,
		gameplayFlag: func(flag uint32) int32 {
			if noxflags.HasGamePlay(noxflags.GameplayFlag(flag)) {
				return 1
			}
			return 0
		},
		teamByID: func(id uint8) *Team {
			return s.Teams.ByID(TeamID(id))
		},
		loadTeamCrown: s.Teams.TeamFlag,
		crownPickup:   runtime.CrownPickup,
		loadTickRate:  s.TickRate,
		applyBuff:     runtime.ApplyBuff,
	}
}

// PlayerRespawn4F7EF0 binds GAME.EXE 004F7EF0 to native-width server data.
// The int16 result is retained for the original C ABI; active Go callers
// intentionally discard it just as the original callers did.
func (s *Server) PlayerRespawn4F7EF0(
	unit *Object,
	runtime PlayerRespawnRuntime4F7EF0,
) int16 {
	return playerRespawnResultLow4F7EF0(
		playerRespawnNative4F7EF0(unit, playerRespawnServerDeps4F7EF0(s, runtime)),
	)
}

type soulGateRespawnPointNativeDeps4F80C0 struct {
	randomPoint   func(float32, *types.Pointf, *types.Pointf)
	allowTeleport func(*types.Pointf) int32
}

func soulGateRespawnPointNative4F80C0(
	gate *Object,
	output *types.Pointf,
	deps soulGateRespawnPointNativeDeps4F80C0,
) int32 {
	return soulGateRespawnPoint4F80C0(gate, output, soulGateRespawnPointHooks4F80C0[*Object]{
		loadPosition: func(gate *Object) types.Pointf {
			return gate.PosVec
		},
		randomPoint: func(radius float32, gate *Object, output *types.Pointf) {
			deps.randomPoint(radius, &gate.PosVec, output)
		},
		allowTeleport: deps.allowTeleport,
	})
}

// SoulGateRespawnPointInto4F80C0 restores the native-pointer SoulGate helper.
func (s *Server) SoulGateRespawnPointInto4F80C0(
	gate *Object,
	output *types.Pointf,
	allowTeleport func(*types.Pointf) int32,
) int32 {
	return soulGateRespawnPointNative4F80C0(gate, output, soulGateRespawnPointNativeDeps4F80C0{
		randomPoint: func(radius float32, center, output *types.Pointf) {
			s.RandomReachablePointAroundInto4ED970(radius, center, output)
		},
		allowTeleport: allowTeleport,
	})
}

// RespawnPlayerImplRuntime53FBC0 supplies the legacy corpse lookup tables,
// direction quantizer, object placement, and one-time initializer.
type RespawnPlayerImplRuntime53FBC0 struct {
	Initialized func() uint32
	Initialize  func()
	Direction   func(int16) int32
	TypeIndex   func(int32, int) uint32
	Offset      func(int32, int) types.Pointf
	NetworkMode func() uint32
	CreateAt    func(*Object, types.Pointf)
}

type respawnPlayerImplNativeDeps53FBC0 struct {
	initialized func() uint32
	initialize  func()
	direction   func(int16) int32
	typeIndex   func(int32, int) uint32
	offset      func(int32, int) types.Pointf
	newObject   func(uint32) *Object
	networkMode func() uint32
	gameFlag    func(uint32) int32
	createAt    func(*Object, types.Pointf)
	randomInt   func(int32, int32) int32
	tickRate    func() uint32
	setDecay    func(*Object, uint32)
}

func respawnPlayerImplNative53FBC0(
	center *types.Pointf,
	direction int16,
	deps respawnPlayerImplNativeDeps53FBC0,
) {
	respawnPlayerImpl53FBC0(direction, respawnPlayerImplHooks53FBC0[*Object]{
		loadInitialized: deps.initialized,
		initialize:      deps.initialize,
		directionIndex:  deps.direction,
		loadTypeIndex:   deps.typeIndex,
		newObject:       deps.newObject,
		loadNetworkMode: deps.networkMode,
		gameFlag:        deps.gameFlag,
		loadObjectFlags: func(object *Object) uint32 {
			return uint32(object.ObjFlags)
		},
		storeObjectFlags: func(objectValue *Object, value uint32) {
			objectValue.ObjFlags = object.Flags(value)
		},
		loadOffsetY: func(direction int32, part int) float32 {
			return deps.offset(direction, part).Y
		},
		loadCenterY: func() float32 {
			return center.Y
		},
		loadOffsetX: func(direction int32, part int) float32 {
			return deps.offset(direction, part).X
		},
		loadCenterX: func() float32 {
			return center.X
		},
		createAt:     deps.createAt,
		randomInt:    deps.randomInt,
		loadTickRate: deps.tickRate,
		setDecay:     deps.setDecay,
	})
}

func respawnPlayerImplServerDeps53FBC0(
	s *Server,
	runtime RespawnPlayerImplRuntime53FBC0,
) respawnPlayerImplNativeDeps53FBC0 {
	return respawnPlayerImplNativeDeps53FBC0{
		initialized: runtime.Initialized,
		initialize:  runtime.Initialize,
		direction:   runtime.Direction,
		typeIndex:   runtime.TypeIndex,
		offset:      runtime.Offset,
		newObject: func(typeIndex uint32) *Object {
			return s.NewObjectByTypeInd(int(typeIndex))
		},
		networkMode: runtime.NetworkMode,
		gameFlag: func(flag uint32) int32 {
			if noxflags.HasGame(noxflags.GameFlag(flag)) {
				return 1
			}
			return 0
		},
		createAt: runtime.CreateAt,
		randomInt: func(minimum, maximum int32) int32 {
			return int32(s.Rand.Logic.IntClamp(int(minimum), int(maximum)))
		},
		tickRate: s.TickRate,
		setDecay: func(object *Object, delay uint32) {
			_ = s.DecaySetTime511660(object, delay)
		},
	}
}

// RespawnPlayerImpl53FBC0 creates corpse pieces without narrowing any newly
// allocated Object pointer through the original int temporary.
func (s *Server) RespawnPlayerImpl53FBC0(
	center *types.Pointf,
	direction int16,
	runtime RespawnPlayerImplRuntime53FBC0,
) {
	respawnPlayerImplNative53FBC0(
		center,
		direction,
		respawnPlayerImplServerDeps53FBC0(s, runtime),
	)
}
