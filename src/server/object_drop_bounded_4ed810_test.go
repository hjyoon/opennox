package server

import (
	"fmt"
	"math"
	"reflect"
	"testing"

	"github.com/opennox/libs/types"
)

type objectDropBoundedTestObject4ED810 struct {
	name     string
	position types.Pointf
	netCode  uint32
	typeInd  uint16
}

type objectDropBoundedTestPoint4ED810 struct {
	name  string
	value types.Pointf
}

type objectDropBoundedTestWorld4ED810 struct {
	ownerArg *objectDropBoundedTestObject4ED810
	itemArg  *objectDropBoundedTestObject4ED810
	pointArg *objectDropBoundedTestPoint4ED810

	traceResult    int32
	gameFlagResult int32
	crownCache     uint32
	crownLookup    uint32
	dispatchResult int32

	events          []string
	faultAt         int
	traceOrigin     *types.Pointf
	traceTarget     *types.Pointf
	traceOriginIn   types.Pointf
	traceTargetIn   types.Pointf
	dispatchTarget  *types.Pointf
	afterTrace      func(*objectDropBoundedTestWorld4ED810, *types.Pointf, *types.Pointf)
	afterMessage    func(*objectDropBoundedTestWorld4ED810)
	afterGameFlag   func(*objectDropBoundedTestWorld4ED810)
	afterCacheStore func(*objectDropBoundedTestWorld4ED810)
}

func objectDropBoundedObjectName4ED810(obj *objectDropBoundedTestObject4ED810) string {
	if obj == nil {
		return "nil"
	}
	return obj.name
}

func objectDropBoundedPointName4ED810(point *objectDropBoundedTestPoint4ED810) string {
	if point == nil {
		return "nil"
	}
	return point.name
}

func objectDropBoundedFloat4ED810(value float32) string {
	return fmt.Sprintf("%08x", math.Float32bits(value))
}

func objectDropBoundedPoint4ED810(point types.Pointf) string {
	return objectDropBoundedFloat4ED810(point.X) + "," + objectDropBoundedFloat4ED810(point.Y)
}

func (w *objectDropBoundedTestWorld4ED810) event(name string) {
	w.events = append(w.events, name)
	if w.faultAt != 0 && len(w.events) == w.faultAt {
		panic(name)
	}
}

func newObjectDropBoundedTestWorld4ED810() *objectDropBoundedTestWorld4ED810 {
	return &objectDropBoundedTestWorld4ED810{
		ownerArg: &objectDropBoundedTestObject4ED810{
			name:     "owner-a",
			position: types.Pointf{X: 10, Y: -20},
			netCode:  0x12345678,
		},
		itemArg: &objectDropBoundedTestObject4ED810{
			name:    "item-a",
			typeInd: 7,
		},
		pointArg: &objectDropBoundedTestPoint4ED810{
			name:  "point-a",
			value: types.Pointf{X: 13, Y: -16},
		},
		traceResult:    1,
		dispatchResult: 0x76543210,
	}
}

