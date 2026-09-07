package server

const (
	buffApplyQuestFlag4FF380        = uint32(0x1000)
	buffApplyCoopFlag4FF380         = uint32(0x0800)
	buffApplyMonsterClass4FF380     = uint8(0x02)
	buffApplyFearSubclass4FF380     = uint32(0x1000)
	buffApplyRejectedFlags4FF380    = uint32(0x8022)
	buffApplyConfused4FF380         = int32(3)
	buffApplyAfraid4FF380           = int32(11)
	buffApplyAntiMagic4FF380        = int32(29)
	buffApplyInvisible4FF380        = int32(0)
	buffApplyHecubahAudio4FF380     = int32(582)
	buffApplyNecromancerAudio4FF380 = int32(595)
	buffApplyOnAudioSelector4FF380  = int32(1)
)

// BuffApplyHooks4FF380 exposes every observable argument, cache, object-field,
// and callback access in GAME.EXE 004FF380. Comparable object tokens retain
// canonical nil and full native pointer identities without inheriting PE32's
// pointer width. Integer fields retain the exact widths loaded or stored by
// the original i386 instructions.
type BuffApplyHooks4FF380[Object comparable] struct {
	LoadHecubahTypeID      func() uint32
	StoreHecubahTypeID     func(uint32)
	LoadNecromancerTypeID  func() uint32
	StoreNecromancerTypeID func(uint32)
	LookupTypeID           func(string) uint32
	LoadBuffArg            func() int32
	LoadUnitArg            func() Object
	LoadDurationArg        func() int16
	LoadPowerArg           func() int8
	LoadTypeIndex          func(Object) uint16
	LoadClassLow           func(Object) uint8
	LoadSubclass           func(Object) uint32
	LoadObjectFlags        func(Object) uint32
	GameFlag               func(uint32) int32
	Audio                  func(int32, Object, int32, int32)
	TestBuff               func(Object, int32) int32
	LoadBuffTimer          func(Object, int32) int32
	BuffOff                func(Object, int32) int32
	StoreDuration          func(Object, int32, uint16)
	StorePower             func(Object, int32, uint8)
	LoadBuffs              func(Object) uint32
	SetBuffFlags           func(Object, uint32)
	EnchantSpell           func(int32) int32
	SpellAudio             func(int32, int32) int32
}

// BuffApply4FF380 preserves GAME.EXE 004FF380's exact gates, access order,
// integer widths, and callback sequence. The Hecubah cache is read before the
// buff argument; x86 SHL masks that signed dword to five bits before type-cache
// initialization and before the unit null gate. A zero Hecubah cache triggers
// both name lookups on every call, with Hecubah stored before Necromancer is
// looked up.
//
// Hecubah rejects AntiMagic unconditionally. In Quest mode Hecubah and
// Necromancer reject Confused with their fixed audio cues. Outside Coop, a
// monster carrying subclass bit 0x1000 rejects Afraid whether or not it is one
// of those named types; the named types additionally emit their cue. Normal
// application then rejects object flags 0x8022 and rejects only a buff that is
// already active with a zero timer. A nonzero buff removes Invisibility before
// the duration and power arguments are loaded. Those arguments are stored as
// exact word and byte values, the live buff dword is loaded after both stores,
// and the on-sound is resolved and emitted after SetBuffFlags. No bounds,
// callback, or nil guard absent from the original is added.
func BuffApply4FF380[Object comparable](h BuffApplyHooks4FF380[Object]) {
	hecubah := h.LoadHecubahTypeID()
	buff := h.LoadBuffArg()
	mask := uint32(1) << (uint32(buff) & 31)
	if hecubah == 0 {
		hecubah = h.LookupTypeID("Hecubah")
		h.StoreHecubahTypeID(hecubah)
		necromancer := h.LookupTypeID("Necromancer")
		h.StoreNecromancerTypeID(necromancer)
	}

	unit := h.LoadUnitArg()
	var nilObject Object
	if unit == nilObject {
		return
	}

	hecubah = h.LoadHecubahTypeID()
	typeIndex := uint32(h.LoadTypeIndex(unit))
	if typeIndex == hecubah && buff == buffApplyAntiMagic4FF380 {
		return
	}

	if h.GameFlag(buffApplyQuestFlag4FF380) != 0 {
		hecubah = h.LoadHecubahTypeID()
		typeIndex = uint32(h.LoadTypeIndex(unit))
		if typeIndex == hecubah && buff == buffApplyConfused4FF380 {
			h.Audio(buffApplyHecubahAudio4FF380, unit, 0, 0)
			return
		}
	}

	if h.GameFlag(buffApplyQuestFlag4FF380) != 0 {
		necromancer := h.LoadNecromancerTypeID()
		typeIndex = uint32(h.LoadTypeIndex(unit))
		if typeIndex == necromancer && buff == buffApplyConfused4FF380 {
			h.Audio(buffApplyNecromancerAudio4FF380, unit, 0, 0)
			return
		}
	}

	if h.LoadClassLow(unit)&buffApplyMonsterClass4FF380 != 0 {
		if h.LoadSubclass(unit)&buffApplyFearSubclass4FF380 != 0 &&
			buff == buffApplyAfraid4FF380 &&
			h.GameFlag(buffApplyCoopFlag4FF380) == 0 {
			typeIndex = uint32(h.LoadTypeIndex(unit))
			hecubah = h.LoadHecubahTypeID()
			if typeIndex == hecubah {
				h.Audio(buffApplyHecubahAudio4FF380, unit, 0, 0)
				return
			}
			if typeIndex == h.LoadNecromancerTypeID() {
				h.Audio(buffApplyNecromancerAudio4FF380, unit, 0, 0)
			}
			return
		}
	}

	if h.LoadObjectFlags(unit)&buffApplyRejectedFlags4FF380 != 0 {
		return
	}
	if h.TestBuff(unit, buff) != 0 && h.LoadBuffTimer(unit, buff) == 0 {
		return
	}
	if buff != buffApplyInvisible4FF380 {
		_ = h.BuffOff(unit, buffApplyInvisible4FF380)
	}

	duration := uint16(h.LoadDurationArg())
	power := uint8(h.LoadPowerArg())
	h.StoreDuration(unit, buff, duration)
	h.StorePower(unit, buff, power)
	h.SetBuffFlags(unit, h.LoadBuffs(unit)|mask)
	spellID := h.EnchantSpell(buff)
	audioID := h.SpellAudio(spellID, buffApplyOnAudioSelector4FF380)
	h.Audio(audioID, unit, 0, 0)
}
