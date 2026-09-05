package server

import "github.com/opennox/libs/types"

const (
	pentagramTeleportRejectedClassMask53C060 = uint32(0x00420000)
	pentagramTeleportPointFX53C060           = uint32(137)
	pentagramTeleportSound53C060             = uint32(147)
)

type pentagramUpdateHooks53BEF0[O comparable, D, P any] struct {
	loadUpdate          func(O) D
	loadState           func(D) uint8
	storeState          func(D, uint8)
	loadTriggered       func(D) uint32
	storeTriggered      func(D, uint32)
	loadAnimationFrame  func(D) uint8
	storeAnimationFrame func(D, uint8)
	loadAnimationTick   func(D) uint8
	storeAnimationTick  func(D, uint8)
	loadAnimationStep   func(D) uint8
	storeAnimationStep  func(D, uint8)
	needSync            func(O)
	loadDestination     func(O, D) O
	loadRadius          func(O) float32
	loadPosX            func(O) float32
	loadPosY            func(O) float32
	cachePosition       func(O) P
	eachInRect          func(types.Rectf, func(O))
	teleportVisible     func(O, P)
	teleportInvisible   func(O, P)
	isEnabled           func(O) bool
	frame               func() uint32
	storeField34        func(O, uint32)
}

type pentagramTeleportHooks53C060[O, P any] struct {
	loadClass     func(O) uint32
	cachePosition func(O) P
	pointFX       func(uint32, P)
	audio         func(uint32, O)
	teleport      func(O, P)
}

// pentagramTeleport53C060 preserves GAME.EXE 0053C060. Both point effects
// receive the same live position alias, so the second observes the completed
// teleport. Monster-generator and immobile objects are rejected before any
// other load.
func pentagramTeleport53C060[O, P any](
	unit O,
	destination P,
	hooks pentagramTeleportHooks53C060[O, P],
) {
	if hooks.loadClass(unit)&pentagramTeleportRejectedClassMask53C060 != 0 {
		return
	}
	position := hooks.cachePosition(unit)
	hooks.pointFX(pentagramTeleportPointFX53C060, position)
	hooks.audio(pentagramTeleportSound53C060, unit)
	hooks.teleport(unit, destination)
	hooks.pointFX(pentagramTeleportPointFX53C060, position)
	hooks.audio(pentagramTeleportSound53C060, unit)
}

// pentagramTeleportInvisible53C140 preserves GAME.EXE 0053C140. Unlike the
// visible callback it emits no effects around the shared teleport gate.
func pentagramTeleportInvisible53C140[O, P any](
	unit O,
	destination P,
	loadClass func(O) uint32,
	teleport func(O, P),
) {
	if loadClass(unit)&pentagramTeleportRejectedClassMask53C060 != 0 {
		return
	}
	teleport(unit, destination)
}

func pentagramAnimate53BEF0[O comparable, D, P any](
	pentagram O,
	update D,
	hooks pentagramUpdateHooks53BEF0[O, D, P],
) {
	state := hooks.loadState(update)
	if state != 0 {
		if state <= 2 {
			tick := hooks.loadAnimationTick(update)
			if tick == hooks.loadAnimationFrame(update) {
				hooks.storeAnimationStep(update, hooks.loadAnimationStep(update)+1)
				hooks.needSync(pentagram)
				if hooks.loadAnimationStep(update) == 9 {
					frame := hooks.loadAnimationFrame(update)
					hooks.storeAnimationStep(update, 1)
					hooks.storeAnimationFrame(update, frame+1)
				}
				hooks.storeAnimationTick(update, 0)
			} else {
				hooks.storeAnimationTick(update, tick+1)
			}
		}
		return
	}
	if hooks.loadAnimationStep(update) != 0 {
		hooks.needSync(pentagram)
	}
	hooks.storeAnimationStep(update, 0)
}

