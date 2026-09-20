package server

import (
	"crypto/sha256"
	"encoding/hex"
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/libs/types"
)

func TestIndexedDirection509E20MatchesOriginalRowsAndResidualResult(t *testing.T) {
	rows := make([]byte, 256)
	var counts [9]int
	for direction := range rows {
		var vector IndexedDirectionVector509E20
		result := IndexedDirection509E20(int32(direction), &vector)
		row := vector.X + 3*vector.Y + 4
		if row < 0 || row >= int32(len(counts)) {
			t.Fatalf("direction %d vector = %+v, row = %d", direction, vector, row)
		}
		if got := DirectionOctant509EA0(int32(direction)); got != row {
			t.Fatalf("direction %d octant = %d, want %d", direction, got, row)
		}
		wantResult := int32(-6)
		if vector.Y > 0 {
			wantResult = 6
		}
		if result != wantResult {
			t.Fatalf("direction %d residual result = %d, want %d", direction, result, wantResult)
		}
		rows[direction] = byte(row)
		counts[row]++
	}
	sum := sha256.Sum256(rows)
	if got, want := hex.EncodeToString(sum[:]), "de97e65036c87ba7407e1c1f346bb443c81e78ae8554636eef7c665960b46e7f"; got != want {
		t.Fatalf("quantized row SHA-256 = %s, want %s", got, want)
	}
	if want := [9]int{25, 37, 27, 37, 0, 37, 27, 37, 29}; counts != want {
		t.Fatalf("row counts = %v, want %v", counts, want)
	}
}

func TestIndexedDirection509E20CanonicalDirections(t *testing.T) {
	tests := []struct {
		direction int32
		want      IndexedDirectionVector509E20
		result    int32
	}{
		{direction: 0, want: IndexedDirectionVector509E20{X: 1}, result: -6},
		{direction: 64, want: IndexedDirectionVector509E20{Y: 1}, result: 6},
		{direction: 128, want: IndexedDirectionVector509E20{X: -1}, result: -6},
		{direction: 192, want: IndexedDirectionVector509E20{Y: -1}, result: -6},
	}
	for _, tc := range tests {
		var got IndexedDirectionVector509E20
		if result := IndexedDirection509E20(tc.direction, &got); result != tc.result || got != tc.want {
			t.Fatalf("direction %d = %+v/%d, want %+v/%d", tc.direction, got, result, tc.want, tc.result)
		}
	}
}

func TestDirectionAngle509E00And509E90MatchSealedTable(t *testing.T) {
	want := [3][3]uint32{
		{160, 192, 224},
		{128, 0, 0},
		{96, 64, 32},
	}
	for y := int32(-1); y <= 1; y++ {
		for x := int32(-1); x <= 1; x++ {
			if got := DirectionToAngle509E00(&DirectionInitData{X: x, Y: y}); got != want[y+1][x+1] {
				t.Fatalf("direction (%d,%d) angle = %d, want %d", x, y, got, want[y+1][x+1])
			}
		}
	}
	flat := [...]uint32{160, 192, 224, 128, 0, 0, 96, 64, 32}
	for index, value := range flat {
		if got := DirectionIndexToAngle509E90(int32(index)); got != value {
			t.Fatalf("index %d angle = %d, want %d", index, got, value)
		}
	}
}

func TestDirectionMath509ExRejectsUnsealedAdjacentData(t *testing.T) {
	for _, test := range []struct {
		name string
		call func()
	}{
		{name: "indexed negative", call: func() { IndexedDirection509E20(-1, new(IndexedDirectionVector509E20)) }},
		{name: "indexed high", call: func() { IndexedDirection509E20(256, new(IndexedDirectionVector509E20)) }},
		{name: "angle negative", call: func() { DirectionIndexToAngle509E90(-1) }},
		{name: "angle high", call: func() { DirectionIndexToAngle509E90(9) }},
	} {
		t.Run(test.name, func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Fatal("call did not panic")
				}
			}()
			test.call()
		})
	}
}

func TestNormalizeVector509F20PublicBoundary(t *testing.T) {
	want := types.Ptf(3, 4)
	chestOpenNormalizeVector509F20(&want)
	got := types.Ptf(3, 4)
	NormalizeVector509F20(&got)
	if math.Float32bits(got.X) != math.Float32bits(want.X) ||
		math.Float32bits(got.Y) != math.Float32bits(want.Y) {
		t.Fatalf("normalized vector bits = %#x/%#x, want %#x/%#x",
			math.Float32bits(got.X), math.Float32bits(got.Y),
			math.Float32bits(want.X), math.Float32bits(want.Y))
	}
}

func TestIndexedDirectionVector509E20Layout(t *testing.T) {
	if got := unsafe.Sizeof(IndexedDirectionVector509E20{}); got != 8 {
		t.Fatalf("size = %d, want 8", got)
	}
	if got := unsafe.Offsetof(IndexedDirectionVector509E20{}.Y); got != 4 {
		t.Fatalf("Y offset = %d, want 4", got)
	}
}
