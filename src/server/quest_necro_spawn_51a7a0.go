package server

import (
	"math"
	"math/big"
)

type questNecroSpawnHooks51A7A0[O, D comparable, U, H, T, P any] struct {
	newObject       func(string) O
	healthScale     func() float64
	loadUpdate      func(O) U
	loadDefinition  func(U) D
	loadQuestHealth func(D) int32
	loadType        func(O) uint16
	lookupType      func(uint16) T
	loadTypeHealth  func(T) H
	loadHealth      func(O) H
	loadCurrent     func(H) uint16
	loadMaximum     func(H) uint16
	setHP           func(O, uint16)
	storeMaximum    func(H, uint16)
	storeAI         func(U, uint32, uint32)
	loadPosition    func() P
	createAt        func(O, P)
	stage           func() uint32
	activateReward  func(O, uint32) O
	inventoryPut    func(O, O, int32)
	freeObject      func(O)
}

// questMinionHealth51A7A0 preserves the signed FILD-dword times binary32
// product. Its at-most-55-bit significand is exact in x87, but not necessarily
// in binary64. Current HP truncates that unspilled product to a signed qword;
// maximum HP rounds the separate binary32 spill to a signed dword. Both
// setters retain only the low word; invalid conversions have a zero low word.
func questMinionHealth51A7A0(base int32, scale float32) (current, maximum uint16) {
	if math.IsNaN(float64(scale)) || math.IsInf(float64(scale), 0) {
		return 0, 0
	}
	var a, b, product big.Float
	a.SetPrec(64).SetInt64(int64(base))
	b.SetPrec(64).SetFloat64(float64(scale))
	product.SetPrec(64).Mul(&a, &b)
	integer, _ := product.Int(nil)
	if integer.IsInt64() {
		current = uint16(integer.Int64())
	}
	spilled, _ := product.Float32()
	maximum = uint16(questInventoryRoundFloat32ToInt32_4F2C30(spilled))
	return current, maximum
}

// questNecroSpawn51A7A0 is GAME.EXE 0051A7A0, including the scale read after
// failed allocation, unordered scale clamp, cached update pointer, live health
// reloads after HP callbacks, late position read and one stage+2 reward.
func questNecroSpawn51A7A0[O, D comparable, U, H, T, P any](h questNecroSpawnHooks51A7A0[O, D, U, H, T, P]) {
	unit := h.newObject("Necromancer")
	scale := float32(h.healthScale())
	var zero O
	if unit == zero {
		return
	}
	update := h.loadUpdate(unit)
	definition := h.loadDefinition(update)
	var zeroDefinition D
	var base int32
	if definition != zeroDefinition {
		base = h.loadQuestHealth(definition)
	} else {
		base = int32(h.loadMaximum(h.loadTypeHealth(h.lookupType(h.loadType(unit)))))
	}
	// FCOM's C0 bit is set for both less-than and unordered operands.
	if scale < 1 || math.IsNaN(float64(scale)) {
		scale = 1
	}
	current, maximum := questMinionHealth51A7A0(base, scale)
	h.setHP(unit, current)
	h.storeMaximum(h.loadHealth(unit), maximum)
	if h.loadCurrent(h.loadHealth(unit)) == 0 {
		h.setHP(unit, 1)
	}
	health := h.loadHealth(unit)
	if h.loadMaximum(health) == 0 {
		h.storeMaximum(health, 1)
	}
	h.storeAI(update, 340, 4)
	h.storeAI(update, 411, 0x10000000)
	h.storeAI(update, 423, 0x10000000)
	h.storeAI(update, 326, 0x3f547ae1)
	h.storeAI(update, 510, 1)
	h.storeAI(update, 410, 0x08000000)
	h.storeAI(update, 444, 0x20000000)
	h.storeAI(update, 415, 0x40000000)
	h.createAt(unit, h.loadPosition())
	marker := h.newObject("RewardMarker")
	if marker == zero {
		return
	}
	if reward := h.activateReward(marker, h.stage()+2); reward != zero {
		h.inventoryPut(unit, reward, 0)
	}
	h.freeObject(marker)
}
