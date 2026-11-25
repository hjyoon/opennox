package legacy

import (
	"cmp"
	"slices"
	"testing"
)

func TestSortBookEntriesKeepsOriginalTiesAndMissingTitles(t *testing.T) {
	// Expected order and comparison counts from GAME.EXE 0056624F in PE32.
	for _, test := range []struct {
		name        string
		keys        []int
		want        []uint32
		comparisons int
	}{
		{"eight entries", []int{0, 1, 2, 0, 1, 2, 0, 1}, []uint32{6, 3, 0, 4, 7, 1, 5, 2}, 28},
		{"nine entries", []int{0, 1, 2, 0, 1, 2, 0, 1, 2}, []uint32{3, 6, 0, 1, 4, 7, 2, 8, 5}, 22},
		{"missing titles", []int{-1, 4, 4, 2, 0, -1, 4, 3, 0, 3, -1, 1}, []uint32{5, 1, 2, 3, 4, 11, 0, 7, 8, 9, 10, 6}, 32},
	} {
		t.Run(test.name, func(t *testing.T) {
			const canary uint32 = 0xff000000
			book := make([]uint32, len(test.keys)+1)
			entries := book[:len(test.keys)]
			for i := range entries {
				entries[i] = uint32(i)
			}
			book[len(entries)] = canary
			comparisons := 0
			sortBookEntries(entries, func(a, b uint32) int {
				comparisons++
				left, right := test.keys[a], test.keys[b]
				if left < 0 || right < 0 {
					return 0
				}
				return cmp.Compare(left, right)
			})
			if !slices.Equal(entries, test.want) || comparisons != test.comparisons || book[len(entries)] != canary {
				t.Fatalf("sort = %v, %d comparisons, tail %#x, want %v, %d comparisons, tail %#x",
					entries, comparisons, book[len(entries)], test.want, test.comparisons, canary)
			}
		})
	}
}
