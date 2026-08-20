package server

const (
	audEventDropRowCapacity4EE2F0 = 50
	audEventDropRowStorage4EE2F0  = audEventDropRowCapacity4EE2F0 + 1
	audEventDropSentinel4EE2F0    = uint16(0xffff)
)

type audEventDropSoundRow4EE2F0 struct {
	typeInd uint16
	sound   uint16
}

// audEventDropHooks4EE2F0 exposes GAME.EXE 004EE2F0's exact argument,
// DefaultDrop, ordered-table, object-field, and audio access order.
type audEventDropHooks4EE2F0[O, P comparable] struct {
	loadOwnerArg func() O
	loadItemArg  func() O
	loadPointArg func() P

	defaultDrop  func(O, O, P) int32
	loadRowType  func(int) uint16
	loadTypeInd  func(O) uint16
	loadRowSound func(int) uint16
	audio        func(uint32, O, int32, uint32)
}

// audEventDrop4EE2F0 preserves GAME.EXE 004EE2F0. Each pointer argument is
// rejected before the next is loaded. Every nonzero DefaultDrop result is
// cached and returned exactly. The first row type is read before the item's
// TypeInd, which is then loaded only once. A matching row emits its sound even
// when that sound is zero; the 0xffff type, not the sound, is the sentinel.
func audEventDrop4EE2F0[O, P comparable](hooks audEventDropHooks4EE2F0[O, P]) int32 {
	var nilObject O
	owner := hooks.loadOwnerArg()
	if owner == nilObject {
		return 0
	}

	item := hooks.loadItemArg()
	if item == nilObject {
		return 0
	}

	var nilPoint P
	point := hooks.loadPointArg()
	if point == nilPoint {
		return 0
	}

	result := hooks.defaultDrop(owner, item, point)
	if result == 0 {
		return result
	}

	row := 0
	rowType := hooks.loadRowType(row)
	if rowType == audEventDropSentinel4EE2F0 {
		return result
	}
	typeInd := hooks.loadTypeInd(item)
	for {
		if rowType == typeInd {
			sound := hooks.loadRowSound(row)
			hooks.audio(uint32(sound), owner, 0, 0)
			return result
		}
		row++
		rowType = hooks.loadRowType(row)
		if rowType == audEventDropSentinel4EE2F0 {
			return result
		}
	}
}

// audEventDropParseHooks536AC0 exposes the observable order of the parser
// bound to AudEventDrop in GAME.EXE's registration record. T models the token
// pointer so a nil token and a non-nil token whose first byte is zero remain
// distinguishable in the contract tests.
type audEventDropParseHooks536AC0[T comparable] struct {
	loadInit      func() uint32
	storeRowType  func(int, uint16)
	storeRowSound func(int, uint16)
	storeInit     func(uint32)

	loadRowType   func(int) uint16
	nextToken     func() T
	loadTokenByte func(T) byte
	resolveSound  func(T) uint16
	loadTypeInd   func() uint16
}

// audEventDropParse536AC0 preserves the parser's table mutations. The parser's
// machine return value is ignored by its only registration caller, so this
// helper reports only whether a row was published. Publication writes sound
// first and type last; the sentinel therefore remains visible until the row is
// complete.
func audEventDropParse536AC0[T comparable](hooks audEventDropParseHooks536AC0[T]) bool {
	if hooks.loadInit() == 0 {
		for row := 0; row < audEventDropRowStorage4EE2F0; row++ {
			hooks.storeRowType(row, audEventDropSentinel4EE2F0)
			hooks.storeRowSound(row, 0)
		}
		hooks.storeInit(1)
	}

	row := 0
	for {
		if hooks.loadRowType(row) == audEventDropSentinel4EE2F0 {
			break
		}
		row++
		if row >= audEventDropRowCapacity4EE2F0 {
			return false
		}
	}

	token := hooks.nextToken()
	var nilToken T
	if token == nilToken {
		return false
	}
	if hooks.loadTokenByte(token) == 0 {
		return false
	}
	sound := hooks.resolveSound(token)
	if sound == 0 {
		return false
	}
	typeInd := hooks.loadTypeInd()
	hooks.storeRowSound(row, sound)
	hooks.storeRowType(row, typeInd)
	return true
}
