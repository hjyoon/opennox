package server

const (
	trapDoorCollidePlayerClass4EAB60  = uint32(0x00000004)
	trapDoorCollideDoorClass4EAB60    = uint32(0x00000080)
	trapDoorCollideEnabled4EAB60      = uint32(0x01000000)
	trapDoorCollideFallFlags4EAB60    = uint32(0x00060000)
	trapDoorCollideTreadLightly4EAB60 = int32(4)
)

type trapDoorCollideHooks4EAB60[O comparable, D any] struct {
	loadCollideData    func(O) D
	loadClass          func(O) uint32
	loadFlags          func(O) uint32
	loadShapeKind      func(O) uint32
	loadBoxWidth       func(O) float32
	loadBoxHeight      func(O) float32
	loadCircleRadius   func(O) float32
	mapPointInBox      func(O, O) bool
	orFlags            func(O, uint32)
	loadFallVelocityX  func(D) int32
	storeFallVelocityX func(O, float32)
	loadFallVelocityY  func(D) int32
	storeFallVelocityY func(O, float32)
	loadPosX           func(O) float32
	storeFallPosX      func(O, float32)
	loadPosY           func(O) float32
	storeFallPosY      func(O, float32)
	loadActivated      func(D) uint32
	abilityActive      func(O, int32) int32
	loadDelay          func(D) uint16
	gameFrame          func() uint32
	storeNextFrame     func(D, uint32)
	scriptCallback     func(D, O, O, ScriptEventType)
	storeActivated     func(D, uint32)
}

// trapDoorCollide4EAB60 preserves GAME.EXE 004EAB60. CollideData is cached
// before the nil-target branch. The target class is then cached before the
// source enabled flag, and the collision argument is never observed.
//
// The enabled path keeps the original x87 comparison behavior: unordered box
// dimensions reject, while an unordered circle diameter comparison continues.
// On a successful point-in-box test, falling flags, X/Y velocities and X/Y
// positions are stored in that exact order.
//
// The inactive path reads Activated first, skips Tread Lightly for non-players,
// reads Delay after the ability callback, calls the script with the cached data
// block, and stores Activated only after that callback returns.
func trapDoorCollide4EAB60[O comparable, D, C any](
	source, target O,
	_ C,
	hooks trapDoorCollideHooks4EAB60[O, D],
) {
	data := hooks.loadCollideData(source)

	var zero O
	if target == zero {
		return
	}
	targetClass := hooks.loadClass(target)
	if targetClass&trapDoorCollideDoorClass4EAB60 != 0 {
		return
	}

	if hooks.loadFlags(source)&trapDoorCollideEnabled4EAB60 != 0 {
		switch hooks.loadShapeKind(target) {
		case uint32(ShapeKindBox):
			if !(hooks.loadBoxWidth(source) >= hooks.loadBoxWidth(target)) {
				return
			}
			if !(hooks.loadBoxHeight(source) >= hooks.loadBoxHeight(target)) {
				return
			}
		case uint32(ShapeKindCircle):
			radius := float64(hooks.loadCircleRadius(target))
			diameter := mapPointInBoxAdd64_57B850(radius, radius)
			if diameter > float64(hooks.loadBoxWidth(source)) {
				return
			}
			if diameter > float64(hooks.loadBoxHeight(source)) {
				return
			}
		}

		if !hooks.mapPointInBox(source, target) {
			return
		}
		hooks.orFlags(target, trapDoorCollideFallFlags4EAB60)
		hooks.storeFallVelocityX(target, float32(hooks.loadFallVelocityX(data)))
		hooks.storeFallVelocityY(target, float32(hooks.loadFallVelocityY(data)))
		hooks.storeFallPosX(target, hooks.loadPosX(source))
		hooks.storeFallPosY(target, hooks.loadPosY(source))
		return
	}

	if hooks.loadActivated(data) != 0 {
		return
	}
	if targetClass&trapDoorCollidePlayerClass4EAB60 != 0 &&
		hooks.abilityActive(target, trapDoorCollideTreadLightly4EAB60) != 0 {
		return
	}
	if delay := hooks.loadDelay(data); delay != 0 {
		hooks.storeNextFrame(data, hooks.gameFrame()+uint32(delay))
	}
	hooks.scriptCallback(data, target, source, NoxEventTrapdoorCollide)
	hooks.storeActivated(data, 1)
}
