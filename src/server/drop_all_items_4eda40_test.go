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

type dropAllItemsTestItem4EDA40 struct {
	name     string
	next     *dropAllItemsTestItem4EDA40
	radius   float32
	eligible int32
}

type dropAllItemsTestOwner4EDA40 struct {
	name string
	head *dropAllItemsTestItem4EDA40
	pos  types.Pointf
}

type dropAllItemsTestDrop4EDA40 struct {
	item  *dropAllItemsTestItem4EDA40
	point *types.Pointf
	value types.Pointf
}

type dropAllItemsTestWorld4EDA40 struct {
	owner *dropAllItemsTestOwner4EDA40

	events        []string
	faultAt       int
	randomResults []float64
	randomCalls   int
	traceResults  []int32
	traceCalls    int
	tracePointers []*dropAllItemsRay4EDA40
	traceValues   []dropAllItemsRay4EDA40
	dropResults   []int32
	drops         []dropAllItemsTestDrop4EDA40
	headLoads     int
	eligibleCalls map[*dropAllItemsTestItem4EDA40]int
	afterHead     func(*dropAllItemsTestWorld4EDA40, int)
	afterNext     func(*dropAllItemsTestWorld4EDA40, *dropAllItemsTestItem4EDA40)
	afterEligible func(*dropAllItemsTestWorld4EDA40, *dropAllItemsTestItem4EDA40, int)
	afterRandom   func(*dropAllItemsTestWorld4EDA40, int)
	afterTrace    func(*dropAllItemsTestWorld4EDA40, int, *dropAllItemsRay4EDA40)
	afterDrop     func(*dropAllItemsTestWorld4EDA40, int, *dropAllItemsTestItem4EDA40, *types.Pointf)
}

func dropAllItemsTestItemName4EDA40(item *dropAllItemsTestItem4EDA40) string {
	if item == nil {
		return "nil"
	}
	return item.name
}

func dropAllItemsTestOwnerName4EDA40(owner *dropAllItemsTestOwner4EDA40) string {
	if owner == nil {
		return "nil"
	}
	return owner.name
}

func dropAllItemsTestFloat32_4EDA40(value float32) string {
	return fmt.Sprintf("%08x", math.Float32bits(value))
}

func dropAllItemsTestFloat64_4EDA40(value float64) string {
	return fmt.Sprintf("%016x", math.Float64bits(value))
}

func dropAllItemsTestPoint4EDA40(point types.Pointf) string {
	return dropAllItemsTestFloat32_4EDA40(point.X) + "," + dropAllItemsTestFloat32_4EDA40(point.Y)
}

func (w *dropAllItemsTestWorld4EDA40) event(name string) {
	w.events = append(w.events, name)
	if w.faultAt != 0 && len(w.events) == w.faultAt {
		panic(name)
	}
}

