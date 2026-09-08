package legacy

import (
	"math"
	"runtime"
	"testing"
	"unsafe"
)

func TestQuestJournalGetFloatExport500770PreservesNativeLinksAndX87Result(t *testing.T) {
	resetQuestJournal500540(t)
	if got := math.Float64bits(questJournalGetFloatExportCall500770("War01a:Missing")); got != 0 {
		t.Fatalf("missing export bits = %#016x, want 0", got)
	}

	entry := questJournalSet500540("War01a:Value", 1, 0)
	if entry == nil {
		t.Fatal("cannot allocate quest-journal entry")
	}
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(entry)) <= math.MaxUint32 {
		t.Fatalf("entry pointer = %p, want native address above 4 GiB", entry)
	}
	tests := []struct {
		stored uint32
		want   uint32
	}{
		{0x00000000, 0x00000000},
		{0x80000000, 0x80000000},
		{0x00000001, 0x00000001},
		{0x7f7fffff, 0x7f7fffff},
		{0x7f800000, 0x7f800000},
		{0xff800000, 0xff800000},
		{0x7fc12345, 0x7fc12345},
		{0xffc12345, 0xffc12345},
		{0x7f812345, 0x7fc12345},
		{0xff812345, 0xffc12345},
	}
	for _, tc := range tests {
		if got := questJournalSet500540("WAR01A:VALUE", 0, tc.stored); got != entry {
			t.Fatalf("update for bits %#08x returned %p, want %p", tc.stored, got, entry)
		}
		got := questJournalGetFloatExportCall500770("war01A:value")
		want := float64(math.Float32frombits(tc.want))
		if gotBits, wantBits := math.Float64bits(got), math.Float64bits(want); gotBits != wantBits {
			t.Fatalf("stored bits %#08x widened to %#016x, want %#016x", tc.stored, gotBits, wantBits)
		}
	}
	if uint32(entry.kind) != 1 {
		t.Fatalf("export getter changed or rejected Boolean entry kind %d", uint32(entry.kind))
	}
	runtime.KeepAlive(entry)
}
