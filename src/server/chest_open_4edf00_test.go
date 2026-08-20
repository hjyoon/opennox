package server

import (
	"fmt"
	"image"
	"math"
	"reflect"
	"testing"

	"github.com/opennox/libs/types"
)

type chestOpenTestObject4EDF00 int

const (
	chestOpenTestNil4EDF00   chestOpenTestObject4EDF00 = 0
	chestOpenTestChest4EDF00 chestOpenTestObject4EDF00 = 1
	chestOpenTestUnit4EDF00  chestOpenTestObject4EDF00 = 2
	chestOpenTestItemA4EDF00 chestOpenTestObject4EDF00 = 10
	chestOpenTestItemB4EDF00 chestOpenTestObject4EDF00 = 11
	chestOpenTestItemC4EDF00 chestOpenTestObject4EDF00 = 12
	chestOpenTestItemD4EDF00 chestOpenTestObject4EDF00 = 13
	chestOpenTestItemE4EDF00 chestOpenTestObject4EDF00 = 14
	chestOpenTestItemF4EDF00 chestOpenTestObject4EDF00 = 15
)

func chestOpenTestObjectName4EDF00(obj chestOpenTestObject4EDF00) string {
	switch obj {
	case chestOpenTestNil4EDF00:
		return "nil"
	case chestOpenTestChest4EDF00:
		return "chest"
	case chestOpenTestUnit4EDF00:
		return "unit"
	case chestOpenTestItemA4EDF00:
		return "item-a"
	case chestOpenTestItemB4EDF00:
		return "item-b"
	case chestOpenTestItemC4EDF00:
		return "item-c"
	case chestOpenTestItemD4EDF00:
		return "item-d"
	case chestOpenTestItemE4EDF00:
		return "item-e"
	case chestOpenTestItemF4EDF00:
		return "item-f"
	default:
		return fmt.Sprintf("object-%d", obj)
	}
}

func chestOpenFloat4EDF00(value float32) string {
	return fmt.Sprintf("%08x", math.Float32bits(value))
}

func chestOpenPoint4EDF00(value types.Pointf) string {
	return chestOpenFloat4EDF00(value.X) + "," + chestOpenFloat4EDF00(value.Y)
}

type chestOpenTestWorld4EDF00 struct {
	chestArg   chestOpenTestObject4EDF00
	unitArg    chestOpenTestObject4EDF00
	count      int32
	subClass   map[chestOpenTestObject4EDF00]uint32
	position   map[chestOpenTestObject4EDF00]types.Pointf
	extent     float64
	first      chestOpenTestObject4EDF00
	next       map[chestOpenTestObject4EDF00]chestOpenTestObject4EDF00
	weight     map[chestOpenTestObject4EDF00]uint8
	classLow   map[chestOpenTestObject4EDF00]uint8
	flags      map[chestOpenTestObject4EDF00]uint32
	traceValue []int32

	events           []string
	faultAt          int
	traceCalls       int
	traceRay         *chestOpenRay4EDF00
	traceInputs      []chestOpenRay4EDF00
	dropPoints       []*types.Pointf
	dropInputs       []types.Pointf
	mutateTrace      bool
	mutateFirstDrop  bool
	normalizedResult *types.Pointf
}

