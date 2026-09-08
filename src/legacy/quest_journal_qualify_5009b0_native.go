package legacy

import "github.com/opennox/opennox/v1/common/memmap"

const (
	questJournalMemoryBase5009B0      = uintptr(0x5D4594)
	questJournalMapNameOffset5009B0   = uintptr(1570008)
	questJournalScratchOffset5009B0   = uintptr(1570140)
	questJournalScratchCapacity5009B0 = 132
)

func questJournalCurrentMapName5009B0() string {
	return memmap.String(questJournalMemoryBase5009B0, questJournalMapNameOffset5009B0)
}

func questJournalScratch5009B0() []byte {
	return memmap.Slice(questJournalMemoryBase5009B0, questJournalScratchOffset5009B0)[:questJournalScratchCapacity5009B0]
}

// questJournalQualifyNative5009B0 materializes GAME.EXE 005009B0's result in
// the original 132-byte shared scratch slot. The original has no bound check;
// the native port preserves every in-domain write and return value while
// leaving the slot unchanged when the terminating NUL would not fit.
func questJournalQualifyNative5009B0(name string) uint32 {
	qualified, result := questJournalQualifyString5009B0(name, questJournalCurrentMapName5009B0)
	if len(qualified) >= questJournalScratchCapacity5009B0 {
		return result
	}
	scratch := questJournalScratch5009B0()
	copy(scratch, qualified)
	scratch[len(qualified)] = 0
	return result
}
