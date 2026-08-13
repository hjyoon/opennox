package server

type objectPlayerMasksRebuildHooks4E8110[O, P comparable] struct {
	playerByInd    func(int32) P
	firstObject    func() O
	loadField36    func(O) uint32
	loadField35    func(O) uint32
	loadClassLow   func(O) uint8
	storeField36   func(O, uint32)
	storeField35   func(O, uint32)
	loadPlayerUnit func(P) O
	isHostile      func(O, O) int32
	nextObject     func(O) O
}

// objectPlayerMasksRebuild4E8110 preserves GAME.EXE 004E8110. The IA-32
// shift count is masked to five bits before the player lookup. For every
// object, Field36, Field35, and the low class byte are cached in that order;
// both player bits are cleared before a qualifying unit is tested again.
func objectPlayerMasksRebuild4E8110[O, P comparable](
	playerInd int32,
	hooks objectPlayerMasksRebuildHooks4E8110[O, P],
) O {
	bit := uint32(1) << (uint32(playerInd) & 31)
	var zeroObject O
	var zeroPlayer P
	player := hooks.playerByInd(playerInd)
	if player == zeroPlayer {
		return zeroObject
	}
	obj := hooks.firstObject()
	if obj == zeroObject {
		return zeroObject
	}
	clear := ^bit
	for obj != zeroObject {
		field36 := hooks.loadField36(obj)
		field35 := hooks.loadField35(obj)
		classLow := hooks.loadClassLow(obj)
		hooks.storeField36(obj, field36&clear)
		hooks.storeField35(obj, field35&clear)

		if classLow&0x6 != 0 {
			unit := hooks.loadPlayerUnit(player)
			if unit != zeroObject {
				hostile := hooks.isHostile(unit, obj)
				liveField36 := hooks.loadField36(obj)
				if hostile == 1 {
					if liveField36&bit == 0 {
						hooks.storeField36(obj, liveField36|bit)
						liveField35 := hooks.loadField35(obj)
						hooks.storeField35(obj, liveField35|bit)
					}
				} else if liveField36&bit != 0 {
					reloadedField36 := hooks.loadField36(obj)
					hooks.storeField36(obj, reloadedField36&clear)
					liveField35 := hooks.loadField35(obj)
					hooks.storeField35(obj, liveField35|bit)
				}
			}
		}
		obj = hooks.nextObject(obj)
	}
	return obj
}