func (w *dropAllItemsTestWorld4EDA40) hooks() dropAllItemsHooks4EDA40[
	*dropAllItemsTestOwner4EDA40,
	*dropAllItemsTestItem4EDA40,
] {
	if w.eligibleCalls == nil {
		w.eligibleCalls = make(map[*dropAllItemsTestItem4EDA40]int)
	}
	return dropAllItemsHooks4EDA40[
		*dropAllItemsTestOwner4EDA40,
		*dropAllItemsTestItem4EDA40,
	]{
		loadOwnerArg: func() *dropAllItemsTestOwner4EDA40 {
			w.event("owner-arg:" + dropAllItemsTestOwnerName4EDA40(w.owner))
			return w.owner
		},
		loadInventoryHead: func(owner *dropAllItemsTestOwner4EDA40) *dropAllItemsTestItem4EDA40 {
			w.headLoads++
			w.event(fmt.Sprintf("head:%d:%s:%s", w.headLoads, dropAllItemsTestOwnerName4EDA40(owner), dropAllItemsTestItemName4EDA40(owner.head)))
			if w.afterHead != nil {
				w.afterHead(w, w.headLoads)
			}
			return owner.head
		},
		loadInventoryNext: func(item *dropAllItemsTestItem4EDA40) *dropAllItemsTestItem4EDA40 {
			w.event("next:" + dropAllItemsTestItemName4EDA40(item) + ":" + dropAllItemsTestItemName4EDA40(item.next))
			next := item.next
			if w.afterNext != nil {
				w.afterNext(w, item)
			}
			return next
		},
		dropEligible: func(owner *dropAllItemsTestOwner4EDA40, item *dropAllItemsTestItem4EDA40) int32 {
			w.eligibleCalls[item]++
			call := w.eligibleCalls[item]
			w.event(fmt.Sprintf("eligible:%s:%s:%d:%d", dropAllItemsTestOwnerName4EDA40(owner), dropAllItemsTestItemName4EDA40(item), call, item.eligible))
			if w.afterEligible != nil {
				w.afterEligible(w, item, call)
			}
			return item.eligible
		},
		loadItemRadius: func(item *dropAllItemsTestItem4EDA40) float32 {
			w.event("radius:" + dropAllItemsTestItemName4EDA40(item) + ":" + dropAllItemsTestFloat32_4EDA40(item.radius))
			return item.radius
		},
		loadOwnerX: func(owner *dropAllItemsTestOwner4EDA40) float32 {
			w.event("owner-x:" + dropAllItemsTestOwnerName4EDA40(owner) + ":" + dropAllItemsTestFloat32_4EDA40(owner.pos.X))
			return owner.pos.X
		},
		loadOwnerY: func(owner *dropAllItemsTestOwner4EDA40) float32 {
			w.event("owner-y:" + dropAllItemsTestOwnerName4EDA40(owner) + ":" + dropAllItemsTestFloat32_4EDA40(owner.pos.Y))
			return owner.pos.Y
		},
		ownerPosition: func(owner *dropAllItemsTestOwner4EDA40) *types.Pointf {
			w.event("owner-position:" + dropAllItemsTestOwnerName4EDA40(owner))
			return &owner.pos
		},
		randomFloat: func(min, max float32, source string, line int32) float64 {
			index := w.randomCalls
			result := float64(0)
			if index < len(w.randomResults) {
				result = w.randomResults[index]
			}
			w.randomCalls++
			w.event(fmt.Sprintf(
				"random:%d:%s:%s:%d:%s:%s",
				w.randomCalls,
				dropAllItemsTestFloat32_4EDA40(min),
				dropAllItemsTestFloat32_4EDA40(max),
				line,
				source,
				dropAllItemsTestFloat64_4EDA40(result),
			))
			if w.afterRandom != nil {
				w.afterRandom(w, w.randomCalls)
			}
			return result
		},
		mapTrace: func(ray *dropAllItemsRay4EDA40, outPoint *types.Pointf, outGrid *image.Point, flags uint8) int32 {
			if outPoint != nil || outGrid != nil {
				panic("non-nil-trace-output")
			}
			index := w.traceCalls
			result := int32(0)
			if index < len(w.traceResults) {
				result = w.traceResults[index]
			}
			w.traceCalls++
			w.event(fmt.Sprintf(
				"trace:%d:%s->%s:%d:%d",
				w.traceCalls,
				dropAllItemsTestPoint4EDA40(ray.Origin),
				dropAllItemsTestPoint4EDA40(ray.Destination),
				flags,
				result,
			))
			w.tracePointers = append(w.tracePointers, ray)
			w.traceValues = append(w.traceValues, *ray)
			if w.afterTrace != nil {
				w.afterTrace(w, w.traceCalls, ray)
			}
			return result
		},
		drop: func(owner *dropAllItemsTestOwner4EDA40, item *dropAllItemsTestItem4EDA40, point *types.Pointf) int32 {
			index := len(w.drops)
			result := int32(0)
			if index < len(w.dropResults) {
				result = w.dropResults[index]
			}
			pointName := "local"
			if point == &owner.pos {
				pointName = "owner-position"
			}
			w.event(fmt.Sprintf(
				"drop:%d:%s:%s:%s:%s:%d",
				index+1,
				dropAllItemsTestOwnerName4EDA40(owner),
				dropAllItemsTestItemName4EDA40(item),
				pointName,
				dropAllItemsTestPoint4EDA40(*point),
				result,
			))
			w.drops = append(w.drops, dropAllItemsTestDrop4EDA40{item: item, point: point, value: *point})
			if w.afterDrop != nil {
				w.afterDrop(w, index+1, item, point)
			}
			return result
		},
	}
}