func pentagramRect53BEF0[O comparable, D, P any](
	pentagram O,
	hooks pentagramUpdateHooks53BEF0[O, D, P],
) types.Rectf {
	radius := hooks.loadRadius(pentagram)
	minX := hooks.loadPosX(pentagram) - radius
	minY := hooks.loadPosY(pentagram) - radius
	maxX := hooks.loadPosX(pentagram) + radius
	maxY := hooks.loadPosY(pentagram) + radius
	return types.Rectf{
		Min: types.Pointf{X: minX, Y: minY},
		Max: types.Pointf{X: maxX, Y: maxY},
	}
}

// pentagramUpdate53BEF0 preserves GAME.EXE 0053BEF0. The update pointer is
// cached once, animation advances before state handling, triggers are consumed
// on every non-early-return path, and paired activation samples gameFrame
// independently for the source and destination.
func pentagramUpdate53BEF0[O comparable, D, P any](
	pentagram O,
	hooks pentagramUpdateHooks53BEF0[O, D, P],
) int32 {
	update := hooks.loadUpdate(pentagram)
	pentagramAnimate53BEF0(pentagram, update, hooks)

	result := int32(hooks.loadState(update))
	if hooks.loadState(update) != 0 {
		stateAfterOne := result - 1
		if stateAfterOne != 0 {
			result = stateAfterOne - 1
			if result == 0 && hooks.loadAnimationFrame(update) >= 4 {
				hooks.storeAnimationStep(update, 0)
				hooks.storeState(update, 0)
				hooks.storeTriggered(update, 0)
				return result
			}
		} else {
			result = int32(hooks.loadAnimationFrame(update))
			if uint8(result) != 0 {
				if uint8(result) == 4 {
					hooks.storeState(update, 0)
					hooks.storeTriggered(update, 0)
					return result
				}
			} else if hooks.loadAnimationStep(update) == 8 {
				destination := hooks.loadDestination(pentagram, update)
				var zero O
				if destination != zero {
					rect := pentagramRect53BEF0(pentagram, hooks)
					destinationPosition := hooks.cachePosition(destination)
					hooks.eachInRect(rect, func(unit O) {
						hooks.teleportVisible(unit, destinationPosition)
					})
					hooks.storeTriggered(update, 0)
					return 1
				}
			}
		}
	} else if hooks.loadTriggered(update) != 0 {
		destination := hooks.loadDestination(pentagram, update)
		var zero O
		if destination != zero && hooks.isEnabled(pentagram) {
			destinationUpdate := hooks.loadUpdate(destination)
			hooks.storeState(update, 1)
			hooks.storeAnimationFrame(update, 0)
			hooks.storeAnimationTick(update, 0)
			hooks.storeField34(pentagram, hooks.frame())
			hooks.storeState(destinationUpdate, 2)
			hooks.storeAnimationFrame(destinationUpdate, 0)
			hooks.storeAnimationTick(destinationUpdate, 0)
			result = int32(hooks.frame())
			hooks.storeField34(destination, uint32(result))
		}
	}
	hooks.storeTriggered(update, 0)
	return result
}

// pentagramInvisibleUpdate53C0C0 preserves GAME.EXE 0053C0C0. Its trigger is
// consumed regardless of destination or enabled state, and accepted units use
// the effect-free callback at 0053C140.
func pentagramInvisibleUpdate53C0C0[O comparable, D, P any](
	pentagram O,
	hooks pentagramUpdateHooks53BEF0[O, D, P],
) int32 {
	update := hooks.loadUpdate(pentagram)
	if hooks.loadTriggered(update) != 0 {
		destination := hooks.loadDestination(pentagram, update)
		var zero O
		if destination != zero && hooks.isEnabled(pentagram) {
			rect := pentagramRect53BEF0(pentagram, hooks)
			destinationPosition := hooks.cachePosition(destination)
			hooks.eachInRect(rect, func(unit O) {
				hooks.teleportInvisible(unit, destinationPosition)
			})
		}
	}
	hooks.storeTriggered(update, 0)
	return 0
}
