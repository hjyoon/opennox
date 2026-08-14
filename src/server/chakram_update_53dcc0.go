package server

import "math"

const chakramUpdateLifetimeSeconds53DCC0 = uint32(5)

type chakramUpdateHooks53DCC0[O comparable, U any] struct {
	loadUpdateData    func(O) U
	inventoryFirst    func(O) O
	loadFlags         func(O) uint32
	loadLastHit       func(U) O
	storeLastHit      func(U, O)
	loadOwner         func(O) O
	loadPosX          func(O) float32
	loadPosY          func(O) float32
	storeOwnerPosX    func(U, float32)
	storeOwnerPosY    func(U, float32)
	loadOwnerPosX     func(U) float32
	loadOwnerPosY     func(U) float32
	mapCheck          func(O, O) bool
	loadReturnState   func(U) uint8
	storeReturnState  func(U, uint8)
	loadReturnTarget  func(U) O
	storeReturnTarget func(U, O)
	loadSpeed         func(O) float32
	storeVelocityX    func(O, float32)
	storeVelocityY    func(O, float32)
	frame             func() uint32
	loadCreateFrame   func(O) uint32
	frameRate         func() uint32
	delayedDelete     func(O)
}

// chakramUpdate53DCC0 preserves GAME.EXE 0053DCC0. The update pointer and
// valid owner used for OwnerPos are cached, while map checks and return-target
// writes reload the live owner. Steering uses the original asymmetric x87 Y
// spill, and lifetime expiry uses wrapping uint32 frame arithmetic.
func chakramUpdate53DCC0[O comparable, U any](source O, hooks chakramUpdateHooks53DCC0[O, U]) {
	update := hooks.loadUpdateData(source)
	item := hooks.inventoryFirst(source)
	var zero O
	if item == zero || hooks.loadFlags(item)&chakramDestroyedFlag4EAF00 != 0 {
		hooks.delayedDelete(source)
		return
	}

	lastHit := hooks.loadLastHit(update)
	if lastHit != zero && hooks.loadFlags(lastHit)&chakramDestroyedFlag4EAF00 != 0 {
		hooks.storeLastHit(update, zero)
	}

	owner := hooks.loadOwner(source)
	if owner == zero || hooks.loadFlags(owner)&chakramDestroyedFlag4EAF00 != 0 {
		hooks.storeReturnState(update, chakramReturnStateDrop4EAF00)
		hooks.storeReturnTarget(update, zero)
	} else {
		hooks.storeOwnerPosX(update, hooks.loadPosX(owner))
		hooks.storeOwnerPosY(update, hooks.loadPosY(owner))
		if !hooks.mapCheck(source, hooks.loadOwner(source)) {
			hooks.storeReturnTarget(update, zero)
		} else if hooks.loadReturnState(update) != chakramReturnStateHome4EAF00 {
			goto lifetime
		} else {
			hooks.storeReturnTarget(update, hooks.loadOwner(source))
		}
		if hooks.loadReturnState(update) != chakramReturnStateHome4EAF00 {
			goto lifetime
		}
		returnTarget := hooks.loadReturnTarget(update)
		if returnTarget != zero && hooks.loadFlags(returnTarget)&chakramInvalidOwnerMask4EAF00 != 0 {
			hooks.storeReturnTarget(update, zero)
			hooks.storeReturnState(update, chakramReturnStateDrop4EAF00)
		} else {
			dx := float64(hooks.loadOwnerPosX(update)) - float64(hooks.loadPosX(source))
			dyExtended := float64(hooks.loadOwnerPosY(update)) - float64(hooks.loadPosY(source))
			dy := float32(dyExtended)
			denominator := float32(math.Sqrt(dyExtended*float64(dy)+dx*dx) + float64(chakramRetargetEpsilon4EB250))
			hooks.storeVelocityX(source, float32(dx*float64(hooks.loadSpeed(source))/float64(denominator)))
			hooks.storeVelocityY(source, float32(float64(dy)*float64(hooks.loadSpeed(source))/float64(denominator)))
		}
	}

lifetime:
	delta := hooks.frame() - hooks.loadCreateFrame(source)
	limit := hooks.frameRate() * chakramUpdateLifetimeSeconds53DCC0
	if delta > limit {
		hooks.storeReturnState(update, chakramReturnStateDrop4EAF00)
		hooks.storeReturnTarget(update, zero)
	}
}
