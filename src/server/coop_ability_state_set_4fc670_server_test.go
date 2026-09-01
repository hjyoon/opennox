package server

import (
	"math"
	"testing"
	"unsafe"
)

func TestCoopAbilityStateSet4FC670NativeWidthStoreAndReturn(t *testing.T) {
	s := new(Server)
	if got := unsafe.Sizeof(s.coopAbilityState4FC670); got != 4 {
		t.Fatalf("cooperative-ability state size = %d, want 4", got)
	}
	for _, value := range []int32{
		math.MinInt32,
		-1,
		0,
		1,
		math.MaxInt32,
		-1985229329, // 0x89abcdef
	} {
		if got := s.SetCoopAbilityState4FC670(value); uint32(got) != uint32(value) {
			t.Fatalf("SetCoopAbilityState4FC670(%#08x) = %#08x", uint32(value), uint32(got))
		}
		if got := s.CoopAbilityState4FC670(); uint32(got) != uint32(value) {
			t.Fatalf("CoopAbilityState4FC670 after %#08x = %#08x", uint32(value), uint32(got))
		}
	}
}
