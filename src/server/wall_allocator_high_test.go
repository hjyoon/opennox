package server

import (
	"image"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

func TestWallNativeAllocatorPolicies(t *testing.T) {
	for _, high := range []bool{false, true} {
		name := "default"
		if high {
			name = "high"
		}
		t.Run(name, func(t *testing.T) {
			var walls serverWalls
			init := walls.Init
			if high {
				init = walls.InitHigh
			}
			if init() != 1 {
				t.Fatal("wall storage allocation failed")
			}
			t.Cleanup(func() {
				if walls.byPos != nil {
					walls.Free()
				}
			})
			if len(walls.byPos) != wallsPerBucket*WallGridSize || len(walls.indexY) != WallGridSize {
				t.Fatal("wall storage capacities changed")
			}
			if high && unsafe.Sizeof(uintptr(0)) > 4 {
				for _, p := range []unsafe.Pointer{
					unsafe.Pointer(unsafe.SliceData(walls.byPos)),
					unsafe.Pointer(unsafe.SliceData(walls.indexY)),
				} {
					if uintptr(p) <= 0xffffffff {
						t.Fatalf("wall index below 4GiB: %p", p)
					}
				}
				count := 0
				for p := walls.freeList; p != nil; p = p.Next20 {
					if uintptr(unsafe.Pointer(p)) <= 0xffffffff {
						t.Fatalf("wall below 4GiB: %p", p)
					}
					count++
				}
				if count != wallsPerBucket*WallGridSize {
					t.Fatalf("wall pool count = %d", count)
				}
			}
			position := image.Pt(14, 16)
			created := walls.CreateAtGrid(position)
			if created == nil || walls.GetWallAtGrid(position) != created || walls.CreateAtGrid(position) != created {
				t.Fatal("indexed wall insertion or reuse changed")
			}
			walls.Reset()
			if walls.GetWallAtGrid(position) != nil || walls.head != nil {
				t.Fatal("wall survived reset")
			}
			recycled := walls.CreateAtGrid(position)
			if recycled == nil {
				t.Fatal("reset lost the wall pool")
			}
			walls.DeleteAtGrid(position)
			replacement := image.Pt(18, 20)
			if got := walls.CreateAtGrid(replacement); got != recycled || walls.GetWallAtGrid(position) != nil || walls.GetWallAtGrid(replacement) != recycled {
				t.Fatal("deleted wall was not reused through the native indexes")
			}
			var poolOwner unsafe.Pointer
			if high {
				poolOwner = unsafe.Pointer(unsafe.SliceData(walls.highStorage))
				if poolOwner == nil {
					t.Fatal("high wall pool has no foreign owner")
				}
			} else if walls.highStorage != nil {
				t.Fatal("ordinary wall allocation inherited the fixture policy")
			}
			// Keep an indexed live wall and free-list entries alive until teardown.
			walls.Free()
			if walls.byPos != nil || walls.indexY != nil || walls.freeList != nil || walls.highStorage != nil {
				t.Fatal("released wall storage retained an owner")
			}
			if poolOwner != nil && alloc.TryFreePtr(poolOwner) {
				t.Fatal("released wall pool remained registered")
			}
		})
	}
}
