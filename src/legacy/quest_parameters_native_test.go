package legacy

import (
	"math"
	"testing"

	"github.com/opennox/opennox/v1/common/memmap"
)

func TestQuestParameterNativeFloatOriginalBits(t *testing.T) {
	for _, offset := range []uintptr{questDifficulty, questDamageScale, questHealthScale} {
		word := memmap.PtrUint32(questParametersBase, offset)
		original := *word
		t.Cleanup(func() { *word = original })
		for _, tc := range []struct {
			input uint32
			want  uint64
		}{
			{0x80000000, 0x8000000000000000},
			{0x00000001, 0x36a0000000000000},
			{0x7f800000, 0x7ff0000000000000},
			{0xff800000, 0xfff0000000000000},
			{0x7f800001, 0x7ff8000020000000},
			{0xffbfffff, 0xffffffffe0000000},
		} {
			*word = tc.input
			if got := math.Float64bits(questParameterFloat(offset)); got != tc.want {
				t.Fatalf("quest word at %#x with bits %#08x = %#016x, want %#016x", offset, tc.input, got, tc.want)
			}
		}
	}
}
