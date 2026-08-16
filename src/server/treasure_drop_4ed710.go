package server

const (
	treasureDropPlayerClass4ED710 = uint8(4)
	treasureDropGameFlag4ED710    = uint32(64)
	treasureDropAudio4ED710       = uint32(308)
)

// treasureDropHooks4ED710 exposes the argument, field, and callback order in
// GAME.EXE 004ED710. The owner UpdateData pointer is loaded after the game-flag
// callback and cached, while its Player pointer is reloaded after the treasure
// maximum callback.
type treasureDropHooks4ED710[O, U, L, P any] struct {
	loadPointArg    func() P
	loadTreasureArg func() O
	loadOwnerArg    func() O
	defaultDrop     func(O, O, P) int32

	loadClassLow func(O) uint8
	gameFlag     func(uint32) int32
	loadUpdate   func(O) U
	loadPlayer   func(U) L
	loadCount    func(L) uint32
	storeCount   func(L, uint32)
	treasureMax  func() uint32
	storeMax     func(L, uint32)
	report       func(O)
	audio        func(uint32, O, int32, uint32)
}

// treasureDrop4ED710 preserves GAME.EXE 004ED710. DefaultDrop gates all later
// work with its whole EAX result. Only a Player owner in Scavenger mode updates
// the two wrapping uint32 treasure counters and emits report/audio callbacks.
func treasureDrop4ED710[O, U, L, P any](hooks treasureDropHooks4ED710[O, U, L, P]) int32 {
	point := hooks.loadPointArg()
	treasure := hooks.loadTreasureArg()
	owner := hooks.loadOwnerArg()
	if hooks.defaultDrop(owner, treasure, point) == 0 {
		return 0
	}
	if hooks.loadClassLow(owner)&treasureDropPlayerClass4ED710 == 0 {
		return 1
	}
	if hooks.gameFlag(treasureDropGameFlag4ED710) == 0 {
		return 1
	}

	update := hooks.loadUpdate(owner)
	player := hooks.loadPlayer(update)
	count := hooks.loadCount(player)
	hooks.storeCount(player, count-1)
	maximum := hooks.treasureMax()
	player = hooks.loadPlayer(update)
	hooks.storeMax(player, maximum)
	hooks.report(owner)
	hooks.audio(treasureDropAudio4ED710, owner, 0, 0)
	return 1
}
