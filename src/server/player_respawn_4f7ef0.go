package server

import "github.com/opennox/libs/types"

const (
	playerRespawnQuestFlag4F7EF0      = uint32(0x1000)
	playerRespawnCrownFlag4F7EF0      = uint32(0x0010)
	playerRespawnCrownPlayFlag4F7EF0  = uint32(0x0004)
	playerRespawnProtectionFlag4F7EF0 = uint32(0x2000)
	playerRespawnMarker4F7EF0         = uint8(0xfa)
	playerRespawnEnchant4F7EF0        = uint32(23)
	playerRespawnEnchantPower4F7EF0   = uint32(5)
	playerRespawnQuestSound4F7EF0     = uint32(1006)
	playerRespawnNormalSound4F7EF0    = uint32(148)
	playerRespawnMessage4F7EF0        = "GeneralPrint:Respawn"
)

type playerRespawnResultKind4F7EF0 uint8

const (
	playerRespawnResultScalar4F7EF0 playerRespawnResultKind4F7EF0 = iota
	playerRespawnResultSettings4F7EF0
)

// playerRespawnResult4F7EF0 keeps the original EAX source native-width until
// the public short ABI is produced. GAME.EXE leaves the settings pointer in
// EAX for a nil unit, the Quest blocker for an early return, and the final
// 0x2000 flag-test result on the normal path.
type playerRespawnResult4F7EF0[S comparable] struct {
	kind     playerRespawnResultKind4F7EF0
	settings S
	value    uint32
}

type playerRespawnHooks4F7EF0[O, U, P, S, T comparable] struct {
	loadSettings        func() S
	loadUnitArg         func() O
	loadUpdateData      func(O) U
	loadPlayer          func(U) P
	gameFlag            func(uint32) int32
	loadQuestBlock      func(U) uint32
	storePlayerDone     func(P, uint32)
	makeDefaultItems    func(O, int32, int32)
	loadPlayerIndex     func(P) uint8
	storeRespawnMarker  func(U, uint8, uint8)
	priorityMessage     func(O, string, uint8)
	audio               func(uint32, O, int32, int32)
	loadNetworkMode     func() uint32
	loadSkeletonSetting func(S) uint32
	respawnCorpse       func(O)
	loadPositionX       func(O) float32
	loadPositionY       func(O) float32
	loadSoulGate        func(U) O
	soulGatePoint       func(O, *types.Pointf)
	mapPlayerStart      func(*types.Pointf, O)
	move                func(O, *types.Pointf)
	gameplayFlag        func(uint32) int32
	loadPlayerUnit      func(P) O
	loadTeamID          func(O) uint8
	teamByID            func(uint8) T
	loadTeamCrown       func(T) O
	loadInventoryHolder func(O) O
	crownPickup         func(O, O, int32, int32)
	loadTickRate        func() uint32
	applyBuff           func(O, uint32, uint16, uint32)
}

// playerRespawn4F7EF0 preserves GAME.EXE 004F7EF0's observable load, branch,
// callback, and caching order. Settings, UpdateData, and the first Player are
// cached. Player is reloaded only for the Quest marker; the crown path keeps
// using the entry-time Player and reloads PlayerUnit for the pickup call.
//
// Original fault behavior is intentional: apart from the unit, entry Player,
// SoulGate, crown, and crown holder checks visible in the machine code, this
// routine adds no nil guards.
func playerRespawn4F7EF0[O, U, P, S, T comparable](
	h playerRespawnHooks4F7EF0[O, U, P, S, T],
) playerRespawnResult4F7EF0[S] {
	settings := h.loadSettings()
	unit := h.loadUnitArg()
	var zeroObject O
	if unit == zeroObject {
		return playerRespawnResult4F7EF0[S]{
			kind:     playerRespawnResultSettings4F7EF0,
			settings: settings,
		}
	}

	update := h.loadUpdateData(unit)
	cachedPlayer := h.loadPlayer(update)
	if h.gameFlag(playerRespawnQuestFlag4F7EF0) != 0 {
		blocker := h.loadQuestBlock(update)
		if blocker != 0 {
			return playerRespawnResult4F7EF0[S]{value: blocker}
		}
	}

	var zeroPlayer P
	if cachedPlayer != zeroPlayer {
		h.storePlayerDone(cachedPlayer, 0)
	}

	if h.gameFlag(playerRespawnQuestFlag4F7EF0) != 0 {
		h.makeDefaultItems(unit, 1, 1)
		player := h.loadPlayer(update)
		h.storeRespawnMarker(update, h.loadPlayerIndex(player), playerRespawnMarker4F7EF0)
		h.priorityMessage(unit, playerRespawnMessage4F7EF0, 0)
	} else {
		h.makeDefaultItems(unit, 1, 0)
	}

	if h.gameFlag(playerRespawnQuestFlag4F7EF0) != 0 {
		h.audio(playerRespawnQuestSound4F7EF0, unit, 0, 0)
	} else {
		h.audio(playerRespawnNormalSound4F7EF0, unit, 0, 0)
	}

	networkMode := h.loadNetworkMode()
	if networkMode == 0 || h.loadSkeletonSetting(settings) != 0 {
		h.respawnCorpse(unit)
	}

	var destination types.Pointf
	destination.X = h.loadPositionX(unit)
	destination.Y = h.loadPositionY(unit)
	if h.gameFlag(playerRespawnQuestFlag4F7EF0) != 0 {
		gate := h.loadSoulGate(update)
		if gate != zeroObject {
			h.soulGatePoint(gate, &destination)
		} else {
			h.mapPlayerStart(&destination, unit)
		}
	} else {
		h.mapPlayerStart(&destination, unit)
	}
	h.move(unit, &destination)

	if h.gameFlag(playerRespawnCrownFlag4F7EF0) != 0 &&
		h.gameplayFlag(playerRespawnCrownPlayFlag4F7EF0) != 0 {
		playerUnit := h.loadPlayerUnit(cachedPlayer)
		teamID := h.loadTeamID(playerUnit)
		team := h.teamByID(teamID)
		crown := h.loadTeamCrown(team)
		if crown != zeroObject && h.loadInventoryHolder(crown) == zeroObject {
			h.crownPickup(h.loadPlayerUnit(cachedPlayer), crown, 1, 1)
		}
	}

	finalFlag := h.gameFlag(playerRespawnProtectionFlag4F7EF0)
	if finalFlag != 0 {
		duration := uint16(uint32(uint16(h.loadTickRate())) * 5)
		h.applyBuff(
			unit,
			playerRespawnEnchant4F7EF0,
			duration,
			playerRespawnEnchantPower4F7EF0,
		)
	}
	return playerRespawnResult4F7EF0[S]{value: uint32(finalFlag)}
}

