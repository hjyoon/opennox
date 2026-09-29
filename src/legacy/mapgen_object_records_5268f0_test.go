package legacy

import (
	"strconv"
	"testing"
)

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
	if strconv.IntSize == 64 && got.recordsAddress <= 1<<32 {
		t.Fatalf("records fixture address = %#x, want address above PE32 range", got.recordsAddress)
	}
}
