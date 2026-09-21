package opennox

import (
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/libs/types"
)

func TestMonsterListenNativeLayout50CD40(t *testing.T) {
	if got := unsafe.Sizeof(MonsterListen{}); got != monsterListenNativeSize {
		t.Fatalf("MonsterListen size = %d, want %d", got, monsterListenNativeSize)
	}
}

func TestMonsterListenShouldEmitNative50CDD0(t *testing.T) {
	tests := []struct {
		name                      string
		eventFrame, previousFrame uint32
		score                     int
		previousScore             uint32
		want                      bool
	}{
		{name: "newer frame", eventFrame: 11, previousFrame: 10, previousScore: math.MaxUint32, want: true},
		{name: "same frame louder", eventFrame: 10, previousFrame: 10, score: 51, previousScore: 50, want: true},
		{name: "same frame equal", eventFrame: 10, previousFrame: 10, score: 50, previousScore: 50},
		{name: "signed stale score", eventFrame: 10, previousFrame: 10, score: 0, previousScore: math.MaxUint32, want: true},
		{name: "older but louder", eventFrame: 9, previousFrame: 10, score: 51, previousScore: 50, want: true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := monsterListenShouldEmitNative50CDD0(tc.eventFrame, tc.previousFrame, tc.score, tc.previousScore); got != tc.want {
				t.Fatalf("emit = %t, want %t", got, tc.want)
			}
		})
	}
}

func TestSoundFadePercentNative501AF0(t *testing.T) {
	tests := []struct {
		name      string
		max, mute int
		from, to  types.Pointf
		want      int
	}{
		{
			name: "sqrt rounds to nearest even",
			max:  10,
			from: types.Ptf(1.5, 0),
			want: 80,
		},
		{
			name: "negative delta reaches distance gate",
			max:  10,
			from: types.Ptf(-10, 0),
			want: 0,
		},
		{
			name: "positive max delta short circuits",
			max:  10,
			from: types.Ptf(10, 0),
			want: 0,
		},
		{
			name: "mute threshold is inclusive",
			max:  10,
			mute: 80,
			from: types.Ptf(1.5, 0),
			want: 0,
		},
		{
			name: "unordered x87 comparisons",
			max:  10,
			from: types.Ptf(float32(math.NaN()), 0),
			want: 100,
		},
		{
			name: "nonpositive max distance",
			max:  0,
			from: types.Ptf(1, 0),
			want: 0,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := soundFadePercentNative501AF0(tc.max, tc.mute, tc.from, tc.to); got != tc.want {
				t.Fatalf("fade = %d, want %d", got, tc.want)
			}
		})
	}
}

func TestX87RoundQwordLowNative501AF0(t *testing.T) {
	tests := []struct {
		value float64
		want  int32
	}{
		{value: 2.5, want: 2},
		{value: 3.5, want: 4},
		{value: -2.5, want: -2},
		{value: math.NaN(), want: 0},
		{value: math.Inf(1), want: 0},
	}
	for _, tc := range tests {
		if got := x87RoundQwordLowNative501AF0(tc.value); got != tc.want {
			t.Fatalf("round(%v) = %d, want %d", tc.value, got, tc.want)
		}
	}
}

func TestSoundOccludedPercentNative50D000(t *testing.T) {
	tests := []struct {
		value int
		want  int
	}{
		{value: 1, want: 0},
		{value: 3, want: 2},
		{value: 5, want: 2},
		{value: 7, want: 4},
		{value: -3, want: -2},
	}
	for _, tc := range tests {
		if got := soundOccludedPercentNative50D000(tc.value); got != tc.want {
			t.Fatalf("occluded(%d) = %d, want %d", tc.value, got, tc.want)
		}
	}
}

func TestSoundPassesThresholdNative50D0C0(t *testing.T) {
	tests := []struct {
		flags, score int
		want         bool
	}{
		{flags: 0, score: 49},
		{flags: 0, score: 50, want: true},
		{flags: 0x20, score: 88},
		{flags: 0x20, score: 89, want: true},
		{flags: 0x40, score: 19},
		{flags: 0x40, score: 20, want: true},
		{flags: 0x60, score: 88},
		{flags: 0x60, score: 89, want: true},
	}
	for _, tc := range tests {
		if got := soundPassesThresholdNative50D0C0(tc.flags, tc.score); got != tc.want {
			t.Fatalf("threshold(%#x, %d) = %t, want %t", tc.flags, tc.score, got, tc.want)
		}
	}
}

func TestMonsterHearingPolygonCoordinateRounding50CF10(t *testing.T) {
	tests := []struct {
		value float32
		want  int32
	}{
		{value: 10.5, want: 10},
		{value: 11.5, want: 12},
		{value: -10.5, want: -10},
		{value: -11.5, want: -12},
	}
	for _, tc := range tests {
		if got := polygonFloatToIntNative4217B0(tc.value); got != tc.want {
			t.Fatalf("polygon coordinate %v = %d, want %d", tc.value, got, tc.want)
		}
	}
}
