package server

import (
	"math"
	"testing"
	"unsafe"
)

func TestMapEntryStateSet4FC580NativeWidthStoreAndReturn(t *testing.T) {
	s := new(Server)
	if got := unsafe.Sizeof(s.mapEntryState4FC580); got != 4 {
		t.Fatalf("map-entry state size = %d, want 4", got)
	}
	for _, value := range []int32{
		math.MinInt32,
		-1,
		0,
		1,
		math.MaxInt32,
		-1985229329, // 0x89abcdef
	} {
		if got := s.SetMapEntryState4FC580(value); uint32(got) != uint32(value) {
			t.Fatalf("SetMapEntryState4FC580(%#08x) = %#08x", uint32(value), uint32(got))
		}
		if got := s.MapEntryState4FC580(); uint32(got) != uint32(value) {
			t.Fatalf("MapEntryState4FC580 after %#08x = %#08x", uint32(value), uint32(got))
		}
	}
}

func TestMapEntryStateSet4FC580ConsumerComparisons(t *testing.T) {
	s := new(Server)
	for _, tc := range []struct {
		name       string
		value      int32
		pending    bool
		playerInit bool
	}{
		{name: "minimum", value: math.MinInt32, pending: true},
		{name: "minus-one", value: -1, pending: true},
		{name: "zero", value: 0},
		{name: "one", value: 1, pending: true, playerInit: true},
		{name: "two", value: 2, pending: true},
		{name: "maximum", value: math.MaxInt32, pending: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s.SetMapEntryState4FC580(tc.value)
			if got := s.MapEntryStatePending4FC600(); got != tc.pending {
				t.Fatalf("pending(%d) = %v, want %v", tc.value, got, tc.pending)
			}
			if got := s.MapEntryStateRequestsPlayerInit4FC6D0(); got != tc.playerInit {
				t.Fatalf("player-init(%d) = %v, want %v", tc.value, got, tc.playerInit)
			}
		})
	}
}
