package server

const (
	breakInitStatusMask4F0570 = uint8(0x0e)
	breakInitStatusBit4F0570  = uint32(2)
	breakInitTypeField536910  = uint32(2)
)

type breakInitHooks4F0570[O any] struct {
	loadStatusLow func(O) uint8
	setXStatus    func(O, uint32)
}

// breakInit4F0570 preserves the observable order of GAME.EXE 004F0570.
// The original tests only the low byte at object offset 20. It sets xstatus
// bit 2 when none of bits 1..3 are present and otherwise returns immediately.
// Init callbacks are called through a void slot, so branch-dependent residual
// EAX values are not part of the public contract. There is no nil guard.
func breakInit4F0570[O any](unit O, hooks breakInitHooks4F0570[O]) {
	status := hooks.loadStatusLow(unit)
	if status&breakInitStatusMask4F0570 != 0 {
		return
	}
	hooks.setXStatus(unit, breakInitStatusBit4F0570)
}

// breakInitParse536910 models the dedicated parser stored in the BreakInit
// registration row. The original ignores its definition-text argument, writes
// full dword 2 to ObjectType offset 36 through its second argument, and returns
// canonical one. In particular, zero init-data size does not suppress it.
func breakInitParse536910[O any](objectType O, storeField9 func(O, uint32)) int32 {
	storeField9(objectType, breakInitTypeField536910)
	return 1
}