func newChestOpenTestWorld4EDF00() *chestOpenTestWorld4EDF00 {
	return &chestOpenTestWorld4EDF00{
		chestArg: chestOpenTestChest4EDF00,
		unitArg:  chestOpenTestUnit4EDF00,
		count:    6,
		subClass: map[chestOpenTestObject4EDF00]uint32{
			chestOpenTestChest4EDF00: chestOpenDirectionSouthEast4EDF00,
		},
		position: map[chestOpenTestObject4EDF00]types.Pointf{
			chestOpenTestChest4EDF00: {X: 10, Y: 20},
			chestOpenTestUnit4EDF00:  {X: 100, Y: 200},
		},
		extent: 1,
		first:  chestOpenTestItemA4EDF00,
		next: map[chestOpenTestObject4EDF00]chestOpenTestObject4EDF00{
			chestOpenTestItemA4EDF00: chestOpenTestItemB4EDF00,
			chestOpenTestItemB4EDF00: chestOpenTestItemC4EDF00,
			chestOpenTestItemC4EDF00: chestOpenTestItemD4EDF00,
			chestOpenTestItemD4EDF00: chestOpenTestItemE4EDF00,
			chestOpenTestItemE4EDF00: chestOpenTestItemF4EDF00,
			chestOpenTestItemF4EDF00: chestOpenTestNil4EDF00,
		},
		weight: map[chestOpenTestObject4EDF00]uint8{
			chestOpenTestItemA4EDF00: 1,
			chestOpenTestItemB4EDF00: chestOpenInvalidWeight4EDF00,
			chestOpenTestItemC4EDF00: 2,
			chestOpenTestItemD4EDF00: 3,
			chestOpenTestItemE4EDF00: 4,
			chestOpenTestItemF4EDF00: 5,
		},
		classLow: map[chestOpenTestObject4EDF00]uint8{
			chestOpenTestItemC4EDF00: chestOpenMonsterClass4EDF00,
		},
		flags: map[chestOpenTestObject4EDF00]uint32{
			chestOpenTestItemA4EDF00: 0x00000120,
			chestOpenTestItemD4EDF00: 0x80000001,
			chestOpenTestItemE4EDF00: 0xffffffbf,
			chestOpenTestItemF4EDF00: 0,
		},
		traceValue: []int32{-1, 0, math.MinInt32},
	}
}

func (w *chestOpenTestWorld4EDF00) event(value string) {
	w.events = append(w.events, value)
	if w.faultAt != 0 && len(w.events) == w.faultAt {
		panic(value)
	}
}

