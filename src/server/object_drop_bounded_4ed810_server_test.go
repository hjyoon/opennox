package server

import (
	"math"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/types"
)

func defaultObjectDropBoundedNativeDeps4ED810() objectDropBoundedNativeDeps4ED810 {
	return objectDropBoundedNativeDeps4ED810{
		mapTrace:            func(*types.Pointf, *types.Pointf) int32 { return 1 },
		priorityMessage:     func(*Object, string, int32) {},
		audio:               func(uint32, *Object, int32, uint32) {},
		gameFlag:            func(uint32) int32 { return 0 },
		loadCrownTypeCache:  func() uint32 { return 0 },
		lookupCrownType:     func() uint32 { return 0 },
		storeCrownTypeCache: func(uint32) {},
		dispatch:            func(*Object, *Object, *types.Pointf) int32 { return 0 },
	}
}

func TestObjectDropBounded4ED810NativeLayout(t *testing.T) {
	wantObjectSize := uintptr(780)
	wantTypeInd := uintptr(4)
	wantNetCode := uintptr(36)
	wantPosition := uintptr(56)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantObjectSize = 928
		wantTypeInd = 8
		wantNetCode = 40
		wantPosition = 60
	}
	checks := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"Object size", unsafe.Sizeof(Object{}), wantObjectSize},
		{"Object.TypeInd", unsafe.Offsetof(Object{}.TypeInd), wantTypeInd},
		{"Object.NetCode", unsafe.Offsetof(Object{}.NetCode), wantNetCode},
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

func TestObjectDropBoundedNative4ED810AllowedUsesNativeObjectsAndLocalPoint(t *testing.T) {
	owner := &Object{PosVec: types.Pointf{X: 10, Y: -20}}
	item := &Object{TypeInd: 0x1234}
	point := &types.Pointf{X: 110, Y: -20}
	originalPoint := *point
	events := make([]string, 0, 3)
	deps := defaultObjectDropBoundedNativeDeps4ED810()
	var tracedTarget *types.Pointf
	deps.mapTrace = func(origin, target *types.Pointf) int32 {
		events = append(events, "trace")
		if *origin != owner.PosVec || *target != (types.Pointf{X: 85, Y: -20}) {
			t.Fatalf("trace segment = %+v -> %+v", *origin, *target)
		}
		if origin == &owner.PosVec || target == point {
			t.Fatal("trace received caller-owned storage")
		}
		tracedTarget = target
		*target = types.Pointf{X: 31, Y: 41}
		return math.MinInt32
	}
	deps.priorityMessage = func(*Object, string, int32) { t.Fatal("allowed trace sent message") }
	deps.audio = func(uint32, *Object, int32, uint32) { t.Fatal("allowed trace sent audio") }
	deps.gameFlag = func(flag uint32) int32 {
		events = append(events, "game")
		if flag != objectDropBoundedKOTRFlag4ED810 {
			t.Fatalf("game flag = %d", flag)
		}
		return 0
	}
	deps.loadCrownTypeCache = func() uint32 { t.Fatal("non-KOTR read Crown cache"); return 0 }
	deps.dispatch = func(gotOwner, gotItem *Object, target *types.Pointf) int32 {
		events = append(events, "dispatch")
		if gotOwner != owner || gotItem != item || target != tracedTarget || *target != (types.Pointf{X: 31, Y: 41}) {
			t.Fatalf("dispatch = %p/%p/%p %+v", gotOwner, gotItem, target, *target)
		}
		return math.MinInt32
	}

	if got := objectDropBoundedNative4ED810(owner, item, point, deps); got != math.MinInt32 {
		t.Fatalf("result = %d, want %d", got, int32(math.MinInt32))
	}
	if *point != originalPoint {
		t.Fatalf("caller point changed to %+v, want %+v", *point, originalPoint)
	}
	if !reflect.DeepEqual(events, []string{"trace", "game", "dispatch"}) {
		t.Fatalf("events = %v", events)
	}
}

