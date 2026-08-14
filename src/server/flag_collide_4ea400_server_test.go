package server

import (
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
)

func defaultFlagCollideNativeDeps4EA400() flagCollideNativeDeps4EA400 {
	return flagCollideNativeDeps4EA400{
		hasGameFlag:       func(uint32) int32 { return 0 },
		loadGameBallCache: func() uint32 { return 0 },
		lookupGameBall:    func(string) uint32 { return 0 },
		storeGameBall:     func(uint32) {},
		pickupCTF:         func(*Object, *Object, *types.Pointf) {},
		pickupGameBall:    func(*Object, *Object, *types.Pointf) {},
	}
}

func TestFlagCollide4EA400NativeLayout(t *testing.T) {
	wantType := uintptr(4)
	wantClass := uintptr(8)
	wantFlags := uintptr(16)
	wantSize := uintptr(780)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantType = 8
		wantClass = 12
		wantFlags = 20
		wantSize = 928
	}
	checks := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"Object size", unsafe.Sizeof(Object{}), wantSize},
		{"Object.TypeInd", unsafe.Offsetof(Object{}.TypeInd), wantType},
		{"Object.ObjClass", unsafe.Offsetof(Object{}.ObjClass), wantClass},
		{"Object.ObjFlags", unsafe.Offsetof(Object{}.ObjFlags), wantFlags},
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Errorf("%s = %d, want %d", check.name, check.got, check.want)
		}
	}
}

func TestFlagCollideNative4EA400UsesLiveFieldsAndExactOrder(t *testing.T) {
	source := &Object{}
	target := &Object{TypeInd: 1, ObjClass: object.ClassMonster}
	collision := &types.Pointf{X: 3.5, Y: -8.25}
	events := make([]string, 0, 8)
	deps := defaultFlagCollideNativeDeps4EA400()
	deps.hasGameFlag = func(mask uint32) int32 {
		if mask == 0x20 {
			events = append(events, "mode32")
			return 0
		}
		events = append(events, "mode64")
		return 1
	}
	deps.loadGameBallCache = func() uint32 {
		events = append(events, "cache")
		return 0
	}
	deps.lookupGameBall = func(name string) uint32 {
		events = append(events, "lookup")
		target.TypeInd = 9
		return 8
	}
	deps.storeGameBall = func(ind uint32) {
		events = append(events, "store")
		target.ObjClass = object.ClassPlayer
	}
	deps.pickupGameBall = func(gotSource, gotTarget *Object, gotCollision *types.Pointf) {
		events = append(events, "ball")
		if gotSource != source || gotTarget != target || gotCollision != collision {
			t.Fatal("native FlagBall handler did not receive original pointers")
		}
	}
	flagCollideNative4EA400(source, target, collision, deps)
	want := []string{"mode32", "mode64", "cache", "lookup", "store", "ball"}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
}

func TestFlagCollide4EA400ServerBindingNilTarget(t *testing.T) {
	(&Server{}).FlagCollide4EA400(nil, nil, nil, FlagCollideRuntime4EA400{})
}
