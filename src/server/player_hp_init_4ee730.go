package server

const (
	playerHPInitPlayerBit4EE730   = uint8(0x04)
	playerHPInitSampleCount4EE730 = 32
)

type playerHPInitHooks4EE730[O, H, U any] struct {
	loadUnitArg        func() O
	loadClassLow       func(O) uint8
	loadHealth         func(O) H
	loadUpdateData     func(O) U
	loadCurrent        func(H) uint16
	storeSample        func(U, int, uint16)
	storeCurrentSample func(U, uint16)
}

// playerHPInit4EE730 preserves GAME.EXE 004EE730. UpdateData is cached after
// the Player-class gate but before the initial HealthData nil gate. Once that
// gate succeeds, HealthData is reloaded independently for every one of the 32
// history samples and once more for the trailing current sample. None of those
// live records or the cached UpdateData pointer has a nil guard.
func playerHPInit4EE730[O, H, U comparable](hooks playerHPInitHooks4EE730[O, H, U]) {
	unit := hooks.loadUnitArg()
	var nilObject O
	if unit == nilObject {
		return
	}
	if hooks.loadClassLow(unit)&playerHPInitPlayerBit4EE730 == 0 {
		return
	}

	initialHealth := hooks.loadHealth(unit)
	update := hooks.loadUpdateData(unit)
	var nilHealth H
	if initialHealth == nilHealth {
		return
	}

	for i := 0; i < playerHPInitSampleCount4EE730; i++ {
		liveHealth := hooks.loadHealth(unit)
		current := hooks.loadCurrent(liveHealth)
		hooks.storeSample(update, i, current)
	}

	liveHealth := hooks.loadHealth(unit)
	current := hooks.loadCurrent(liveHealth)
	hooks.storeCurrentSample(update, current)
}
