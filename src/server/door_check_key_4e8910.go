package server

const (
	doorKeyClassByte4E8910    = uint8(0x40)
	doorPlayerClassByte4E8910 = uint8(0x04)
)

type doorCheckKeyHooks4E8910[O comparable, D any] struct {
	loadDoorData         func(O) D
	loadLockCode         func(D) uint8
	loadOwner            func(O) O
	firstItem            func(O) O
	loadClassByte        func(O) uint8
	loadTypeName         func(O) string
	nextItem             func(O) O
	hasQuestGameMode     func() bool
	loadQuestKeyState    func() int32
	playersHaveSilverKey func() O
}

// doorCheckKey4E8910 preserves GAME.EXE 004E8910. Door update data is cached,
// but its lock byte is reloaded after every eligible item-name lookup and at
// the final Quest fallback. The unit class byte is read after the inventory
// walk even when a matching key was already found.
func doorCheckKey4E8910[O comparable, D any](unit, door O, hooks doorCheckKeyHooks4E8910[O, D]) O {
	data := hooks.loadDoorData(door)
	if hooks.loadLockCode(data) == 5 {
		var zero O
		return zero
	}

	var zero O
	if hooks.loadOwner(door) != zero {
		return zero
	}

	var found O
	item := hooks.firstItem(unit)
	for item != zero {
		if hooks.loadClassByte(item)&doorKeyClassByte4E8910 != 0 {
			name := hooks.loadTypeName(item)
			lockCode := hooks.loadLockCode(data)
			if doorKeyNameMatches4E8910(lockCode, name) {
				found = item
				break
			}
		}
		item = hooks.nextItem(item)
	}

	if hooks.loadClassByte(unit)&doorPlayerClassByte4E8910 == 0 {
		return found
	}
	if found != zero {
		return found
	}
	if !hooks.hasQuestGameMode() {
		return zero
	}
	if hooks.loadQuestKeyState() != 1 {
		return zero
	}
	if hooks.loadLockCode(data) != 1 {
		return zero
	}
	return hooks.playersHaveSilverKey()
}

func doorKeyNameMatches4E8910(lockCode uint8, name string) bool {
	var expected string
	switch lockCode {
	case 1:
		expected = "SilverKey"
	case 2:
		expected = "GoldKey"
	case 3:
		expected = "RubyKey"
	case 4:
		expected = "SapphireKey"
	default:
		return false
	}
	if len(name) < len(expected) || name[:len(expected)] != expected {
		return false
	}
	// The original rep cmpsb includes the terminating NUL. A Go object-type ID
	// normally omits it, while an explicit embedded NUL models the same bytes.
	return len(name) == len(expected) || name[len(expected)] == 0
}

type playersHaveSilverKeyHooks4E8A10[O comparable] struct {
	loadCachedTypeID  func() uint32
	lookupTypeID      func() uint32
	storeCachedTypeID func(uint32)
	firstPlayerUnit   func() O
	nextPlayerUnit    func(O) O
	firstItem         func(O) O
	nextItem          func(O) O
	loadTypeInd       func(O) uint16
}

// playersHaveSilverKey4E8A10 preserves private GAME.EXE 004E8A10. The first
// pass remembers the last player unit with a positive signed 32-bit key count;
// the second pass returns that unit's first matching item. The global type-ID
// cache is read before every item TypeInd load in both passes.
func playersHaveSilverKey4E8A10[O comparable](hooks playersHaveSilverKeyHooks4E8A10[O]) O {
	if hooks.loadCachedTypeID() == 0 {
		id := hooks.lookupTypeID()
		hooks.storeCachedTypeID(id)
	}

	var zero O
	var foundUnit O
	unit := hooks.firstPlayerUnit()
	for unit != zero {
		count := int32(0)
		item := hooks.firstItem(unit)
		for item != zero {
			cached := hooks.loadCachedTypeID()
			typeInd := hooks.loadTypeInd(item)
			if uint32(typeInd) == cached {
				count++
			}
			item = hooks.nextItem(item)
		}
		if count > 0 {
			foundUnit = unit
		}
		unit = hooks.nextPlayerUnit(unit)
	}

	if foundUnit == zero {
		return zero
	}
	item := hooks.firstItem(foundUnit)
	for item != zero {
		cached := hooks.loadCachedTypeID()
		typeInd := hooks.loadTypeInd(item)
		if uint32(typeInd) == cached {
			return item
		}
		item = hooks.nextItem(item)
	}
	return zero
}
