package opennox

import (
	"fmt"
	"image"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/client"
)

func TestUpdateCloudDrawable4CE1D0HighAddress(t *testing.T) {
	for _, tc := range []struct {
		name   string
		radius int
	}{
		{name: "cloud", radius: 75},
		{name: "small cloud", radius: 35},
	} {
		t.Run(tc.name, func(t *testing.T) {
			source := &client.Drawable{PosVec: image.Pt(1234, 5678)}
			if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(source)) <= uintptr(^uint32(0)) {
				t.Skipf("allocator returned a low address: %p", source)
			}
			particle := &client.Drawable{ZVal: 99}
			particle.UnionEffect().Field_108 = 0xaabbccdd
			var riseMarker byte
			riseUpdate := unsafe.Pointer(&riseMarker)
			var calls []string
			randomCall := 0
			hooks := cloudDrawableHooks4CE200{
				frame: func() uint32 {
					calls = append(calls, "frame")
					return 101
				},
				typeID: func(name string) int {
					calls = append(calls, "type:"+name)
					switch name {
					case "GreenPuff":
						return 40
					case "GreenSmoke":
						return 41
					default:
						t.Fatalf("unexpected type %q", name)
						return 0
					}
				},
				random: func(min, max int) int {
					randomCall++
					calls = append(calls, fmt.Sprintf("random:%d:%d", min, max))
					switch randomCall {
					case 1:
						if min != 0 || max != 255 {
							t.Fatalf("angle range = (%d, %d)", min, max)
						}
						return 64
					case 2:
						if min != 0 || max != tc.radius {
							t.Fatalf("radius range = (%d, %d), want (0, %d)", min, max, tc.radius)
						}
						return tc.radius
					case 3:
						if min != 0 || max != 10 {
							t.Fatalf("sprite choice range = (%d, %d)", min, max)
						}
						return 2
					case 4:
						if min != 1 || max != 3 {
							t.Fatalf("rise range = (%d, %d)", min, max)
						}
						return 3
					case 5:
						if min != 10 || max != 32 {
							t.Fatalf("lifetime range = (%d, %d)", min, max)
						}
						return 22
					default:
						t.Fatalf("unexpected random call %d", randomCall)
						return 0
					}
				},
				spawn: func(typ int, pos image.Point) *client.Drawable {
					calls = append(calls, "spawn")
					dir := sincosTable16[64]
					wantPos := source.PosVec.Add(image.Pt(tc.radius*dir.X/16, tc.radius*dir.Y/16))
					if typ != 41 || pos != wantPos {
						t.Fatalf("spawn = type %d at %v, want type 41 at %v", typ, pos, wantPos)
					}
					return particle
				},
				activate: func(got *client.Drawable) {
					calls = append(calls, "activate")
					if got != particle || got.ZVal != 0 {
						t.Fatalf("activated particle = %p, z = %d", got, got.ZVal)
					}
				},
				riseUpdate: riseUpdate,
				decay: func(got *client.Drawable, lifetime int) {
					calls = append(calls, "decay")
					if got != particle || lifetime != 22 {
						t.Fatalf("decay = %p, %d", got, lifetime)
					}
				},
				updateList: func(got *client.Drawable) {
					calls = append(calls, "update-list")
					if got != particle {
						t.Fatalf("update-list particle = %p", got)
					}
				},
				sightDestroy: func(got *client.Drawable) {
					calls = append(calls, "sight-destroy")
					if got != particle {
						t.Fatalf("sight-destroy particle = %p", got)
					}
				},
			}
			if got := updateCloudDrawable4CE1D0(source, tc.radius, hooks); got != 1 {
				t.Fatalf("update = %d, want 1", got)
			}
			if particle.ZVal != 0 || particle.UnionEffect().Field_108 != 0xaabbcc03 || particle.Field_115 != riseUpdate {
				t.Fatalf("particle state = z %d, union %#x, update %p", particle.ZVal, particle.UnionEffect().Field_108, particle.Field_115)
			}
			if randomCall != 5 {
				t.Fatalf("random calls = %d, want 5", randomCall)
			}
			wantCalls := []string{
				"frame", "type:GreenPuff", "type:GreenSmoke", "random:0:255",
				fmt.Sprintf("random:0:%d", tc.radius), "random:0:10", "spawn", "activate",
				"random:1:3", "random:10:32", "decay", "update-list", "sight-destroy",
			}
			if fmt.Sprint(calls) != fmt.Sprint(wantCalls) {
				t.Fatalf("calls = %v, want %v", calls, wantCalls)
			}
		})
	}
}

func TestUpdateCloudDrawable4CE1D0EvenFrame(t *testing.T) {
	source := &client.Drawable{}
	called := false
	hooks := cloudDrawableHooks4CE200{
		frame: func() uint32 { return 100 },
		typeID: func(string) int {
			called = true
			return 0
		},
	}
	if got := updateCloudDrawable4CE1D0(source, 75, hooks); got != 1 {
		t.Fatalf("update = %d, want 1", got)
	}
	if called {
		t.Fatal("even frame spawned a cloud particle")
	}
}

func TestUpdateCloudParticleRise4CE340HighAddress(t *testing.T) {
	particle := &client.Drawable{ZVal: 65534}
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(particle)) <= uintptr(^uint32(0)) {
		t.Skipf("allocator returned a low address: %p", particle)
	}
	particle.UnionEffect().Field_108 = 0xaabbcc03
	if got := updateCloudParticleRise4CE340(particle); got != 1 {
		t.Fatalf("update = %d, want 1", got)
	}
	if particle.ZVal != 1 {
		t.Fatalf("z = %d, want wrapped value 1", particle.ZVal)
	}
}
