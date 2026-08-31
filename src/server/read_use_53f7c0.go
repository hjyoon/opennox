package server

const readUsePlayerClass53F7C0 = uint8(0x04)

type readUseHooks53F7C0[O any, D any] struct {
	loadOwnerArg    func() O
	loadClassLow    func(O) uint8
	loadReadableArg func() O
	loadFPS         func() uint32
	loadFrame       func() uint32
	loadUseData     func(O) D
	loadReadState   func(D) uint32
	mapCheck        func(O, O) int32
	primaryMessage  func(O, D, uint8)
	storeReadState  func(D, uint32)
}

// readUse53F7C0 preserves GAME.EXE 0053F7C0. The owner class byte is the
// first dereference; a non-Player returns canonical one without observing the
// readable object or timing state. The Player path caches the readable object
// and UseData pointer, then applies the original unsigned three-second
// cooldown. A zero state bypasses the cooldown.
//
// Visibility must return exactly one. On success, the primary message uses
// the cached UseData pointer and exact value one. The current frame is loaded
// again after that callback and stored through the same cached data pointer.
func readUse53F7C0[O any, D any](hooks readUseHooks53F7C0[O, D]) int32 {
	owner := hooks.loadOwnerArg()
	if hooks.loadClassLow(owner)&readUsePlayerClass53F7C0 == 0 {
		return 1
	}

	readable := hooks.loadReadableArg()
	fps := hooks.loadFPS()
	frame := hooks.loadFrame()
	data := hooks.loadUseData(readable)
	state := hooks.loadReadState(data)
	if state != 0 && frame-state <= 3*fps {
		return 1
	}
	if hooks.mapCheck(owner, readable) != 1 {
		return 1
	}

	hooks.primaryMessage(owner, data, 1)
	frame = hooks.loadFrame()
	hooks.storeReadState(data, frame)
	return 1
}
