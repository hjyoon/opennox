package server

import (
	"math"

	"github.com/opennox/opennox/v1/common/unit/ai"
)

const (
	unitOrderMonsterClass533900 = uint8(0x02)
	unitOrderPlayerClass533900  = uint8(0x04)
	unitOrderBlockedFlag5339A0  = uint32(0x00008000)
	unitOrderSummonedFlag533900 = uint32(0x00000080)
	unitOrderShootFlag5339A0    = uint32(0x00000040)
)

var (
	unitOrderMovingSpeed5339A0 = math.Float32frombits(0x3c23d70a)
	unitOrderCalm5339A0        = math.Float32frombits(0x3f000000)
	unitOrderAggressive5339A0  = math.Float32frombits(0x3f547ae1)
	unitOrderGuardSight5339A0  = math.Float32frombits(0x437a0000)
)

// unitOrderDispatchHooks533900 describes the observable object-list accesses
// and calls made by GAME.EXE 00533900. O and U remain native-width handles;
// in particular, the owned-list successor is loaded after enact returns.
type unitOrderDispatchHooks533900[O comparable, U any] struct {
	loadClassLow func(O) uint8
	firstOwned   func(O) O
	nextOwned    func(O) O
	loadUpdate   func(O) U
	loadStatus   func(U) uint32
	localOrder   func(O, uint32)
	enactOrder   func(O, O, uint32)
}

func unitOrder533900[O comparable, U any](owner, creature O, order uint32, hooks unitOrderDispatchHooks533900[O, U]) {
	var nilObject O
	if owner == nilObject {
		return
	}
	if creature != nilObject {
		hooks.enactOrder(owner, creature, order)
		return
	}
	if hooks.loadClassLow(owner)&unitOrderPlayerClass533900 != 0 && (order == 4 || order == 3 || order == 5) {
		hooks.localOrder(owner, order)
	}
	for unit := hooks.firstOwned(owner); unit != nilObject; unit = hooks.nextOwned(unit) {
		if hooks.loadClassLow(unit)&unitOrderMonsterClass533900 == 0 {
			continue
		}
		update := hooks.loadUpdate(unit)
		if hooks.loadStatus(update)&unitOrderSummonedFlag533900 != 0 {
			hooks.enactOrder(owner, unit, order)
		}
	}
}

// unitOrderEnactHooks5339A0 exposes each load, store, and callback from
// GAME.EXE 005339A0. The first UpdateData value is intentionally cached before
// source/class validation. ESCORT alone reloads it at case entry, matching the
// original binary.
type unitOrderEnactHooks5339A0[O, U, D, A comparable] struct {
	loadUpdate       func(O) U
	loadClassLow     func(O) uint8
	isZombie         func(O) bool
	loadObjectFlags  func(O) uint32
	loadMonsterDef   func(U) D
	storeMonsterDef  func(U, D)
	loadTypeIndex    func(O) uint16
	monsterDefByType func(uint16) D
	loadOwner        func(O) O
	banish           func(O)
	observe          func(O, O)
	monsterCommand   func(O, O, string, int16)
	playOrderSound   func(O)
	loadStatus       func(U) uint32
	storeStatus      func(U, uint32)
	storeAggression  func(U, float32)
	storeSightRange  func(U, float32)
	isMoving         func(O) bool
	canShoot         func(O) bool
	clearActionStack func(O)
	pushAction       func(O, ai.ActionType) A
	setGuardArgs     func(A, O)
	setEscortArgs    func(A, O)
}

func enactUnitOrder5339A0[O, U, D, A comparable](source, unit O, order uint32, hooks unitOrderEnactHooks5339A0[O, U, D, A]) {
	update := hooks.loadUpdate(unit)
	var nilObject O
	if source == nilObject {
		return
	}
	if hooks.loadClassLow(unit)&unitOrderMonsterClass533900 == 0 {
		return
	}
	if !hooks.isZombie(unit) || order != 0 {
		if hooks.loadObjectFlags(unit)&unitOrderBlockedFlag5339A0 != 0 {
			return
		}
	}

	def := hooks.loadMonsterDef(update)
	var nilDef D
	if def == nilDef {
		def = hooks.monsterDefByType(hooks.loadTypeIndex(unit))
		hooks.storeMonsterDef(update, def)
		if def == nilDef {
			return
		}
	}

	var nilAction A
	switch order {
	case 0: // BANISH
		if hooks.loadOwner(unit) == source {
			hooks.banish(unit)
		}
	case 1: // OBSERVE
		if hooks.loadClassLow(source)&unitOrderPlayerClass533900 != 0 {
			hooks.observe(source, unit)
		}
	case 2: // IDLE
		if hooks.loadClassLow(source)&unitOrderPlayerClass533900 != 0 {
			hooks.monsterCommand(unit, source, "MonUtil.c:idle", 0)
		}
		hooks.playOrderSound(unit)
		status := hooks.loadStatus(update)
		hooks.storeAggression(update, unitOrderCalm5339A0)
		hooks.storeStatus(update, status&^unitOrderShootFlag5339A0)
		hooks.clearActionStack(unit)
		hooks.pushAction(unit, ai.ACTION_IDLE)
	case 3: // GUARD
		if !hooks.isMoving(unit) {
			return
		}
		if hooks.loadClassLow(source)&unitOrderPlayerClass533900 != 0 {
			hooks.monsterCommand(unit, source, "MonUtil.c:guarding", 0)
		}
		hooks.playOrderSound(unit)
		hooks.storeAggression(update, unitOrderCalm5339A0)
		if hooks.canShoot(unit) {
			status := hooks.loadStatus(update)
			hooks.storeStatus(update, status|unitOrderShootFlag5339A0)
		}
		hooks.storeSightRange(update, unitOrderGuardSight5339A0)
		hooks.clearActionStack(unit)
		action := hooks.pushAction(unit, ai.ACTION_GUARD)
		if action != nilAction {
			hooks.setGuardArgs(action, unit)
		}
	case 4: // ESCORT
		escortUpdate := hooks.loadUpdate(unit)
		if !hooks.isMoving(unit) {
			return
		}
		if hooks.loadClassLow(source)&unitOrderPlayerClass533900 != 0 {
			hooks.monsterCommand(unit, source, "MonUtil.c:escorting", 0)
		}
		hooks.playOrderSound(unit)
		status := hooks.loadStatus(escortUpdate)
		hooks.storeAggression(escortUpdate, unitOrderAggressive5339A0)
		hooks.storeStatus(escortUpdate, status&^unitOrderShootFlag5339A0)
		hooks.clearActionStack(unit)
		action := hooks.pushAction(unit, ai.ACTION_ESCORT)
		if action != nilAction {
			hooks.setEscortArgs(action, source)
		}
	case 5: // HUNT
		if !hooks.isMoving(unit) {
			return
		}
		hooks.playOrderSound(unit)
		if hooks.loadClassLow(source)&unitOrderPlayerClass533900 != 0 {
			hooks.monsterCommand(unit, source, "MonUtil.c:Hunting", 0)
		}
		status := hooks.loadStatus(update)
		hooks.storeAggression(update, unitOrderAggressive5339A0)
		hooks.storeStatus(update, status&^unitOrderShootFlag5339A0)
		hooks.clearActionStack(unit)
		hooks.pushAction(unit, ai.ACTION_HUNT)
	default:
		return
	}
}