func (w *chestOpenTestWorld4EDF00) hooks() chestOpenHooks4EDF00[chestOpenTestObject4EDF00] {
	return chestOpenHooks4EDF00[chestOpenTestObject4EDF00]{
		loadChestArg: func() chestOpenTestObject4EDF00 {
			w.event("chest-arg:" + chestOpenTestObjectName4EDF00(w.chestArg))
			return w.chestArg
		},
		loadUnitArg: func() chestOpenTestObject4EDF00 {
			w.event("unit-arg:" + chestOpenTestObjectName4EDF00(w.unitArg))
			return w.unitArg
		},
		countInventory: func(chest chestOpenTestObject4EDF00, query int32) int32 {
			w.event(fmt.Sprintf("count:%s:%08x=%08x", chestOpenTestObjectName4EDF00(chest), uint32(query), uint32(w.count)))
			return w.count
		},
		loadSubClass: func(obj chestOpenTestObject4EDF00) uint32 {
			value := w.subClass[obj]
			w.event(fmt.Sprintf("subclass:%s=%08x", chestOpenTestObjectName4EDF00(obj), value))
			return value
		},
		loadPosX: func(obj chestOpenTestObject4EDF00) float32 {
			value := w.position[obj].X
			w.event("x:" + chestOpenTestObjectName4EDF00(obj) + "=" + chestOpenFloat4EDF00(value))
			return value
		},
		loadPosY: func(obj chestOpenTestObject4EDF00) float32 {
			value := w.position[obj].Y
			w.event("y:" + chestOpenTestObjectName4EDF00(obj) + "=" + chestOpenFloat4EDF00(value))
			return value
		},
		normalize: func(direction *types.Pointf) {
			w.event("normalize:" + chestOpenPoint4EDF00(*direction))
			if w.normalizedResult != nil {
				*direction = *w.normalizedResult
			}
		},
		shapeExtent: func(obj chestOpenTestObject4EDF00) float64 {
			w.event(fmt.Sprintf("extent:%s=%016x", chestOpenTestObjectName4EDF00(obj), math.Float64bits(w.extent)))
			return w.extent
		},
		mapTrace: func(ray *chestOpenRay4EDF00, point *types.Pointf, grid *image.Point, flag uint8) int32 {
			call := w.traceCalls
			value := int32(1)
			if call < len(w.traceValue) {
				value = w.traceValue[call]
			}
			w.event(fmt.Sprintf(
				"trace:%d:%s->%s:%t:%t:%02x=%08x",
				call,
				chestOpenPoint4EDF00(ray.Origin),
				chestOpenPoint4EDF00(ray.Destination),
				point == nil,
				grid == nil,
				flag,
				uint32(value),
			))
			if w.traceRay == nil {
				w.traceRay = ray
			} else if w.traceRay != ray {
				panic("trace ray identity changed")
			}
			w.traceInputs = append(w.traceInputs, *ray)
			if w.mutateTrace {
				ray.Origin.X++
				ray.Origin.Y += 2
				ray.Destination.X += 1000
				ray.Destination.Y += 2000
				w.position[chestOpenTestUnit4EDF00] = types.Pointf{
					X: 101 + float32(call),
					Y: 201 + float32(call),
				}
			}
			w.traceCalls++
			return value
		},
		firstItem: func(chest chestOpenTestObject4EDF00) chestOpenTestObject4EDF00 {
			w.event("first:" + chestOpenTestObjectName4EDF00(chest) + "=" + chestOpenTestObjectName4EDF00(w.first))
			return w.first
		},
		nextItem: func(item chestOpenTestObject4EDF00) chestOpenTestObject4EDF00 {
			value := w.next[item]
			w.event("next:" + chestOpenTestObjectName4EDF00(item) + "=" + chestOpenTestObjectName4EDF00(value))
			return value
		},
		loadWeight: func(item chestOpenTestObject4EDF00) uint8 {
			value := w.weight[item]
			w.event(fmt.Sprintf("weight:%s=%02x", chestOpenTestObjectName4EDF00(item), value))
			return value
		},
		loadClassLow: func(item chestOpenTestObject4EDF00) uint8 {
			value := w.classLow[item]
			w.event(fmt.Sprintf("class:%s=%02x", chestOpenTestObjectName4EDF00(item), value))
			return value
		},
		loadFlags: func(item chestOpenTestObject4EDF00) uint32 {
			value := w.flags[item]
			w.event(fmt.Sprintf("flags:%s=%08x", chestOpenTestObjectName4EDF00(item), value))
			return value
		},
		storeFlags: func(item chestOpenTestObject4EDF00, value uint32) {
			w.event(fmt.Sprintf("store-flags:%s=%08x", chestOpenTestObjectName4EDF00(item), value))
			w.flags[item] = value
		},
		refresh: func(item chestOpenTestObject4EDF00) {
			w.event("refresh:" + chestOpenTestObjectName4EDF00(item))
		},
		drop: func(chest, item chestOpenTestObject4EDF00, point *types.Pointf) int32 {
			w.event("drop:" + chestOpenTestObjectName4EDF00(chest) + ":" + chestOpenTestObjectName4EDF00(item) + ":" + chestOpenPoint4EDF00(*point))
			w.dropPoints = append(w.dropPoints, point)
			w.dropInputs = append(w.dropInputs, *point)
			if w.mutateFirstDrop && len(w.dropInputs) == 1 {
				point.X++
				point.Y++
				w.next[item] = chestOpenTestNil4EDF00
			}
			return math.MinInt32
		},
	}
}

func verifyChestOpenFaultPrefixes4EDF00(
	t *testing.T,
	want []string,
	build func() *chestOpenTestWorld4EDF00,
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
			chestOpen4EDF00(w.hooks())
		})
	}
}

