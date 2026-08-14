package server

const (
	flagPickupBuffSlots4EA7A0   = uint32(32)
	flagPickupRemovedFlag4EA7A0 = uint32(0x00080000)
)

type flagPickupBuffPurgeHooks4EA7A0[O any] struct {
	hasBuff       func(O, uint32) int32
	enchantSpell  func(uint32) int32
	spellHasFlags func(int32, uint32) int32
	buffOff       func(O, uint32) int32
}

// flagPickupBuffPurge4EA7A0 preserves GAME.EXE 004EA7A0. It visits all 32
// enchant slots and removes active enchants whose spells carry flag 0x80000.
// The original calls the enchant-to-spell lookup twice: the first result is a
// nonzero gate and the second, freshly loaded result is passed to HasFlags.
// EAX from the last operation in slot 31 is returned unchanged.
func flagPickupBuffPurge4EA7A0[O any](
	obj O,
	hooks flagPickupBuffPurgeHooks4EA7A0[O],
) int32 {
	var result int32
	for enchant := uint32(0); enchant < flagPickupBuffSlots4EA7A0; enchant++ {
		result = hooks.hasBuff(obj, enchant)
		if result == 0 {
			continue
		}
		result = hooks.enchantSpell(enchant)
		if result == 0 {
			continue
		}
		spell := hooks.enchantSpell(enchant)
		result = hooks.spellHasFlags(spell, flagPickupRemovedFlag4EA7A0)
		if result == 0 {
			continue
		}
		result = hooks.buffOff(obj, enchant)
	}
	return result
}
