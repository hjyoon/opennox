package server

const (
	teleportWakeAnchoredEnchant4EAE30  = uint32(14)
	teleportWakeInvisibleEnchant4EAE30 = uint32(0)
	teleportWakeOwnerPlayerBit4EAE30   = uint8(0x04)
	teleportWakeTargetClassMask4EAE30  = uint32(0x03001016)
	teleportWakePreFX4EAE30            = uint32(138)
	teleportWakePostFX4EAE30           = uint32(137)
	teleportWakeSound4EAE30            = uint32(147)
)

type teleportWakeCollideHooks4EAE30[O comparable, D, P any] struct {
	loadCollideData  func(O) D
	hasEnchant       func(O, uint32) bool
	questMode        func() bool
	loadOwner        func(O) O
	loadOwnerClassLo func(O) uint8
	loadTargetClass  func(O) uint32
	position         func(O) P
	pointFX          func(uint32, P)
	audio            func(uint32, O)
	teleport         func(O, D)
}

// teleportWakeCollide4EAE30 preserves GAME.EXE 004EAE30. The source
// collide-data pointer is cached before the nil-target guard. Anchored targets
// return before the Quest owner gate; in Quest mode a non-nil owner must have
// the low-byte Player bit. The complete target class word is then tested
// against the original acceptance mask.
//
// Invisibility is queried independently before and after teleport. Each
// accepted query takes a live target-position reference for its point effect.
// Both sounds are unconditional within the accepted class path, the teleport
// receives the entry-cached destination pointer, and collision is never read.
func teleportWakeCollide4EAE30[O comparable, D, P, C any](
	source, target O,
	_ C,
	hooks teleportWakeCollideHooks4EAE30[O, D, P],
) {
	destination := hooks.loadCollideData(source)

	var zero O
	if target == zero {
		return
	}
	if hooks.hasEnchant(target, teleportWakeAnchoredEnchant4EAE30) {
		return
	}
	if hooks.questMode() {
		owner := hooks.loadOwner(source)
		if owner != zero && hooks.loadOwnerClassLo(owner)&teleportWakeOwnerPlayerBit4EAE30 == 0 {
			return
		}
	}
	if hooks.loadTargetClass(target)&teleportWakeTargetClassMask4EAE30 == 0 {
		return
	}

	if !hooks.hasEnchant(target, teleportWakeInvisibleEnchant4EAE30) {
		hooks.pointFX(teleportWakePreFX4EAE30, hooks.position(target))
	}
	hooks.audio(teleportWakeSound4EAE30, target)
	hooks.teleport(target, destination)
	if !hooks.hasEnchant(target, teleportWakeInvisibleEnchant4EAE30) {
		hooks.pointFX(teleportWakePostFX4EAE30, hooks.position(target))
	}
	hooks.audio(teleportWakeSound4EAE30, target)
}
