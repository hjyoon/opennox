package opennox

import (
	"image"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/client"
)

func TestUpdateFireballDrawable4CCE70NativeWidth(t *testing.T) {
	ball := &client.Drawable{
		PosVec:  image.Pt(108, 206),
		Field_8: 100,
		Field_9: 200,
	}
	spark := &client.Drawable{ZVal2: 99}
	if unsafe.Sizeof(uintptr(0)) == 8 &&
		(uintptr(unsafe.Pointer(ball)) <= uintptr(^uint32(0)) ||
			uintptr(unsafe.Pointer(spark)) <= uintptr(^uint32(0))) {
		t.Skip("allocator returned a low address")
	}

	randomValues := []int{25, 10, 150, 35, -2, 75}
	randomRanges := [][2]int{
		{0, 100}, {-25, 25}, {100, 300}, {30, 45}, {-2, 4}, {0, 100},
	}
	randomIndex := 0
	spawned := 0
	frames := 0
	activated := false
	hooks := fireballDrawableHooks4CCEA0{
		paused: func() bool { return false },
		typeID: func(name string) int {
			if name != "Spark" {
				t.Fatalf("type name = %q", name)
			}
			return 73
		},
		random: func(min, max int) int {
			if randomIndex >= len(randomValues) {
				t.Fatalf("unexpected random call (%d, %d)", min, max)
			}
			if got := [2]int{min, max}; got != randomRanges[randomIndex] {
				t.Fatalf("random range %d = %v, want %v", randomIndex, got, randomRanges[randomIndex])
			}
			value := randomValues[randomIndex]
			randomIndex++
			return value
		},
		frame: func() uint32 {
			value := uint32(1000 + frames)
			frames++
			return value
		},
		direction: func(v image.Point) int {
			if v != (image.Pt(-8, -6)) {
				t.Fatalf("direction vector = %v", v)
			}
			return 250
		},
		spawn: func(typ int, pos image.Point) *client.Drawable {
			spawned++
			if typ != 73 {
				t.Fatalf("spawn type = %d", typ)
			}
			switch spawned {
			case 1:
				if pos != (image.Pt(102, 201)) {
					t.Fatalf("first spawn position = %v", pos)
				}
				return spark
			case 2:
				if pos != (image.Pt(106, 204)) {
					t.Fatalf("second spawn position = %v", pos)
				}
				return nil
			default:
				t.Fatalf("unexpected spawn %d", spawned)
				return nil
			}
		},
		activate: func(got *client.Drawable) {
			if got != spark {
				t.Fatalf("activated %p, want %p", got, spark)
			}
			activated = true
		},
	}

	if got := updateFireballDrawable4CCE70(ball, 3, hooks); got != 1 {
		t.Fatalf("update = %d, want 1", got)
	}
	if spawned != 2 || randomIndex != len(randomValues) || frames != 2 || !activated {
		t.Fatalf("spawned=%d random=%d frames=%d activated=%t", spawned, randomIndex, frames, activated)
	}
	effect := spark.UnionEffect()
	if effect.Field_108 != uint32(102)<<12 || effect.Field_109 != uint32(201)<<12 ||
		effect.Field_110 != 450 || effect.Field_111 != 1001 || effect.Field_112 != 1035 ||
		spark.Field_74_4 != 4 || spark.ZVal != 28 || spark.ZVal2 != 0 || spark.VelZ != -2 {
		t.Fatalf("spark not initialized: %+v, %+v", spark, effect)
	}
}

func TestUpdateFireballDrawable4CCE70Guards(t *testing.T) {
	t.Run("field 120 short circuits pause", func(t *testing.T) {
		ball := &client.Drawable{Field_120: 1}
		pauseCalls := 0
		got := updateFireballDrawable4CCE70(ball, 5, fireballDrawableHooks4CCEA0{
			paused: func() bool {
				pauseCalls++
				return false
			},
		})
		if got != 1 || pauseCalls != 0 {
			t.Fatalf("update=%d pause calls=%d", got, pauseCalls)
		}
	})

	t.Run("pause suppresses sparks", func(t *testing.T) {
		ball := &client.Drawable{PosVec: image.Pt(20, 20), Field_8: 10, Field_9: 10}
		typeCalls := 0
		got := updateFireballDrawable4CCE70(ball, 5, fireballDrawableHooks4CCEA0{
			paused: func() bool { return true },
			typeID: func(string) int {
				typeCalls++
				return 1
			},
		})
		if got != 1 || typeCalls != 0 {
			t.Fatalf("update=%d type calls=%d", got, typeCalls)
		}
	})
}

func TestUpdateFireballDrawable4CCE70Stationary(t *testing.T) {
	ball := &client.Drawable{PosVec: image.Pt(20, 30), Field_8: 20, Field_9: 30}
	typeCalls := 0
	spawned := false
	got := updateFireballDrawable4CCE70(ball, 1, fireballDrawableHooks4CCEA0{
		paused: func() bool { return false },
		typeID: func(name string) int {
			typeCalls++
			return 1
		},
		spawn: func(int, image.Point) *client.Drawable {
			spawned = true
			return nil
		},
	})
	if got != 1 || typeCalls != 1 || spawned {
		t.Fatalf("update=%d type calls=%d spawned=%t", got, typeCalls, spawned)
	}
}