func TestObjectDropBoundedNative4ED810RejectUsesPostMessageNetCode(t *testing.T) {
	owner := &Object{PosVec: types.Pointf{X: 1, Y: 2}, NetCode: 0x11111111}
	events := make([]string, 0, 3)
	deps := defaultObjectDropBoundedNativeDeps4ED810()
	deps.mapTrace = func(*types.Pointf, *types.Pointf) int32 {
		events = append(events, "trace")
		owner.NetCode = 0x22222222
		return 0
	}
	deps.priorityMessage = func(gotOwner *Object, message string, kind int32) {
		events = append(events, "message")
		if gotOwner != owner || message != objectDropBoundedRejectMessage4ED810 || kind != 0 {
			t.Fatalf("message = %p/%q/%d", gotOwner, message, kind)
		}
		owner.NetCode = 0xfedcba98
	}
	deps.audio = func(id uint32, gotOwner *Object, kind int32, code uint32) {
		events = append(events, "audio")
		if id != 925 || gotOwner != owner || kind != 2 || code != 0xfedcba98 {
			t.Fatalf("audio = %d/%p/%d/%08x", id, gotOwner, kind, code)
		}
	}
	deps.gameFlag = func(uint32) int32 { t.Fatal("rejected trace checked game flag"); return 0 }
	deps.dispatch = func(*Object, *Object, *types.Pointf) int32 { t.Fatal("rejected trace dispatched"); return 0 }

	if got := objectDropBoundedNative4ED810(owner, &Object{}, &types.Pointf{X: 3, Y: 4}, deps); got != 0 {
		t.Fatalf("result = %d, want 0", got)
	}
	if !reflect.DeepEqual(events, []string{"trace", "message", "audio"}) {
		t.Fatalf("events = %v", events)
	}
}

func TestObjectDropBoundedNative4ED810KOTRUsesSharedCache(t *testing.T) {
	owner := &Object{}
	item := &Object{TypeInd: 0}
	events := make([]string, 0, 5)
	deps := defaultObjectDropBoundedNativeDeps4ED810()
	deps.gameFlag = func(uint32) int32 {
		events = append(events, "game")
		return -1
	}
	deps.loadCrownTypeCache = func() uint32 {
		events = append(events, "load")
		return 0
	}
	deps.lookupCrownType = func() uint32 {
		events = append(events, "lookup")
		return 0
	}
	deps.storeCrownTypeCache = func(value uint32) {
		events = append(events, "store")
		if value != 0 {
			t.Fatalf("stored Crown type = %d", value)
		}
	}
	deps.dispatch = func(*Object, *Object, *types.Pointf) int32 { t.Fatal("Crown dispatched"); return 0 }

	if got := objectDropBoundedNative4ED810(owner, item, &types.Pointf{}, deps); got != 0 {
		t.Fatalf("result = %d, want 0", got)
	}
	if !reflect.DeepEqual(events, []string{"game", "load", "lookup", "store"}) {
		t.Fatalf("events = %v", events)
	}
}

func TestObjectDropBoundedNative4ED810NilItemBoundary(t *testing.T) {
	owner := &Object{}
	point := &types.Pointf{}
	deps := defaultObjectDropBoundedNativeDeps4ED810()
	deps.dispatch = func(gotOwner, gotItem *Object, gotPoint *types.Pointf) int32 {
		if gotOwner != owner || gotItem != nil || gotPoint == point {
			t.Fatalf("dispatch = %p/%p/%p", gotOwner, gotItem, gotPoint)
		}
		return 0x76543210
	}
	if got := objectDropBoundedNative4ED810(owner, nil, point, deps); got != 0x76543210 {
		t.Fatalf("non-KOTR result = %d", got)
	}

	deps.gameFlag = func(uint32) int32 { return 1 }
	deps.loadCrownTypeCache = func() uint32 { return 7 }
	defer func() {
		if recover() == nil {
			t.Fatal("KOTR nil item did not fault at TypeInd")
		}
	}()
	_ = objectDropBoundedNative4ED810(owner, nil, point, deps)
}
