package server

import (
	"image"
	"math"
	"reflect"
	"testing"
	"unsafe"
)

func TestForEachGroup502670NativeObjectsAndHighIDs(t *testing.T) {
	groupID := uint32(0xf1234567)
	const extent = uint32(0xe1234567)
	s := &Server{}
	obj := &Object{Extent: extent}
	s.Objs.List = obj
	child := &MapGroup{typ: byte(MapGroupObjects), Ind: groupID, List: &MapGroupItem{Raw0: extent, Next8: &MapGroupItem{Raw0: extent + 1}}}
	root := &MapGroup{typ: byte(MapGroupGroups), List: &MapGroupItem{Raw0: groupID}}
	s.MapGroups.AddNewMapGroup57C3B0(child)
	if got := s.MapGroups.GroupByInd(int(int32(groupID))); got != child {
		t.Fatalf("signed ABI32 group ID resolved to %p, want %p", got, child)
	}
	var got []any
	s.ForEachGroup502670(root, MapGroupObjects, func(v any) { got = append(got, v) })
	if len(got) != 1 || got[0] != obj {
		t.Fatalf("object visits = %v, want exact pointer %p", got, obj)
	}
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(obj)) <= math.MaxUint32 {
		t.Fatalf("object pointer = %p, want native address above 4 GiB", obj)
	}
}

func TestForEachGroup502670NativeWaypointAndWall(t *testing.T) {
	s := &Server{}
	wp := &Waypoint{Index: 0xfedcba98}
	s.WPs.List = wp
	wpGroup := &MapGroup{typ: byte(MapGroupWaypoints), List: &MapGroupItem{Raw0: wp.Index}}
	var got []any
	s.ForEachGroup502670(wpGroup, MapGroupWaypoints, func(v any) { got = append(got, v) })
	if !reflect.DeepEqual(got, []any{wp}) {
		t.Fatalf("waypoint visits = %v, want exact pointer %p", got, wp)
	}

	s.Walls.byPos = make([]*Wall, 1<<13)
	wall1 := &Wall{X5: 2, Y6: 4}
	wall2 := &Wall{X5: 8, Y6: 4}
	for _, w := range []*Wall{wall1, wall2} {
		ind, ok := wallArrayInd(w.GridPos())
		if !ok {
			t.Fatalf("invalid wall grid position: %v", w.GridPos())
		}
		s.Walls.byPos[ind] = w
	}
	child := &MapGroup{typ: byte(MapGroupWalls), Ind: 2, List: &MapGroupItem{Raw0: 8, Raw4: 4}}
	s.MapGroups.AddNewMapGroup57C3B0(child)
	wallGroup := &MapGroup{typ: byte(MapGroupWalls), List: &MapGroupItem{Raw0: 2, Raw4: 4}}
	got = nil
	s.ForEachGroup502670(wallGroup, MapGroupWalls, func(v any) { got = append(got, v) })
	if !reflect.DeepEqual(got, []any{wall1, wall2}) {
		t.Fatalf("wall visits and original fallthrough = %v, want [%p %p]", got, wall1, wall2)
	}
	if s.Walls.GetWallAtGrid(image.Pt(2, 4)) != wall1 {
		t.Fatal("wall fixture did not resolve through the normal server index")
	}
}
