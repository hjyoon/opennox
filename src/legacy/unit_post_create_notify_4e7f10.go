package legacy

type unitPostCreateNotifyHooks4E7F10[O, P comparable] struct {
	storeField35   func(O, uint32)
	storeField36   func(O, uint32)
	firstPlayer    func() P
	loadPlayerInd  func(P) uint8
	loadPlayerUnit func(P) O
	isHostile      func(O, O) int32
	loadField35    func(O) uint32
	loadField36    func(O) uint32
	nextPlayer     func(P) P
}

// unitPostCreateNotify4E7F10 preserves GAME.EXE 004E7F10. The two object
// masks are cleared before the player-list head is requested. Per player, the
// index is read before the unit pointer, and only an exact hostile result of
// one sets the cached x86 shift bit in both live masks. The successor is read
// after every callback and mask update.
func unitPostCreateNotify4E7F10[O, P comparable](
	obj O,
	hooks unitPostCreateNotifyHooks4E7F10[O, P],
) P {
	hooks.storeField35(obj, 0)
	hooks.storeField36(obj, 0)

	var zeroObject O
	var zeroPlayer P
	for player := hooks.firstPlayer(); player != zeroPlayer; player = hooks.nextPlayer(player) {
		playerInd := hooks.loadPlayerInd(player)
		unit := hooks.loadPlayerUnit(player)
		// IA-32 masks a variable shift count to five bits. Preserve that
		// behavior for malformed or synthetic player indices as well.
		bit := uint32(1) << (playerInd & 31)
		if unit != zeroObject && hooks.isHostile(unit, obj) == 1 {
			field35 := hooks.loadField35(obj)
			field36 := hooks.loadField36(obj)
			hooks.storeField35(obj, field35|bit)
			hooks.storeField36(obj, field36|bit)
		}
	}
	return zeroPlayer
}
