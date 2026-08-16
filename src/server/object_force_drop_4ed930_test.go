package server

import (
	"fmt"
	"math"
	"reflect"
	"testing"
)

type objectForceDropTestObject4ED930 struct {
	name string
}

type objectForceDropTestPoint4ED930 struct {
	x uint32
	y uint32
}

type objectForceDropTestWorld4ED930 struct {
	ownerArg *objectForceDropTestObject4ED930
	itemArg  *objectForceDropTestObject4ED930

	helperPoint  objectForceDropTestPoint4ED930
	helperReturn *objectForceDropTestPoint4ED930
	result       int32
	events       []string
	faultAt      int
	local        *objectForceDropTestPoint4ED930
	afterHelper  func(*objectForceDropTestWorld4ED930)
}

func objectForceDropTestObjectName4ED930(obj *objectForceDropTestObject4ED930) string {
	if obj == nil {
		return "nil"
	}
	return obj.name
}

func (w *objectForceDropTestWorld4ED930) event(event string) {
	w.events = append(w.events, event)
	if w.faultAt != 0 && len(w.events) == w.faultAt {
		panic(event)
	}
}

func newObjectForceDropTestWorld4ED930() *objectForceDropTestWorld4ED930 {
	return &objectForceDropTestWorld4ED930{
		ownerArg:    &objectForceDropTestObject4ED930{name: "owner-a"},
		itemArg:     &objectForceDropTestObject4ED930{name: "item-a"},
		helperPoint: objectForceDropTestPoint4ED930{x: 0x3f800000, y: 0xc0200000},
		result:      math.MinInt32,
	}
}

func (w *objectForceDropTestWorld4ED930) hooks() objectForceDropHooks4ED930[
	*objectForceDropTestObject4ED930,
	objectForceDropTestPoint4ED930,
] {
	return objectForceDropHooks4ED930[
		*objectForceDropTestObject4ED930,
		objectForceDropTestPoint4ED930,
	]{
		loadOwnerArg: func() *objectForceDropTestObject4ED930 {
			w.event("owner-arg:" + objectForceDropTestObjectName4ED930(w.ownerArg))
			return w.ownerArg
		},
		randomReachable: func(radius float32, owner *objectForceDropTestObject4ED930, output *objectForceDropTestPoint4ED930) *objectForceDropTestPoint4ED930 {
			w.event(fmt.Sprintf("helper:%08x:%s", math.Float32bits(radius), objectForceDropTestObjectName4ED930(owner)))
			w.local = output
			*output = w.helperPoint
			if w.afterHelper != nil {
				w.afterHelper(w)
			}
			return w.helperReturn
		},
		loadItemArg: func() *objectForceDropTestObject4ED930 {
			w.event("item-arg:" + objectForceDropTestObjectName4ED930(w.itemArg))
			return w.itemArg
		},
		dispatch: func(owner, item *objectForceDropTestObject4ED930, point *objectForceDropTestPoint4ED930) int32 {
			w.event(fmt.Sprintf(
				"dispatch:%s:%s:%08x,%08x",
				objectForceDropTestObjectName4ED930(owner),
				objectForceDropTestObjectName4ED930(item),
				point.x,
				point.y,
			))
			if point != w.local {
				panic("dispatch-point-identity")
			}
			return w.result
		},
	}
}

func verifyObjectForceDropFaultPrefixes4ED930(
	t *testing.T,
	want []string,
	build func() *objectForceDropTestWorld4ED930,
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
			_ = objectForceDrop4ED930(w.hooks())
		})
	}
}

func TestObjectForceDrop4ED930ExactOrderAndLocalIdentity(t *testing.T) {
	build := func() *objectForceDropTestWorld4ED930 {
		w := newObjectForceDropTestWorld4ED930()
		ownerB := &objectForceDropTestObject4ED930{name: "owner-b"}
		itemB := &objectForceDropTestObject4ED930{name: "item-b"}
		returned := &objectForceDropTestPoint4ED930{x: 0xaaaaaaaa, y: 0xbbbbbbbb}
		w.helperReturn = returned
		w.afterHelper = func(w *objectForceDropTestWorld4ED930) {
			w.ownerArg = ownerB
			w.itemArg = itemB
		}
		return w
	}
	want := []string{
		"owner-arg:owner-a",
		"helper:42480000:owner-a",
		"item-arg:item-b",
		"dispatch:owner-a:item-b:3f800000,c0200000",
	}
	w := build()
	if got := objectForceDrop4ED930(w.hooks()); got != math.MinInt32 {
		t.Fatalf("result = %d, want %d", got, int32(math.MinInt32))
	}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %v, want %v", w.events, want)
	}
	if w.local == nil || w.local == w.helperReturn {
		t.Fatalf("local point = %p, helper return = %p", w.local, w.helperReturn)
	}
	verifyObjectForceDropFaultPrefixes4ED930(t, want, build)
}

func TestObjectForceDrop4ED930NilItemStillCallsHelperFirst(t *testing.T) {
	w := newObjectForceDropTestWorld4ED930()
	w.itemArg = nil
	w.result = 0
	if got := objectForceDrop4ED930(w.hooks()); got != 0 {
		t.Fatalf("result = %d, want 0", got)
	}
	want := []string{
		"owner-arg:owner-a",
		"helper:42480000:owner-a",
		"item-arg:nil",
		"dispatch:owner-a:nil:3f800000,c0200000",
	}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %v, want %v", w.events, want)
	}
}

func TestObjectForceDrop4ED930PreservesWholeDispatcherResult(t *testing.T) {
	for _, result := range []int32{0, 1, -1, math.MaxInt32, math.MinInt32, 0x76543210} {
		w := newObjectForceDropTestWorld4ED930()
		w.result = result
		if got := objectForceDrop4ED930(w.hooks()); got != result {
			t.Fatalf("result = %#x, want %#x", uint32(got), uint32(result))
		}
	}
}
