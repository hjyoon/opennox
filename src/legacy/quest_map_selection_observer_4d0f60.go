package legacy

import (
	"unsafe"

	"github.com/opennox/opennox/v1/common/memmap"
)

// QuestMapHistory4D0F60 is a value copy, not a view of the PE32 history table.
type QuestMapHistory4D0F60 struct {
	Group    uint32
	Name     [20]byte
	Uses     uint32
	LastUsed uint32
}

type QuestMapSelectionState4D0F60 struct {
	Count, Clock, LastIndex, Frame uint32
	LogicIndex, OtherIndex         int
	Entries                        []QuestMapHistory4D0F60
}

type QuestMapSelection4D0F60 struct {
	Before, After QuestMapSelectionState4D0F60
	ResultPointer uintptr
	ResultIndex   int32 // -2: null; -3: not a name slot; -1: original fallback
}

func questMapSelectionSnapshot4D0F60() QuestMapSelectionState4D0F60 {
	count, clock := questMapFileGlobals4D0F60()
	s := GetServer().S()
	state := QuestMapSelectionState4D0F60{
		Count: *count, Clock: *clock, LastIndex: memmap.Uint32(0x587000, 191880),
		Frame: s.Frame(), LogicIndex: s.Rand.Logic.Index(), OtherIndex: s.Rand.Other.Index(),
	}
	// The original table has 128 records. An invalid count remains observable,
	// but must not make diagnostic reads run beyond that allocation.
	if state.Count > 128 {
		return state
	}
	for index := uint32(0); index < state.Count; index++ {
		entry := *questMapEntryNative4D0F60(int32(index))
		state.Entries = append(state.Entries, QuestMapHistory4D0F60{
			Group: entry.Group, Name: entry.Name, Uses: entry.Uses, LastUsed: entry.LastUsed,
		})
	}
	return state
}

func questMapSelectionIndex4D0F60(result unsafe.Pointer) int32 {
	if result == nil {
		return -2
	}
	// Compare full native pointers, including the original ordinal -1 fallback.
	// Never dereference an arbitrary returned pointer for diagnostic purposes.
	for index := int32(-1); index < 128; index++ {
		if result == questMapNameNative4D0F60(index) {
			return index
		}
	}
	return -3
}

// ObserveQuestMapSelection4D0F60 installs a main-game-thread observer for E2E
// diagnostics. It calls the existing selector exactly once, preserves its
// result and never advances either RNG or edits history. The callback receives
// detached scalar records. Stop observers in reverse installation order.
// No observer is installed by ordinary game startup.
func ObserveQuestMapSelection4D0F60(observe func(QuestMapSelection4D0F60)) func() {
	if observe == nil {
		return func() {}
	}
	previous := questMapFileCall4D0F60
	questMapFileCall4D0F60 = func() unsafe.Pointer {
		before := questMapSelectionSnapshot4D0F60()
		result := previous()
		after := questMapSelectionSnapshot4D0F60()
		observe(QuestMapSelection4D0F60{
			Before: before, After: after,
			ResultPointer: uintptr(result), ResultIndex: questMapSelectionIndex4D0F60(result),
		})
		return result
	}
	stopped := false
	return func() {
		if !stopped {
			questMapFileCall4D0F60 = previous
			stopped = true
		}
	}
}
