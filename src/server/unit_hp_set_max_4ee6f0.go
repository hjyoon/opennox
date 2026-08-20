package server

const unitHPSetOnMaxMonsterBit4EE6F0 = uint8(0x02)

type unitHPSetOnMaxHooks4EE6F0[O, H any] struct {
	loadUnitArg  func() O
	loadHealth   func(O) H
	loadMaximum  func(H) uint16
	setHP        func(O, uint16)
	loadCurrent  func(H) uint16
	storeField2  func(H, uint16)
	loadClassLow func(O) uint8
	informOwner  func(O)
}

// unitHPSetOnMax4EE6F0 preserves GAME.EXE 004EE6F0. The initial HealthData
// record is used only for the nil gate and maximum value. SetHP may mutate the
// object, so HealthData is reloaded before Cur is copied to Field2. That live
// record intentionally has no nil guard. The Monster class is read only after
// the copy, allowing either callback to affect the final owner notification.
func unitHPSetOnMax4EE6F0[O, H comparable](hooks unitHPSetOnMaxHooks4EE6F0[O, H]) {
	unit := hooks.loadUnitArg()
	var nilObject O
	if unit == nilObject {
		return
	}

	initialHealth := hooks.loadHealth(unit)
	var nilHealth H
	if initialHealth == nilHealth {
		return
	}

	maximum := hooks.loadMaximum(initialHealth)
	hooks.setHP(unit, maximum)

	liveHealth := hooks.loadHealth(unit)
	current := hooks.loadCurrent(liveHealth)
	hooks.storeField2(liveHealth, current)

	if hooks.loadClassLow(unit)&unitHPSetOnMaxMonsterBit4EE6F0 != 0 {
		hooks.informOwner(unit)
	}
}
