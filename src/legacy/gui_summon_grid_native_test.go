package legacy

import (
	"testing"

	"github.com/opennox/opennox/v1/common/memmap"
)

func TestSummonGrid4C2F70PacksBySizeAndRetainsRecordPointers(t *testing.T) {
	// The application loads this original table when it initializes its data blobs.
	positions := memmap.PtrT[[4][2]int32](0x587000, 184456)
	savedPositions := *positions
	t.Cleanup(func() { *positions = savedPositions })
	*positions = [4][2]int32{{0, 0}, {0, 1}, {1, 0}, {1, 1}}

	for i := 0; i < summonRecordCount; i++ {
		record := summonRecordAt(i)
		old := *record
		t.Cleanup(func() { *record = old })
		*record = summonRecord{}
	}
	for x := int32(0); x < summonGridSide; x++ {
		for y := int32(0); y < summonGridSide; y++ {
			cell := summonGridCell(x, y)
			old := *cell
			t.Cleanup(func() { *cell = old })
			*cell = nil
		}
	}

	const wireCode = 0x80001234
	tall := sub_4C2F20()
	tall.size, tall.net_code = 2, wireCode
	small := sub_4C2F20()
	small.size = 1
	large := sub_4C2F20()
	large.size = 4

	checkGrid := func(want [4]*summonRecord) {
		t.Helper()
		got := [4]*summonRecord{
			*summonGridCell(0, 0), *summonGridCell(0, 1), *summonGridCell(1, 0), *summonGridCell(1, 1),
		}
		if got != want {
			t.Fatalf("grid pointers = %v, want %v", got, want)
		}
	}
	sub_4C2F70()
	checkGrid([4]*summonRecord{small, nil, tall, tall})
	if tall.grid_x != 1 || tall.grid_y != 0 || sub_4C31D0(wireCode-(1<<32)) != tall {
		t.Fatalf("tall summon at (%d, %d) with net code %#x", tall.grid_x, tall.grid_y, tall.net_code)
	}

	sub_4C3210(small)
	sub_4C2F70()
	checkGrid([4]*summonRecord{tall, tall, nil, nil})

	sub_4C3210(tall)
	sub_4C2F70()
	checkGrid([4]*summonRecord{large, large, large, large})
	if first := sub_4C2D60(); first != large || sub_4C2D90(first) != nil {
		t.Fatalf("first active record = %p, want only %p", first, large)
	}
	sub_4C2BF0()
	checkGrid([4]*summonRecord{})
}
