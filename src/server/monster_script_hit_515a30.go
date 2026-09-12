package server

import (
	"math"

	"github.com/opennox/libs/types"
)

const (
	monsterScriptHitClassLow515A30 = uint8(0x02)
	monsterScriptHitDeadFlag515A30 = uint32(0x8000)
)

// The two script attack entrypoints share their eligibility and action-stack
// services, but not their action sequences. O and A remain native-width handles;
// only action numbers and position payloads are PE32 dwords.
type monsterScriptHitHooks515A30[O, A comparable] struct {
	loadClassLow     func(O) uint8
	loadFlags        func(O) uint32
	canAttack        func(O) bool
	clearActionStack func(O)
	pushAction       func(O, uint32) A
	storeArgBits     func(A, int, uint32)
	loadMeleeRange   func(O) float32
	loadRadius       func(O) float32
}

// monsterScriptHitMelee515A30 restores GAME.EXE 00515A30. The PE32 code does
// not check pos for nil: it reads the location only after the action pushes.
func monsterScriptHitMelee515A30[O, A comparable](unit O, pos *types.Pointf, h monsterScriptHitHooks515A30[O, A]) {
	var nilUnit O
	if unit == nilUnit || h.loadClassLow(unit)&monsterScriptHitClassLow515A30 == 0 ||
		h.loadFlags(unit)&monsterScriptHitDeadFlag515A30 != 0 || !h.canAttack(unit) {
		return
	}
	h.clearActionStack(unit)
	var nilAction A
	if item := h.pushAction(unit, 32); item != nilAction {
		h.storeArgBits(item, 0, 16)
	}
	h.pushAction(unit, 16)
	if item := h.pushAction(unit, 51); item != nilAction {
		// x87 loads both binary32 values, adds them at extended precision,
		// then stores one binary32 result into the dependency argument.
		rangeValue := h.loadMeleeRange(unit)
		radius := h.loadRadius(unit)
		h.storeArgBits(item, 0, math.Float32bits(float32(float64(rangeValue)+float64(radius))))
		h.storeArgBits(item, 2, math.Float32bits(pos.X))
		h.storeArgBits(item, 3, math.Float32bits(pos.Y))
	}
	if item := h.pushAction(unit, 7); item != nilAction {
		h.storeArgBits(item, 0, math.Float32bits(pos.X))
		h.storeArgBits(item, 1, math.Float32bits(pos.Y))
		h.storeArgBits(item, 2, 0)
	}
}

// monsterScriptHitMissile515B80 restores GAME.EXE 00515B80. Unlike melee,
// the original checks a nil location before reading any object fields.
func monsterScriptHitMissile515B80[O, A comparable](unit O, pos *types.Pointf, h monsterScriptHitHooks515A30[O, A]) {
	var nilUnit O
	if unit == nilUnit || pos == nil || h.loadClassLow(unit)&monsterScriptHitClassLow515A30 == 0 ||
		h.loadFlags(unit)&monsterScriptHitDeadFlag515A30 != 0 || !h.canAttack(unit) {
		return
	}
	h.clearActionStack(unit)
	var nilAction A
	if item := h.pushAction(unit, 32); item != nilAction {
		h.storeArgBits(item, 0, 17)
	}
	if item := h.pushAction(unit, 17); item != nilAction {
		h.storeArgBits(item, 0, math.Float32bits(pos.X))
		h.storeArgBits(item, 1, math.Float32bits(pos.Y))
		h.storeArgBits(item, 2, 0)
	}
}
