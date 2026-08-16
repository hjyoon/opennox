package server

import (
	"fmt"
	"reflect"
	"testing"
)

type objectDropDispatchTestObject4ED790 struct {
	name  string
	class uint32
	flags uint32
	drop  string
}

type objectDropDispatchTestPoint4ED790 struct {
	name string
}

type objectDropDispatchTestWorld4ED790 struct {
	owner, item *objectDropDispatchTestObject4ED790
	point       *objectDropDispatchTestPoint4ED790

	itemArg  *objectDropDispatchTestObject4ED790
	ownerArg *objectDropDispatchTestObject4ED790
	pointArg *objectDropDispatchTestPoint4ED790
	online   int32
	quest    int32

	callbackResult int32
	defaultResult  int32
	events         []string
	faultAt        int
	afterRefresh   func(*objectDropDispatchTestWorld4ED790)
	afterLoadDrop  func(*objectDropDispatchTestWorld4ED790)
}

func objectDropDispatchObjectName4ED790(obj *objectDropDispatchTestObject4ED790) string {
	if obj == nil {
		return "nil"
	}
	return obj.name
}

func objectDropDispatchPointName4ED790(point *objectDropDispatchTestPoint4ED790) string {
	if point == nil {
		return "nil"
	}
	return point.name
}

func (w *objectDropDispatchTestWorld4ED790) event(name string) {
	w.events = append(w.events, name)
	if w.faultAt != 0 && len(w.events) == w.faultAt {
		panic(name)
	}
}

func newObjectDropDispatchTestWorld4ED790() *objectDropDispatchTestWorld4ED790 {
	owner := &objectDropDispatchTestObject4ED790{name: "owner"}
	item := &objectDropDispatchTestObject4ED790{
		name:  "item",
		class: objectDropClassMask4ED790,
		flags: 0x80000005,
		drop:  "entry-handler",
	}
	point := &objectDropDispatchTestPoint4ED790{name: "point"}
	return &objectDropDispatchTestWorld4ED790{
		owner: owner, item: item, point: point,
		ownerArg: owner, itemArg: item, pointArg: point,
		online: 1, quest: 0,
		callbackResult: int32(-0x1234567),
		defaultResult:  int32(0x76543210),
	}
}

func (w *objectDropDispatchTestWorld4ED790) hooks() objectDropDispatchHooks4ED790[
	*objectDropDispatchTestObject4ED790,
	*objectDropDispatchTestPoint4ED790,
	string,
] {
	return objectDropDispatchHooks4ED790[
		*objectDropDispatchTestObject4ED790,
		*objectDropDispatchTestPoint4ED790,
		string,
	]{
		loadItemArg: func() *objectDropDispatchTestObject4ED790 {
			w.event("item-arg:" + objectDropDispatchObjectName4ED790(w.itemArg))
			return w.itemArg
		},
		gameFlag: func(flag uint32) int32 {
			w.event(fmt.Sprintf("game:%#x", flag))
			if flag == objectDropOnlineFlag4ED790 {
				return w.online
			}
			if flag == objectDropQuestFlag4ED790 {
				return w.quest
			}
			panic(fmt.Sprintf("unexpected game flag %#x", flag))
		},
		loadClass: func(obj *objectDropDispatchTestObject4ED790) uint32 {
			w.event(fmt.Sprintf("class:%s:%#x", obj.name, obj.class))
			return obj.class
		},
		loadFlags: func(obj *objectDropDispatchTestObject4ED790) uint32 {
			w.event(fmt.Sprintf("flags:%s:%#x", obj.name, obj.flags))
			return obj.flags
		},
		storeFlags: func(obj *objectDropDispatchTestObject4ED790, flags uint32) {
			w.event(fmt.Sprintf("store-flags:%s:%#x", obj.name, flags))
			obj.flags = flags
		},
		refreshUnit: func(obj *objectDropDispatchTestObject4ED790) {
			w.event("refresh:" + obj.name)
			if w.afterRefresh != nil {
				w.afterRefresh(w)
			}
		},
		loadDrop: func(obj *objectDropDispatchTestObject4ED790) string {
			drop := obj.drop
			w.event("load-drop:" + obj.name + ":" + drop)
			if w.afterLoadDrop != nil {
				w.afterLoadDrop(w)
			}
			return drop
		},
		hasDrop: func(drop string) bool {
			w.event("has-drop:" + drop)
			return drop != ""
		},
		loadPointArg: func() *objectDropDispatchTestPoint4ED790 {
			w.event("point-arg:" + objectDropDispatchPointName4ED790(w.pointArg))
			return w.pointArg
		},
		loadOwnerArg: func() *objectDropDispatchTestObject4ED790 {
			w.event("owner-arg:" + objectDropDispatchObjectName4ED790(w.ownerArg))
			return w.ownerArg
		},
		callDrop: func(drop string, owner, item *objectDropDispatchTestObject4ED790, point *objectDropDispatchTestPoint4ED790) int32 {
			w.event(fmt.Sprintf(
				"call:%s:%s:%s:%s",
				drop,
				objectDropDispatchObjectName4ED790(owner),
				objectDropDispatchObjectName4ED790(item),
				objectDropDispatchPointName4ED790(point),
			))
			return w.callbackResult
		},
		defaultDrop: func(owner, item *objectDropDispatchTestObject4ED790, point *objectDropDispatchTestPoint4ED790) int32 {
			w.event(fmt.Sprintf(
				"default:%s:%s:%s",
				objectDropDispatchObjectName4ED790(owner),
				objectDropDispatchObjectName4ED790(item),
				objectDropDispatchPointName4ED790(point),
			))
			return w.defaultResult
		},
	}
}

