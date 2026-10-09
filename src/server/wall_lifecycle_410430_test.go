package server

import (
	"fmt"
	"image"
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

func wallLifecycleFixture410430(t *testing.T) *serverWalls {
	t.Helper()
	byPos, freePos := alloc.Make([]*Wall{}, wallsPerBucket*WallGridSize)
	indexY, freeY := alloc.Make([]*Wall{}, WallGridSize)
	t.Cleanup(freePos)
	t.Cleanup(freeY)
	s := &serverWalls{byPos: byPos, indexY: indexY}
	for i := 0; i < 4; i++ {
		w, free := alloc.New(Wall{})
		t.Cleanup(free)
		if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(w)) <= math.MaxUint32 {
			t.Fatal("wall fixture must retain a native address above 4 GiB")
		}
		w.Next20, s.freeList = s.freeList, w
	}
	return s
}

func checkWallLifecycle410430(t *testing.T, s *serverWalls, live map[image.Point]*Wall) {
	t.Helper()
	seen := make(map[*Wall]bool)
	s.EachWallRaw(func(w *Wall) bool {
		if seen[w] || live[w.GridPos()] != w {
			t.Fatalf("global list contains duplicate/deleted wall %p at %v", w, w.GridPos())
		}
		seen[w] = true
		return true
	})
	if len(seen) != len(live) {
		t.Fatalf("global list has %d walls, want %d; live=%v", len(seen), len(live), live)
	}
	for pos, w := range live {
		if !seen[w] || s.GetWallAtGridRaw(pos) != w || s.GetWallAtGrid(pos) != w ||
			w.Tile1 != 71 || w.Dir0 != 1 || w.Field8 != 0x8123 || w.Field12 != 0xdeadbeef {
			t.Fatalf("live wall at %v lost its index or scalar state: %p", pos, w)
		}
	}
	rowSeen := make(map[*Wall]bool)
	for y := 0; y < WallGridSize; y++ {
		lastX := -1
		for w := s.IndexByY(y); w != nil; w = w.NextByY24 {
			if rowSeen[w] || !seen[w] || int(w.Y6) != y || int(w.X5) <= lastX {
				t.Fatalf("minimap row %d has stale/duplicate/out-of-order wall %p", y, w)
			}
			rowSeen[w], lastX = true, int(w.X5)
		}
	}
	if len(rowSeen) != len(live) {
		t.Fatalf("minimap row index has %d walls, want %d", len(rowSeen), len(live))
	}
	freeSeen := make(map[*Wall]bool)
	for w := s.freeList; w != nil; w = w.Next20 {
		if freeSeen[w] || seen[w] {
			t.Fatalf("wall %p occurs in both live/free lists or twice in free list", w)
		}
		freeSeen[w] = true
	}
	if len(freeSeen)+len(live) != 4 {
		t.Fatalf("wall pool lost entries: live=%d free=%d", len(live), len(freeSeen))
	}
}

func TestWallLifecycle410430KeepsGlobalAndMinimapIndices(t *testing.T) {
	// The first two coordinates share a bucket and row; the others use
	// different rows. All 24 orders exercise global head/middle/tail deletes.
	positions := []image.Point{image.Pt(2, 2), image.Pt(34, 2), image.Pt(4, 4), image.Pt(6, 6)}
	var permute func([]int, []int)
	permute = func(order, rest []int) {
		if len(rest) != 0 {
			for i, index := range rest {
				next := append([]int(nil), rest[:i]...)
				next = append(next, rest[i+1:]...)
				permute(append(append([]int(nil), order...), index), next)
			}
			return
		}
		t.Run(fmt.Sprint(order), func(t *testing.T) {
			s := wallLifecycleFixture410430(t)
			live := make(map[image.Point]*Wall)
			for _, pos := range positions {
				w := s.CreateAtGrid(pos)
				if w == nil {
					t.Fatal("wall pool unexpectedly exhausted")
				}
				w.Tile1, w.Dir0, w.Field8, w.Field12 = 71, 1, 0x8123, 0xdeadbeef
				live[pos] = w
				checkWallLifecycle410430(t, s, live)
			}
			for _, index := range order {
				pos, deleted := positions[index], live[positions[index]]
				s.DeleteAtGrid(pos)
				delete(live, pos)
				if s.GetWallAtGridRaw(pos) != nil {
					t.Fatalf("deleted wall remains at %v", pos)
				}
				checkWallLifecycle410430(t, s, live)
				// Reuse through the ordinary free-list path, with no stale
				// geometry, scalar data, bucket or minimap-row links.
				reusedPos := image.Pt(120+2*index, 120)
				reused := s.CreateAtGrid(reusedPos)
				if reused != deleted || reused.Dir0 != 0 || reused.Tile1 != 0 || reused.Data != nil ||
					reused.Field8 != 0 || reused.Field12 != 0 || reused.NextByY24 != nil {
					t.Fatalf("wall reuse at %v retained stale state: %p, want %p", reusedPos, reused, deleted)
				}
				reused.Tile1, reused.Dir0, reused.Field8, reused.Field12 = 71, 1, 0x8123, 0xdeadbeef
				live[reusedPos] = reused
				checkWallLifecycle410430(t, s, live)
				s.DeleteAtGrid(reusedPos)
				delete(live, reusedPos)
				checkWallLifecycle410430(t, s, live)
				s.DeleteAtGrid(reusedPos) // absent deletion must be harmless
				checkWallLifecycle410430(t, s, live)
			}
		})
	}
	permute(nil, []int{0, 1, 2, 3})
}
