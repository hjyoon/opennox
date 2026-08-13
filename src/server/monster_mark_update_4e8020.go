package server

type monsterMarkUpdateHooks4E8020[O, P comparable] struct {
	firstPlayer    func() P
	loadPlayerInd  func(P) uint8
	loadPlayerUnit func(P) O
	isHostile      func(O, O) int32
	loadField36    func(O) uint32
	loadField35    func(O) uint32
	storeField36   func(O, uint32)
	storeField35   func(O, uint32)
	nextPlayer     func(P) P
}

// monsterMarkUpdate4E8020 preserves GAME.EXE 004E8020. The player-list head
// is requested before the object is touched. Per player, the index precedes
// the unit pointer and the IA-32 variable shift masks the count to five bits.
// A nil unit clears both cached masks unconditionally; for a non-nil unit,
// Field35 is marked only when the Field36 state bit changes.
func monsterMarkUpdate4E8020[O, P comparable](
	obj O,
	hooks monsterMarkUpdateHooks4E8020[O, P],
) {
	var zeroPlayer P
	var zeroObject O
	for player := hooks.firstPlayer(); player != zeroPlayer; player = hooks.nextPlayer(player) {
		playerInd := hooks.loadPlayerInd(player)
		unit := hooks.loadPlayerUnit(player)
		bit := uint32(1) << (playerInd & 31)

		if unit == zeroObject {
			field36 := hooks.loadField36(obj)
			field35 := hooks.loadField35(obj)
			clear := ^bit
			hooks.storeField36(obj, field36&clear)
			hooks.storeField35(obj, field35&clear)
			continue
		}

		hostile := hooks.isHostile(unit, obj)
		field36 := hooks.loadField36(obj)
		if hostile == 1 {
			if field36&bit == 0 {
				hooks.storeField36(obj, field36|bit)
				field35 := hooks.loadField35(obj)
				hooks.storeField35(obj, field35|bit)
			}
		} else if field36&bit != 0 {
			hooks.storeField36(obj, field36&^bit)
			field35 := hooks.loadField35(obj)
			hooks.storeField35(obj, field35|bit)
		}
	}
}
