package server

import (
	"fmt"
	"image"
	"math"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/types"
)

type randomReachablePointTestPoint4ED970 struct {
	name string
	X    float32
	Y    float32
}

type randomReachablePointTestWorld4ED970 struct {
	radius    float32
	centerArg *randomReachablePointTestPoint4ED970
	outputArg *randomReachablePointTestPoint4ED970

	randomResult float64
	cosineResult float64
	sineResult   float64
	traceResults []int32

	events            []string
	faultAt           int
	cosineInputs      []float64
	sineInputs        []float64
	traceRays         []*randomReachablePointRay4ED970
	traceValues       []randomReachablePointRay4ED970
	afterRandom       func(*randomReachablePointTestWorld4ED970)
	afterCosine       func(*randomReachablePointTestWorld4ED970)
	afterSine         func(*randomReachablePointTestWorld4ED970)
	afterTrace        func(*randomReachablePointTestWorld4ED970, int, *randomReachablePointRay4ED970)
	afterStoreOutputX func(*randomReachablePointTestWorld4ED970)
}

func randomReachablePointTestName4ED970(point *randomReachablePointTestPoint4ED970) string {
	if point == nil {
		return "nil"
	}
	return point.name
}

func randomReachablePointTestFloat32_4ED970(value float32) string {
	return fmt.Sprintf("%08x", math.Float32bits(value))
}

func randomReachablePointTestFloat64_4ED970(value float64) string {
	return fmt.Sprintf("%016x", math.Float64bits(value))
}

func randomReachablePointTestPointValue4ED970(point types.Pointf) string {
	return randomReachablePointTestFloat32_4ED970(point.X) + "," +
		randomReachablePointTestFloat32_4ED970(point.Y)
}

func (w *randomReachablePointTestWorld4ED970) event(name string) {
	w.events = append(w.events, name)
	if w.faultAt != 0 && len(w.events) == w.faultAt {
		panic(name)
	}
}

func newRandomReachablePointTestWorld4ED970() *randomReachablePointTestWorld4ED970 {
	return &randomReachablePointTestWorld4ED970{
		radius: 8,
		centerArg: &randomReachablePointTestPoint4ED970{
			name: "center",
			X:    1,
			Y:    2,
		},
		outputArg:    &randomReachablePointTestPoint4ED970{name: "output-a"},
		randomResult: 0.25 + 0x1p-40,
		cosineResult: 0.5,
		sineResult:   -0.25,
		traceResults: []int32{1},
	}
}