func (w *objectDropBoundedTestWorld4ED810) hooks() objectDropBoundedHooks4ED810[
	*objectDropBoundedTestObject4ED810,
	*objectDropBoundedTestPoint4ED810,
] {
	return objectDropBoundedHooks4ED810[
		*objectDropBoundedTestObject4ED810,
		*objectDropBoundedTestPoint4ED810,
	]{
		loadOwnerArg: func() *objectDropBoundedTestObject4ED810 {
			w.event("owner-arg:" + objectDropBoundedObjectName4ED810(w.ownerArg))
			return w.ownerArg
		},
		loadOwnerX: func(owner *objectDropBoundedTestObject4ED810) float32 {
			w.event("owner-x:" + objectDropBoundedObjectName4ED810(owner) + ":" + objectDropBoundedFloat4ED810(owner.position.X))
			return owner.position.X
		},
		loadOwnerY: func(owner *objectDropBoundedTestObject4ED810) float32 {
			w.event("owner-y:" + objectDropBoundedObjectName4ED810(owner) + ":" + objectDropBoundedFloat4ED810(owner.position.Y))
			return owner.position.Y
		},
		loadPointArg: func() *objectDropBoundedTestPoint4ED810 {
			w.event("point-arg:" + objectDropBoundedPointName4ED810(w.pointArg))
			return w.pointArg
		},
		loadPointX: func(point *objectDropBoundedTestPoint4ED810) float32 {
			w.event("point-x:" + objectDropBoundedPointName4ED810(point) + ":" + objectDropBoundedFloat4ED810(point.value.X))
			return point.value.X
		},
		loadPointY: func(point *objectDropBoundedTestPoint4ED810) float32 {
			w.event("point-y:" + objectDropBoundedPointName4ED810(point) + ":" + objectDropBoundedFloat4ED810(point.value.Y))
			return point.value.Y
		},
		mapTrace: func(origin, target *types.Pointf) int32 {
			w.event("trace:" + objectDropBoundedPoint4ED810(*origin) + "->" + objectDropBoundedPoint4ED810(*target))
			w.traceOrigin = origin
			w.traceTarget = target
			w.traceOriginIn = *origin
			w.traceTargetIn = *target
			if w.afterTrace != nil {
				w.afterTrace(w, origin, target)
			}
			return w.traceResult
		},
		priorityMessage: func(owner *objectDropBoundedTestObject4ED810, message string, kind int32) {
			w.event(fmt.Sprintf("message:%s:%s:%d", objectDropBoundedObjectName4ED810(owner), message, kind))
			if w.afterMessage != nil {
				w.afterMessage(w)
			}
		},
		loadNetCode: func(owner *objectDropBoundedTestObject4ED810) uint32 {
			w.event(fmt.Sprintf("net-code:%s:%08x", objectDropBoundedObjectName4ED810(owner), owner.netCode))
			return owner.netCode
		},
		audio: func(id uint32, owner *objectDropBoundedTestObject4ED810, kind int32, code uint32) {
			w.event(fmt.Sprintf("audio:%d:%s:%d:%08x", id, objectDropBoundedObjectName4ED810(owner), kind, code))
		},
		gameFlag: func(flag uint32) int32 {
			w.event(fmt.Sprintf("game-flag:%d:%d", flag, w.gameFlagResult))
			if w.afterGameFlag != nil {
				w.afterGameFlag(w)
			}
			return w.gameFlagResult
		},
		loadItemArg: func() *objectDropBoundedTestObject4ED810 {
			w.event("item-arg:" + objectDropBoundedObjectName4ED810(w.itemArg))
			return w.itemArg
		},
		loadCrownTypeCache: func() uint32 {
			w.event(fmt.Sprintf("crown-cache:%08x", w.crownCache))
			return w.crownCache
		},
		lookupCrownType: func() uint32 {
			w.event(fmt.Sprintf("crown-lookup:%08x", w.crownLookup))
			return w.crownLookup
		},
		storeCrownTypeCache: func(value uint32) {
			w.event(fmt.Sprintf("crown-store:%08x", value))
			w.crownCache = value
			if w.afterCacheStore != nil {
				w.afterCacheStore(w)
			}
		},
		loadTypeIndex: func(item *objectDropBoundedTestObject4ED810) uint16 {
			w.event("type-index:" + objectDropBoundedObjectName4ED810(item))
			return item.typeInd
		},
		dispatch: func(owner, item *objectDropBoundedTestObject4ED810, target *types.Pointf) int32 {
			w.event("dispatch:" + objectDropBoundedObjectName4ED810(owner) + ":" + objectDropBoundedObjectName4ED810(item) + ":" + objectDropBoundedPoint4ED810(*target))
			w.dispatchTarget = target
			return w.dispatchResult
		},
	}
}

func verifyObjectDropBoundedFaultPrefixes4ED810(
	t *testing.T,
	want []string,
	build func() *objectDropBoundedTestWorld4ED810,
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
			_ = objectDropBounded4ED810(w.hooks())
		})
	}
}

