package legacy

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"math"
	"testing"

	"github.com/opennox/libs/types"
)

func TestIndexedDirectionRow4E6E50OriginalVectors(t *testing.T) {
	rows := make([]byte, 256)
	var counts [9]int
	for dir := range rows {
		row := indexedDirectionRow4E6E50(int32(dir))
		if row < 0 || row >= len(counts) {
			t.Fatalf("direction %d returned row %d", dir, row)
		}
		rows[dir] = byte(row)
		counts[row]++
	}
	sum := sha256.Sum256(rows)
	if got, want := hex.EncodeToString(sum[:]), "de97e65036c87ba7407e1c1f346bb443c81e78ae8554636eef7c665960b46e7f"; got != want {
		t.Fatalf("quantized row SHA-256 = %s, want %s", got, want)
	}
	if want := [9]int{25, 37, 27, 37, 0, 37, 27, 37, 29}; counts != want {
		t.Fatalf("row counts = %v, want %v", counts, want)
	}
	for _, dir := range []int32{math.MinInt32, -1, 256, math.MaxInt32} {
		if got := indexedDirectionRow4E6E50(dir); got != 4 {
			t.Fatalf("invalid direction %d returned row %d, want neutral row 4", dir, got)
		}
	}
}

func TestTwoPointsDirectionResults4E6E50OriginalTable(t *testing.T) {
	raw := make([]byte, 0, 9*16*4)
	var word [4]byte
	for _, row := range twoPointsDirectionResults4E6E50 {
		for _, value := range row {
			binary.LittleEndian.PutUint32(word[:], value)
			raw = append(raw, word[:]...)
		}
	}
	sum := sha256.Sum256(raw)
	if got, want := hex.EncodeToString(sum[:]), "fdee48935cdb090ac1b446fa5e1cd5c1c909592421edce6eb155d1c24e7d644c"; got != want {
		t.Fatalf("result table SHA-256 = %s, want %s", got, want)
	}
}

func TestTwoPointsAndDirection4E6E50CanonicalFacings(t *testing.T) {
	tests := []struct {
		name  string
		dir   int32
		front types.Pointf
		back  types.Pointf
	}{
		{name: "right", dir: 0, front: types.Ptf(1, 0), back: types.Ptf(-1, 0)},
		{name: "upper right", dir: 32, front: types.Ptf(1, 1), back: types.Ptf(-1, -1)},
		{name: "up", dir: 64, front: types.Ptf(0, 1), back: types.Ptf(0, -1)},
		{name: "upper left", dir: 96, front: types.Ptf(-1, 1), back: types.Ptf(1, -1)},
		{name: "left", dir: 128, front: types.Ptf(-1, 0), back: types.Ptf(1, 0)},
		{name: "lower left", dir: 160, front: types.Ptf(-1, -1), back: types.Ptf(1, 1)},
		{name: "down", dir: 192, front: types.Ptf(0, -1), back: types.Ptf(0, 1)},
		{name: "lower right", dir: 224, front: types.Ptf(1, -1), back: types.Ptf(-1, 1)},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			origin := types.Ptf(17, -23)
			if got := twoPointsAndDirection4E6E50(origin, tc.dir, origin.Add(tc.front)); got != 1 {
				t.Fatalf("front result = %d, want 1", got)
			}
			if got := twoPointsAndDirection4E6E50(origin, tc.dir, origin.Add(tc.back)); got != 2 {
				t.Fatalf("back result = %d, want 2", got)
			}
		})
	}
}

func TestTwoPointsAndDirection4E6E50SideAndInvalidCases(t *testing.T) {
	origin := types.Pointf{}
	if got := twoPointsAndDirection4E6E50(origin, 0, types.Ptf(0, 1)); got != 8 {
		t.Fatalf("right-facing upper side = %d, want 8", got)
	}
	if got := twoPointsAndDirection4E6E50(origin, 0, types.Ptf(0, -1)); got != 4 {
		t.Fatalf("right-facing lower side = %d, want 4", got)
	}
	if got := twoPointsAndDirection4E6E50(origin, 0, origin); got != 4 {
		t.Fatalf("coincident points = %d, want 4", got)
	}
	nan := float32(math.NaN())
	if got := twoPointsAndDirection4E6E50(origin, 0, types.Ptf(nan, 0)); got != 8 {
		t.Fatalf("unordered sector = %d, want 8", got)
	}
	for _, dir := range []int32{-1, 256} {
		if got := twoPointsAndDirection4E6E50(origin, dir, types.Ptf(1, 0)); got != 0 {
			t.Fatalf("invalid direction %d result = %d, want 0", dir, got)
		}
	}
}
