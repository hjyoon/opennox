package server

const (
	spellManaChargePlayerClass4FCF90 = uint8(0x04)
	spellManaChargeSummonFirst4FCF90 = int32(75)
	spellManaChargeSummonLast4FCF90  = int32(114)
)

type spellManaChargeHooks4FCF90[Unit, Update any] struct {
	loadUnitArg        func() Unit
	loadClassLow       func(Unit) uint8
	loadUpdateData     func(Unit) Update
	loadSpellArg       func() int32
	loadGodMode        func() bool
	loadCostTypeArg    func() int32
	summonCost         func(int32, Unit) int32
	spellManaCost      func(int32, int32) int32
	loadCurrentMana    func(Update) uint16
	subtractMana       func(Unit, int32)
	storeRechargeCost  func(Update, uint16)
	loadTickRate       func() uint32
	storeRechargeFrame func(Update, uint16)
}

// spellManaCharge4FCF90 preserves GAME.EXE 004FCF90's observation and
// callback order while allowing Unit and Update to retain native pointer
// width. The class byte is read before the update pointer is cached, and the
// cached pointer remains authoritative across every later callback. There are
// deliberately no nil guards: the original faults at the first dereference
// reached by the selected path.
//
// Spell IDs and costs remain whole signed 32-bit values. Summon IDs 75..114
// do not read the cost-type argument. On insufficient mana, the ordinary type
// 1 cost is recomputed even for summons, and only its low word is stored before
// the live tick rate is read and truncated to a word.
func spellManaCharge4FCF90[Unit, Update any](
	hooks spellManaChargeHooks4FCF90[Unit, Update],
) int32 {
	unit := hooks.loadUnitArg()
	classLow := hooks.loadClassLow(unit)
	update := hooks.loadUpdateData(unit)
	if classLow&spellManaChargePlayerClass4FCF90 == 0 {
		return -1
	}

	spellID := hooks.loadSpellArg()
	if spellID == 0 {
		return -1
	}
	if hooks.loadGodMode() {
		return 0
	}

	var cost int32
	if spellID >= spellManaChargeSummonFirst4FCF90 && spellID <= spellManaChargeSummonLast4FCF90 {
		cost = hooks.summonCost(spellID, unit)
	} else {
		costType := hooks.loadCostTypeArg()
		cost = hooks.spellManaCost(spellID, costType)
	}

	if int32(hooks.loadCurrentMana(update)) >= cost {
		hooks.subtractMana(unit, cost)
		return cost
	}

	rechargeCost := hooks.spellManaCost(spellID, 1)
	hooks.storeRechargeCost(update, uint16(rechargeCost))
	frames := hooks.loadTickRate()
	hooks.storeRechargeFrame(update, uint16(frames))
	return -1
}