func TestObjectDropBounded4ED810AllowedDelaysItemAndPreservesLocalTarget(t *testing.T) {
	build := func() *objectDropBoundedTestWorld4ED810 {
		w := newObjectDropBoundedTestWorld4ED810()
		w.traceResult = math.MinInt32
		w.dispatchResult = math.MinInt32
		itemB := &objectDropBoundedTestObject4ED810{name: "item-b"}
		itemC := &objectDropBoundedTestObject4ED810{name: "item-c"}
		w.afterTrace = func(w *objectDropBoundedTestWorld4ED810, origin, target *types.Pointf) {
			*origin = types.Pointf{X: -1, Y: -2}
			*target = types.Pointf{X: 31, Y: 41}
			w.itemArg = itemB
		}
		w.afterGameFlag = func(w *objectDropBoundedTestWorld4ED810) {
			w.itemArg = itemC
		}
		return w
	}
	want := []string{
		"owner-arg:owner-a",
		"owner-x:owner-a:41200000",
		"owner-y:owner-a:c1a00000",
		"point-arg:point-a",
		"point-x:point-a:41500000",
		"point-y:point-a:c1800000",
		"trace:41200000,c1a00000->41500000,c1800000",
		"game-flag:16:0",
		"item-arg:item-c",
		"dispatch:owner-a:item-c:41f80000,42240000",
	}
	w := build()
	originalPoint := w.pointArg.value
	if got := objectDropBounded4ED810(w.hooks()); got != math.MinInt32 {
		t.Fatalf("result = %d, want %d", got, int32(math.MinInt32))
	}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %v, want %v", w.events, want)
	}
	if w.pointArg.value != originalPoint {
		t.Fatalf("caller point changed to %+v, want %+v", w.pointArg.value, originalPoint)
	}
	if w.traceTarget == nil || w.dispatchTarget != w.traceTarget {
		t.Fatalf("dispatch target %p does not reuse trace target %p", w.dispatchTarget, w.traceTarget)
	}
	verifyObjectDropBoundedFaultPrefixes4ED810(t, want, build)
}

func TestObjectDropBounded4ED810RejectedTraceUsesCachedOwnerLiveNetCode(t *testing.T) {
	build := func() *objectDropBoundedTestWorld4ED810 {
		w := newObjectDropBoundedTestWorld4ED810()
		ownerA := w.ownerArg
		w.ownerArg.position = types.Pointf{}
		w.pointArg.value = types.Pointf{X: 100}
		w.traceResult = 0
		ownerB := &objectDropBoundedTestObject4ED810{name: "owner-b", netCode: 0xaaaaaaaa}
		w.afterTrace = func(w *objectDropBoundedTestWorld4ED810, _, _ *types.Pointf) {
			w.ownerArg = ownerB
		}
		w.afterMessage = func(w *objectDropBoundedTestWorld4ED810) {
			w.ownerArg = ownerB
			ownerA.netCode = 0xfedcba98
		}
		return w
	}
	want := []string{
		"owner-arg:owner-a",
		"owner-x:owner-a:00000000",
		"owner-y:owner-a:00000000",
		"point-arg:point-a",
		"point-x:point-a:42c80000",
		"point-y:point-a:00000000",
		"trace:00000000,00000000->42960000,00000000",
		"message:owner-a:drop.c:DropNotAllowed:0",
		"net-code:owner-a:fedcba98",
		"audio:925:owner-a:2:fedcba98",
	}
	w := build()
	if got := objectDropBounded4ED810(w.hooks()); got != 0 {
		t.Fatalf("result = %d, want 0", got)
	}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %v, want %v", w.events, want)
	}
	verifyObjectDropBoundedFaultPrefixes4ED810(t, want, build)
}

