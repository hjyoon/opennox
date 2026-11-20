package legacy

import (
	"reflect"
	"testing"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/server"
)

func testMapgenMoveNativeR311(t *testing.T) {
	vectors := []struct {
		fixed mapgenWorldPoint
		door  types.Pointf
	}{
		{mapgenWorldPoint{7.25, -9.5}, types.Pointf{}},
		{mapgenWorldPoint{24, 47}, types.Pointf{X: 23, Y: 46}},
		{mapgenWorldPoint{-47, -24}, types.Pointf{X: -23}},
		{mapgenWorldPoint{}, types.Pointf{}},
	}
	input := mapgenWorldPoint{123, 456}
	type calls struct {
		fix, lookup, create, pending int
	}
	count := 0
	for index, vector := range vectors {
		for _, door := range []bool{false, true} {
			for _, valid := range []bool{false, true} {
				obj := server.Object{ObjFlags: 0xa5000003, Extent: 83}
				want := obj
				extent := uint32(123)
				if index == len(vectors)-1 {
					extent = ^uint32(0)
				}
				expectedExtent := extent
				var observed, expected calls
				if valid {
					want.Extent = extent
					want.ObjFlags |= object.FlagEnabled
					want.PosVec = types.Pointf{X: vector.fixed.x, Y: vector.fixed.y}
					if door {
						want.PosVec = vector.door
					}
					expected = calls{1, 1, 1, 1}
					expectedExtent++
				} else {
					expected.fix = 1
				}
				access := mapgenObjectPlacementAccess{
					fix: func(from, to *mapgenWorldPoint) bool {
						if from != &input {
							t.Fatal("transform received the wrong coordinate pointer")
						}
						observed.fix++
						*to = vector.fixed
						return valid
					},
					usesDoorXfer: func(unit *server.Object) bool {
						if unit != &obj {
							t.Fatal("type inspection lost object identity")
						}
						observed.lookup++
						return door
					},
					create: func(unit *server.Object) {
						if unit != &obj || unit.PosVec != want.PosVec {
							t.Fatal("creation did not receive the moved object and snapped position")
						}
						observed.create++
					},
					clearPending: func() { observed.pending++ },
				}
				after := mapgenMoveObject(&obj, &input, &extent, access)
				if after != &obj || !reflect.DeepEqual(obj, want) || input != (mapgenWorldPoint{123, 456}) ||
					extent != expectedExtent || observed != expected {
					t.Fatalf("move changed the original contract for vector %d, door=%t, valid=%t", index, door, valid)
				}
				count++
			}
		}
	}
	for _, valid := range []bool{false, true} {
		extent := uint32(89)
		var seen calls
		access := mapgenObjectPlacementAccess{fix: func(_, out *mapgenWorldPoint) bool {
			seen.fix++
			*out = mapgenWorldPoint{24, 47}
			return valid
		}}
		if obj := mapgenMoveObject(nil, &input, &extent, access); obj != nil || seen != (calls{fix: 1}) || extent != 89 {
			t.Fatalf("nil object changed state for valid=%t", valid)
		}
		count++
	}
	t.Logf("Mapgen move native R311: %d original-domain cases PASS", count)
}
