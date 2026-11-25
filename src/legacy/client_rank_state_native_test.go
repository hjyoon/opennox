package legacy

import (
	"testing"

	"github.com/opennox/opennox/v1/common/memmap"
)

func TestClientRankObjectiveUpdatesPreserveNativeReturns(t *testing.T) {
	carriers := memmap.PtrT[[3]uint16](0x5D4594, 1090128)
	previous := *carriers
	t.Cleanup(func() { *carriers = previous })
	*carriers = [3]uint16{10, 11, 12}
	if got := sub_4705F0(1, 1, -19017); uint8(got) != 0xb7 || *carriers != [3]uint16{0xb5b7, 11, 12} {
		t.Fatalf("first flag pickup = %#x, carriers %#x", uint8(got), *carriers)
	}
	if got := sub_4705F0(1, 2, 256); got != 1 || *carriers != [3]uint16{0xb5b7, 256, 12} {
		t.Fatalf("second flag pickup = %d, carriers %#x", got, *carriers)
	}
	if got := sub_4705F0(7, 3, 1234); got != 3 || *carriers != [3]uint16{0xb5b7, 256, 12} {
		t.Fatalf("unknown flag team = %d, carriers %#x", got, *carriers)
	}
	if got := sub_4705F0(2, 1, 9); got != 2 || carriers[0] != 0 {
		t.Fatalf("flag drop = %d, carriers %#x", got, *carriers)
	}
	if got := sub_470650(4, 0x73bd); uint8(got) != 0xbd || carriers[2] != 0x73bd {
		t.Fatalf("ball pickup = %#x, carriers %#x", uint8(got), *carriers)
	}
	if got := sub_470650(1, 33); got != 1 || carriers[2] != 0 {
		t.Fatalf("ball drop = %d, carriers %#x", got, *carriers)
	}
	if got := sub_470680(); got != 0 || *carriers != [3]uint16{} {
		t.Fatalf("reset objective markers = %d, carriers %#x", got, *carriers)
	}
}
