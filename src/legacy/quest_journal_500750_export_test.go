package legacy

import (
	"math"
	"runtime"
	"testing"
	"unsafe"
)

func TestQuestJournalGetIntExport500750PreservesNativeLinksAndDword(t *testing.T) {
	resetQuestJournal500540(t)
	if got := questJournalGetIntExportCall500750("War01a:Missing"); got != 0 {
		t.Fatalf("missing export value = %d, want 0", got)
	}

	entry := questJournalSet500540("War01a:Value", 1, uint32(0x80000000))
	if entry == nil {
		t.Fatal("cannot allocate quest-journal entry")
	}
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(entry)) <= math.MaxUint32 {
		t.Fatalf("entry pointer = %p, want native address above 4 GiB", entry)
	}
	for _, value := range []int32{math.MinInt32, -1, 0, 1, math.MaxInt32} {
		if got := questJournalSet500540("WAR01A:VALUE", 0, uint32(value)); got != entry {
			t.Fatalf("update for value %d returned %p, want %p", value, got, entry)
		}
		if got := questJournalGetIntExportCall500750("war01A:value"); got != value {
			t.Fatalf("export value = %d, want %d", got, value)
		}
	}
	if uint32(entry.kind) != 1 {
		t.Fatalf("export getter changed or rejected Boolean entry kind %d", uint32(entry.kind))
	}
	runtime.KeepAlive(entry)
}