func TestDropAllItems4EDA40ConstantsLayoutSpacingAndSpiral(t *testing.T) {
	if math.Float32bits(dropAllItemsSpacingExtra4EDA40) != 0x40c00000 ||
		math.Float32bits(dropAllItemsRandomMin4EDA40) != 0xc0400000 ||
		math.Float32bits(dropAllItemsRandomMax4EDA40) != 0x40400000 {
		t.Fatalf("float constants = %#x/%#x/%#x", math.Float32bits(dropAllItemsSpacingExtra4EDA40), math.Float32bits(dropAllItemsRandomMin4EDA40), math.Float32bits(dropAllItemsRandomMax4EDA40))
	}
	if dropAllItemsSource4EDA40 != `C:\NoxPost\src\Server\Object\pickdrop\drop.c` ||
		dropAllItemsSourceLineX4EDA40 != 823 || dropAllItemsSourceLineY4EDA40 != 824 ||
		dropAllItemsTraceFlag4EDA40 != 1 {
		t.Fatalf("source/lines/flag = %q/%d/%d/%d", dropAllItemsSource4EDA40, dropAllItemsSourceLineX4EDA40, dropAllItemsSourceLineY4EDA40, dropAllItemsTraceFlag4EDA40)
	}
	var ray dropAllItemsRay4EDA40
	if unsafe.Sizeof(ray) != 16 || unsafe.Offsetof(ray.Origin) != 0 || unsafe.Offsetof(ray.Destination) != 8 {
		t.Fatalf("ray size/origin/destination = %d/%d/%d", unsafe.Sizeof(ray), unsafe.Offsetof(ray.Origin), unsafe.Offsetof(ray.Destination))
	}
	if got := math.Float32bits(dropAllItemsSpacing4EDA40(7.25)); got != math.Float32bits(20.5) {
		t.Fatalf("spacing bits = %#08x, want %#08x", got, math.Float32bits(20.5))
	}

	spiral := newDropAllItemsSpiral4EDA40()
	want := [][2]int32{{0, -1}, {1, -1}, {2, -1}, {2, 0}, {2, 1}, {1, 1}, {0, 1}, {0, 0}, {0, -1}}
	for i, point := range want {
		if got := [2]int32{spiral.x, spiral.y}; got != point {
			t.Fatalf("candidate %d = %v, want %v", i+1, got, point)
		}
		if got := spiral.advance4EDA40(); got != (i != len(want)-1) {
			t.Fatalf("advance %d = %t", i+1, got)
		}
	}

	spiral = newDropAllItemsSpiral4EDA40()
	spiral.ringHadSuccess = true
	for range want {
		if !spiral.advance4EDA40() {
			t.Fatal("successful perimeter selected fallback")
		}
	}
	if spiral.gridSize != 5 || spiral.sideLimit != 4 || spiral.segmentProgress != 1 ||
		spiral.direction != 0 || spiral.x != -1 || spiral.y != -2 || spiral.ringHadSuccess {
		t.Fatalf("expanded spiral = %+v", spiral)
	}
}