func (w *randomReachablePointTestWorld4ED970) hooks() randomReachablePointHooks4ED970[
	*randomReachablePointTestPoint4ED970,
	*randomReachablePointTestPoint4ED970,
] {
	return randomReachablePointHooks4ED970[
		*randomReachablePointTestPoint4ED970,
		*randomReachablePointTestPoint4ED970,
	]{
		loadRadiusArg: func() float32 {
			w.event("radius:" + randomReachablePointTestFloat32_4ED970(w.radius))
			return w.radius
		},
		loadCenterArg: func() *randomReachablePointTestPoint4ED970 {
			w.event("center-arg:" + randomReachablePointTestName4ED970(w.centerArg))
			return w.centerArg
		},
		loadCenterX: func(center *randomReachablePointTestPoint4ED970) float32 {
			w.event("center-x:" + randomReachablePointTestName4ED970(center) + ":" + randomReachablePointTestFloat32_4ED970(center.X))
			return center.X
		},
		loadCenterY: func(center *randomReachablePointTestPoint4ED970) float32 {
			w.event("center-y:" + randomReachablePointTestName4ED970(center) + ":" + randomReachablePointTestFloat32_4ED970(center.Y))
			return center.Y
		},
		randomFloat: func(min, max float32, source string, line int32) float64 {
			w.event(fmt.Sprintf(
				"random:%s:%s:%d:%s",
				randomReachablePointTestFloat32_4ED970(min),
				randomReachablePointTestFloat32_4ED970(max),
				line,
				source,
			))
			if w.afterRandom != nil {
				w.afterRandom(w)
			}
			return w.randomResult
		},
		cosine: func(value float64) float64 {
			w.event("cos:" + randomReachablePointTestFloat64_4ED970(value))
			w.cosineInputs = append(w.cosineInputs, value)
			if w.afterCosine != nil {
				w.afterCosine(w)
			}
			return w.cosineResult
		},
		sine: func(value float64) float64 {
			w.event("sin:" + randomReachablePointTestFloat64_4ED970(value))
			w.sineInputs = append(w.sineInputs, value)
			if w.afterSine != nil {
				w.afterSine(w)
			}
			return w.sineResult
		},
		mapTrace: func(ray *randomReachablePointRay4ED970, outPoint *types.Pointf, outGrid *image.Point, flags uint8) int32 {
			if outPoint != nil || outGrid != nil {
				panic("non-nil-trace-output")
			}
			index := len(w.traceRays)
			w.event(fmt.Sprintf(
				"trace:%d:%s->%s:%d",
				index+1,
				randomReachablePointTestPointValue4ED970(ray.Origin),
				randomReachablePointTestPointValue4ED970(ray.Destination),
				flags,
			))
			w.traceRays = append(w.traceRays, ray)
			w.traceValues = append(w.traceValues, *ray)
			if w.afterTrace != nil {
				w.afterTrace(w, index+1, ray)
			}
			if index < len(w.traceResults) {
				return w.traceResults[index]
			}
			return 0
		},
		loadOutputArg: func() *randomReachablePointTestPoint4ED970 {
			w.event("output-arg:" + randomReachablePointTestName4ED970(w.outputArg))
			return w.outputArg
		},
		storeOutputX: func(output *randomReachablePointTestPoint4ED970, value float32) {
			w.event("output-x:" + randomReachablePointTestName4ED970(output) + ":" + randomReachablePointTestFloat32_4ED970(value))
			output.X = value
			if w.afterStoreOutputX != nil {
				w.afterStoreOutputX(w)
			}
		},
		storeOutputY: func(output *randomReachablePointTestPoint4ED970, value float32) {
			w.event("output-y:" + randomReachablePointTestName4ED970(output) + ":" + randomReachablePointTestFloat32_4ED970(value))
			output.Y = value
		},
	}
}

func verifyRandomReachablePointFaultPrefixes4ED970(
	t *testing.T,
	want []string,
	build func() *randomReachablePointTestWorld4ED970,
) {
	t.Helper()
	for faultAt := 1; faultAt <= len(want); faultAt++ {
		t.Run(fmt.Sprintf("fault-%d", faultAt), func(t *testing.T) {
			w := build()
			w.faultAt = faultAt
			defer func() {
				if got := recover(); got != want[faultAt-1] {
					t.Fatalf("panic = %v, want %q", got, want[faultAt-1])
				}
				if !reflect.DeepEqual(w.events, want[:faultAt]) {
					t.Fatalf("events = %v, want %v", w.events, want[:faultAt])
				}
			}()
			_ = randomReachablePoint4ED970(w.hooks())
		})
	}
}

func TestRandomReachablePoint4ED970ConstantsAndLocalLayout(t *testing.T) {
	if got := math.Float32bits(randomReachablePointPi4ED970); got != 0x40490fdb {
		t.Fatalf("pi bits = %#08x, want 0x40490fdb", got)
	}
	if got := math.Float32bits(randomReachablePointAngleStep4ED970); got != 0x3ff1463b {
		t.Fatalf("angle step bits = %#08x, want 0x3ff1463b", got)
	}
	if got := math.Float32bits(randomReachablePointRadiusStep4ED970); got != 0x3c800000 {
		t.Fatalf("radius step bits = %#08x, want 0x3c800000", got)
	}
	if randomReachablePointAttemptLimit4ED970 != 64 || randomReachablePointSourceLine4ED970 != 728 {
		t.Fatalf("attempts/line = %d/%d, want 64/728", randomReachablePointAttemptLimit4ED970, randomReachablePointSourceLine4ED970)
	}
	if randomReachablePointSource4ED970 != `C:\NoxPost\src\Server\Object\pickdrop\drop.c` {
		t.Fatalf("source path = %q", randomReachablePointSource4ED970)
	}
	var ray randomReachablePointRay4ED970
	if got := unsafe.Sizeof(ray); got != 16 {
		t.Fatalf("ray size = %d, want 16", got)
	}
	if got := unsafe.Offsetof(ray.Origin); got != 0 {
		t.Fatalf("origin offset = %d, want 0", got)
	}
	if got := unsafe.Offsetof(ray.Destination); got != 8 {
		t.Fatalf("destination offset = %d, want 8", got)
	}
}

