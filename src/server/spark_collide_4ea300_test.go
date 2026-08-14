package server

import (
	"reflect"
	"testing"
)

type sparkCollideTestData4EA300 struct {
	kind uint32
}

type sparkCollideTestObject4EA300 struct {
	name      string
	update    *sparkCollideTestData4EA300
	class     uint8
	slowCount uint8
	slowTimer uint16
}

func defaultSparkCollideHooks4EA300() sparkCollideHooks4EA300[
	*sparkCollideTestObject4EA300,
	*uint32,
	*sparkCollideTestData4EA300,
] {
	return sparkCollideHooks4EA300[
		*sparkCollideTestObject4EA300,
		*uint32,
		*sparkCollideTestData4EA300,
	]{
		loadUpdateData: func(obj *sparkCollideTestObject4EA300) *sparkCollideTestData4EA300 {
			return obj.update
		},
		loadKind: func(data *sparkCollideTestData4EA300) uint32 {
			return data.kind
		},
		wallReflect:   func(*sparkCollideTestObject4EA300, *sparkCollideTestObject4EA300, *uint32) {},
		audio:         func(uint32, *sparkCollideTestObject4EA300) {},
		delayedDelete: func(*sparkCollideTestObject4EA300) {},
		loadSlowCount: func(obj *sparkCollideTestObject4EA300) uint8 {
			return obj.slowCount
		},
		loadClassLow: func(obj *sparkCollideTestObject4EA300) uint8 {
			return obj.class
		},
		storeSlowCount: func(obj *sparkCollideTestObject4EA300, count uint8) {
			obj.slowCount = count
		},
		storeSlowTimer: func(obj *sparkCollideTestObject4EA300, timer uint16) {
			obj.slowTimer = timer
		},
		priorityMessage: func(*sparkCollideTestObject4EA300, string) {},
	}
}

func TestSparkCollide4EA300OtherKindForwardsOriginalArguments(t *testing.T) {
	data := &sparkCollideTestData4EA300{kind: 3}
	source := &sparkCollideTestObject4EA300{name: "source", update: data}
	target := &sparkCollideTestObject4EA300{name: "target"}
	collision := uint32(0xa5a5a5a5)
	events := make([]string, 0, 3)
	hooks := defaultSparkCollideHooks4EA300()
	hooks.loadUpdateData = func(got *sparkCollideTestObject4EA300) *sparkCollideTestData4EA300 {
		events = append(events, "update")
		if got != source {
			t.Fatalf("update source = %p", got)
		}
		return got.update
	}
	hooks.loadKind = func(got *sparkCollideTestData4EA300) uint32 {
		events = append(events, "kind")
		if got != data {
			t.Fatalf("update data = %p", got)
		}
		got.kind = sparkCollideNoEffectKind4EA300
		return 3
	}
	hooks.wallReflect = func(gotSource, gotTarget *sparkCollideTestObject4EA300, gotCollision *uint32) {
		events = append(events, "wall-reflect")
		if gotSource != source || gotTarget != target || gotCollision != &collision {
			t.Fatalf("forward = %p/%p/%p", gotSource, gotTarget, gotCollision)
		}
	}
	sparkCollide4EA300(source, target, &collision, hooks)
	if !reflect.DeepEqual(events, []string{"update", "kind", "wall-reflect"}) {
		t.Fatalf("events = %v", events)
	}
}

func TestSparkCollide4EA300KindsFourAndFiveNilTargetReturn(t *testing.T) {
	tests := []struct {
		name string
		kind uint32
	}{
		{name: "kind four", kind: sparkCollideNoEffectKind4EA300},
		{name: "kind five nil target", kind: sparkCollideWebbingKind4EA300},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			data := &sparkCollideTestData4EA300{kind: tc.kind}
			source := &sparkCollideTestObject4EA300{update: data}
			events := make([]string, 0, 2)
			hooks := defaultSparkCollideHooks4EA300()
			hooks.loadUpdateData = func(obj *sparkCollideTestObject4EA300) *sparkCollideTestData4EA300 {
				events = append(events, "update")
				return obj.update
			}
			hooks.loadKind = func(got *sparkCollideTestData4EA300) uint32 {
				events = append(events, "kind")
				return got.kind
			}
			hooks.wallReflect = func(*sparkCollideTestObject4EA300, *sparkCollideTestObject4EA300, *uint32) {
				t.Fatal("early return forwarded to WallReflect")
			}
			hooks.audio = func(uint32, *sparkCollideTestObject4EA300) {
				t.Fatal("early return emitted audio")
			}
			sparkCollide4EA300(source, nil, (*uint32)(nil), hooks)
			if !reflect.DeepEqual(events, []string{"update", "kind"}) {
				t.Fatalf("events = %v", events)
			}
		})
	}
}

