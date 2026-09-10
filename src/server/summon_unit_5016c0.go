package server

import "github.com/opennox/libs/types"

const (
	summonUnitPlayerClass5016C0    = uint8(0x04)
	summonUnitStatus5016C0         = uint32(0x80)
	summonUnitMonitor5016C0        = uint32(0x80)
	summonUnitMigrate5016C0        = uint32(0x100)
	summonUnitNonPlayerOrder5016C0 = uint32(4)
	summonUnitInvalidAction5016C0  = uint32(0x26)
	summonUnitMinimapFlag5016C0    = uint32(1)
)

// summonUnitHooks5016C0 exposes every observable load, call, and store made
// by GAME.EXE 005016C0. Pointer-like values remain generic native-width
// handles; only fields that were fixed-width in the PE32 executable use
// fixed-width integers here.
type summonUnitHooks5016C0[Object comparable, Position, MonsterUpdate, PlayerUpdate, Player, Team any] struct {
	newObject          func(int32) Object
	loadPositionY      func(Position) float32
	loadPositionX      func(Position) float32
	createObjectAt     func(Object, Object, types.Pointf)
	loadDirection      func(uint8) uint16
	loadMonsterUpdate  func(Object) MonsterUpdate
	storeDirection1    func(Object, uint16)
	storeDirection2    func(Object, uint16)
	loadMonsterStatus  func(MonsterUpdate) uint32
	storeMonsterStatus func(MonsterUpdate, uint32)
	loadClassLow       func(Object) uint8
	loadPlayerUpdate   func(Object) PlayerUpdate
	loadPlayer         func(PlayerUpdate) Player
	loadSummonOrder    func(Player) uint32
	orderUnit          func(Object, Object, uint32)
	storeAIAction      func(MonsterUpdate, uint32)
	loadSubclass       func(Object) uint32
	storeSubclass      func(Object, uint32)
	loadPlayerIndex    func(Player) uint8
	reportAcquire      func(uint8, Object)
	markMinimap        func(uint8, Object, uint32)
	sendSimpleObject   func(uint8, Object)
	loadTeam           func(Object) Team
	hasTeam            func(Team) bool
	loadNetCode        func(Object) uint32
	loadTeamID         func(Object) uint8
	createTeam         func(uint8, Team, uint32, uint32, uint32)
}

// summonUnitAt5016C0 preserves GAME.EXE 005016C0's allocation and callback
// order. In particular, allocation failure does not touch the position
// pointer, the spawned MonsterUpdate handle survives the order callback, the
// owner's PlayerUpdate handle is cached while its Player pointer is reloaded
// for every report, and the ObjectTeam handle survives the team predicate.
func summonUnitAt5016C0[Object comparable, Position, MonsterUpdate, PlayerUpdate, Player, Team any](
	typeID int32,
	position Position,
	owner Object,
	direction uint8,
	h summonUnitHooks5016C0[Object, Position, MonsterUpdate, PlayerUpdate, Player, Team],
) Object {
	created := h.newObject(typeID)
	var nilObject Object
	if created == nilObject {
		return nilObject
	}

	y := h.loadPositionY(position)
	x := h.loadPositionX(position)
	h.createObjectAt(created, owner, types.Pointf{X: x, Y: y})
	direction16 := h.loadDirection(direction)
	monsterUpdate := h.loadMonsterUpdate(created)
	h.storeDirection2(created, direction16)
	h.storeDirection1(created, direction16)
	status := h.loadMonsterStatus(monsterUpdate)
	h.storeMonsterStatus(monsterUpdate, status|summonUnitStatus5016C0)

	if owner == nilObject {
		return created
	}
	if h.loadClassLow(owner)&summonUnitPlayerClass5016C0 == 0 {
		h.orderUnit(owner, created, summonUnitNonPlayerOrder5016C0)
		h.storeAIAction(monsterUpdate, summonUnitInvalidAction5016C0)
		return created
	}

	playerUpdate := h.loadPlayerUpdate(owner)
	player := h.loadPlayer(playerUpdate)
	h.orderUnit(owner, created, h.loadSummonOrder(player))
	h.storeAIAction(monsterUpdate, summonUnitInvalidAction5016C0)

	subclass := h.loadSubclass(created)
	h.storeSubclass(created, subclass|summonUnitMonitor5016C0)

	player = h.loadPlayer(playerUpdate)
	h.reportAcquire(h.loadPlayerIndex(player), created)
	player = h.loadPlayer(playerUpdate)
	h.markMinimap(h.loadPlayerIndex(player), created, summonUnitMinimapFlag5016C0)
	player = h.loadPlayer(playerUpdate)
	h.sendSimpleObject(h.loadPlayerIndex(player), created)

	team := h.loadTeam(owner)
	if h.hasTeam(team) {
		netCode := h.loadNetCode(created)
		teamID := h.loadTeamID(owner)
		h.createTeam(teamID, team, 1, netCode, 0)
	}
	subclass = h.loadSubclass(created)
	h.storeSubclass(created, subclass|summonUnitMigrate5016C0)
	return created
}
