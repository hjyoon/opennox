package server

import "math"

type playerReadValuesStat4EEDC0 uint8

const (
	playerReadValuesHealth4EEDC0 playerReadValuesStat4EEDC0 = iota
	playerReadValuesMana4EEDC0
	playerReadValuesSpeed4EEDC0
	playerReadValuesStrength4EEDC0
)

const playerReadValuesMaxLevel4EEDC0 = int8(10)

type playerReadValuesHooks4EEDC0[O, U, P, S, H, N any, I comparable] struct {
	loadUnitArg       func() O
	loadUpdateData    func(O) U
	loadPlayer        func(U) P
	loadBaseStats     func() S
	loadPlayerClass   func(P) uint8
	loadClassStats    func(uint8) S
	loadStat          func(S, playerReadValuesStat4EEDC0) float32
	gameFlagsCheck    func(uint32) int32
	floatToInt        func(float32) int32
	floatToInt16Abs   func(float32) int16
	loadHealthData    func(O) H
	storeHealthMax    func(H, uint16)
	loadHealthMax     func(H) uint16
	setHP             func(O, uint16)
	storeManaMax      func(U, uint16)
	loadManaMax       func(U) uint16
	storeManaCurrent  func(U, uint16)
	storeStrength     func(P, uint32)
	loadStrength      func(P) uint32
	storeSpeedStat    func(P, uint32)
	loadSpeedStat     func(P) uint32
	storeSpeedBase    func(O, float32)
	loadLevel         func(P) uint8
	soloMode          func() int32
	loadRewardArg     func() int32
	abilityGiveAll    func(O, int8, int32)
	storeMass         func(O, float32)
	floatToInt64Trunc func(float64) int64
	storeCapacityWord func(P, uint16)
	loadCapacityWord  func(P) uint16
	storeCarry        func(O, uint16)
	loadStrengthToken func(P) uint32
	loadSpeedToken    func(P) uint32
	loadManaMaxToken  func(P) uint32
	loadHealthToken   func(P) uint32
	protectInt        func(uint32, uint32)
	protectUint16     func(uint32, uint16)
	loadFirstItem     func(O) I
	loadItemWeight    func(I) uint8
	loadNextItem      func(I) I
	loadCarry         func(O) uint16
	storeOverweight   func(P, uint32)
	loadNameToken     func(P) uint32
	loadName          func(P) N
	wideLen           func(N) uint32
	protectName       func(N, uint32, uint32) int32
	storeInitialized  func(P, uint8)
}

