package legacy

import "testing"

func TestMapgenAdjacentPortalScanUsesSignedCells(t *testing.T) {
	room := &mapgenRankedRoom{}
	room.grid[0], room.grid[1] = 0xfffffff9, 19
	room.cellDimensions[0], room.cellDimensions[1] = 7, 7

	prefab, hall, ordinary := &mapgenRankedRoom{flags: 2}, &mapgenRankedRoom{flags: 2}, &mapgenRankedRoom{}
	prefab.kind, hall.kind, ordinary.kind = 1, 5, 1

	for _, tc := range []struct {
		name         string
		direction    int
		start, step  [2]int32
		neighbor     *mapgenRankedRoom
		from, to     int
		midpoint     int32
		double, want bool
	}{
		{"north prefab", 0, [2]int32{-7, 18}, [2]int32{1, 0}, prefab, 1, 4, 2, false, true},
		{"south prefab skipped", 1, [2]int32{-7, 26}, [2]int32{1, 0}, prefab, 1, 4, 0, false, false},
		{"east hall through last cell", 2, [2]int32{0, 19}, [2]int32{0, 1}, hall, 3, 7, 4, true, true},
		{"west hall", 3, [2]int32{-8, 19}, [2]int32{0, 1}, hall, 0, 4, 2, true, true},
		{"unmarked room skipped", 2, [2]int32{0, 19}, [2]int32{0, 1}, ordinary, 0, 7, 0, false, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			visited, placed := 0, 0
			mapgenScanAdjacentPortals(room, tc.direction, func(cell [2]int32) *mapgenRankedRoom {
				i := int32(visited)
				want := [2]int32{tc.start[0] + i*tc.step[0], tc.start[1] + i*tc.step[1]}
				if cell != want {
					t.Fatalf("visited grid coordinate %v, want %v", cell, want)
				}
				visited++
				if visited-1 >= tc.from && visited-1 < tc.to {
					return tc.neighbor
				}
				return nil
			}, func(neighbor *mapgenRankedRoom, midpoint int32, double bool) {
				placed++
				if neighbor != tc.neighbor || midpoint != tc.midpoint || double != tc.double {
					t.Fatalf("portal position %d, two leaves = %v, want %d and %v", midpoint, double, tc.midpoint, tc.double)
				}
			})
			wantPlaced := 0
			if tc.want {
				wantPlaced = 1
			}
			if visited != 7 || placed != wantPlaced {
				t.Fatalf("visited %d grid cells and placed %d portals, want 7 and %d", visited, placed, wantPlaced)
			}
		})
	}
}