func objectDropDispatchSuccessEvents4ED790() []string {
	return []string{
		"item-arg:item",
		"game:0x2000",
		"game:0x1000",
		"class:item:0x3001010",
		"flags:item:0x80000005",
		"store-flags:item:0x80000045",
		"refresh:item",
		"load-drop:item:live-handler",
		"has-drop:live-handler",
		"point-arg:point",
		"owner-arg:owner",
		"call:live-handler:owner:item:point",
	}
}

func TestObjectDropDispatch4ED790ExactTraceAndFaultPrefixes(t *testing.T) {
	w := newObjectDropDispatchTestWorld4ED790()
	w.afterRefresh = func(w *objectDropDispatchTestWorld4ED790) {
		w.item.drop = "live-handler"
	}
	if got := objectDropDispatch4ED790(w.hooks()); got != w.callbackResult {
		t.Fatalf("result = %#x, want %#x", uint32(got), uint32(w.callbackResult))
	}
	want := objectDropDispatchSuccessEvents4ED790()
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %#v, want %#v", w.events, want)
	}
	if w.item.flags != 0x80000045 {
		t.Fatalf("flags = %#x, want 0x80000045", w.item.flags)
	}

	for faultAt := 1; faultAt <= len(want); faultAt++ {
		t.Run(fmt.Sprintf("fault-%d", faultAt), func(t *testing.T) {
			w := newObjectDropDispatchTestWorld4ED790()
			w.afterRefresh = func(w *objectDropDispatchTestWorld4ED790) {
				w.item.drop = "live-handler"
			}
			w.faultAt = faultAt
			defer func() {
				if got := recover(); got != want[faultAt-1] {
					t.Fatalf("panic = %v, want %q", got, want[faultAt-1])
				}
				if !reflect.DeepEqual(w.events, want[:faultAt]) {
					t.Fatalf("events = %#v, want %#v", w.events, want[:faultAt])
				}
			}()
			objectDropDispatch4ED790(w.hooks())
		})
	}
}

func TestObjectDropDispatch4ED790NilItemStopsBeforeAllOtherReads(t *testing.T) {
	w := newObjectDropDispatchTestWorld4ED790()
	w.itemArg = nil
	if got := objectDropDispatch4ED790(w.hooks()); got != 0 {
		t.Fatalf("result = %d, want 0", got)
	}
	if want := []string{"item-arg:nil"}; !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %#v, want %#v", w.events, want)
	}
}

func TestObjectDropDispatch4ED790ModeAndClassGatesUseWholeDwords(t *testing.T) {
	tests := []struct {
		name   string
		online int32
		quest  int32
		class  uint32
		prefix []string
	}{
		{
			name: "offline",
			prefix: []string{
				"item-arg:item", "game:0x2000",
			},
		},
		{
			name:   "quest-whole-eax",
			online: -1, quest: -0x80000000,
			prefix: []string{
				"item-arg:item", "game:0x2000", "game:0x1000",
			},
		},
		{
			name:   "class-miss",
			online: -0x80000000, class: 0x80000004,
			prefix: []string{
				"item-arg:item", "game:0x2000", "game:0x1000", "class:item:0x80000004",
			},
		},
		{
			name:   "high-class-hit",
			online: -1, class: 0x01000000,
			prefix: []string{
				"item-arg:item", "game:0x2000", "game:0x1000", "class:item:0x1000000",
				"flags:item:0x80000005", "store-flags:item:0x80000045", "refresh:item",
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			w := newObjectDropDispatchTestWorld4ED790()
			w.online = tc.online
			w.quest = tc.quest
			w.item.class = tc.class
			_ = objectDropDispatch4ED790(w.hooks())
			want := append(append([]string(nil), tc.prefix...),
				"load-drop:item:entry-handler", "has-drop:entry-handler",
				"point-arg:point", "owner-arg:owner",
				"call:entry-handler:owner:item:point",
			)
			if !reflect.DeepEqual(w.events, want) {
				t.Fatalf("events = %#v, want %#v", w.events, want)
			}
		})
	}
}

func TestObjectDropDispatch4ED790DefaultFallbackAndDelayedArgs(t *testing.T) {
	w := newObjectDropDispatchTestWorld4ED790()
	w.online = 0
	w.item.drop = ""
	w.ownerArg = nil
	w.pointArg = nil
	if got := objectDropDispatch4ED790(w.hooks()); got != w.defaultResult {
		t.Fatalf("result = %#x, want %#x", uint32(got), uint32(w.defaultResult))
	}
	want := []string{
		"item-arg:item", "game:0x2000", "load-drop:item:", "has-drop:",
		"point-arg:nil", "owner-arg:nil", "default:nil:item:nil",
	}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %#v, want %#v", w.events, want)
	}
}

func TestObjectDropDispatch4ED790CachesHandlerBeforeArgumentLoads(t *testing.T) {
	w := newObjectDropDispatchTestWorld4ED790()
	w.online = 0
	w.item.drop = "cached-handler"
	w.afterLoadDrop = func(w *objectDropDispatchTestWorld4ED790) {
		w.item.drop = "replacement-handler"
		w.itemArg = &objectDropDispatchTestObject4ED790{name: "replacement-item"}
	}
	if got := objectDropDispatch4ED790(w.hooks()); got != w.callbackResult {
		t.Fatalf("result = %#x, want %#x", uint32(got), uint32(w.callbackResult))
	}
	wantTail := []string{
		"load-drop:item:cached-handler", "has-drop:cached-handler",
		"point-arg:point", "owner-arg:owner", "call:cached-handler:owner:item:point",
	}
	if got := w.events[len(w.events)-len(wantTail):]; !reflect.DeepEqual(got, wantTail) {
		t.Fatalf("event tail = %#v, want %#v", got, wantTail)
	}
}