// playerReadValues4EEDC0 preserves GAME.EXE 004EEDC0. The routine has no
// pointer guards: Unit.UpdateData and the initial Player are cached before any
// mode branch. Class/base stat handles are likewise cached, while individual
// values remain live memory reads. The interpolation path retains the signed
// low-byte level rule, binary32 spills, x87-style conversion boundaries,
// inventory int32 wrapping, repeated UpdateData.Player loads, and the final
// cached-Player initialized-byte store after name protection.
func playerReadValues4EEDC0[O, U, P, S, H, N any, I comparable](
	hooks playerReadValuesHooks4EEDC0[O, U, P, S, H, N, I],
) int32 {
	unit := hooks.loadUnitArg()
	update := hooks.loadUpdateData(unit)
	var inventoryWeight int32
	player := hooks.loadPlayer(update)
	baseStats := hooks.loadBaseStats()
	class := hooks.loadPlayerClass(player)
	maxStats := hooks.loadClassStats(class)
	baseClassStats := hooks.loadClassStats(0)

	if hooks.gameFlagsCheck(0x2000) != 0 {
		healthMaximum := uint16(hooks.floatToInt(hooks.loadStat(maxStats, playerReadValuesHealth4EEDC0)))
		health := hooks.loadHealthData(unit)
		hooks.storeHealthMax(health, healthMaximum)

		healthCurrent := uint16(hooks.floatToInt16Abs(hooks.loadStat(maxStats, playerReadValuesHealth4EEDC0)))
		hooks.setHP(unit, healthCurrent)

		manaMaximum := uint16(hooks.floatToInt(hooks.loadStat(maxStats, playerReadValuesMana4EEDC0)))
		hooks.storeManaMax(update, manaMaximum)
		manaCurrent := uint16(hooks.floatToInt(hooks.loadStat(maxStats, playerReadValuesMana4EEDC0)))
		hooks.storeManaCurrent(update, manaCurrent)

		strength := uint32(hooks.floatToInt(hooks.loadStat(maxStats, playerReadValuesStrength4EEDC0)))
		hooks.storeStrength(player, strength)

		speedValue := hooks.loadStat(maxStats, playerReadValuesSpeed4EEDC0)
		hooks.storeSpeedBase(unit, playerReadValuesScaleSpeed4EEDC0(speedValue))
		speed := uint32(hooks.floatToInt(speedValue))
		hooks.storeSpeedStat(player, speed)

		if hooks.loadPlayerClass(player) == 0 &&
			hooks.gameFlagsCheck(0x1000) == 0 &&
			hooks.soloMode() == 0 {
			hooks.abilityGiveAll(unit, playerReadValuesMaxLevel4EEDC0, 0)
		}
	} else {
		level := playerReadValuesClampLevel4EEDC0(hooks.loadLevel(player))
		factor := float32(int32(level) - 1)

		healthValue := playerReadValuesInterpolate4EEDC0(
			hooks.loadStat(maxStats, playerReadValuesHealth4EEDC0),
			hooks.loadStat(baseStats, playerReadValuesHealth4EEDC0),
			hooks.loadStat(baseStats, playerReadValuesHealth4EEDC0),
			factor,
		)
		healthRounded := float32(healthValue + float64(playerReadValuesFloat32Constant4EEDC0(0x3f000000)))
		healthMaximum := uint16(hooks.floatToInt(healthRounded))
		health := hooks.loadHealthData(unit)
		hooks.storeHealthMax(health, healthMaximum)
		health = hooks.loadHealthData(unit)
		hooks.setHP(unit, hooks.loadHealthMax(health))

		manaValue := playerReadValuesInterpolate4EEDC0(
			hooks.loadStat(maxStats, playerReadValuesMana4EEDC0),
			hooks.loadStat(baseStats, playerReadValuesMana4EEDC0),
			hooks.loadStat(baseStats, playerReadValuesMana4EEDC0),
			factor,
		)
		manaMaximumLive := float64(hooks.loadStat(maxStats, playerReadValuesMana4EEDC0))
		if manaValue > manaMaximumLive {
			manaValue = manaMaximumLive
		}
		manaRounded := float32(manaValue + float64(playerReadValuesFloat32Constant4EEDC0(0x3f000000)))
		mana := uint16(hooks.floatToInt(manaRounded))
		hooks.storeManaMax(update, mana)
		hooks.storeManaCurrent(update, mana)

		strengthValue := playerReadValuesInterpolate4EEDC0(
			hooks.loadStat(maxStats, playerReadValuesStrength4EEDC0),
			hooks.loadStat(baseStats, playerReadValuesStrength4EEDC0),
			hooks.loadStat(baseStats, playerReadValuesStrength4EEDC0),
			factor,
		)
		strengthRounded := float32(strengthValue + float64(playerReadValuesFloat32Constant4EEDC0(0x3f000000)))
		hooks.storeStrength(player, uint32(hooks.floatToInt(strengthRounded)))

		speedValue := playerReadValuesInterpolate4EEDC0(
			hooks.loadStat(maxStats, playerReadValuesSpeed4EEDC0),
			hooks.loadStat(baseStats, playerReadValuesSpeed4EEDC0),
			hooks.loadStat(baseStats, playerReadValuesSpeed4EEDC0),
			factor,
		)
		hooks.storeSpeedBase(unit, playerReadValuesScaleSpeedExtended4EEDC0(speedValue))
		speedRounded := float32(speedValue + float64(playerReadValuesFloat32Constant4EEDC0(0x3f000000)))
		hooks.storeSpeedStat(player, uint32(hooks.floatToInt(speedRounded)))

		if hooks.loadPlayerClass(player) == 0 {
			rewardArg := hooks.loadRewardArg()
			hooks.abilityGiveAll(unit, level, rewardArg)
		}
	}

	strength := hooks.loadStrength(player)
	baseStrength := hooks.loadStat(baseClassStats, playerReadValuesStrength4EEDC0)
	hooks.storeMass(unit, playerReadValuesMass4EEDC0(strength, baseStrength))

	strength = hooks.loadStrength(player)
	baseStrength = hooks.loadStat(baseClassStats, playerReadValuesStrength4EEDC0)
	capacityValue := playerReadValuesCapacity4EEDC0(strength, baseStrength)
	capacity := uint16(hooks.floatToInt64Trunc(capacityValue))
	capacityPlayer := hooks.loadPlayer(update)
	hooks.storeCapacityWord(capacityPlayer, capacity)
	capacityPlayer = hooks.loadPlayer(update)
	capacity = hooks.loadCapacityWord(capacityPlayer)
	hooks.storeCarry(unit, capacity)

	protectedPlayer := hooks.loadPlayer(update)
	strength = hooks.loadStrength(player)
	strengthToken := hooks.loadStrengthToken(protectedPlayer)
	hooks.protectInt(strengthToken, strength)

	protectedPlayer = hooks.loadPlayer(update)
	speed := hooks.loadSpeedStat(player)
	speedToken := hooks.loadSpeedToken(protectedPlayer)
	hooks.protectInt(speedToken, speed)

	protectedPlayer = hooks.loadPlayer(update)
	manaMaximum := hooks.loadManaMax(update)
	manaToken := hooks.loadManaMaxToken(protectedPlayer)
	hooks.protectUint16(manaToken, manaMaximum)

	health := hooks.loadHealthData(unit)
	protectedPlayer = hooks.loadPlayer(update)
	healthMaximum := hooks.loadHealthMax(health)
	healthToken := hooks.loadHealthToken(protectedPlayer)
	hooks.protectUint16(healthToken, healthMaximum)

	item := hooks.loadFirstItem(unit)
	var nilItem I
	for item != nilItem {
		weight := hooks.loadItemWeight(item)
		next := hooks.loadNextItem(item)
		inventoryWeight = playerReadValuesAddWeight4EEDC0(inventoryWeight, weight)
		item = next
	}
	carry := hooks.loadCarry(unit)
	if playerReadValuesOverweight4EEDC0(inventoryWeight, carry) {
		hooks.storeOverweight(player, 1)
	} else {
		hooks.storeOverweight(player, 0)
	}

	namePlayer := hooks.loadPlayer(update)
	nameToken := hooks.loadNameToken(namePlayer)
	name := hooks.loadName(namePlayer)
	nameLength := hooks.wideLen(name)
	namePlayer = hooks.loadPlayer(update)
	name = hooks.loadName(namePlayer)
	result := hooks.protectName(name, nameLength<<1, nameToken)
	hooks.storeInitialized(player, 1)
	return result
}

