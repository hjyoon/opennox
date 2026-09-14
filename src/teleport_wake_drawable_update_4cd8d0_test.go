package opennox

import (
	"image"
	"reflect"
	"testing"

	"github.com/opennox/opennox/v1/client"
)

func TestUpdateTeleportWakeDrawable4CD8D0(t *testing.T) {
	wake := &client.Drawable{PosVec: image.Pt(100, 200)}
	spark := &client.Drawable{PosVec: image.Pt(95, 195), ZVal: 20, VelZ: 9}
	var bounds [][2]int
	activated := false
	got := updateTeleportWakeDrawable4CD8D0(wake, teleportWakeDrawableHooks4CD8D0{
		typeID: func() int { return 17 },
		random: func(min, max int) int {
			bounds = append(bounds, [2]int{min, max})
			return min
		},
		frame: func() uint32 { return 100 },
		spawn: func(typ int, pos image.Point) *client.Drawable {
			if typ != 17 || pos != spark.PosVec {
				t.Fatalf("spawn = %d/%v", typ, pos)
			}
			return spark
		},
		activate: func(dr *client.Drawable) {
			activated = dr == spark
		},
	})
	wantBounds := [][2]int{{-5, 5}, {-5, 5}, {0, 255}, {1, 100}, {10, 32}, {3, 8}}
	if got != 1 || !activated || !reflect.DeepEqual(bounds, wantBounds) {
		t.Fatalf("result/activated/bounds = %d/%t/%v", got, activated, bounds)
	}
	effect := spark.UnionEffect()
	if effect.Field_108 != uint32(spark.PosVec.X)<<12 || effect.Field_109 != uint32(spark.PosVec.Y)<<12 ||
		effect.Field_110 != 1 || effect.Field_111 != 100 || effect.Field_112 != 110 ||
		spark.Field_74_4 != 0 || spark.ZVal != 0 || spark.VelZ != 3 {
		t.Fatalf("spark = %+v, effect = %+v", spark, effect)
	}
}

func TestUpdateTeleportWakeDrawable4CD8D0SpawnFailure(t *testing.T) {
	got := updateTeleportWakeDrawable4CD8D0(&client.Drawable{}, teleportWakeDrawableHooks4CD8D0{
		typeID: func() int { return 17 },
		random: func(min, max int) int { return min },
		spawn:  func(int, image.Point) *client.Drawable { return nil },
	})
	if got != 1 {
		t.Fatalf("result = %d", got)
	}
}