type soulGateRespawnPointHooks4F80C0[O comparable] struct {
	loadPosition  func(O) types.Pointf
	randomPoint   func(float32, O, *types.Pointf)
	allowTeleport func(*types.Pointf) int32
}

// soulGateRespawnPoint4F80C0 copies the gate position first and then performs
// at most 32 random-point/teleport tests. The gate's live position is supplied
// to every random callback, matching the original pointer into Object.PosVec.
func soulGateRespawnPoint4F80C0[O comparable](
	gate O,
	output *types.Pointf,
	h soulGateRespawnPointHooks4F80C0[O],
) int32 {
	*output = h.loadPosition(gate)
	var result int32
	for attempt := 0; attempt < 32; attempt++ {
		h.randomPoint(60, gate, output)
		result = h.allowTeleport(output)
		if result == 0 {
			break
		}
	}
	return result
}

const respawnPlayerCorpsePieceCount53FBC0 = 11

type respawnPlayerImplHooks53FBC0[O comparable] struct {
	loadInitialized  func() uint32
	initialize       func()
	directionIndex   func(int16) int32
	loadTypeIndex    func(int32, int) uint32
	newObject        func(uint32) O
	loadNetworkMode  func() uint32
	gameFlag         func(uint32) int32
	loadObjectFlags  func(O) uint32
	storeObjectFlags func(O, uint32)
	loadOffsetY      func(int32, int) float32
	loadCenterY      func() float32
	loadOffsetX      func(int32, int) float32
	loadCenterX      func() float32
	createAt         func(O, types.Pointf)
	randomInt        func(int32, int32) int32
	loadTickRate     func() uint32
	setDecay         func(O, uint32)
}

// respawnPlayerImpl53FBC0 restores the eleven-piece corpse burst without the
// original IA-32 object-pointer round trip through int. The original function
// result is discarded by every caller, so the native contract is deliberately
// void while all observable callbacks retain their machine-code order.
func respawnPlayerImpl53FBC0[O comparable](
	direction int16,
	h respawnPlayerImplHooks53FBC0[O],
) {
	if h.loadInitialized() == 0 {
		h.initialize()
	}
	directionIndex := h.directionIndex(direction)
	var zeroObject O
	for part := 0; part < respawnPlayerCorpsePieceCount53FBC0; part++ {
		typeIndex := h.loadTypeIndex(directionIndex, part)
		obj := h.newObject(typeIndex)
		if obj == zeroObject {
			break
		}
		if h.loadNetworkMode() != 0 && h.gameFlag(playerRespawnProtectionFlag4F7EF0) != 0 {
			flags := h.loadObjectFlags(obj)
			h.storeObjectFlags(obj, flags|0x40)
		}
		position := types.Pointf{
			Y: h.loadOffsetY(directionIndex, part) + h.loadCenterY(),
			X: h.loadOffsetX(directionIndex, part) + h.loadCenterX(),
		}
		h.createAt(obj, position)
		seconds := h.randomInt(10, 20)
		h.setDecay(obj, h.loadTickRate()*uint32(seconds))
	}
}
