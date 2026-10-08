package opennox

import (
	"math"
	"testing"
)

func TestNativeAudioSamplePan452FA0(t *testing.T) {
	for pan := -128; pan <= 127; pan++ {
		original := math.Max(-50, math.Min(50, float64(pan)))
		fixed := int(math.Trunc(original*8192/50)) + 8192
		want := int(uint32(127*fixed) >> 14)
		if got := nativeAudioSamplePan(pan); got != want {
			t.Fatalf("wire pan=%d: sample pan=%d, want original signed IDIV result %d", pan, got, want)
		}
	}
	for pan, want := range map[int]int{
		math.MinInt: 0, -51: 0, -50: 0, -49: 1, -1: 62,
		0: 63, 1: 64, 49: 125, 50: 127, 51: 127, math.MaxInt: 127,
	} {
		if got := nativeAudioSamplePan(pan); got != want {
			t.Errorf("pan=%d: sample pan=%d, want %d", pan, got, want)
		}
	}
}
