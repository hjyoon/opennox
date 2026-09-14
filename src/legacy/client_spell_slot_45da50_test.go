package legacy

import "testing"

func TestQuickbarSelectedSlot45DA50UsesNativeRowAddress(t *testing.T) {
	for _, tc := range []struct {
		row, slot int
	}{
		{0, 0},
		{2, 2},
		{4, 4},
	} {
		got := quickbarSelectedSlotFixture45DA50(tc.row, tc.slot, false)
		wantOffset := 40*tc.row + 8*tc.slot
		if got.Offset != wantOffset || got.Spell != 0x12345678 || got.Flags != 3 {
			t.Errorf("row %d slot %d: got %+v, want offset %d spell 0x12345678 flags 3", tc.row, tc.slot, got, wantOffset)
		}
	}
}

func TestQuickbarSelectedSlot45DA50RejectsInvalidInput(t *testing.T) {
	for _, tc := range []struct {
		row, slot int
		nilBase   bool
	}{
		{0, 2, true},
		{5, 2, false},
		{-1, 2, false},
		{2, -1, false},
		{2, 5, false},
	} {
		got := quickbarSelectedSlotFixture45DA50(tc.row, tc.slot, tc.nilBase)
		if got.Offset != -1 {
			t.Errorf("row %d slot %d nil base %t: got %+v, want no slot", tc.row, tc.slot, tc.nilBase, got)
		}
	}
}
