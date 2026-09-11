package server

// audioEventInsertHooks501EA0 describes the observable loads, stores, and
// calls made by GAME.EXE 00501EA0. Descriptor and event handles remain
// native-width; only fields that were dwords in the PE32 oracle use int32.
type audioEventInsertHooks501EA0[D, E comparable] struct {
	loadEventSound    func(E) int32
	descriptor        func(int32) D
	testBitmap        func(int32) uint32
	setBitmap         func(int32)
	storeListHead     func(D, E)
	storeEventPercent func(E, int32)
	insert            func(D, E)
}

// audioEventInsert501EA0 preserves GAME.EXE 00501EA0's access and callback
// order. The sound ID and descriptor are cached before the bitmap test. A
// sound first seen in the current player pass has its bit set and its stale
// list head cleared before the event percentage is stored.
func audioEventInsert501EA0[D, E comparable](
	event E,
	percentage int32,
	hooks audioEventInsertHooks501EA0[D, E],
) {
	soundID := hooks.loadEventSound(event)
	descriptor := hooks.descriptor(soundID)
	if hooks.testBitmap(soundID) == 0 {
		hooks.setBitmap(soundID)
		var nilEvent E
		hooks.storeListHead(descriptor, nilEvent)
	}
	hooks.storeEventPercent(event, percentage)
	hooks.insert(descriptor, event)
}

func soundBitmapWordAndMask501EF0(soundID int32) (uint32, uint32) {
	id := uint32(soundID)
	return id >> 5, uint32(1) << (id & 0x1f)
}

// testSoundBitmap501EF0 preserves GAME.EXE 00501EF0, including its raw
// masked-bit return rather than canonicalizing the result to a boolean.
func testSoundBitmap501EF0(bitmap *soundBitmap501E80, soundID int32) uint32 {
	word, mask := soundBitmapWordAndMask501EF0(soundID)
	return bitmap[word] & mask
}

// setSoundBitmap501F10 preserves GAME.EXE 00501F10's logical dword index and
// low-five-bit selection.
func setSoundBitmap501F10(bitmap *soundBitmap501E80, soundID int32) {
	word, mask := soundBitmapWordAndMask501EF0(soundID)
	bitmap[word] |= mask
}

// audioEventListHooks501F30 exposes the pointer-bearing list accesses from
// GAME.EXE 00501F30 without representing a native pointer as a PE32 dword.
type audioEventListHooks501F30[D, E comparable] struct {
	loadHead     func(D) E
	storeHead    func(D, E)
	loadPercent  func(E) int32
	loadSound    func(E) int32
	descriptor   func(int32) D
	loadFlagsLow func(D) uint8
	loadLimit    func(D) int32
	loadNext     func(E) E
	storeNext    func(E, E)
}

// insertAudioEventList501F30 preserves GAME.EXE 00501F30's signed-dword
// ordering and capacity behavior. In particular, subtraction and negation
// intentionally retain int32 overflow, the new percentage is cached once,
// and the new event's sound is loaded only on the near-percentage branch.
func insertAudioEventList501F30[D, E comparable](
	descriptor D,
	event E,
	hooks audioEventListHooks501F30[D, E],
) {
	current := hooks.loadHead(descriptor)
	initialHead := current
	previous := current
	var nilEvent E
	if current == nilEvent {
		hooks.storeNext(event, current)
		hooks.storeHead(descriptor, event)
		return
	}

	count := int32(1)
	percentage := hooks.loadPercent(event)
	for {
		currentPercentage := hooks.loadPercent(current)
		difference := percentage - currentPercentage
		if difference < 0 {
			difference = -difference
		}
		insertBefore := false
		if difference >= 5 {
			insertBefore = percentage > currentPercentage
		} else {
			soundID := hooks.loadSound(event)
			soundDescriptor := hooks.descriptor(soundID)
			insertBefore = hooks.loadFlagsLow(soundDescriptor)&0x10 != 0
		}
		if insertBefore {
			break
		}

		count++
		if count > hooks.loadLimit(descriptor) {
			return
		}
		previous = current
		current = hooks.loadNext(current)
		if current == nilEvent {
			break
		}
	}

	hooks.storeNext(event, current)
	if current == initialHead {
		hooks.storeHead(descriptor, event)
	} else {
		hooks.storeNext(previous, event)
	}
}
