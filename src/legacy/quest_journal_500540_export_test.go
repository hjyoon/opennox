package legacy

import (
	"math"
	"runtime"
	"testing"
	"unsafe"
)

func TestQuestJournalSetExport500540PreservesNativePointerDwordAndResult(t *testing.T) {
	resetQuestJournal500540(t)

	if got := questJournalSetExportCall500540("War01a:First", math.MinInt32); got != nil {
		t.Fatalf("first insertion result = %p, want nil previous head", got)
	}
	first := questJournalHead500540
	if first == nil {
		t.Fatal("first export call did not allocate an entry")
	}
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(first)) <= math.MaxUint32 {
		t.Fatalf("first entry pointer = %p, want native address above 4 GiB", first)
	}
	if uint32(first.value) != uint32(0x80000000) {
		t.Fatalf("minimum-dword value = %#08x, want 0x80000000", uint32(first.value))
	}

	if got := questJournalSetExportCall500540("War01a:Second", math.MaxInt32); got != first {
		t.Fatalf("second insertion result = %p, want full previous-head identity %p", got, first)
	}
	second := questJournalHead500540
	if second == nil || second == first || second.next != first || first.prev != second {
		t.Fatal("second export call did not prepend the native-width list")
	}
	if uint32(second.value) != uint32(math.MaxInt32) {
		t.Fatalf("maximum-dword value = %#08x, want 0x7fffffff", uint32(second.value))
	}

	if got := questJournalSetExportCall500540("war01A:first", -1); got != first {
		t.Fatalf("existing-entry result = %p, want %p", got, first)
	}
	if uint32(first.value) != math.MaxUint32 {
		t.Fatalf("negative-one value = %#08x, want 0xffffffff", uint32(first.value))
	}
	runtime.KeepAlive(first)
	runtime.KeepAlive(second)
}
