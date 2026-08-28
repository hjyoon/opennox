package server

const (
	pickupAbilityBookGameFlags4F3CE0        = uint32(0x1800)
	pickupAbilityBookDestroyedFlagLow4F3CE0 = uint8(0x20)
	pickupAbilityBookAudio4F3CE0            = uint32(826)
)

// pickupAbilityBookHooks4F3CE0 exposes each observable field load and call in
// GAME.EXE 004F3CE0. The registered callback's trailing arguments remain
// delayed loads because a use-by-net-code callback can destroy the item before
// DefaultPickup is reached.
type pickupAbilityBookHooks4F3CE0[O any] struct {
	gameFlagsCheck   func(uint32) int32
	useByNetCode     func(O, O)
	loadItemFlagsLow func(O) uint8
	loadArg4         func() int32
	loadArg3         func() int32
	defaultPickup    func(O, O, int32, int32) int32
	audio            func(uint32, O, int32, uint32)
}

// pickupAbilityBook4F3CE0 preserves GAME.EXE 004F3CE0.
//
// Game modes selected by 0x1800 first invoke UseByNetCode. Item flags are then
// read live; a destroyed item returns canonical one without loading either
// trailing callback argument. Otherwise arg4 is loaded before arg3 and both
// are forwarded to DefaultPickup. Its full signed int32 result is preserved,
// and any nonzero result emits sound 826. The original has no nil guards.
func pickupAbilityBook4F3CE0[O any](
	owner, item O,
	hooks pickupAbilityBookHooks4F3CE0[O],
) int32 {
	if hooks.gameFlagsCheck(pickupAbilityBookGameFlags4F3CE0) != 0 {
		hooks.useByNetCode(owner, item)
	}
	if hooks.loadItemFlagsLow(item)&pickupAbilityBookDestroyedFlagLow4F3CE0 != 0 {
		return 1
	}

	arg4 := hooks.loadArg4()
	arg3 := hooks.loadArg3()
	result := hooks.defaultPickup(owner, item, arg3, arg4)
	if result != 0 {
		hooks.audio(pickupAbilityBookAudio4F3CE0, owner, 0, 0)
	}
	return result
}