func TestRandomReachablePoint4ED970SuccessOrderSpillAndOutputIdentity(t *testing.T) {
	build := func() *randomReachablePointTestWorld4ED970 {
		w := newRandomReachablePointTestWorld4ED970()
		outputB := &randomReachablePointTestPoint4ED970{name: "output-b"}
		w.traceResults = []int32{math.MinInt32}
		w.afterCosine = func(w *randomReachablePointTestWorld4ED970) {
			w.centerArg.X = 10
		}
		w.afterSine = func(w *randomReachablePointTestWorld4ED970) {
			w.centerArg.Y = 20
		}
		w.afterTrace = func(w *randomReachablePointTestWorld4ED970, _ int, ray *randomReachablePointRay4ED970) {
			ray.Destination = types.Pointf{X: 31, Y: 41}
			w.outputArg = outputB
		}
		return w
	}

	randomSpill := float32(0.25 + 0x1p-40)
	cosineInput := randomReachablePointAdd64_4ED970(
		float64(randomSpill),
		float64(randomReachablePointAngleStep4ED970),
	)
	sineInput := float64(float32(cosineInput))
	want := []string{
		"radius:41000000",
		"center-arg:center",
		"center-x:center:3f800000",
		"center-y:center:40000000",
		"random:c0490fdb:40490fdb:728:" + randomReachablePointSource4ED970,
		"cos:" + randomReachablePointTestFloat64_4ED970(cosineInput),
		"center-x:center:41200000",
		"sin:" + randomReachablePointTestFloat64_4ED970(sineInput),
		"center-y:center:41a00000",
		"trace:1:3f800000,40000000->41600000,41900000:1",
		"output-arg:output-b",
		"output-x:output-b:41f80000",
		"output-y:output-b:42240000",
	}

	w := build()
	outputA := w.outputArg
	got := randomReachablePoint4ED970(w.hooks())
	if got != w.outputArg || got == outputA {
		t.Fatalf("result = %p, delayed output = %p, entry output = %p", got, w.outputArg, outputA)
	}
	if got.X != 31 || got.Y != 41 {
		t.Fatalf("output = (%v,%v), want (31,41)", got.X, got.Y)
	}
	if outputA.X != 0 || outputA.Y != 0 {
		t.Fatalf("entry output changed to (%v,%v)", outputA.X, outputA.Y)
	}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %v, want %v", w.events, want)
	}
	if len(w.cosineInputs) != 1 || len(w.sineInputs) != 1 || w.cosineInputs[0] == w.sineInputs[0] {
		t.Fatalf("cos/sin inputs = %v/%v; want distinct unspilled/spilled values", w.cosineInputs, w.sineInputs)
	}
	verifyRandomReachablePointFaultPrefixes4ED970(t, want, build)
}

func TestRandomReachablePoint4ED970FailureKeepsRayAndReloadsCenter(t *testing.T) {
	w := newRandomReachablePointTestWorld4ED970()
	w.radius = 64
	w.randomResult = 0.1
	w.cosineResult = 1
	w.sineResult = 0
	w.traceResults = []int32{0, -1}
	w.afterTrace = func(w *randomReachablePointTestWorld4ED970, attempt int, ray *randomReachablePointRay4ED970) {
		if attempt == 1 {
			ray.Origin = types.Pointf{X: 9, Y: 10}
			w.centerArg.X = 3
			w.centerArg.Y = 4
		}
	}

	got := randomReachablePoint4ED970(w.hooks())
	if got != w.outputArg || got.X != 66 || got.Y != 4 {
		t.Fatalf("result = %p (%v,%v), want output (66,4)", got, got.X, got.Y)
	}
	if len(w.traceValues) != 2 || len(w.traceRays) != 2 {
		t.Fatalf("trace count = %d/%d, want 2", len(w.traceValues), len(w.traceRays))
	}
	if w.traceRays[0] != w.traceRays[1] {
		t.Fatalf("ray identity changed from %p to %p", w.traceRays[0], w.traceRays[1])
	}
	if got := w.traceValues[0]; got.Origin != (types.Pointf{X: 1, Y: 2}) || got.Destination != (types.Pointf{X: 65, Y: 2}) {
		t.Fatalf("first trace = %+v", got)
	}
	if got := w.traceValues[1]; got.Origin != (types.Pointf{X: 9, Y: 10}) || got.Destination != (types.Pointf{X: 66, Y: 4}) {
		t.Fatalf("second trace = %+v", got)
	}
}

