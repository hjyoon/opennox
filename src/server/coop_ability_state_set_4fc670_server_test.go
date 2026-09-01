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

func TestCoopAbilityStateSet4FC670ConsumerComparisonAndClear(t *testing.T) {
	s := new(Server)
	for _, tc := range []struct {
		name    string
		value   int32
		pending bool
	}{
		{name: "minimum", value: math.MinInt32, pending: true},
		{name: "minus-one", value: -1, pending: true},
		{name: "zero", value: 0},
		{name: "one", value: 1, pending: true},
		{name: "two", value: 2, pending: true},
		{name: "maximum", value: math.MaxInt32, pending: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s.SetCoopAbilityState4FC670(tc.value)
			if got := s.CoopAbilityStatePending4FC680(); got != tc.pending {
				t.Fatalf("pending(%d) = %v, want %v", tc.value, got, tc.pending)
			}
			if got := s.CoopAbilityState4FC670(); got != tc.value {
				t.Fatalf("state(%d) = %d", tc.value, got)
			}
			s.ClearCoopAbilityState4FC680()
			if got := s.CoopAbilityState4FC670(); got != 0 {
				t.Fatalf("state after clear = %d, want 0", got)
			}
		})
	}
}
