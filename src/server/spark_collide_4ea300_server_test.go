package server

import (
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
)

func defaultSparkCollideNativeDeps4EA300() sparkCollideNativeDeps4EA300 {
	return sparkCollideNativeDeps4EA300{
		wallReflect:     func(*Object, *Object, *types.Pointf) {},
		audio:           func(uint32, *Object) {},
		delayedDelete:   func(*Object) {},
		priorityMessage: func(*Object, string) {},
	}
}

func TestSparkCollide4EA300NativeLayout(t *testing.T) {
	wantClass := uintptr(8)
	wantSlowCount := uintptr(541)
	wantSlowTimer := uintptr(542)
	wantCollideData := uintptr(700)
	wantUpdateData := uintptr(748)
	wantSize := uintptr(780)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantClass = 12
		wantSlowCount = 601
		wantSlowTimer = 602
		wantCollideData = 776
		wantUpdateData = 872
		wantSize = 928
	}
	checks := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"Object size", unsafe.Sizeof(Object{}), wantSize},
		{"Object.ObjClass", unsafe.Offsetof(Object{}.ObjClass), wantClass},
		{"Object.Field541", unsafe.Offsetof(Object{}.Field541), wantSlowCount},
		{"Object.Field542", unsafe.Offsetof(Object{}.Field542), wantSlowTimer},
		{"Object.CollideData", unsafe.Offsetof(Object{}.CollideData), wantCollideData},
		{"Object.UpdateData", unsafe.Offsetof(Object{}.UpdateData), wantUpdateData},
		{"SparkUpdateData size", unsafe.Sizeof(SparkUpdateData{}), 16},
		{"SparkUpdateData.Kind", unsafe.Offsetof(SparkUpdateData{}.Kind), 12},
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Errorf("%s = %d, want %d", check.name, check.got, check.want)
		}
	}
}

func TestSparkCollideNative4EA300WebbingUsesLiveFieldsAfterCallbacks(t *testing.T) {
	update := &SparkUpdateData{Kind: sparkCollideWebbingKind4EA300}
	source := &Object{UpdateData: unsafe.Pointer(update)}
	target := &Object{Field541: 1}
	events := make([]string, 0, 3)
	deps := defaultSparkCollideNativeDeps4EA300()
	deps.audio = func(id uint32, obj *Object) {
		events = append(events, "audio")
		if id != 351 || obj != source {
			t.Fatalf("audio = %d/%p", id, obj)
		}
		target.Field541 = 0x7f
	}
	deps.delayedDelete = func(obj *Object) {
		events = append(events, "delete")
		if obj != source {
			t.Fatalf("deleted = %p", obj)
		}
		target.Field541 = 0xff
		target.ObjClass = object.ClassPlayer
	}
	deps.priorityMessage = func(obj *Object, message string) {
		events = append(events, "message")
		if obj != target || message != sparkCollideWebbingMessage4EA300 {
			t.Fatalf("message = %p/%q", obj, message)
		}
	}
	sparkCollideNative4EA300(source, target, nil, deps)
	if !reflect.DeepEqual(events, []string{"audio", "delete", "message"}) {
		t.Fatalf("events = %v", events)
	}
	if target.Field541 != 0 || target.Field542 != 1000 {
		t.Fatalf("slow state = %d/%d", target.Field541, target.Field542)
	}
}

func TestSparkCollideNative4EA300ForwardsGenericBranch(t *testing.T) {
	update := &SparkUpdateData{Kind: 9}
	source := &Object{UpdateData: unsafe.Pointer(update)}
	target := &Object{}
	collision := &types.Pointf{X: 3.5, Y: -8.25}
	deps := defaultSparkCollideNativeDeps4EA300()
	called := false
	deps.wallReflect = func(gotSource, gotTarget *Object, gotCollision *types.Pointf) {
		called = true
		if gotSource != source || gotTarget != target || gotCollision != collision {
			t.Fatalf("forward = %p/%p/%p", gotSource, gotTarget, gotCollision)
		}
	}
	sparkCollideNative4EA300(source, target, collision, deps)
	if !called {
		t.Fatal("generic branch did not reach WallReflect")
	}
}

func TestSparkCollide4EA300ServerBindingKindFour(t *testing.T) {
	update := &SparkUpdateData{Kind: sparkCollideNoEffectKind4EA300}
	source := &Object{UpdateData: unsafe.Pointer(update)}
	(&Server{}).SparkCollide4EA300(source, nil, nil, SparkCollideRuntime4EA300{})
}
