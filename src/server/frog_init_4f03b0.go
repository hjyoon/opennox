package server

const (
	frogInitDelayMinimum4F03B0        = int32(55)
	frogInitDelayMaximum4F03B0        = int32(60)
	frogInitDelayRandomPath4F03B0     = `C:\NoxPost\src\Server\Object\init\Init.c`
	frogInitDelayRandomLine4F03B0     = int32(943)
	frogInitDirectionMinimum4F03B0    = int32(0)
	frogInitDirectionMaximum4F03B0    = int32(255)
	frogInitDirectionRandomPath4F03B0 = `C:\NoxPost\src\Server\Object\init\Init.c`
	frogInitDirectionRandomLine4F03B0 = int32(947)
	frogInitByte1Value4F03B0          = uint8(1)
	frogInitByte2Value4F03B0          = uint8(0)
)

type frogInitHooks4F03B0[O, D any] struct {
	loadUpdateData func(O) D
	randomInt      func(int32, int32, string, int32) int32
	storeDelay     func(D, uint8)
	storeByte1     func(D, uint8)
	storeByte2     func(D, uint8)
	storeDirection func(O, uint16)
}

// frogInit4F03B0 preserves the exact observable order of GAME.EXE 004F03B0.
// The update-data pointer and unit pointer are entry-cached. The first logic
// RNG result is narrowed through AL, the second through AX for Direction2,
// while the full second RNG result remains the function return. The original
// has no nil guards.
func frogInit4F03B0[O, D any](unit O, hooks frogInitHooks4F03B0[O, D]) int32 {
	update := hooks.loadUpdateData(unit)
	delay := hooks.randomInt(
		frogInitDelayMinimum4F03B0,
		frogInitDelayMaximum4F03B0,
		frogInitDelayRandomPath4F03B0,
		frogInitDelayRandomLine4F03B0,
	)
	hooks.storeDelay(update, uint8(delay))
	hooks.storeByte1(update, frogInitByte1Value4F03B0)
	hooks.storeByte2(update, frogInitByte2Value4F03B0)
	direction := hooks.randomInt(
		frogInitDirectionMinimum4F03B0,
		frogInitDirectionMaximum4F03B0,
		frogInitDirectionRandomPath4F03B0,
		frogInitDirectionRandomLine4F03B0,
	)
	hooks.storeDirection(unit, uint16(direction))
	return direction
}