func TestDropAllItems4EDA40EmptyInventoryOrderAndFaultPrefixes(t *testing.T) {
	build := func() *dropAllItemsTestWorld4EDA40 {
		return &dropAllItemsTestWorld4EDA40{
			owner: &dropAllItemsTestOwner4EDA40{
				name: "owner",
				pos:  types.Pointf{X: 1, Y: 2},
			},
		}
	}
	want := []string{
		"owner-arg:owner",
		"head:1:owner:nil",
		"head:2:owner:nil",
		"owner-x:owner:3f800000",
		"owner-y:owner:40000000",
	}
	w := build()
	if got := dropAllItems4EDA40(w.hooks()); got != 0 {
		t.Fatalf("result = %d, want 0", got)
	}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %v, want %v", w.events, want)
	}
	for fault := 1; fault <= len(want); fault++ {
		w := build()
		w.faultAt = fault
		func() {
			defer func() {
				if got := recover(); got != want[fault-1] {
					t.Fatalf("fault %d panic = %v, want %q", fault, got, want[fault-1])
				}
			}()
			_ = dropAllItems4EDA40(w.hooks())
		}()
		if !reflect.DeepEqual(w.events, want[:fault]) {
			t.Fatalf("fault %d events = %v, want %v", fault, w.events, want[:fault])
		}
	}
}

func TestDropAllItems4EDA40PrimaryPassOrderLiveStateAndCachedSuccessor(t *testing.T) {
	skip := &dropAllItemsTestItem4EDA40{name: "skip", eligible: 0, radius: 1000}
	first := &dropAllItemsTestItem4EDA40{name: "first", eligible: 1, radius: float32(math.NaN())}
	second := &dropAllItemsTestItem4EDA40{name: "second", eligible: 1, radius: 4}
	skip.next = first
	first.next = second
	owner := &dropAllItemsTestOwner4EDA40{name: "owner", head: skip, pos: types.Pointf{X: 100, Y: 200}}
	w := &dropAllItemsTestWorld4EDA40{
		owner:         owner,
		randomResults: []float64{0, 0, 0, 0, 0, 0},
		traceResults:  []int32{0, math.MinInt32, 7},
		dropResults:   []int32{math.MinInt32, 99},
	}
	w.afterTrace = func(w *dropAllItemsTestWorld4EDA40, call int, ray *dropAllItemsRay4EDA40) {
		switch call {
		case 1:
			ray.Origin = types.Pointf{X: 7, Y: 8}
			owner.pos = types.Pointf{X: 110, Y: 220}
		case 2:
			ray.Destination = types.Pointf{X: 31, Y: 41}
		case 3:
			ray.Destination = types.Pointf{X: 51, Y: 61}
		}
	}
	w.afterDrop = func(_ *dropAllItemsTestWorld4EDA40, call int, item *dropAllItemsTestItem4EDA40, _ *types.Pointf) {
		if call == 1 && item == first {
			first.next = nil
		}
	}

	if got := dropAllItems4EDA40(w.hooks()); got != 0 {
		t.Fatalf("result = %d, want normal-exhaustion zero", got)
	}
	if len(w.traceValues) != 3 || len(w.drops) != 2 || w.randomCalls != 6 {
		t.Fatalf("trace/drop/random counts = %d/%d/%d", len(w.traceValues), len(w.drops), w.randomCalls)
	}
	for i := 1; i < len(w.tracePointers); i++ {
		if w.tracePointers[i] != w.tracePointers[0] {
			t.Fatalf("trace pointer %d changed", i+1)
		}
	}
	wantTrace := []dropAllItemsRay4EDA40{
		{Origin: types.Pointf{X: 100, Y: 200}, Destination: types.Pointf{X: 100, Y: 214}},
		{Origin: types.Pointf{X: 7, Y: 8}, Destination: types.Pointf{X: 124, Y: 234}},
		{Origin: types.Pointf{X: 7, Y: 8}, Destination: types.Pointf{X: 138, Y: 234}},
	}
	if !reflect.DeepEqual(w.traceValues, wantTrace) {
		t.Fatalf("trace values = %+v, want %+v", w.traceValues, wantTrace)
	}
	if w.drops[0].item != first || w.drops[0].value != (types.Pointf{X: 31, Y: 41}) ||
		w.drops[1].item != second || w.drops[1].value != (types.Pointf{X: 51, Y: 61}) ||
		w.drops[0].point != w.drops[1].point {
		t.Fatalf("drops = %+v", w.drops)
	}
	if w.eligibleCalls[skip] != 2 || w.eligibleCalls[first] != 2 || w.eligibleCalls[second] != 2 {
		t.Fatalf("eligibility calls = skip:%d first:%d second:%d", w.eligibleCalls[skip], w.eligibleCalls[first], w.eligibleCalls[second])
	}
	for _, event := range w.events {
		if event == "radius:skip:447a0000" {
			t.Fatal("ineligible item radius was loaded")
		}
	}
}