func TestChestShapeExtent4EE2A0ExactBranchesReloadsAndSignedZero(t *testing.T) {
	t.Run("circle", func(t *testing.T) {
		var events []string
		got := chestShapeExtent4EE2A0(7, chestShapeExtentHooks4EE2A0[int]{
			loadShapeKind: func(obj int) uint32 {
				events = append(events, fmt.Sprintf("kind:%d", obj))
				return chestShapeCircle4EE2A0
			},
			loadCircleR: func(obj int) float32 {
				events = append(events, fmt.Sprintf("radius:%d", obj))
				return 12.5
			},
		})
		if got != 12.5 || !reflect.DeepEqual(events, []string{"kind:7", "radius:7"}) {
			t.Fatalf("result/events = (%v, %v)", got, events)
		}
	})

	t.Run("other-positive-zero", func(t *testing.T) {
		var events []string
		got := chestShapeExtent4EE2A0(8, chestShapeExtentHooks4EE2A0[int]{
			loadShapeKind: func(obj int) uint32 {
				events = append(events, fmt.Sprintf("kind:%d", obj))
				return math.MaxUint32
			},
			loadCircleR: func(int) float32 { panic("unexpected circle load") },
			loadBoxExtentW: func(int) float32 {
				panic("unexpected box W load")
			},
			loadBoxExtentH: func(int) float32 {
				panic("unexpected box H load")
			},
		})
		if math.Float64bits(got) != 0 || !reflect.DeepEqual(events, []string{"kind:8"}) {
			t.Fatalf("result bits/events = (%016x, %v)", math.Float64bits(got), events)
		}
	})

	tests := []struct {
		name       string
		first      float32
		second     float32
		reloadW    float32
		reloadH    float32
		want       float64
		wantEvents []string
	}{
		{name: "ordered-greater-reloads-W", first: 8, second: 4, reloadW: 10, reloadH: 20, want: 5, wantEvents: []string{"kind", "W:1", "H:1", "W:2"}},
		{name: "less-reloads-H", first: 2, second: 4, reloadW: 10, reloadH: 14, want: 7, wantEvents: []string{"kind", "W:1", "H:1", "H:2"}},
		{name: "equal-reloads-H", first: 4, second: 4, reloadW: 10, reloadH: 18, want: 9, wantEvents: []string{"kind", "W:1", "H:1", "H:2"}},
		{name: "first-unordered-reloads-H", first: float32(math.NaN()), second: 4, reloadW: 10, reloadH: 22, want: 11, wantEvents: []string{"kind", "W:1", "H:1", "H:2"}},
		{name: "second-unordered-reloads-H", first: 8, second: float32(math.NaN()), reloadW: 10, reloadH: 24, want: 12, wantEvents: []string{"kind", "W:1", "H:1", "H:2"}},
		{name: "negative-zero-second", first: 0, second: math.Float32frombits(1 << 31), reloadW: 1, reloadH: math.Float32frombits(1 << 31), want: math.Float64frombits(1 << 63), wantEvents: []string{"kind", "W:1", "H:1", "H:2"}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var events []string
			wLoads, hLoads := 0, 0
			got := chestShapeExtent4EE2A0(9, chestShapeExtentHooks4EE2A0[int]{
				loadShapeKind: func(int) uint32 {
					events = append(events, "kind")
					return chestShapeBox4EE2A0
				},
				loadBoxExtentW: func(int) float32 {
					wLoads++
					events = append(events, fmt.Sprintf("W:%d", wLoads))
					if wLoads == 1 {
						return tc.first
					}
					return tc.reloadW
				},
				loadBoxExtentH: func(int) float32 {
					hLoads++
					events = append(events, fmt.Sprintf("H:%d", hLoads))
					if hLoads == 1 {
						return tc.second
					}
					return tc.reloadH
				},
			})
			if math.Float64bits(got) != math.Float64bits(tc.want) || !reflect.DeepEqual(events, tc.wantEvents) {
				t.Fatalf("result/events = (%016x, %v), want (%016x, %v)", math.Float64bits(got), events, math.Float64bits(tc.want), tc.wantEvents)
			}
		})
	}
}

