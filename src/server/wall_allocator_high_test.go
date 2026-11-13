package server

import (
	"image"
	"testing"
	"unsafe"
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
			walls.Free()
			if walls.byPos != nil || walls.indexY != nil || walls.freeList != nil {
				t.Fatal("released wall storage retained an owner")
			}
		})
	}
}