func TestDropAllItems4EDA40ReloadsHeadAfterRadiusScan(t *testing.T) {
	scanned := &dropAllItemsTestItem4EDA40{name: "scanned", eligible: 1, radius: 5}
	reloaded := &dropAllItemsTestItem4EDA40{name: "reloaded", eligible: 1, radius: 100}
	owner := &dropAllItemsTestOwner4EDA40{name: "owner", head: scanned, pos: types.Pointf{X: 10, Y: 20}}
	w := &dropAllItemsTestWorld4EDA40{owner: owner, traceResults: []int32{1}}
	w.afterNext = func(w *dropAllItemsTestWorld4EDA40, item *dropAllItemsTestItem4EDA40) {
		if item == scanned && w.eligibleCalls[scanned] == 1 {
			owner.head = reloaded
		}
	}

	if got := dropAllItems4EDA40(w.hooks()); got != 0 {
		t.Fatalf("result = %d, want 0", got)
	}
	if len(w.traceValues) != 1 || len(w.drops) != 1 || w.drops[0].item != reloaded {
		t.Fatalf("trace/drops = %d/%+v", len(w.traceValues), w.drops)
	}
	// Spacing comes from the first pass's radius 5, not the reloaded head's
	// unscanned radius 100: 2*5+6 = 16 at initial spiral coordinate (0,-1).
	if got, want := w.traceValues[0].Destination, (types.Pointf{X: 10, Y: 36}); got != want {
		t.Fatalf("destination = %v, want %v", got, want)
	}
	if w.eligibleCalls[scanned] != 1 || w.eligibleCalls[reloaded] != 1 {
		t.Fatalf("eligibility calls = scanned:%d reloaded:%d", w.eligibleCalls[scanned], w.eligibleCalls[reloaded])
	}
}

func TestDropAllItems4EDA40PrimaryFaultPrefixes(t *testing.T) {
	build := func() *dropAllItemsTestWorld4EDA40 {
		item := &dropAllItemsTestItem4EDA40{name: "item", radius: 2, eligible: 1}
		return &dropAllItemsTestWorld4EDA40{
			owner: &dropAllItemsTestOwner4EDA40{
				name: "owner",
				head: item,
				pos:  types.Pointf{X: 1, Y: 2},
			},
			traceResults: []int32{1},
			dropResults:  []int32{-77},
		}
	}
	w := build()
	if got := dropAllItems4EDA40(w.hooks()); got != 0 {
		t.Fatalf("result = %d, want 0", got)
	}
	want := []string{
		"owner-arg:owner",
		"head:1:owner:item",
		"eligible:owner:item:1:1",
		"radius:item:40000000",
		"next:item:nil",
		"head:2:owner:item",
		"owner-x:owner:3f800000",
		"owner-y:owner:40000000",
		"next:item:nil",
		"eligible:owner:item:2:1",
		"owner-x:owner:3f800000",
		"owner-y:owner:40000000",
		"random:1:c0400000:40400000:823:" + dropAllItemsSource4EDA40 + ":0000000000000000",
		"random:2:c0400000:40400000:824:" + dropAllItemsSource4EDA40 + ":0000000000000000",
		"trace:1:3f800000,40000000->3f800000,41400000:1:1",
		"drop:1:owner:item:local:3f800000,41400000:-77",
	}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %v, want %v", w.events, want)
	}
	for fault := 1; fault <= len(want); fault++ {
		w := build()
		w.faultAt = fault
		func() {
			defer func() {
				if got := recover(); got != want[fault-1] {
					t.Fatalf("fault %d panic = %v, want %q", fault, got, want[fault-1])
				}
			}()
			_ = dropAllItems4EDA40(w.hooks())
		}()
		if !reflect.DeepEqual(w.events, want[:fault]) {
			t.Fatalf("fault %d events = %v, want %v", fault, w.events, want[:fault])
		}
	}
}