func TestRandomReachablePoint4ED970KeepsTrigPrecisionUntilDestinationSpill(t *testing.T) {
	w := newRandomReachablePointTestWorld4ED970()
	w.radius = 0x1p24
	w.centerArg.X = 1
	w.centerArg.Y = 1
	w.cosineResult = 1 + 0x1p-24
	w.sineResult = 1 + 0x1p-24
	w.traceResults = []int32{1}

	got := randomReachablePoint4ED970(w.hooks())
	if x, y := math.Float32bits(got.X), math.Float32bits(got.Y); x != 0x4b800001 || y != 0x4b800001 {
		t.Fatalf("output bits = %#08x/%#08x, want 0x4b800001/0x4b800001", x, y)
	}
	premature := float32(float32(w.cosineResult)*w.radius + w.centerArg.X)
	if math.Float32bits(premature) != 0x4b800000 {
		t.Fatalf("control path bits = %#08x, want 0x4b800000", math.Float32bits(premature))
	}
}

func TestRandomReachablePoint4ED970SixtyFourFailuresUseLiveAliasedFallback(t *testing.T) {
	w := newRandomReachablePointTestWorld4ED970()
	w.radius = 64
	w.centerArg.X = 0
	w.centerArg.Y = 0
	w.outputArg = w.centerArg
	w.randomResult = 0
	w.cosineResult = 1
	w.sineResult = 0
	w.traceResults = nil
	w.afterTrace = func(w *randomReachablePointTestWorld4ED970, attempt int, _ *randomReachablePointRay4ED970) {
		if attempt == randomReachablePointAttemptLimit4ED970 {
			w.centerArg.X = 30
			w.centerArg.Y = 40
		}
	}
	w.afterStoreOutputX = func(w *randomReachablePointTestWorld4ED970) {
		w.centerArg.Y = 50
	}

	got := randomReachablePoint4ED970(w.hooks())
	if got != w.centerArg || got.X != 30 || got.Y != 50 {
		t.Fatalf("fallback = %p (%v,%v), want aliased center (30,50)", got, got.X, got.Y)
	}
	if len(w.traceValues) != randomReachablePointAttemptLimit4ED970 {
		t.Fatalf("trace count = %d, want %d", len(w.traceValues), randomReachablePointAttemptLimit4ED970)
	}
	if first, last := w.traceValues[0].Destination.X, w.traceValues[len(w.traceValues)-1].Destination.X; first != 64 || last != 1 {
		t.Fatalf("candidate radius sequence endpoints = %v/%v, want 64/1", first, last)
	}
	if len(w.cosineInputs) != 64 || len(w.sineInputs) != 64 {
		t.Fatalf("trig counts = %d/%d, want 64/64", len(w.cosineInputs), len(w.sineInputs))
	}
	wantTail := []string{
		"output-arg:center",
		"center-x:center:41f00000",
		"output-x:center:41f00000",
		"center-y:center:42480000",
		"output-y:center:42480000",
	}
	if gotTail := w.events[len(w.events)-len(wantTail):]; !reflect.DeepEqual(gotTail, wantTail) {
		t.Fatalf("fallback tail = %v, want %v", gotTail, wantTail)
	}
}

func TestRandomReachablePoint4ED970AcceptsEveryWholeNonzeroTraceResult(t *testing.T) {
	for _, result := range []int32{-1, 1, math.MaxInt32, math.MinInt32, 0x76543210} {
		t.Run(fmt.Sprintf("%08x", uint32(result)), func(t *testing.T) {
			w := newRandomReachablePointTestWorld4ED970()
			w.traceResults = []int32{result}
			got := randomReachablePoint4ED970(w.hooks())
			if got != w.outputArg || len(w.traceValues) != 1 {
				t.Fatalf("result pointer/trace count = %p/%d, want %p/1", got, len(w.traceValues), w.outputArg)
			}
		})
	}
}
