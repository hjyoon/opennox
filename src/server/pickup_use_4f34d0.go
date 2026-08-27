package server

const pickupUseDestroyedFlagsLow4F34D0 = uint8(0x20)

type pickupUseHooks4F34D0[O any] struct {
	useByNetCode  func(O, O) int32
	loadFlagsLow  func(O) uint8
	defaultPickup func(O, O, int32, int32) int32
}

// pickupUse4F34D0 preserves GAME.EXE 004F34D0. The use-by-net-code helper is
// called unconditionally and its result is discarded. The item flags are
// loaded afterwards, so a use callback may destroy the item and turn the
// wrapper into a canonical-success path. Otherwise all four registered
// callback arguments reach DefaultPickup and its full int32 result is returned.
// There are deliberately no owner or item nil guards in this wrapper.
func pickupUse4F34D0[O any](
	owner, item O,
	arg3, arg4 int32,
	hooks pickupUseHooks4F34D0[O],
) int32 {
	_ = hooks.useByNetCode(owner, item)
	if hooks.loadFlagsLow(item)&pickupUseDestroyedFlagsLow4F34D0 != 0 {
		return 1
	}
	return hooks.defaultPickup(owner, item, arg3, arg4)
}
