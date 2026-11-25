package legacy

import (
	"bytes"
	"strconv"
	"strings"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/legacy/internal/cfixture"
)

func TestMapgenAreaRecordNamesAndAttachmentFlags(t *testing.T) {
	archive := cfixture.NewGame4ArchiveBank()
	if archive == nil {
		t.Fatal("allocate area archive")
	}
	t.Cleanup(archive.Close)
	archive.Seed()
	names := [3]string{"Test--Room---NSW-C", "Second-Room-w-NSE", strings.Repeat("R", 60) + "P"}
	for i, name := range names {
		buf := unsafe.Slice((*byte)(game4MapRecord(int32(i))), mapgenAreaRecordBytes)
		alloc.StrCopy(buf, name)
	}
	markers := memmap.Slice(0x587000, 255032)[:6]
	old := bytes.Clone(markers)
	copy(markers, "NSEWC\x00")
	t.Cleanup(func() { copy(markers, old) })
	restore := mapgenHighFixtureMemory()
	defer restore()
	got, free := mapgenLoadAreaRecordsCFixture()
	defer free()
	if got != uintptr(len(names)) {
		t.Fatalf("area reader returned word %#x, want processed count %d", got, len(names))
	}
	for i, flags := range []uint32{11, 0, 'P'} {
		record := mapgenAreaRecordAt(int32(i))
		if record == nil || record.flags != flags || mapgenFindAreaRecord(names[i]) != int32(i) {
			t.Fatalf("area record %d did not preserve its name and marker flags %#x", i, flags)
		}
	}
	if mapgenFindAreaRecord(strings.ToLower(names[0])) != -1 || mapgenAreaRecordAt(-1) != nil {
		t.Fatal("area record accepted a mismatched name or negative index")
	}
}

func TestMapgenObjectRecords5268F0PreserveNativePointer(t *testing.T) {
	got := mapgenObjectRecordsFixture5268F0()
	if !got.setupClean || got.recordsAddress == 0 || got.recordsToken == 0 {
		t.Fatalf("fixture setup failed: clean=%v address=%#x token=%#x",
			got.setupClean, got.recordsAddress, got.recordsToken)
	}
	if got.resolvedAddress != got.recordsAddress {
		t.Fatalf("resolved records = %#x, want %#x", got.resolvedAddress, got.recordsAddress)
	}
	for i, address := range got.entries {
		want := got.recordsAddress + uintptr(i*64)
		if address != want {
			t.Fatalf("entry[%d] = %#x, want %#x", i, address, want)
		}
	}
	if got.indices != [4]int{0, 1, 2, -1} {
		t.Fatalf("record lookups = %v, want [0 1 2 -1]", got.indices)
	}
	if got.tokenAfterCleanup != 0 || got.lookupAfterCleanup != -1 {
		t.Fatalf("record state survived cleanup: token=%#x lookup=%d",
			got.tokenAfterCleanup, got.lookupAfterCleanup)
	}
	if strconv.IntSize == 64 && uint64(got.recordsAddress) <= 1<<32 {
		t.Fatalf("records fixture address = %#x, want address above PE32 range", got.recordsAddress)
	}
}
