package server

const (
	unitBuffUpdateFlagSlot4FF620     = int32(7)
	unitBuffUpdateSlowSlot4FF620     = int32(9)
	unitBuffUpdateDeathSlot4FF620    = int32(16)
	unitBuffUpdateFlagMask4FF620     = uint32(0x40)
	unitBuffUpdateDeathDamage4FF620  = int32(9999999)
	unitBuffUpdateDeathType4FF620    = uint32(13)
	unitBuffUpdateWarningAudio4FF620 = int32(26)
	unitBuffUpdateExpireAudio4FF620  = int32(779)
	unitBuffUpdateSpeedFactor4FF620  = float32(1.25)
	unitBuffUpdateObjectClass4FF620  = uint8(4)
	unitBuffUpdateFirstSlot4FF620    = int32(0)
	unitBuffUpdateSlotCount4FF620    = int32(32)
)

// UnitBuffUpdateHooks4FF620 exposes every observable object-field and callback
// access in GAME.EXE 004FF620. Object tokens retain their full native identity;
// buff masks, damage metadata, and callback scalars retain their original
// fixed widths.
type UnitBuffUpdateHooks4FF620[Object any] struct {
	LoadUnitArg        func() Object
	LoadBuffs          func(Object) uint32
	LoadFPS            func() uint32
	LoadDuration       func(Object, int32) uint16
	StoreDuration      func(Object, int32, uint16)
	Audio              func(int32, Object, int32, int32)
	LoadFlags          func(Object) uint32
	StoreFlags         func(Object, uint32)
	StoreObj130        func(Object, Object)
	StoreDamageType    func(Object, uint32)
	DamageClear        func(Object, int32)
	LoadClassLow       func(Object) uint8
	IncrementElimDeath func(Object)
	ReportLesson       func(Object)
	BuffOff            func(Object, int32) int32
	StorePower         func(Object, int32, uint8)
	TestBuff           func(Object, int32) int32
	LoadSpeed          func(Object) float32
	StoreSpeed         func(Object, float32)
}

// UnitBuffUpdate4FF620 preserves GAME.EXE 004FF620's exact live-read and
// callback order. The initial buff dword gates the whole routine, but every
// loop iteration reloads that dword so callbacks can affect later slots. Slot
// 16 reloads its duration after the periodic warning callback; the FPS divisor
// is loaded only once and a zero divisor deliberately faults.
//
// An expiring slot stores its decremented duration before any special action.
// Slot 7 clears object flag 0x40. Slot 16 clears native-width Obj130, stores
// damage type 13, applies 9999999 damage, emits sound 779, then reloads the low
// class byte before optional elimination and lesson callbacks. BuffOff runs
// before the original redundant power-byte clear. Finally, buff 9 is tested
// even if callbacks cleared the live mask, and a present buff scales the live
// speed value by 1.25. The original has no nil or callback guards.
func UnitBuffUpdate4FF620[Object any](h UnitBuffUpdateHooks4FF620[Object]) {
	unit := h.LoadUnitArg()
	if h.LoadBuffs(unit) == 0 {
		return
	}

	for buff := unitBuffUpdateFirstSlot4FF620; buff < unitBuffUpdateSlotCount4FF620; buff++ {
		mask := uint32(1) << uint32(buff)
		if h.LoadBuffs(unit)&mask == 0 {
			continue
		}

		if buff == unitBuffUpdateDeathSlot4FF620 {
			fps := h.LoadFPS()
			duration := h.LoadDuration(unit, buff)
			if uint32(duration)%fps == fps-1 {
				h.Audio(unitBuffUpdateWarningAudio4FF620, unit, 0, 0)
			}
		}

		duration := h.LoadDuration(unit, buff)
		if duration == 0 {
			continue
		}
		duration--
		h.StoreDuration(unit, buff, duration)
		if duration != 0 {
			continue
		}

		switch buff {
		case unitBuffUpdateFlagSlot4FF620:
			flags := h.LoadFlags(unit)
			h.StoreFlags(unit, flags&^unitBuffUpdateFlagMask4FF620)
		case unitBuffUpdateDeathSlot4FF620:
			var zero Object
			h.StoreObj130(unit, zero)
			h.StoreDamageType(unit, unitBuffUpdateDeathType4FF620)
			h.DamageClear(unit, unitBuffUpdateDeathDamage4FF620)
			h.Audio(unitBuffUpdateExpireAudio4FF620, unit, 0, 0)
			if h.LoadClassLow(unit)&unitBuffUpdateObjectClass4FF620 != 0 {
				h.IncrementElimDeath(unit)
				h.ReportLesson(unit)
			}
		}

		_ = h.BuffOff(unit, buff)
		h.StorePower(unit, buff, 0)
	}

	if h.TestBuff(unit, unitBuffUpdateSlowSlot4FF620) != 0 {
		speed := h.LoadSpeed(unit)
		h.StoreSpeed(unit, speed*unitBuffUpdateSpeedFactor4FF620)
	}
}
