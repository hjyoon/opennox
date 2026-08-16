package server

const (
	decayPendingFlag511660 = uint32(0x00010000)
	decayListedFlag511660  = uint32(0x00400000)
	decayDeleteFlag511750  = uint32(0x00000080)
)

type decayRemoveResultKind5116F0 uint8

const (
	decayRemoveWord5116F0 decayRemoveResultKind5116F0 = iota
	decayRemoveObject5116F0
)

// decayRemoveResult5116F0 keeps the original mixed EAX domains distinct.
// An unlisted object returns its flags, an empty/missed search returns zero,
// a head match returns the object itself, and a later match returns its next
// object. Native pointers therefore never need to pass through a 32-bit word.
type decayRemoveResult5116F0[O any] struct {
	kind   decayRemoveResultKind5116F0
	word   uint32
	object O
}

// decayHooks511660 separates the original ABI32 link at object offset 468
// from native object identity. Argument loads remain hooks because 00511660
// reads the delay only after its pending/listed gates and optional removal.
type decayHooks511660[O comparable] struct {
	loadSetObjectArg func() O
	loadSetDelayArg  func() uint32

	loadObjectFlags  func(O) uint32
	storeObjectFlags func(O, uint32)
	loadFrame        func() uint32
	loadDeadline     func(O) uint32
	storeDeadline    func(O, uint32)

	loadHead  func() O
	storeHead func(O)
	loadNext  func(O) O
	storeNext func(O, O)

	loadHolder       func(O) O
	loadDeleteFlags  func(O) uint32
	storeDeleteFlags func(O, uint32)
	delayedDelete    func(O)
}

// decaySetTime511660 preserves GAME.EXE 00511660. Deadlines and comparisons
// are unsigned 32-bit values. Equal deadlines are inserted after existing
// entries. The tail path deliberately reloads flags before storing a null
// next link, unlike the head and middle paths.
func decaySetTime511660[O comparable](hooks decayHooks511660[O]) uint32 {
	obj := hooks.loadSetObjectArg()
	result := hooks.loadObjectFlags(obj)
	if result&decayPendingFlag511660 != 0 {
		return result
	}
	if result&decayListedFlag511660 != 0 {
		decayRemove5116F0(obj, hooks)
	}

	delay := hooks.loadSetDelayArg()
	deadline := hooks.loadFrame() + delay
	hooks.storeDeadline(obj, deadline)

	var zero, previous O
	current := hooks.loadHead()
	for current != zero {
		if deadline < hooks.loadDeadline(current) {
			break
		}
		previous = current
		current = hooks.loadNext(current)
	}

	if previous != zero {
		hooks.storeNext(previous, obj)
		if current == zero {
			result = hooks.loadObjectFlags(obj)
			hooks.storeNext(obj, zero)
			result |= decayListedFlag511660
			hooks.storeObjectFlags(obj, result)
			return result
		}
	} else {
		hooks.storeHead(obj)
	}

	hooks.storeNext(obj, current)
	result = hooks.loadObjectFlags(obj)
	result |= decayListedFlag511660
	hooks.storeObjectFlags(obj, result)
	return result
}

// decayRemove5116F0 preserves GAME.EXE 005116F0. The listed bit is cleared
// before the head is read, even when the object is absent. Removal never
// clears the removed object's own next link.
func decayRemove5116F0[O comparable](item O, hooks decayHooks511660[O]) decayRemoveResult5116F0[O] {
	flags := hooks.loadObjectFlags(item)
	if flags&decayListedFlag511660 == 0 {
		return decayRemoveResult5116F0[O]{kind: decayRemoveWord5116F0, word: flags}
	}

	hooks.storeObjectFlags(item, flags&^decayListedFlag511660)
	var zero, previous O
	current := hooks.loadHead()
	if current == zero {
		return decayRemoveResult5116F0[O]{kind: decayRemoveWord5116F0}
	}
	for current != item {
		previous = current
		current = hooks.loadNext(current)
		if current == zero {
			return decayRemoveResult5116F0[O]{kind: decayRemoveWord5116F0}
		}
	}

	if previous != zero {
		next := hooks.loadNext(item)
		hooks.storeNext(previous, next)
		return decayRemoveResult5116F0[O]{kind: decayRemoveObject5116F0, object: next}
	}
	next := hooks.loadNext(item)
	hooks.storeHead(next)
	return decayRemoveResult5116F0[O]{kind: decayRemoveObject5116F0, object: item}
}

// decayTick511750 preserves GAME.EXE 00511750. Holder and next are loaded in
// that order before any branch. A future deadline stops the whole sorted-list
// scan. Due-node deletion continues through the next pointer cached before
// removal and before the delayed-delete callback.
func decayTick511750[O comparable](hooks decayHooks511660[O]) {
	var zero O
	current := hooks.loadHead()
	for current != zero {
		holder := hooks.loadHolder(current)
		next := hooks.loadNext(current)
		if holder != zero {
			decayRemove5116F0(current, hooks)
		} else {
			deadline := hooks.loadDeadline(current)
			if deadline > hooks.loadFrame() {
				return
			}
			decayRemove5116F0(current, hooks)
			flags := hooks.loadDeleteFlags(current)
			hooks.storeDeleteFlags(current, flags|decayDeleteFlag511750)
			hooks.delayedDelete(current)
		}
		current = next
	}
}

// decayDestroy5117B0 preserves GAME.EXE 005117B0. Each next link is cached
// before removal and the head receives a final null store even when it was
// already empty or the last removal already cleared it.
func decayDestroy5117B0[O comparable](hooks decayHooks511660[O]) uint32 {
	var zero O
	current := hooks.loadHead()
	if current != zero {
		for {
			next := hooks.loadNext(current)
			decayRemove5116F0(current, hooks)
			current = next
			if current == zero {
				break
			}
		}
	}
	hooks.storeHead(zero)
	return 0
}
