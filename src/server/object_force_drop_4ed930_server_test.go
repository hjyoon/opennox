package server

import (
	"math"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/types"
)

func defaultObjectForceDropNativeDeps4ED930() objectForceDropNativeDeps4ED930 {
	return objectForceDropNativeDeps4ED930{
		randomReachable: func(_ float32, _ *Object, output *types.Pointf) *types.Pointf {
			*output = types.Pointf{}
			return output
		},
		dispatch: func(*Object, *Object, *types.Pointf) int32 { return 0 },
	}
}

func TestObjectForceDrop4ED930NativeLayout(t *testing.T) {
	wantObjectSize := uintptr(780)
	wantPosition := uintptr(56)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantObjectSize = 928
		wantPosition = 60
	}
	checks := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"Object size", unsafe.Sizeof(Object{}), wantObjectSize},
		{"Object.PosVec", unsafe.Offsetof(Object{}.PosVec), wantPosition},
		{"Pointf size", unsafe.Sizeof(types.Pointf{}), 8},
		{"Pointf.X", unsafe.Offsetof(types.Pointf{}.X), 0},
		{"Pointf.Y", unsafe.Offsetof(types.Pointf{}.Y), 4},
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Errorf("%s = %d, want %d", check.name, check.got, check.want)
		}
	}
}

func TestObjectForceDropNative4ED930UsesNativeObjectsAndLocalPoint(t *testing.T) {
	owner := &Object{PosVec: types.Pointf{X: 10, Y: -20}}
	item := &Object{TypeInd: 0x1234}
	events := make([]string, 0, 2)
	deps := defaultObjectForceDropNativeDeps4ED930()
	returned := &types.Pointf{X: -1, Y: -2}
	var helperOutput *types.Pointf
	deps.randomReachable = func(radius float32, gotOwner *Object, output *types.Pointf) *types.Pointf {
		events = append(events, "helper")
		if math.Float32bits(radius) != 0x42480000 || gotOwner != owner {
			t.Fatalf("helper = %08x/%p, want 42480000/%p", math.Float32bits(radius), gotOwner, owner)
		}
		if output == &owner.PosVec {
			t.Fatal("helper output aliases owner position")
		}
		helperOutput = output
		*output = types.Pointf{X: 31, Y: 41}
		owner.PosVec = types.Pointf{X: 99, Y: 100}
		return returned
	}
	deps.dispatch = func(gotOwner, gotItem *Object, point *types.Pointf) int32 {
		events = append(events, "dispatch")
		if gotOwner != owner || gotItem != item || point != helperOutput || point == returned {
			t.Fatalf("dispatch = %p/%p/%p, helper=%p return=%p", gotOwner, gotItem, point, helperOutput, returned)
		}
		if *point != (types.Pointf{X: 31, Y: 41}) {
			t.Fatalf("dispatch point = %+v", *point)
		}
		return math.MinInt32
	}

	if got := objectForceDropNative4ED930(owner, item, deps); got != math.MinInt32 {
		t.Fatalf("result = %d, want %d", got, int32(math.MinInt32))
	}
	if !reflect.DeepEqual(events, []string{"helper", "dispatch"}) {
		t.Fatalf("events = %v", events)
	}
}

func TestObjectForceDropNative4ED930NilItemStillRunsHelper(t *testing.T) {
	owner := &Object{}
	events := make([]string, 0, 2)
	deps := defaultObjectForceDropNativeDeps4ED930()
	deps.randomReachable = func(_ float32, gotOwner *Object, output *types.Pointf) *types.Pointf {
		events = append(events, "helper")
		if gotOwner != owner {
			t.Fatalf("helper owner = %p, want %p", gotOwner, owner)
		}
		return output
	}
	deps.dispatch = func(gotOwner, gotItem *Object, _ *types.Pointf) int32 {
		events = append(events, "dispatch")
		if gotOwner != owner || gotItem != nil {
			t.Fatalf("dispatch = %p/%p", gotOwner, gotItem)
		}
		return 0
	}
	if got := objectForceDropNative4ED930(owner, nil, deps); got != 0 {
		t.Fatalf("result = %d, want 0", got)
	}
	if !reflect.DeepEqual(events, []string{"helper", "dispatch"}) {
		t.Fatalf("events = %v", events)
	}
}
