package legacy

import (
	"math"
	"testing"
)

func TestRandomSpellSelectionExport4FE060PreservesUnsignedMasksAndSignedResult(t *testing.T) {
	oldSelect := Nox_xxx_unused_4FE060
	t.Cleanup(func() { Nox_xxx_unused_4FE060 = oldSelect })

	var gotFirst, gotSecond uint32
	Nox_xxx_unused_4FE060 = func(firstMask, secondMask uint32) int32 {
		gotFirst, gotSecond = firstMask, secondMask
		return math.MinInt32
	}
	if got := randomSpellSelectionExportCall4FE060(math.MaxUint32, 0x80000000); got != math.MinInt32 {
		t.Fatalf("export result = %d, want %d", got, int32(math.MinInt32))
	}
	if gotFirst != math.MaxUint32 || gotSecond != 0x80000000 {
		t.Fatalf("export masks = %#x/%#x, want %#x/0x80000000", gotFirst, gotSecond, uint32(math.MaxUint32))
	}

	Nox_xxx_unused_4FE060 = func(firstMask, secondMask uint32) int32 {
		gotFirst, gotSecond = firstMask, secondMask
		return math.MaxInt32
	}
	if got := randomSpellSelectionExportCall4FE060(0, math.MaxUint32); got != math.MaxInt32 {
		t.Fatalf("second export result = %d, want %d", got, int32(math.MaxInt32))
	}
	if gotFirst != 0 || gotSecond != math.MaxUint32 {
		t.Fatalf("second export masks = %#x/%#x, want 0/%#x", gotFirst, gotSecond, uint32(math.MaxUint32))
	}
}

func TestRandomSpellExcludedExport4FE100MatchesExactSelector(t *testing.T) {
	excluded := map[int32]bool{
		1: true, 2: true, 6: true, 13: true, 15: true, 18: true,
		19: true, 20: true, 30: true, 32: true, 33: true, 34: true,
		38: true, 51: true, 57: true, 68: true, 69: true, 70: true,
		73: true, 129: true, 133: true,
	}
	for spellID := int32(1); spellID <= 133; spellID++ {
		want := int32(0)
		if excluded[spellID] {
			want = 1
		}
		if got := randomSpellExcludedExportCall4FE100(spellID); got != want {
			t.Errorf("spell %d = %d, want %d", spellID, got, want)
		}
	}
	for _, spellID := range []int32{math.MinInt32, -1, 0, 134, math.MaxInt32} {
		if got := randomSpellExcludedExportCall4FE100(spellID); got != 0 {
			t.Errorf("out-of-range spell %d = %d, want 0", spellID, got)
		}
	}
}