func TestChestOpenDirection4EDF00BitPriorityAndFallbackOrder(t *testing.T) {
	bits := []struct {
		name  string
		value uint32
		want  types.Pointf
	}{
		{name: "100", value: 0x100, want: types.Pointf{X: -1, Y: -1}},
		{name: "200", value: 0x200, want: types.Pointf{X: 1, Y: -1}},
		{name: "400", value: 0x400, want: types.Pointf{X: 1, Y: 1}},
		{name: "800", value: 0x800, want: types.Pointf{X: -1, Y: 1}},
		{name: "priority", value: 0xf00, want: types.Pointf{X: -1, Y: -1}},
	}
	for _, tc := range bits {
		t.Run(tc.name, func(t *testing.T) {
			var events []string
			hooks := chestOpenHooks4EDF00[int]{
				loadSubClass: func(int) uint32 {
					events = append(events, "subclass")
					return tc.value
				},
				loadPosX: func(int) float32 { panic("unexpected X load") },
				loadPosY: func(int) float32 { panic("unexpected Y load") },
				normalize: func(point *types.Pointf) {
					events = append(events, "normalize:"+chestOpenPoint4EDF00(*point))
				},
			}
			if got := chestOpenDirection4EDF00(1, 2, hooks); got != tc.want {
				t.Fatalf("direction = %+v, want %+v", got, tc.want)
			}
			wantEvents := []string{"subclass", "normalize:" + chestOpenPoint4EDF00(tc.want)}
			if !reflect.DeepEqual(events, wantEvents) {
				t.Fatalf("events = %v, want %v", events, wantEvents)
			}
		})
	}

	var events []string
	negativeZero := math.Float32frombits(1 << 31)
	position := map[int]types.Pointf{
		1: {X: -1, Y: 0},
		2: {X: 16777216, Y: negativeZero},
	}
	hooks := chestOpenHooks4EDF00[int]{
		loadSubClass: func(int) uint32 {
			events = append(events, "subclass")
			return 0
		},
		loadPosX: func(obj int) float32 {
			events = append(events, fmt.Sprintf("X:%d", obj))
			return position[obj].X
		},
		loadPosY: func(obj int) float32 {
			events = append(events, fmt.Sprintf("Y:%d", obj))
			return position[obj].Y
		},
		normalize: func(point *types.Pointf) {
			events = append(events, "normalize:"+chestOpenPoint4EDF00(*point))
		},
	}
	got := chestOpenDirection4EDF00(1, 2, hooks)
	if math.Float32bits(got.X) != math.Float32bits(float32(16777216)) || math.Float32bits(got.Y) != 1<<31 {
		t.Fatalf("fallback direction bits = %s", chestOpenPoint4EDF00(got))
	}
	wantEvents := []string{
		"subclass", "X:2", "X:1", "Y:2", "Y:1",
		"normalize:4b800000,80000000",
	}
	if !reflect.DeepEqual(events, wantEvents) {
		t.Fatalf("events = %v, want %v", events, wantEvents)
	}
}

