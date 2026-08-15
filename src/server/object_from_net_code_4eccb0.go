package server

const objectDeadFlagLow4ECCB0 = uint8(0x20)

// objectFromNetCodeHooks4ECCB0 separates the cache, active-object list,
// inventory chains, pending-object list, and player-info list used by
// GAME.EXE 004ECCB0. Object and player identities remain pointer-width values;
// only the original 32-bit network code is represented as uint32.
type objectFromNetCodeHooks4ECCB0[O, P comparable] struct {
	cacheLookup    func(uint32) O
	cacheAdd       func(O)
	firstObject    func() O
	nextObject     func(O) O
	firstItem      func(O) O
	nextItem       func(O) O
	firstPending   func() O
	nextPending    func(O) O
	firstPlayer    func() P
	nextPlayer     func(P) P
	loadPlayerUnit func(P) O
	loadFlagsLow   func(O) uint8
	loadNetCode    func(O) uint32
}

func objectMatchesNetCode4ECCB0[O, P comparable](
	obj O,
	code uint32,
	hooks objectFromNetCodeHooks4ECCB0[O, P],
) bool {
	if hooks.loadFlagsLow(obj)&objectDeadFlagLow4ECCB0 != 0 {
		return false
	}
	return hooks.loadNetCode(obj) == code
}

// objectFromNetCode4ECCB0 preserves GAME.EXE 004ECCB0.
//
// The cache is queried first and a hit is returned without checking object
// flags or reloading the network code. A miss scans active top-level objects
// and each object's inventory, then pending objects. Matches from those three
// domains are inserted into the cache before their cached local identity is
// returned. Each candidate reads the low flags byte before its full 32-bit
// network code, and dead candidates skip the network-code load.
//
// Player-info records are scanned last. Their PlayerUnit pointer is loaded
// before the object fields and is loaded again on a match; that live second
// value is returned and player-unit matches are not inserted into the cache.
func objectFromNetCode4ECCB0[O, P comparable](
	code uint32,
	hooks objectFromNetCodeHooks4ECCB0[O, P],
) O {
	var zeroObject O
	if obj := hooks.cacheLookup(code); obj != zeroObject {
		return obj
	}

	for obj := hooks.firstObject(); obj != zeroObject; obj = hooks.nextObject(obj) {
		if objectMatchesNetCode4ECCB0(obj, code, hooks) {
			hooks.cacheAdd(obj)
			return obj
		}
		for item := hooks.firstItem(obj); item != zeroObject; item = hooks.nextItem(item) {
			if objectMatchesNetCode4ECCB0(item, code, hooks) {
				hooks.cacheAdd(item)
				return item
			}
		}
	}

	for obj := hooks.firstPending(); obj != zeroObject; obj = hooks.nextPending(obj) {
		if objectMatchesNetCode4ECCB0(obj, code, hooks) {
			hooks.cacheAdd(obj)
			return obj
		}
	}

	var zeroPlayer P
	for player := hooks.firstPlayer(); player != zeroPlayer; player = hooks.nextPlayer(player) {
		unit := hooks.loadPlayerUnit(player)
		if unit != zeroObject && objectMatchesNetCode4ECCB0(unit, code, hooks) {
			return hooks.loadPlayerUnit(player)
		}
	}
	return zeroObject
}
