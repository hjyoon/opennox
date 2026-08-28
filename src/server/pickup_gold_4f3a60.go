package server

const (
	pickupGoldPlayerClassLow4F3A60 = uint8(0x04)
	pickupGoldAudio4F3A60          = uint32(307)
	pickupGoldMessageKey4F3A60     = "GoldPickup"
	pickupGoldMessagePath4F3A60    = `C:\NoxPost\src\Server\Object\pickdrop\pickup.c`
	pickupGoldMessageLine4F3A60    = 709
)

// pickupGoldHooks4F3A60 exposes every observable field load and call in
// GAME.EXE 004F3A60. GoldInitData is a distinct token so tests can prove that
// its pointer is cached while Amount is loaded once before gold addition and
// again after delayed deletion. The registered callback's trailing arguments
// remain delayed loads because the Player branch never reads them.
type pickupGoldHooks4F3A60[O, D, M any] struct {
	loadOwnerClassLow func(O) uint8
	loadGoldInitData  func(O) D
	loadGoldAmount    func(D) uint32
	addGold           func(O, uint32)
	delayedDelete     func(O)
	loadString        func(string, string, int) M
	sendLineMessage   func(O, M, uint32)
	audio             func(uint32, O, int32, uint32)
	loadArg4          func() int32
	loadArg3          func() int32
	defaultPickup     func(O, O, int32, int32) int32
}

// pickupGold4F3A60 preserves GAME.EXE 004F3A60.
//
// A Player caches the item's GoldInitData pointer, adds its first live Amount,
// schedules deletion, then reloads Amount through the cached pointer for the
// localized line message. The callback returns canonical one after pickup
// audio. A non-Player loads arg4 before arg3, forwards all four registered
// callback arguments to DefaultPickup, emits audio for any nonzero result, and
// returns the full signed int32 unchanged. The original has no nil guards.
func pickupGold4F3A60[O, D, M any](
	owner, item O,
	hooks pickupGoldHooks4F3A60[O, D, M],
) int32 {
	if hooks.loadOwnerClassLow(owner)&pickupGoldPlayerClassLow4F3A60 != 0 {
		data := hooks.loadGoldInitData(item)
		amount := hooks.loadGoldAmount(data)
		hooks.addGold(owner, amount)
		hooks.delayedDelete(item)
		amount = hooks.loadGoldAmount(data)
		message := hooks.loadString(
			pickupGoldMessageKey4F3A60,
			pickupGoldMessagePath4F3A60,
			pickupGoldMessageLine4F3A60,
		)
		hooks.sendLineMessage(owner, message, amount)
		hooks.audio(pickupGoldAudio4F3A60, owner, 0, 0)
		return 1
	}

	arg4 := hooks.loadArg4()
	arg3 := hooks.loadArg3()
	result := hooks.defaultPickup(owner, item, arg3, arg4)
	if result != 0 {
		hooks.audio(pickupGoldAudio4F3A60, owner, 0, 0)
	}
	return result
}
