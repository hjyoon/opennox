package server

const (
	unitDamageClearGodModeFlag4EE5E0 = uint32(0x20)
	unitDamageClearMonsterBit4EE5E0  = uint8(0x02)
	unitDamageClearPlayerBit4EE5E0   = uint8(0x04)
	unitDamageClearDeadFlag4EE5E0    = uint32(0x00008000)
	unitDamageClearHarpoonBuff4EE5E0 = int32(16)
)

type unitDamageClearHooks4EE5E0[O, H, U, P, D any] struct {
	loadUnitArg       func() O
	loadHealth        func(O) H
	loadMaximum       func(H) uint16
	engineFlag        func(uint32) int32
	loadClassLow      func(O) uint8
	loadUpdateData    func(O) U
	loadPlayer        func(U) P
	loadPlayerClass   func(P) uint8
	loadHarpoonTarget func(U) O
	breakHarpoon      func(O)
	loadDamageArg     func() int32
	loadCurrent       func(H) uint16
	setHP             func(O, uint16)
	loadFlags         func(O) uint32
	storeFlags        func(O, uint32)
	buffOff           func(O, int32)
	isZombie          func(O) int32
	soloReward        func(O)
	monsterDie        func(O)
	loadDeath         func(O) D
	callDeath         func(D, O)
	delayedDelete     func(O)
	informOwnerHP     func(O)
}

// unitDamageClear4EE5E0 preserves GAME.EXE 004EE5E0. Reads remain separated
// wherever an intervening callback can mutate Object state: health is reloaded
// after the optional harpoon break, flags after SetHP, class before death and
// again before the final owner-HP report, and Death only on the live non-Monster
// branch. Nil UpdateData is guarded, while its Player and the reloaded health
// record deliberately retain the original unguarded dereferences.
func unitDamageClear4EE5E0[O, H, U, P, D comparable](
	hooks unitDamageClearHooks4EE5E0[O, H, U, P, D],
) {
	unit := hooks.loadUnitArg()
	var nilObject O
	if unit == nilObject {
		return
	}

	initialHealth := hooks.loadHealth(unit)
	var nilHealth H
	if initialHealth == nilHealth || hooks.loadMaximum(initialHealth) == 0 {
		return
	}

	if hooks.engineFlag(unitDamageClearGodModeFlag4EE5E0) != 0 &&
		hooks.loadClassLow(unit)&unitDamageClearPlayerBit4EE5E0 != 0 {
		return
	}

	if hooks.loadClassLow(unit)&unitDamageClearPlayerBit4EE5E0 != 0 {
		update := hooks.loadUpdateData(unit)
		var nilUpdate U
		if update != nilUpdate {
			player := hooks.loadPlayer(update)
			if hooks.loadPlayerClass(player) == 0 && hooks.loadHarpoonTarget(update) != nilObject {
				hooks.breakHarpoon(unit)
			}
		}
	}

	// The entry health record only gates the function. Cur is read from the
	// live record after HarpoonBreak, and damage is not observed before then.
	health := hooks.loadHealth(unit)
	damage := hooks.loadDamageArg()
	current := hooks.loadCurrent(health)
	if int32(current) > damage {
		// The setter consumes AX, so negative damage and large differences wrap
		// modulo 2^16 exactly as the IA-32 subtraction does.
		hooks.setHP(unit, uint16(int64(current)-int64(damage)))
	} else {
		hooks.setHP(unit, 0)
		flags := hooks.loadFlags(unit)
		if flags&unitDamageClearDeadFlag4EE5E0 == 0 {
			hooks.storeFlags(unit, flags|unitDamageClearDeadFlag4EE5E0)
			hooks.buffOff(unit, unitDamageClearHarpoonBuff4EE5E0)
			if hooks.isZombie(unit) == 0 {
				hooks.soloReward(unit)
			}

			if hooks.loadClassLow(unit)&unitDamageClearMonsterBit4EE5E0 != 0 {
				hooks.monsterDie(unit)
			} else {
				death := hooks.loadDeath(unit)
				var nilDeath D
				if death != nilDeath {
					hooks.callDeath(death, unit)
				} else {
					hooks.delayedDelete(unit)
				}
			}
		}
	}

	if hooks.loadClassLow(unit)&unitDamageClearMonsterBit4EE5E0 != 0 {
		hooks.informOwnerHP(unit)
	}
}
