package legacy

import "testing"

func TestUnitFreezeGateSet4E79B0StoresAndReturnsAllBits(t *testing.T) {
	for _, value := range []uint32{0, 1, 0x7fffffff, 0x80000000, 0xffffffff} {
		var stored uint32
		stores := 0
		got := unitFreezeGateSet4E79B0(value, func(v uint32) {
			stores++
			stored = v
		})
		if stores != 1 || stored != value || got != value {
			t.Fatalf("set(%#08x) = %#08x, stored %#08x in %d writes", value, got, stored, stores)
		}
	}
}
