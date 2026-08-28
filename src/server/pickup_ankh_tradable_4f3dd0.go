package server

const (
	pickupAnkhTradablePlayerClassLow4F3DD0 = uint8(0x04)
	pickupAnkhTradableAudio4F3DD0          = uint32(1004)
)

// pickupAnkhTradableHooks4F3DD0 exposes every observable field access and
// call in GAME.EXE 004F3DD0. The item remains a delayed argument load because
// the original does not read it until after ExtraLives has been stored.
type pickupAnkhTradableHooks4F3DD0[O, U any] struct {
	loadOwnerClassLow func(O) uint8
	loadOwnerUpdate   func(O) U
	loadExtraLives    func(U) uint32
	storeExtraLives   func(U, uint32)
	loadItemArg       func() O
	delayedDelete     func(O)
	audio             func(uint32, O, int32, uint32)
}

// pickupAnkhTradable4F3DD0 preserves GAME.EXE 004F3DD0.
//
// A non-player owner returns canonical zero after only the low class-byte
// load. A player caches UpdateData, increments its uint32 ExtraLives with
// wrapping arithmetic, then loads and delayed-deletes the item and emits
// sound 1004 for the cached owner. The original has no nil guards.
func pickupAnkhTradable4F3DD0[O, U any](
	owner O,
	hooks pickupAnkhTradableHooks4F3DD0[O, U],
) int32 {
	if hooks.loadOwnerClassLow(owner)&pickupAnkhTradablePlayerClassLow4F3DD0 == 0 {
		return 0
	}

	update := hooks.loadOwnerUpdate(owner)
	extraLives := hooks.loadExtraLives(update)
	hooks.storeExtraLives(update, extraLives+1)
	item := hooks.loadItemArg()
	hooks.delayedDelete(item)
	hooks.audio(pickupAnkhTradableAudio4F3DD0, owner, 0, 0)
	return 1
}
