package server

const (
	spellBuffOffNoAudio16_4FF5B0 = int32(16)
	spellBuffOffNoAudio30_4FF5B0 = int32(30)
	spellBuffOffAudioField4FF5B0 = int32(2)
)

// SpellBuffOffHooks4FF5B0 exposes every observable argument, object-field,
// and callback access in GAME.EXE 004FF5B0. Object tokens retain their full
// native identity while the enchant argument keeps its original signed dword
// width.
type SpellBuffOffHooks4FF5B0[Object any] struct {
	LoadBuffArg   func() int32
	LoadUnitArg   func() Object
	LoadBuffs     func(Object) uint32
	SetBuffFlags  func(Object, uint32)
	StoreDuration func(Object, int32, uint16)
	StorePower    func(Object, int32, uint8)
	EnchantSpell  func(int32) int32
	SpellAudio    func(int32, int32) int32
	Audio         func(int32, Object, int32, int32)
}

// SpellBuffOff4FF5B0 preserves GAME.EXE 004FF5B0's exact argument and field
// access order. The signed dword enchant argument and native-width unit token
// are loaded before the unit's buff dword. Like x86 SHL, only the low five
// bits of the enchant argument select the flag bit. If that bit is absent the
// selected mask is returned unchanged and no mutation callback runs.
//
// An active enchant clears flags from the cached pre-callback buff dword,
// then clears its duration word before its power byte. Enchantments 16 and 30
// stop after those stores. Every other enchant resolves audio field two and
// emits it for the cached unit. The original has no nil, bounds, or callback
// guard; bindings retain the same fail-fast boundary.
func SpellBuffOff4FF5B0[Object any](h SpellBuffOffHooks4FF5B0[Object]) int32 {
	buff := h.LoadBuffArg()
	unit := h.LoadUnitArg()
	mask := uint32(1) << (uint32(buff) & 31)
	buffs := h.LoadBuffs(unit)
	if buffs&mask == 0 {
		return int32(mask)
	}

	h.SetBuffFlags(unit, buffs&^mask)
	h.StoreDuration(unit, buff, 0)
	h.StorePower(unit, buff, 0)
	if buff == spellBuffOffNoAudio16_4FF5B0 || buff == spellBuffOffNoAudio30_4FF5B0 {
		return 0
	}
	spellID := h.EnchantSpell(buff)
	audioID := h.SpellAudio(spellID, spellBuffOffAudioField4FF5B0)
	h.Audio(audioID, unit, 0, 0)
	return 0
}
