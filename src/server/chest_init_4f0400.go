package server

const (
	chestInitStatusMask4F0400 = uint8(0x0e)
	chestInitStatusBit4F0400  = uint32(2)
)

type chestInitHooks4F0400[O any] struct {
	loadStatusLow func(O) uint8
	setXStatus    func(O, uint32)
}

// chestInit4F0400 preserves the observable order of GAME.EXE 004F0400.
// The original tests only the low byte at object offset 20. It sets xstatus
// bit 2 when none of bits 1..3 are present and otherwise returns immediately.
// Init callbacks are invoked through a void slot, so the original residual
// EAX values are not part of this semantic contract. There is no nil guard.
func chestInit4F0400[O any](unit O, hooks chestInitHooks4F0400[O]) {
	status := hooks.loadStatusLow(unit)
	if status&chestInitStatusMask4F0400 != 0 {
		return
	}
	hooks.setXStatus(unit, chestInitStatusBit4F0400)
}
