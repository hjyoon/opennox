package server

import (
	"image"
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/libs/types"
)

func TestRandomReachablePointNative4ED970Layout(t *testing.T) {
	checks := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"Pointf size", unsafe.Sizeof(types.Pointf{}), 8},
		{"Pointf.X", unsafe.Offsetof(types.Pointf{}.X), 0},
		{"Pointf.Y", unsafe.Offsetof(types.Pointf{}.Y), 4},
		{"ray size", unsafe.Sizeof(randomReachablePointRay4ED970{}), 16},
		{"ray origin", unsafe.Offsetof(randomReachablePointRay4ED970{}.Origin), 0},
		{"ray destination", unsafe.Offsetof(randomReachablePointRay4ED970{}.Destination), 8},
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Errorf("%s = %d, want %d", check.name, check.got, check.want)
		}
	}
}

func TestRandomReachablePointNative4ED970UsesLivePointersAndTraceLocal(t *testing.T) {
	center := &types.Pointf{X: 1, Y: 2}
	output := &types.Pointf{X: -1, Y: -2}
	events := make([]string, 0, 4)
	var traceRay *randomReachablePointRay4ED970
	deps := randomReachablePointNativeDeps4ED970{
		randomFloat: func(min, max float32, source string, line int32) float64 {
			events = append(events, "random")
			if math.Float32bits(min) != 0xc0490fdb || math.Float32bits(max) != 0x40490fdb ||
				source != randomReachablePointSource4ED970 || line != 728 {
				t.Fatalf("random boundary = %08x/%08x/%q/%d", math.Float32bits(min), math.Float32bits(max), source, line)
			}
			return 0.25
		},
		cosine: func(float64) float64 {
			events = append(events, "cos")
			center.X = 10
			return 0.5
		},
		sine: func(float64) float64 {
			events = append(events, "sin")
			center.Y = 20
			return -0.25
		},
		mapTrace: func(ray *randomReachablePointRay4ED970, outPoint *types.Pointf, outGrid *image.Point, flags uint8) int32 {
			events = append(events, "trace")
			traceRay = ray
			if outPoint != nil || outGrid != nil || flags != 1 {
				t.Fatalf("trace optional outputs/flags = %p/%p/%d", outPoint, outGrid, flags)
			}
			if ray.Origin != (types.Pointf{X: 1, Y: 2}) || ray.Destination != (types.Pointf{X: 14, Y: 18}) {
				t.Fatalf("trace ray = %+v", *ray)
			}
			ray.Destination = types.Pointf{X: 31, Y: 41}
			return math.MinInt32
		},
	}

	got := randomReachablePointNative4ED970(8, center, output, deps)
	if got != output || *got != (types.Pointf{X: 31, Y: 41}) {
		t.Fatalf("result = %p %+v, want %p {31 41}", got, *got, output)
	}
	if traceRay == nil || unsafe.Pointer(traceRay) == unsafe.Pointer(center) || unsafe.Pointer(traceRay) == unsafe.Pointer(output) {
		t.Fatalf("trace local identity = %p, center/output = %p/%p", traceRay, center, output)
	}
	want := []string{"random", "cos", "sin", "trace"}
	if len(events) != len(want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
	for i := range want {
		if events[i] != want[i] {
			t.Fatalf("events = %v, want %v", events, want)
		}
	}
}

func TestRandomReachablePointNative4ED970FallsBackToAliasedLiveCenter(t *testing.T) {
	center := &types.Pointf{}
	attempts := 0
	deps := randomReachablePointNativeDeps4ED970{
		randomFloat: func(float32, float32, string, int32) float64 { return 0 },
		cosine:      func(float64) float64 { return 1 },
		sine:        func(float64) float64 { return 0 },
		mapTrace: func(*randomReachablePointRay4ED970, *types.Pointf, *image.Point, uint8) int32 {
			attempts++
			if attempts == 64 {
				center.X = 30
				center.Y = 40
			}
			return 0
		},
	}
	got := randomReachablePointNative4ED970(64, center, center, deps)
	if attempts != 64 || got != center || *got != (types.Pointf{X: 30, Y: 40}) {
		t.Fatalf("attempts/result = %d/%p %+v, want 64/%p {30 40}", attempts, got, *got, center)
	}
}