func TestChestOpenSortCandidates4EDF00PreservesX87PrecisionAndC0Unordered(t *testing.T) {
	tests := []struct {
		name      string
		distance0 float32
		distance1 float64
		distance2 float32
		want      [3]float32
	}{
		{
			name:      "unspilled-distance-one",
			distance0: 1,
			distance1: 1 + math.Ldexp(1, -24),
			distance2: 0.5,
			want:      [3]float32{1, 0, 2},
		},
		{
			name:      "ties-stable",
			distance0: 3,
			distance1: 3,
			distance2: 3,
			want:      [3]float32{0, 1, 2},
		},
		{
			name:      "unordered-uses-C0",
			distance0: float32(math.NaN()),
			distance1: 3,
			distance2: 2,
			want:      [3]float32{1, 2, 0},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			points := [3]types.Pointf{{X: 0}, {X: 1}, {X: 2}}
			chestOpenSortCandidates4EDF00(&points, tc.distance0, tc.distance1, tc.distance2)
			got := [3]float32{points[0].X, points[1].X, points[2].X}
			if got != tc.want {
				t.Fatalf("order = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestChestOpen4EDF00NilAndEmptyGates(t *testing.T) {
	tests := []struct {
		name      string
		configure func(*chestOpenTestWorld4EDF00)
		want      []string
	}{
		{
			name: "nil-chest",
			configure: func(w *chestOpenTestWorld4EDF00) {
				w.chestArg = chestOpenTestNil4EDF00
			},
			want: []string{"chest-arg:nil"},
		},
		{
			name: "nil-unit",
			configure: func(w *chestOpenTestWorld4EDF00) {
				w.unitArg = chestOpenTestNil4EDF00
			},
			want: []string{"chest-arg:chest", "unit-arg:nil"},
		},
		{
			name: "empty",
			configure: func(w *chestOpenTestWorld4EDF00) {
				w.count = 0
			},
			want: []string{"chest-arg:chest", "unit-arg:unit", "count:chest:00000000=00000000"},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			w := newChestOpenTestWorld4EDF00()
			tc.configure(w)
			chestOpen4EDF00(w.hooks())
			if !reflect.DeepEqual(w.events, tc.want) {
				t.Fatalf("events = %v, want %v", w.events, tc.want)
			}
		})
	}
}

func chestOpenFullEvents4EDF00() []string {
	point := func(x, y float32) string { return chestOpenPoint4EDF00(types.Pointf{X: x, Y: y}) }
	return []string{
		"chest-arg:chest",
		"unit-arg:unit",
		"count:chest:00000000=00000006",
		"subclass:chest=00000400",
		"normalize:" + point(1, 1),
		"extent:chest=3ff0000000000000",
		"x:chest=" + chestOpenFloat4EDF00(10),
		"y:chest=" + chestOpenFloat4EDF00(20),
		"x:unit=" + chestOpenFloat4EDF00(100),
		"y:unit=" + chestOpenFloat4EDF00(200),
		"x:unit=" + chestOpenFloat4EDF00(100),
		"y:unit=" + chestOpenFloat4EDF00(200),
		"x:unit=" + chestOpenFloat4EDF00(100),
		"y:unit=" + chestOpenFloat4EDF00(200),
		"y:chest=" + chestOpenFloat4EDF00(20),
		"x:chest=" + chestOpenFloat4EDF00(10),
		"trace:0:" + point(10, 20) + "->" + point(0, 70) + ":true:true:01=ffffffff",
		"trace:1:" + point(11, 22) + "->" + point(30, 40) + ":true:true:01=00000000",
		"x:unit=" + chestOpenFloat4EDF00(102),
		"y:unit=" + chestOpenFloat4EDF00(202),
		"trace:2:" + point(12, 24) + "->" + point(60, 10) + ":true:true:01=80000000",
		"first:chest=item-a",
		"next:item-a=item-b",
		"weight:item-a=01",
		"class:item-a=00",
		"flags:item-a=00000120",
		"store-flags:item-a=00000160",
		"refresh:item-a",
		"drop:chest:item-a:" + point(60, 10),
		"next:item-b=item-c",
		"weight:item-b=ff",
		"next:item-c=item-d",
		"weight:item-c=02",
		"class:item-c=02",
		"next:item-d=item-e",
		"weight:item-d=03",
		"class:item-d=00",
		"flags:item-d=80000001",
		"store-flags:item-d=80000041",
		"refresh:item-d",
		"drop:chest:item-d:" + point(102, 202),
		"next:item-e=item-f",
		"weight:item-e=04",
		"class:item-e=00",
		"flags:item-e=ffffffbf",
		"store-flags:item-e=ffffffff",
		"refresh:item-e",
		"drop:chest:item-e:" + point(0, 70),
		"next:item-f=nil",
		"weight:item-f=05",
		"class:item-f=00",
		"flags:item-f=00000000",
		"store-flags:item-f=00000040",
		"refresh:item-f",
		"drop:chest:item-f:" + point(61, 11),
	}
}

func TestChestOpen4EDF00ExactTraceIterationMutationAndFaultPrefixes(t *testing.T) {
	build := func() *chestOpenTestWorld4EDF00 {
		w := newChestOpenTestWorld4EDF00()
		w.mutateTrace = true
		w.mutateFirstDrop = true
		return w
	}
	want := chestOpenFullEvents4EDF00()
	w := build()
	chestOpen4EDF00(w.hooks())
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %v, want %v", w.events, want)
	}
	if w.traceCalls != 3 || w.traceRay == nil || len(w.traceInputs) != 3 {
		t.Fatalf("trace calls/ray/inputs = (%d, %p, %d)", w.traceCalls, w.traceRay, len(w.traceInputs))
	}
	if got, wantInputs := w.dropInputs, []types.Pointf{{X: 60, Y: 10}, {X: 102, Y: 202}, {X: 0, Y: 70}, {X: 61, Y: 11}}; !reflect.DeepEqual(got, wantInputs) {
		t.Fatalf("drop inputs = %+v, want %+v", got, wantInputs)
	}
	if len(w.dropPoints) != 4 || w.dropPoints[0] != w.dropPoints[3] {
		t.Fatalf("cyclic point identities = %p and %p", w.dropPoints[0], w.dropPoints[3])
	}
	verifyChestOpenFaultPrefixes4EDF00(t, want, build)
}

func TestChestOpen4EDF00CountOneHasNoFirstItemNilGuardAndOtherNonzeroCountsIterate(t *testing.T) {
	t.Run("count-one-nil-first-faults-at-weight", func(t *testing.T) {
		w := newChestOpenTestWorld4EDF00()
		w.count = 1
		w.first = chestOpenTestNil4EDF00
		w.traceValue = []int32{1, 1, 1}
		hooks := w.hooks()
		hooks.loadWeight = func(item chestOpenTestObject4EDF00) uint8 {
			w.event("weight:" + chestOpenTestObjectName4EDF00(item))
			panic("nil-item-dereference")
		}
		defer func() {
			if got := recover(); got != "nil-item-dereference" {
				t.Fatalf("panic = %v", got)
			}
			if got := w.events[len(w.events)-2:]; !reflect.DeepEqual(got, []string{"first:chest=nil", "weight:nil"}) {
				t.Fatalf("last events = %v", got)
			}
		}()
		chestOpen4EDF00(hooks)
	})

	for _, count := range []int32{2, -1, math.MinInt32} {
		t.Run(fmt.Sprintf("other-%08x", uint32(count)), func(t *testing.T) {
			w := newChestOpenTestWorld4EDF00()
			w.count = count
			w.first = chestOpenTestNil4EDF00
			w.traceValue = []int32{1, 1, 1}
			chestOpen4EDF00(w.hooks())
			if got := w.events[len(w.events)-1]; got != "first:chest=nil" {
				t.Fatalf("last event = %q, want multi-item first; all=%v", got, w.events)
			}
		})
	}
}

func TestChestOpenDropItem4EDF00EligibilityOrderFlagORAndIgnoredDropResult(t *testing.T) {
	tests := []struct {
		name       string
		weight     uint8
		classLow   uint8
		flags      uint32
		wantResult bool
		want       []string
	}{
		{name: "invalid-weight", weight: 0xff, wantResult: false, want: []string{"weight"}},
		{name: "monster", weight: 1, classLow: 0x82, wantResult: false, want: []string{"weight", "class"}},
		{name: "eligible", weight: 1, classLow: 0x80, flags: 0xffffffbf, wantResult: true, want: []string{"weight", "class", "flags", "store:ffffffff", "refresh", "drop"}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var events []string
			point := types.Pointf{X: 3, Y: 4}
			hooks := chestOpenHooks4EDF00[int]{
				loadWeight: func(int) uint8 {
					events = append(events, "weight")
					return tc.weight
				},
				loadClassLow: func(int) uint8 {
					events = append(events, "class")
					return tc.classLow
				},
				loadFlags: func(int) uint32 {
					events = append(events, "flags")
					return tc.flags
				},
				storeFlags: func(_ int, value uint32) {
					events = append(events, fmt.Sprintf("store:%08x", value))
				},
				refresh: func(int) {
					events = append(events, "refresh")
				},
				drop: func(_, _ int, gotPoint *types.Pointf) int32 {
					events = append(events, "drop")
					if gotPoint != &point {
						panic("point identity changed")
					}
					return math.MinInt32
				},
			}
			if got := chestOpenDropItem4EDF00(1, 2, &point, hooks); got != tc.wantResult {
				t.Fatalf("result = %t, want %t", got, tc.wantResult)
			}
			if !reflect.DeepEqual(events, tc.want) {
				t.Fatalf("events = %v, want %v", events, tc.want)
			}
		})
	}
}
