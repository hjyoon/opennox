package server

import (
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
)

// SummonUnitRuntime5016C0 supplies services that remain owned by the root or
// legacy boundary. Object, update-data, player, and team pointers stay at
// native width across every callback.
type SummonUnitRuntime5016C0 struct {
	CreateObjectAt   func(*Object, *Object, types.Pointf)
	OrderUnit        func(*Object, *Object, uint32)
	ReportAcquire    func(uint8, *Object)
	MarkMinimap      func(uint8, *Object, uint32)
	SendSimpleObject func(uint8, *Object)
	CreateTeam       func(TeamID, *ObjectTeam, int32, uint32, int32)
}

type summonUnitNativeDeps5016C0 struct {
	newObject      func(int32) *Object
	createObjectAt func(*Object, *Object, types.Pointf)
	runtime        SummonUnitRuntime5016C0
}

func summonUnitNative5016C0(
	typeID int32,
	position *types.Pointf,
	owner *Object,
	direction uint8,
	deps summonUnitNativeDeps5016C0,
) *Object {
	return summonUnitAt5016C0(typeID, position, owner, direction, summonUnitHooks5016C0[
		*Object,
		*types.Pointf,
		*MonsterUpdateData,
		*PlayerUpdateData,
		*Player,
		*ObjectTeam,
	]{
		newObject: deps.newObject,
		loadPositionY: func(position *types.Pointf) float32 {
			return position.Y
		},
		loadPositionX: func(position *types.Pointf) float32 {
			return position.X
		},
		createObjectAt: deps.createObjectAt,
		loadDirection: func(direction uint8) uint16 {
			return uint16(direction)
		},
		loadMonsterUpdate: func(obj *Object) *MonsterUpdateData {
			return (*MonsterUpdateData)(obj.UpdateData)
		},
		storeDirection1: func(obj *Object, direction uint16) {
			obj.Direction1 = Dir16(direction)
		},
		storeDirection2: func(obj *Object, direction uint16) {
			obj.Direction2 = Dir16(direction)
		},
		loadMonsterStatus: func(update *MonsterUpdateData) uint32 {
			return uint32(update.StatusFlags)
		},
		storeMonsterStatus: func(update *MonsterUpdateData, status uint32) {
			update.StatusFlags = object.MonsterStatus(status)
		},
		loadClassLow: func(obj *Object) uint8 {
			return uint8(obj.ObjClass)
		},
		loadPlayerUpdate: func(obj *Object) *PlayerUpdateData {
			return (*PlayerUpdateData)(obj.UpdateData)
		},
		loadPlayer: func(update *PlayerUpdateData) *Player {
			return update.Player
		},
		loadSummonOrder: func(player *Player) uint32 {
			return player.SummonOrderAll
		},
		orderUnit: deps.runtime.OrderUnit,
		storeAIAction: func(update *MonsterUpdateData, action uint32) {
			update.AIAction340 = action
		},
		loadSubclass: func(obj *Object) uint32 {
			return uint32(obj.ObjSubClass)
		},
		storeSubclass: func(obj *Object, subclass uint32) {
			obj.ObjSubClass = object.SubClass(subclass)
		},
		loadPlayerIndex: func(player *Player) uint8 {
			return player.PlayerInd
		},
		reportAcquire:    deps.runtime.ReportAcquire,
		markMinimap:      deps.runtime.MarkMinimap,
		sendSimpleObject: deps.runtime.SendSimpleObject,
		loadTeam: func(obj *Object) *ObjectTeam {
			return &obj.TeamVal
		},
		hasTeam: func(team *ObjectTeam) bool {
			return team.Has()
		},
		loadNetCode: func(obj *Object) uint32 {
			return obj.NetCode
		},
		loadTeamID: func(obj *Object) uint8 {
			return uint8(obj.TeamVal.ID)
		},
		createTeam: func(id uint8, team *ObjectTeam, active, netCode, flags uint32) {
			deps.runtime.CreateTeam(TeamID(id), team, int32(active), netCode, int32(flags))
		},
	})
}

// SummonUnitAt5016C0 binds GAME.EXE 005016C0 to native-width server state.
// The position remains a pointer until after allocation succeeds, matching
// the original function's fault and callback ordering.
//
//go:noinline
func (s *Server) SummonUnitAt5016C0(
	typeID int32,
	position *types.Pointf,
	owner *Object,
	direction uint8,
	runtime SummonUnitRuntime5016C0,
) *Object {
	return summonUnitNative5016C0(typeID, position, owner, direction, summonUnitNativeDeps5016C0{
		newObject: func(typeID int32) *Object {
			return s.Objs.NewObject(s.Types.ByInd(int(typeID)))
		},
		createObjectAt: runtime.CreateObjectAt,
		runtime:        runtime,
	})
}

var (
	_ = [1]struct{}{}[4-unsafe.Sizeof(Object{}.ObjClass)]
	_ = [1]struct{}{}[4-unsafe.Sizeof(Object{}.ObjSubClass)]
	_ = [1]struct{}{}[4-unsafe.Sizeof(Object{}.NetCode)]
	_ = [1]struct{}{}[2-unsafe.Sizeof(Object{}.Direction1)]
	_ = [1]struct{}{}[2-unsafe.Sizeof(Object{}.Direction2)]
	_ = [1]struct{}{}[4-unsafe.Sizeof(MonsterUpdateData{}.AIAction340)]
	_ = [1]struct{}{}[4-unsafe.Sizeof(MonsterUpdateData{}.StatusFlags)]
	_ = [1]struct{}{}[4-unsafe.Sizeof(Player{}.SummonOrderAll)]
	_ = [1]struct{}{}[1-unsafe.Sizeof(Player{}.PlayerInd)]
	_ = [1]struct{}{}[8-unsafe.Sizeof(ObjectTeam{})]
	_ = [1]struct{}{}[4-unsafe.Offsetof(ObjectTeam{}.ID)]
)