func TestSparkCollide4EA300WebbingOrderLiveLoadsWrapAndCachedClass(t *testing.T) {
	source := &sparkCollideTestObject4EA300{
		name:   "source",
		update: &sparkCollideTestData4EA300{kind: sparkCollideWebbingKind4EA300},
	}
	target := &sparkCollideTestObject4EA300{name: "target", slowCount: 1}
	events := make([]string, 0, 9)
	hooks := defaultSparkCollideHooks4EA300()
	hooks.loadUpdateData = func(obj *sparkCollideTestObject4EA300) *sparkCollideTestData4EA300 {
		events = append(events, "update")
		return obj.update
	}
	hooks.loadKind = func(data *sparkCollideTestData4EA300) uint32 {
		events = append(events, "kind")
		return data.kind
	}
	hooks.audio = func(id uint32, obj *sparkCollideTestObject4EA300) {
		events = append(events, "audio")
		if id != sparkCollideWebbingAudio4EA300 || obj != source {
			t.Fatalf("audio = %d/%p", id, obj)
		}
		target.slowCount = 0x7f
		target.class = 0
	}
	hooks.delayedDelete = func(obj *sparkCollideTestObject4EA300) {
		events = append(events, "delete")
		if obj != source {
			t.Fatalf("deleted = %p", obj)
		}
		target.slowCount = 0xff
		target.class = 0x84
	}
	hooks.loadSlowCount = func(obj *sparkCollideTestObject4EA300) uint8 {
		events = append(events, "count")
		return obj.slowCount
	}
	hooks.loadClassLow = func(obj *sparkCollideTestObject4EA300) uint8 {
		events = append(events, "class")
		return obj.class
	}
	hooks.storeSlowCount = func(obj *sparkCollideTestObject4EA300, count uint8) {
		events = append(events, "store-count")
		obj.slowCount = count
		obj.class = 0
	}
	hooks.storeSlowTimer = func(obj *sparkCollideTestObject4EA300, timer uint16) {
		events = append(events, "store-timer")
		obj.slowTimer = timer
	}
	hooks.priorityMessage = func(obj *sparkCollideTestObject4EA300, message string) {
		events = append(events, "message")
		if obj != target || message != sparkCollideWebbingMessage4EA300 {
			t.Fatalf("message = %p/%q", obj, message)
		}
	}
	sparkCollide4EA300(source, target, (*uint32)(nil), hooks)
	want := []string{
		"update", "kind", "audio", "delete", "count", "class",
		"store-count", "store-timer", "message",
	}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
	if target.slowCount != 0 || target.slowTimer != sparkCollideWebbingTimer4EA300 {
		t.Fatalf("slow state = %d/%d", target.slowCount, target.slowTimer)
	}
}

func TestSparkCollide4EA300WebbingNonPlayerStillStores(t *testing.T) {
	source := &sparkCollideTestObject4EA300{update: &sparkCollideTestData4EA300{kind: 5}}
	target := &sparkCollideTestObject4EA300{class: 0x80, slowCount: 8}
	hooks := defaultSparkCollideHooks4EA300()
	hooks.priorityMessage = func(*sparkCollideTestObject4EA300, string) {
		t.Fatal("non-Player target received priority message")
	}
	sparkCollide4EA300(source, target, (*uint32)(nil), hooks)
	if target.slowCount != 9 || target.slowTimer != 1000 {
		t.Fatalf("slow state = %d/%d", target.slowCount, target.slowTimer)
	}
}

func TestSparkCollide4EA300NilSourceFaultsBeforeBranch(t *testing.T) {
	hooks := defaultSparkCollideHooks4EA300()
	defer func() {
		if recover() == nil {
			t.Fatal("nil source did not fault while loading update data")
		}
	}()
	sparkCollide4EA300(nil, nil, (*uint32)(nil), hooks)
}
