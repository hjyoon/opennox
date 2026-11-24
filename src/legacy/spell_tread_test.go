package legacy

import (
	"math"
	"testing"

	"github.com/opennox/opennox/v1/common/memmap"
)

func TestMonsterSightFrameCopyUsesBothFrameReads(t *testing.T) {
	slot := memmap.PtrUint32(0x5D4594, 2487684)
	saved, oldHook := *slot, gameFrameHook
	t.Cleanup(func() { *slot, gameFrameHook = saved, oldHook })
	for _, tc := range []struct {
		name   string
		call   func() int32
		frames [2]uint32
		result int32
		stored uint32
	}{
		{"current", func() int32 { return int32(nox_xxx_frameCounterSetCopy_5281E0()) },
			[2]uint32{0x80000000, 17}, math.MinInt32, 17},
		{"next with wraparound", func() int32 { return int32(nox_xxx_frameCounterSetCopyToNextFrame_5281D0()) },
			[2]uint32{0xffffffff, 19}, 0, 20},
	} {
		t.Run(tc.name, func(t *testing.T) {
			reads := 0
			gameFrameHook = func() uint32 {
				if reads >= len(tc.frames) {
					t.Fatal("unexpected additional frame read")
				}
				frame := tc.frames[reads]
				reads++
				return frame
			}
			if got := tc.call(); got != tc.result || *slot != tc.stored || reads != 2 {
				t.Fatalf("returned %d, stored %d and read %d times, want %d, %d and 2", got, *slot, reads, tc.result, tc.stored)
			}
		})
	}
}
