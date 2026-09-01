package server

import (
	"math"
	"testing"
	"unsafe"
)

func TestMapInitStateSet4FC570NativeWidthStoreAndReturn(t *testing.T) {
	s := new(Server)
	if got := unsafe.Sizeof(s.mapInitState4FC570); got != 4 {
		t.Fatalf("map-init state size = %d, want 4", got)
	}
	for _, value := range []int32{
		math.MinInt32,
		-1,
		0,
		1,
		math.MaxInt32,
		-1985229329, // 0x89abcdef
	} {
		if got := s.SetMapInitState4FC570(value); uint32(got) != uint32(value) {
			t.Fatalf("SetMapInitState4FC570(%#08x) = %#08x", uint32(value), uint32(got))
		}
		if got := s.MapInitState4FC570(); uint32(got) != uint32(value) {
			t.Fatalf("MapInitState4FC570 after %#08x = %#08x", uint32(value), uint32(got))
		}
	}
}