func playerReadValuesFloat32Constant4EEDC0(bits uint32) float32 {
	return math.Float32frombits(bits)
}

func playerReadValuesClampLevel4EEDC0(level uint8) int8 {
	signed := int8(level)
	if signed > playerReadValuesMaxLevel4EEDC0 {
		return playerReadValuesMaxLevel4EEDC0
	}
	return signed
}

// playerReadValuesInterpolate4EEDC0 models the x87 FSUB/FMUL/FMUL/FADD
// sequence. The two base operands are separate because GAME.EXE reads that
// memory twice and an intervening callback may make the second value differ.
func playerReadValuesInterpolate4EEDC0(maximum, baseSubtract, baseAdd, factor float32) float64 {
	value := float64(maximum) - float64(baseSubtract)
	value *= float64(playerReadValuesFloat32Constant4EEDC0(0x3de38e39))
	value *= float64(factor)
	value += float64(baseAdd)
	return value
}

func playerReadValuesScaleSpeed4EEDC0(speed float32) float32 {
	return playerReadValuesScaleSpeedExtended4EEDC0(float64(speed))
}

func playerReadValuesScaleSpeedExtended4EEDC0(speed float64) float32 {
	return float32(speed * float64(playerReadValuesFloat32Constant4EEDC0(0x38d1b717)))
}

func playerReadValuesMass4EEDC0(strength uint32, baseStrength float32) float32 {
	value := float64(int32(strength)) / float64(baseStrength)
	value *= float64(playerReadValuesFloat32Constant4EEDC0(0x41a00000))
	value += float64(playerReadValuesFloat32Constant4EEDC0(0x41200000))
	return float32(value)
}

func playerReadValuesCapacity4EEDC0(strength uint32, baseStrength float32) float64 {
	value := float64(int32(strength)) / float64(baseStrength)
	value *= float64(playerReadValuesFloat32Constant4EEDC0(0x449c4000))
	value += float64(playerReadValuesFloat32Constant4EEDC0(0x443b8000))
	value *= 1.5
	return value
}

func playerReadValuesAddWeight4EEDC0(sum int32, weight uint8) int32 {
	return sum + int32(weight)
}

func playerReadValuesOverweight4EEDC0(sum int32, capacity uint16) bool {
	return sum > int32(capacity)
}
