package server

const (
	audEventPickupRowCapacity4F3D50 = 50
	audEventPickupRowStorage4F3D50  = audEventPickupRowCapacity4F3D50 + 1
	audEventPickupSentinel4F3D50    = uint16(0xffff)
)

type audEventPickupSoundRow4F3D50 struct {
	typeInd uint16
	sound   uint16
}

// audEventPickupHooks4F3D50 exposes GAME.EXE 004F3D50's exact argument,
// DefaultPickup, ordered-table, object-field, and audio access order.
type audEventPickupHooks4F3D50[O comparable] struct {
	loadOwnerArg func() O
	loadItemArg  func() O
	loadArg4     func() int32
	loadArg3     func() int32

	defaultPickup func(O, O, int32, int32) int32
	loadRowType   func(int) uint16
	loadTypeInd   func(O) uint16
	loadRowSound  func(int) uint16
	audio         func(uint32, O, int32, uint32)
}

// audEventPickup4F3D50 preserves GAME.EXE 004F3D50. Each object argument is
// rejected before the next argument is read. The fourth scalar is loaded
// before the third, then both are forwarded to DefaultPickup. Every nonzero
// result is cached and returned exactly. The first row type is read before the
// item's TypeInd, which is then loaded only once. A matching row emits its
// sound even when that sound is zero; the 0xffff type is the sentinel.
func audEventPickup4F3D50[O comparable](hooks audEventPickupHooks4F3D50[O]) int32 {
	var nilObject O
	owner := hooks.loadOwnerArg()
	if owner == nilObject {
		return 0
	}

	item := hooks.loadItemArg()
	if item == nilObject {
		return 0
	}

	arg4 := hooks.loadArg4()
	arg3 := hooks.loadArg3()
	result := hooks.defaultPickup(owner, item, arg3, arg4)
	if result == 0 {
		return result
	}

	row := 0
	rowType := hooks.loadRowType(row)
	if rowType == audEventPickupSentinel4F3D50 {
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
		if rowType == audEventPickupSentinel4F3D50 {
			return result
		}
	}
}

// audEventPickupParseHooks5367B0 exposes the observable order of the parser
// bound to AudEventPickup in GAME.EXE's registration record. T models the
// token pointer so nil and a non-nil token whose first byte is zero remain
// distinguishable in contract tests.
type audEventPickupParseHooks5367B0[T comparable] struct {
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

// audEventPickupParse5367B0 preserves the parser's table mutations. Its
// machine return value is ignored by the registration caller, so this helper
// reports only whether a row was published. Publication writes sound first and
// type last, leaving the sentinel visible until the row is complete.
func audEventPickupParse5367B0[T comparable](hooks audEventPickupParseHooks5367B0[T]) bool {
	if hooks.loadInit() == 0 {
		for row := 0; row < audEventPickupRowStorage4F3D50; row++ {
			hooks.storeRowType(row, audEventPickupSentinel4F3D50)
			hooks.storeRowSound(row, 0)
		}
		hooks.storeInit(1)
	}

	row := 0
	for {
		if hooks.loadRowType(row) == audEventPickupSentinel4F3D50 {
			break
		}
		row++
		if row >= audEventPickupRowCapacity4F3D50 {
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