func TestObjectDropBounded4ED810KOTRCrownCache(t *testing.T) {
	tests := []struct {
		name  string
		build func() *objectDropBoundedTestWorld4ED810
		want  []string
		ret   int32
	}{
		{
			name: "cache hit suppresses crown",
			build: func() *objectDropBoundedTestWorld4ED810 {
				w := newObjectDropBoundedTestWorld4ED810()
				w.gameFlagResult = -1
				w.crownCache = 7
				return w
			},
			want: []string{"crown-cache:00000007", "type-index:item-a"},
		},
		{
			name: "zero lookup is stored and matches type zero",
			build: func() *objectDropBoundedTestWorld4ED810 {
				w := newObjectDropBoundedTestWorld4ED810()
				w.gameFlagResult = 1
				w.itemArg.typeInd = 0
				return w
			},
			want: []string{"crown-cache:00000000", "crown-lookup:00000000", "crown-store:00000000", "type-index:item-a"},
		},
		{
			name: "stored lookup is used without cache reload",
			build: func() *objectDropBoundedTestWorld4ED810 {
				w := newObjectDropBoundedTestWorld4ED810()
				w.gameFlagResult = math.MinInt32
				w.crownLookup = 9
				w.itemArg.typeInd = 10
				w.dispatchResult = math.MinInt32
				w.afterCacheStore = func(w *objectDropBoundedTestWorld4ED810) {
					w.crownCache = 10
				}
				return w
			},
			want: []string{
				"crown-cache:00000000", "crown-lookup:00000009", "crown-store:00000009",
				"type-index:item-a", "dispatch:owner-a:item-a:41500000,c1800000",
			},
			ret: math.MinInt32,
		},
	}
	common := []string{
		"owner-arg:owner-a", "owner-x:owner-a:41200000", "owner-y:owner-a:c1a00000",
		"point-arg:point-a", "point-x:point-a:41500000", "point-y:point-a:c1800000",
		"trace:41200000,c1a00000->41500000,c1800000",
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			w := test.build()
			want := append(append([]string{}, common...), fmt.Sprintf("game-flag:16:%d", w.gameFlagResult), "item-arg:item-a")
			want = append(want, test.want...)
			if got := objectDropBounded4ED810(w.hooks()); got != test.ret {
				t.Fatalf("result = %d, want %d", got, test.ret)
			}
			if !reflect.DeepEqual(w.events, want) {
				t.Fatalf("events = %v, want %v", w.events, want)
			}
		})
	}
}

func TestObjectDropBounded4ED810WholeEAXGates(t *testing.T) {
	for _, traceResult := range []int32{1, -1, math.MinInt32} {
		w := newObjectDropBoundedTestWorld4ED810()
		w.traceResult = traceResult
		w.dispatchResult = traceResult
		if got := objectDropBounded4ED810(w.hooks()); got != traceResult {
			t.Fatalf("trace result %d: dispatch result = %d", traceResult, got)
		}
	}
	for _, flagResult := range []int32{1, -1, math.MinInt32} {
		w := newObjectDropBoundedTestWorld4ED810()
		w.gameFlagResult = flagResult
		w.crownCache = uint32(w.itemArg.typeInd)
		if got := objectDropBounded4ED810(w.hooks()); got != 0 {
			t.Fatalf("game flag result %d: result = %d, want 0", flagResult, got)
		}
		if w.dispatchTarget != nil {
			t.Fatalf("game flag result %d dispatched a Crown", flagResult)
		}
	}
}

func TestObjectDropBoundedDestination4ED810Binary32Spills(t *testing.T) {
	nan := math.Float32frombits(0x7fc12345)
	negativeZero := math.Float32frombits(0x80000000)
	spillOrigin := types.Pointf{
		X: math.Float32frombits(0x40442378),
		Y: math.Float32frombits(0x4265d7b8),
	}
	spillRequested := types.Pointf{
		X: math.Float32frombits(0xc269b70a),
		Y: math.Float32frombits(0xc1a06189),
	}
	tests := []struct {
		name      string
		origin    types.Pointf
		requested types.Pointf
		wantX     uint32
		wantY     uint32
	}{
		{"within", types.Pointf{X: 10, Y: -20}, types.Pointf{X: 13, Y: -16}, 0x41500000, 0xc1800000},
		{"equal", types.Pointf{}, types.Pointf{X: 45, Y: 60}, 0x42340000, 0x42700000},
		{"clamped", types.Pointf{}, types.Pointf{X: 100}, 0x42960000, 0x00000000},
		{"rounded Y delta", spillOrigin, spillRequested, 0xc22e330f, 0xbfa59af0},
		{"unordered", types.Pointf{X: 1, Y: 2}, types.Pointf{X: nan, Y: negativeZero}, 0x7fc12345, 0x80000000},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := objectDropBoundedDestination4ED810(test.origin, test.requested)
			if bits := math.Float32bits(got.X); bits != test.wantX {
				t.Fatalf("X bits = %08x, want %08x", bits, test.wantX)
			}
			if bits := math.Float32bits(got.Y); bits != test.wantY {
				t.Fatalf("Y bits = %08x, want %08x", bits, test.wantY)
			}
		})
	}
}
