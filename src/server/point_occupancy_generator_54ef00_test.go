package server

import (
	"testing"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/strman"
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/common/memmap"
)

func TestPointOccupancyGeneratorDefault54EF00(t *testing.T) {
	for _, tc := range []struct {
		name  string
		class object.Class
		flags uint32
		point types.Pointf
		want  bool
	}{
		{"hit", object.ClassMonster, 0, types.Ptf(300, 300), true},
		{"miss", object.ClassMonster, 0, types.Ptf(312, 300), false},
		{"closed-circle-edge", object.ClassMonster, 0, types.Ptf(310, 300), true},
		{"generator-excluded", object.ClassMonsterGenerator, 0, types.Ptf(300, 300), false},
		{"disabled-field5", object.ClassMonster, 0x800, types.Ptf(300, 300), false},
		{"other-field5", object.ClassMonster, 0x80000000, types.Ptf(300, 300), true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			// Standalone server tests do not import the legacy blob provider.
			// Supply the scratch storage while preserving any smaller test view.
			const need = 2491712
			if blob := memmap.BlobByAddr(0x5D4594); blob == nil {
				memmap.RegisterBlobData(0x5D4594, "point_occupancy_test", make([]byte, need))
			} else if len(blob.Data) < need {
				previous := *blob
				data := make([]byte, need)
				copy(data, blob.Data)
				blob.Data, blob.Size = data, uintptr(len(data))
				t.Cleanup(func() { *blob = previous })
			}
			s := NewHigh(nil, nil, strman.New())
			t.Cleanup(s.Close)
			s.Map.Init()
			t.Cleanup(s.Map.Free)
			unit := &Object{ObjClass: tc.class, ObjFlags: object.FlagActive | object.FlagEnabled,
				PosVec: types.Ptf(300, 300), NewPos: types.Ptf(300, 300), serverHandle: s.handle, Field5: tc.flags}
			unit.Shape.Kind = ShapeKindCircle
			unit.Shape.Circle.R, unit.Shape.Circle.R2 = 10, 100
			s.Map.AddObjectToIndex(unit)
			before := *unit
			before.Field62 = [2]uint32{}
			scratch := memmap.PtrUint32(0x5D4594, 2491708)
			old := *scratch
			t.Cleanup(func() { *scratch = old })
			*scratch = 0x89abcdef
			deps := s.monsterGeneratorNativeDeps54E930(MonsterGeneratorRuntime54E930{})
			if got := deps.occupied(tc.point); got != tc.want {
				t.Fatal("actual ordinary generator occupancy dependency", got, tc.want)
			}
			want := uint32(0)
			if tc.want {
				want = 1
			}
			after := *unit
			after.Field62 = [2]uint32{}
			if *scratch != want || after != before {
				t.Fatal("original global scratch or unrelated indexed object fields differ", *scratch, want)
			}
		})
	}
}
