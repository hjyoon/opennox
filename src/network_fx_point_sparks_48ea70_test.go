package opennox

import (
	"image"
	"testing"

	"github.com/opennox/libs/noxnet/netmsg"

	"github.com/opennox/opennox/v1/client"
)

func TestHandlePointSparkFXNative48EA70(t *testing.T) {
	for _, tc := range []struct {
		op      netmsg.Op
		name    string
		count   int
		speed   int
		minLife int
	}{
		{netmsg.MSG_FX_BLUE_SPARKS, "BlueSpark", 50, 1000, 30},
		{netmsg.MSG_FX_YELLOW_SPARKS, "YellowSpark", 25, 500, 25},
		{netmsg.MSG_FX_CYAN_SPARKS, "CyanSpark", 25, 500, 25},
		{netmsg.MSG_FX_VIOLET_SPARKS, "VioletSpark", 25, 500, 25},
	} {
		t.Run(tc.name, func(t *testing.T) {
			data := []byte{byte(tc.op), 0x34, 0x12, 0xfe, 0xff}
			pos := image.Pt(0x1234, -2)
			var drawables []*client.Drawable
			var randomCalls [][2]int
			got := handlePointSparkFXNative48EA70(tc.op, data, pointSparkFXHooks48EA70{
				connected: func() bool { return true },
				typeID: func(spec pointSparkFXSpec48EA70) int {
					if spec.name != tc.name || spec.count != tc.count || spec.speed != tc.speed || spec.minLife != tc.minLife {
						t.Fatalf("spec = %+v", spec)
					}
					return 77
				},
				random: func(min, max int) int {
					randomCalls = append(randomCalls, [2]int{min, max})
					return min
				},
				frame: func() uint32 { return 100 },
				spawn: func(typ int, at image.Point) *client.Drawable {
					if typ != 77 || at != pos {
						t.Fatalf("spawn = %d/%v", typ, at)
					}
					return &client.Drawable{PosVec: at, ZVal: 99, VelZ: 99}
				},
				activate: func(dr *client.Drawable) { drawables = append(drawables, dr) },
			})
			if got != 5 || len(drawables) != tc.count || len(randomCalls) != tc.count*4 {
				t.Fatalf("result/drawables/random = %d/%d/%d", got, len(drawables), len(randomCalls))
			}
			if want := [][2]int{{0, 255}, {1, tc.speed}, {tc.minLife, 64}, {2, 10}}; randomCalls[0] != want[0] || randomCalls[1] != want[1] || randomCalls[2] != want[2] || randomCalls[3] != want[3] {
				t.Fatalf("random bounds = %v, want %v", randomCalls[:4], want)
			}
			for _, dr := range drawables {
				effect := dr.UnionEffect()
				if effect.Field_108 != uint32(pos.X)<<12 || effect.Field_109 != uint32(pos.Y)<<12 ||
					effect.Field_110 != 1 || effect.Field_111 != 100 || effect.Field_112 != 100+uint32(tc.minLife) ||
					dr.Field_74_4 != 0 || dr.ZVal != 0 || dr.VelZ != 2 {
					t.Fatalf("spark = %+v, effect = %+v", dr, effect)
				}
			}
		})
	}
}

func TestHandlePointSparkFXNative48EA70Gates(t *testing.T) {
	data := []byte{byte(netmsg.MSG_FX_BLUE_SPARKS), 1, 0, 2, 0}
	if got := handlePointSparkFXNative48EA70(netmsg.MSG_FX_BLUE_SPARKS, data[:4], pointSparkFXHooks48EA70{}); got != -1 {
		t.Fatalf("short packet = %d", got)
	}
	if got := handlePointSparkFXNative48EA70(netmsg.MSG_FX_TELEPORT, data, pointSparkFXHooks48EA70{}); got != -1 {
		t.Fatalf("unsupported packet = %d", got)
	}
	if got := handlePointSparkFXNative48EA70(netmsg.MSG_FX_BLUE_SPARKS, data, pointSparkFXHooks48EA70{
		connected: func() bool { return false },
	}); got != 5 {
		t.Fatalf("disconnected packet = %d", got)
	}
}