func TestDropAllItems4EDA40FailedPerimeterFallsBackWithLivePointAndFinalResult(t *testing.T) {
	first := &dropAllItemsTestItem4EDA40{name: "first", eligible: 1}
	second := &dropAllItemsTestItem4EDA40{name: "second", eligible: 1}
	first.next = second
	owner := &dropAllItemsTestOwner4EDA40{name: "owner", head: first, pos: types.Pointf{X: 10, Y: 20}}
	w := &dropAllItemsTestWorld4EDA40{
		owner:        owner,
		traceResults: make([]int32, 9),
		dropResults:  []int32{7, math.MinInt32},
	}
	w.afterTrace = func(_ *dropAllItemsTestWorld4EDA40, call int, ray *dropAllItemsRay4EDA40) {
		ray.Origin.X = float32(call)
		if call == 9 {
			owner.pos = types.Pointf{X: 77, Y: 88}
		}
	}
	w.afterDrop = func(_ *dropAllItemsTestWorld4EDA40, call int, item *dropAllItemsTestItem4EDA40, _ *types.Pointf) {
		if call == 1 && item == first {
			first.next = nil
		}
	}

	if got := dropAllItems4EDA40(w.hooks()); got != math.MinInt32 {
		t.Fatalf("result = %d, want %d", got, int32(math.MinInt32))
	}
	if w.traceCalls != 9 || w.randomCalls != 18 || len(w.drops) != 2 {
		t.Fatalf("trace/random/drop counts = %d/%d/%d", w.traceCalls, w.randomCalls, len(w.drops))
	}
	wantDestinations := []types.Pointf{
		{X: 10, Y: 26},
		{X: 16, Y: 26},
		{X: 22, Y: 26},
		{X: 22, Y: 20},
		{X: 22, Y: 14},
		{X: 16, Y: 14},
		{X: 10, Y: 14},
		{X: 10, Y: 20},
		{X: 10, Y: 26},
	}
	for i, want := range wantDestinations {
		if got := w.traceValues[i].Destination; got != want {
			t.Fatalf("destination %d = %v, want %v", i+1, got, want)
		}
		if i != 0 && w.traceValues[i].Origin.X != float32(i) {
			t.Fatalf("origin mutation at trace %d = %v", i+1, w.traceValues[i].Origin)
		}
	}
	for i, drop := range w.drops {
		if drop.point != &owner.pos || drop.value != (types.Pointf{X: 77, Y: 88}) {
			t.Fatalf("fallback drop %d = %+v", i+1, drop)
		}
	}
}

func TestDropAllItems4EDA40SuccessfulPerimeterExpandsBeforeFallback(t *testing.T) {
	first := &dropAllItemsTestItem4EDA40{name: "first", eligible: 1}
	second := &dropAllItemsTestItem4EDA40{name: "second", eligible: 1}
	first.next = second
	owner := &dropAllItemsTestOwner4EDA40{name: "owner", head: first, pos: types.Pointf{X: 10, Y: 20}}
	traceResults := make([]int32, 10)
	traceResults[0] = 1
	traceResults[9] = 1
	w := &dropAllItemsTestWorld4EDA40{owner: owner, traceResults: traceResults}

	if got := dropAllItems4EDA40(w.hooks()); got != 0 {
		t.Fatalf("result = %d, want 0", got)
	}
	if w.traceCalls != 10 || len(w.drops) != 2 {
		t.Fatalf("trace/drop counts = %d/%d", w.traceCalls, len(w.drops))
	}
	if got, want := w.traceValues[9].Destination, (types.Pointf{X: 4, Y: 32}); got != want {
		t.Fatalf("first expanded candidate = %v, want %v", got, want)
	}
	for _, drop := range w.drops {
		if drop.point == &owner.pos {
			t.Fatal("successful expanded perimeter used fallback point")
		}
	}
}
