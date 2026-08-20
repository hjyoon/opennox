package server

const (
	unitAdjustHPDisabledFlag4EE460  = uint32(0x04000000)
	unitAdjustHPMonsterClass4EE460  = uint8(0x02)
	mobInformOwnerPlayerClass4EE4C0 = uint8(0x04)
)

type unitAdjustHPHooks4EE460[O, H any] struct {
	gameFlag      func(uint32) int32
	loadUnitArg   func() O
	loadHealth    func(O) H
	loadCurrent   func(H) uint16
	loadMaximum   func(H) uint16
	loadDeltaArg  func() int32
	setHP         func(O, uint16)
	loadClassLow  func(O) uint8
	informOwnerHP func(O)
}

// unitAdjustHP4EE460 preserves GAME.EXE 004EE460. The game flag is observed
// before the unit argument. A zero flag result deliberately leaves the unit
// unguarded, so loadHealth must retain the original null-object fault.
func unitAdjustHP4EE460[O, H comparable](hooks unitAdjustHPHooks4EE460[O, H]) {
	if hooks.gameFlag(unitAdjustHPDisabledFlag4EE460) != 0 {
		return
	}

	unit := hooks.loadUnitArg()
	health := hooks.loadHealth(unit)
	var nilHealth H
	if health == nilHealth {
		return
	}

	current := hooks.loadCurrent(health)
	maximum := hooks.loadMaximum(health)
	if current >= maximum {
		return
	}

	// IA-32 ADD wraps in 32 bits. The following JGE is signed, but the HP
	// setter consumes only the low uint16 when the wrapped value is below Max.
	delta := hooks.loadDeltaArg()
	adjusted := int32(uint32(delta) + uint32(current))
	value := uint16(adjusted)
	if adjusted >= int32(maximum) {
		value = maximum
	}

	hooks.setHP(unit, value)
	if hooks.loadClassLow(unit)&unitAdjustHPMonsterClass4EE460 != 0 {
		hooks.informOwnerHP(unit)
	}
}

type mobInformOwnerHPHooks4EE4C0[O, U, P any] struct {
	loadObjectArg  func() O
	loadOwner      func(O) O
	loadClassLow   func(O) uint8
	loadUpdateData func(O) U
	loadPlayer     func(U) P
	loadPlayerInd  func(P) uint8
	reportHP       func(uint8, O)
}

// mobInformOwnerHP4EE4C0 preserves GAME.EXE 004EE4C0. Only object and owner
// are guarded. UpdateData and Player are intentionally dereferenced without
// null checks before the PlayerInd byte is reported with the original object.
func mobInformOwnerHP4EE4C0[O comparable, U, P any](hooks mobInformOwnerHPHooks4EE4C0[O, U, P]) {
	obj := hooks.loadObjectArg()
	var nilObject O
	if obj == nilObject {
		return
	}

	owner := hooks.loadOwner(obj)
	if owner == nilObject {
		return
	}
	if hooks.loadClassLow(owner)&mobInformOwnerPlayerClass4EE4C0 == 0 {
		return
	}

	update := hooks.loadUpdateData(owner)
	player := hooks.loadPlayer(update)
	playerInd := hooks.loadPlayerInd(player)
	hooks.reportHP(playerInd, obj)
}
