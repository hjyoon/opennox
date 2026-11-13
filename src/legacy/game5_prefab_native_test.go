package legacy

import (
	"math"
	"testing"
	"unsafe"
)

func TestGame5PrefabPublicCConnectorsAndRollback(t *testing.T) {
	for _, tc := range []struct {
		name        string
		entry, side int
		a, b, c     [2]int32
		count       int
		blocked     bool
	}{
		{"north-straight", 0, 0, [2]int32{0, 0}, [2]int32{0, -9}, [2]int32{0, -10}, 1, false},
		{"north-right", 0, 0, [2]int32{0, 0}, [2]int32{5, -9}, [2]int32{5, -10}, 3, false},
		{"north-left", 0, 0, [2]int32{0, 0}, [2]int32{-5, -9}, [2]int32{-5, -10}, 3, false},
		{"east-straight", 1, 0, [2]int32{1, 0}, [2]int32{10, 0}, [2]int32{10, 0}, 1, false},
		{"east-south", 1, 0, [2]int32{1, 0}, [2]int32{10, 5}, [2]int32{10, 5}, 3, false},
		{"east-north", 1, 0, [2]int32{1, 0}, [2]int32{10, -5}, [2]int32{10, -5}, 3, false},
		{"elbow-north-east", 2, 0, [2]int32{0, 0}, [2]int32{5, -9}, [2]int32{5, -10}, 2, false},
		{"elbow-north-west", 2, 0, [2]int32{0, 0}, [2]int32{-5, -9}, [2]int32{-6, -10}, 2, false},
		{"elbow-south-east", 3, 0, [2]int32{0, 1}, [2]int32{5, 10}, [2]int32{5, 10}, 2, false},
		{"elbow-south-west", 3, 0, [2]int32{0, 1}, [2]int32{-5, 10}, [2]int32{-6, 10}, 2, false},
		{"prefab-north", 4, 0, [2]int32{0, 0}, [2]int32{}, [2]int32{0, -10}, 1, false},
		{"prefab-south", 4, 1, [2]int32{0, 1}, [2]int32{}, [2]int32{0, 10}, 1, false},
		{"prefab-east", 4, 2, [2]int32{1, 0}, [2]int32{}, [2]int32{10, 0}, 1, false},
		{"prefab-west", 4, 3, [2]int32{0, 0}, [2]int32{}, [2]int32{-10, 0}, 1, false},
		{"occupied-rollback", 0, 0, [2]int32{0, 0}, [2]int32{0, -9}, [2]int32{0, -10}, 0, true},
		{"narrow-offset-rejection", 0, 0, [2]int32{0, 0}, [2]int32{2, -9}, [2]int32{2, -10}, 0, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := game5PrefabProbe(tc.entry, tc.side, tc.a, tc.b, tc.c, tc.blocked)
			base := 2
			if tc.blocked {
				base++
			}
			if got.ok != (tc.count > 0) || got.count != tc.count || got.total != base+tc.count {
				t.Fatalf("actual room graph %+v", got)
			}
			if tc.count == 0 {
				return
			}
			index := map[uintptr]int{}
			for i, p := range got.owners[:2+got.count] {
				if p == 0 || unsafe.Sizeof(uintptr(0)) == 8 && p <= math.MaxUint32 {
					t.Fatal("high room owner", i, p)
				}
				index[p] = i
			}
			// Both outer rooms must be connected through exactly the newly added rooms,
			// and every actual adjacency edge must have its native-width reverse edge.
			visited := map[int]bool{0: true}
			queue := []int{0}
			for len(queue) > 0 {
				current := queue[0]
				queue = queue[1:]
				for _, ptr := range got.neighbors[current] {
					if ptr == 0 {
						continue
					}
					neighbor, ok := index[ptr]
					if !ok {
						t.Fatal("unknown or truncated neighbor", ptr)
					}
					reverse := false
					for _, p := range got.neighbors[neighbor] {
						reverse = reverse || p == got.owners[current]
					}
					if !reverse {
						t.Fatal("missing reverse adjacency", current, neighbor)
					}
					if !visited[neighbor] {
						visited[neighbor] = true
						queue = append(queue, neighbor)
					}
				}
			}
			if len(visited) != 2+tc.count || !visited[1] {
				t.Fatal("disconnected native corridor", visited)
			}
		})
	}
}
