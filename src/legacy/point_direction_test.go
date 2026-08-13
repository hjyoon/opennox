package legacy

import (
	"math"
	"testing"

	"github.com/opennox/libs/types"
)

func TestPointDirection4E6CE0CanonicalSectors(t *testing.T) {
	tests := []struct {
		name string
		to   types.Pointf
		want int
	}{
		{name: "up", to: types.Ptf(0, 1), want: 2},
		{name: "upper right", to: types.Ptf(1, 1), want: 10},
		{name: "right", to: types.Ptf(1, 0), want: 8},
		{name: "lower right", to: types.Ptf(1, -1), want: 9},
		{name: "down", to: types.Ptf(0, -1), want: 1},
		{name: "lower left", to: types.Ptf(-1, -1), want: 5},
		{name: "left", to: types.Ptf(-1, 0), want: 4},
		{name: "upper left", to: types.Ptf(-1, 1), want: 6},
		{name: "coincident", to: types.Ptf(0, 0), want: 1},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := pointDirection4E6CE0(types.Pointf{}, tc.to); got != tc.want {
				t.Fatalf("got %d, want %d", got, tc.want)
			}
		})
	}
}

func TestPointDirection4E6CE0SlopeBoundaries(t *testing.T) {
	small := pointDirectionSlopeSmall4E6CE0
	large := pointDirectionSlopeLarge4E6CE0
	tests := []struct {
		name string
		to   types.Pointf
		want int
	}{
		{name: "small slope equal", to: types.Ptf(1, small), want: 8},
		{name: "small slope above", to: types.Ptf(1, math.Nextafter32(small, float32(math.Inf(1)))), want: 10},
		{name: "large slope equal", to: types.Ptf(1, large), want: 10},
		{name: "large slope above", to: types.Ptf(1, math.Nextafter32(large, float32(math.Inf(1)))), want: 2},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := pointDirection4E6CE0(types.Pointf{}, tc.to); got != tc.want {
				t.Fatalf("got %d, want %d", got, tc.want)
			}
		})
	}
}

func TestPointDirection4E6CE0X87SpillAsymmetry(t *testing.T) {
	min := float32(math.SmallestNonzeroFloat32)
	tests := []struct {
		name string
		to   types.Pointf
		want int
	}{
		{
			name: "small projection rounds negative to minus zero",
			to:   types.Ptf(2*min, min),
			want: 8,
		},
		{
			name: "large projection remains negative in x87",
			to:   types.Ptf(2*min, 5*min),
			want: 2,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := pointDirection4E6CE0(types.Pointf{}, tc.to); got != tc.want {
				t.Fatalf("got %d, want %d", got, tc.want)
			}
		})
	}
}

func TestPointDirection4E6CE0UnorderedAndInfinity(t *testing.T) {
	nan := float32(math.NaN())
	inf := float32(math.Inf(1))
	tests := []struct {
		name string
		to   types.Pointf
		want int
	}{
		{name: "NaN X clears all bits", to: types.Ptf(nan, 0), want: 2},
		{name: "NaN Y clears all bits", to: types.Ptf(0, nan), want: 2},
		{name: "positive X infinity", to: types.Ptf(inf, 0), want: 8},
		{name: "negative X infinity", to: types.Ptf(-inf, 0), want: 4},
		{name: "positive Y infinity", to: types.Ptf(0, inf), want: 2},
		{name: "negative Y infinity", to: types.Ptf(0, -inf), want: 1},
		{name: "infinity minus infinity unordered", to: types.Ptf(inf, inf), want: 2},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := pointDirection4E6CE0(types.Pointf{}, tc.to); got != tc.want {
				t.Fatalf("got %d, want %d", got, tc.want)
			}
		})
	}
}

func TestPointDirectionResult4E6CE0OriginalTables(t *testing.T) {
	if got := math.Float32bits(pointDirectionSlopeSmall4E6CE0); got != 0x3ed37a5f {
		t.Fatalf("small slope bits = %#x, want 0x3ed37a5f", got)
	}
	if got := math.Float32bits(pointDirectionSlopeLarge4E6CE0); got != 0x401af288 {
		t.Fatalf("large slope bits = %#x, want 0x401af288", got)
	}
	want := [...]int{2, 0, 6, 4, 10, 0, 0, 0, 0, 0, 0, 5, 8, 9, 0, 1}
	for mask, expected := range want {
		if got := pointDirectionResult4E6CE0(uint8(mask)); got != expected {
			t.Fatalf("mask %#x = %d, want %d", mask, got, expected)
		}
	}
}
