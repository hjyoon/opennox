package server

const (
	pickupTrapPlayerClassLow4F3510 = uint8(0x04)
	pickupTrapSuccessAudio4F3510   = uint32(824)
	pickupTrapRejectAudio4F3510    = uint32(925)
)

// pickupTrapHooks4F3510 exposes every observable call and delayed load in
// GAME.EXE 004F3510. In particular, the registered callback's fourth scalar
// argument is loaded before its third argument and only after the owner-chain
// predicate accepts the trap.
type pickupTrapHooks4F3510[O any] struct {
	hasOwner func(O, O) int32

	loadArg4      func() int32
	loadArg3      func() int32
	defaultPickup func(O, O, int32, int32) int32

	loadOwnerClassLow func(O) uint8
	loadOwnerNetCode  func(O) uint32
	audio             func(uint32, O, int32, uint32)
}

// pickupTrap4F3510 preserves GAME.EXE 004F3510. The trap must be the picking
// object itself or have it in its owner chain. Admission forwards all four
// registered callback arguments to DefaultPickup, returns its full int32
// result, and only a nonzero result emits TrapPickup audio. Rejection reads the
// owner's live class and, for a Player, its live NetCode before emitting
// NoCanDo. There are deliberately no owner or item nil guards in this wrapper.
func pickupTrap4F3510[O any](
	owner, item O,
	hooks pickupTrapHooks4F3510[O],
) int32 {
	if hooks.hasOwner(item, owner) != 0 {
		arg4 := hooks.loadArg4()
		arg3 := hooks.loadArg3()
		result := hooks.defaultPickup(owner, item, arg3, arg4)
		if result != 0 {
			hooks.audio(pickupTrapSuccessAudio4F3510, owner, 0, 0)
		}
		return result
	}
	if hooks.loadOwnerClassLow(owner)&pickupTrapPlayerClassLow4F3510 != 0 {
		code := hooks.loadOwnerNetCode(owner)
		hooks.audio(pickupTrapRejectAudio4F3510, owner, 2, code)
	}
	return 0
}
