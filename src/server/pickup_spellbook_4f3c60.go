package server

const (
	pickupSpellBookGameFlags4F3C60        = uint32(0x1800)
	pickupSpellBookDestroyedFlagLow4F3C60 = uint8(0x20)
	pickupSpellBookSpellSubClassLow4F3C60 = uint8(0x01)
	pickupSpellBookLearnedAudio4F3C60     = uint32(826)
	pickupSpellBookInventoryAudio4F3C60   = uint32(828)
)

// pickupSpellBookHooks4F3C60 exposes each observable field load and call in
// GAME.EXE 004F3C60. The registered callback's trailing arguments remain
// delayed loads because a use-by-net-code callback can destroy the item before
// DefaultPickup is reached.
type pickupSpellBookHooks4F3C60[O any] struct {
	gameFlagsCheck      func(uint32) int32
	useByNetCode        func(O, O)
	loadItemFlagsLow    func(O) uint8
	loadArg4            func() int32
	loadArg3            func() int32
	defaultPickup       func(O, O, int32, int32) int32
	loadItemSubClassLow func(O) uint8
	audio               func(uint32, O, int32, uint32)
}

// pickupSpellBook4F3C60 preserves GAME.EXE 004F3C60.
//
// Game modes selected by 0x1800 first invoke UseByNetCode. Item flags are then
// read live; a destroyed item returns canonical one without loading either
// trailing callback argument. Otherwise arg4 is loaded before arg3 and both
// are forwarded to DefaultPickup. Its full signed int32 result is preserved.
// A nonzero result reloads the item's subclass and emits sound 826 when its
// low Spell bit is set, or 828 otherwise. The original has no nil guards.
func pickupSpellBook4F3C60[O any](
	owner, item O,
	hooks pickupSpellBookHooks4F3C60[O],
) int32 {
	if hooks.gameFlagsCheck(pickupSpellBookGameFlags4F3C60) != 0 {
		hooks.useByNetCode(owner, item)
	}
	if hooks.loadItemFlagsLow(item)&pickupSpellBookDestroyedFlagLow4F3C60 != 0 {
		return 1
	}

	arg4 := hooks.loadArg4()
	arg3 := hooks.loadArg3()
	result := hooks.defaultPickup(owner, item, arg3, arg4)
	if result != 0 {
		sound := pickupSpellBookInventoryAudio4F3C60
		if hooks.loadItemSubClassLow(item)&pickupSpellBookSpellSubClassLow4F3C60 != 0 {
			sound = pickupSpellBookLearnedAudio4F3C60
		}
		hooks.audio(sound, owner, 0, 0)
	}
	return result
}
