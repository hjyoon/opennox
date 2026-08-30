package server

import (
	"fmt"
	"reflect"
	"testing"
)

type winkGameBallReleaseTestObject4F7DF0 struct {
	name    string
	typeInd uint16
	flags   uint32
	first   *winkGameBallReleaseTestObject4F7DF0
	next    *winkGameBallReleaseTestObject4F7DF0
	obj130  *winkGameBallReleaseTestObject4F7DF0
	owner   *winkGameBallReleaseTestObject4F7DF0
}

type winkGameBallReleaseTestWorld4F7DF0 struct {
	cache       uint32
	lookupValue uint32
	player      *winkGameBallReleaseTestObject4F7DF0
	wrong       *winkGameBallReleaseTestObject4F7DF0
	ball        *winkGameBallReleaseTestObject4F7DF0
	marker      *winkGameBallReleaseTestObject4F7DF0
	events      []string
	faultAt     int
}

func newWinkGameBallReleaseTestWorld4F7DF0() *winkGameBallReleaseTestWorld4F7DF0 {
	marker := &winkGameBallReleaseTestObject4F7DF0{name: "marker"}
	player := &winkGameBallReleaseTestObject4F7DF0{name: "player"}
	ball := &winkGameBallReleaseTestObject4F7DF0{
		name:    "ball",
		typeInd: 0x2468,
		flags:   0xa5b6c7ff,
		obj130:  marker,
		owner:   player,
	}
	wrong := &winkGameBallReleaseTestObject4F7DF0{
		name:    "wrong",
		typeInd: 7,
		next:    ball,
	}
	player.first = wrong
	return &winkGameBallReleaseTestWorld4F7DF0{
		lookupValue: 0x2468,
		player:      player,
		wrong:       wrong,
		ball:        ball,
		marker:      marker,
	}
}

func winkGameBallReleaseTestObjectName4F7DF0(obj *winkGameBallReleaseTestObject4F7DF0) string {
	if obj == nil {
		return "nil"
	}
	return obj.name
}

func (w *winkGameBallReleaseTestWorld4F7DF0) event(event string) {
	w.events = append(w.events, event)
	if w.faultAt != 0 && len(w.events) == w.faultAt {
		panic(event)
	}
}

func (w *winkGameBallReleaseTestWorld4F7DF0) hooks() winkGameBallReleaseHooks4F7DF0[*winkGameBallReleaseTestObject4F7DF0] {
	return winkGameBallReleaseHooks4F7DF0[*winkGameBallReleaseTestObject4F7DF0]{
		loadTypeCache: func() uint32 {
			w.event(fmt.Sprintf("cache:%#x", w.cache))
			return w.cache
		},
		lookupType: func(name string) uint32 {
			w.event("lookup:" + name)
			return w.lookupValue
		},
		storeTypeCache: func(value uint32) {
			w.event(fmt.Sprintf("store-cache:%#x", value))
			w.cache = value
		},
		loadFirstOwned: func(owner *winkGameBallReleaseTestObject4F7DF0) *winkGameBallReleaseTestObject4F7DF0 {
			w.event("first:" + winkGameBallReleaseTestObjectName4F7DF0(owner))
			return owner.first
		},
		loadTypeInd: func(obj *winkGameBallReleaseTestObject4F7DF0) uint16 {
			w.event(fmt.Sprintf("type:%s:%#x", winkGameBallReleaseTestObjectName4F7DF0(obj), obj.typeInd))
			return obj.typeInd
		},
		loadNextOwned: func(obj *winkGameBallReleaseTestObject4F7DF0) *winkGameBallReleaseTestObject4F7DF0 {
			w.event("next:" + obj.name + ":" + winkGameBallReleaseTestObjectName4F7DF0(obj.next))
			return obj.next
		},
		loadFlags: func(obj *winkGameBallReleaseTestObject4F7DF0) uint32 {
			w.event(fmt.Sprintf("flags:%s:%#x", obj.name, obj.flags))
			return obj.flags
		},
		storeFlags: func(obj *winkGameBallReleaseTestObject4F7DF0, flags uint32) {
			w.event(fmt.Sprintf("store-flags:%s:%#x", obj.name, flags))
			obj.flags = flags
		},
		applyForce: func(player, ball *winkGameBallReleaseTestObject4F7DF0, force float32) {
			w.event(fmt.Sprintf("force:%s:%s:%g:%#x", player.name, ball.name, force, ball.flags))
		},
		storeObj130: func(obj, value *winkGameBallReleaseTestObject4F7DF0) {
			w.event("store-obj130:" + obj.name + ":" + winkGameBallReleaseTestObjectName4F7DF0(value))
			obj.obj130 = value
		},
		clearOwner: func(obj *winkGameBallReleaseTestObject4F7DF0) {
			w.event("clear-owner:" + obj.name + ":" + winkGameBallReleaseTestObjectName4F7DF0(obj.obj130))
			obj.owner = nil
		},
		audio: func(id uint32, obj *winkGameBallReleaseTestObject4F7DF0, kind int32, code uint32) {
			w.event(fmt.Sprintf("audio:%d:%s:%d:%d", id, obj.name, kind, code))
		},
		ballStatus: func(state uint8, netCode uint16) int32 {
			w.event(fmt.Sprintf("status:%d:%d", state, netCode))
			return -0x1234567
		},
	}
}

func winkGameBallReleaseSuccessEvents4F7DF0() []string {
	return []string{
		"cache:0x0",
		"lookup:GameBall",
		"store-cache:0x2468",
		"first:player",
		"type:wrong:0x7",
		"next:wrong:ball",
		"type:ball:0x2468",
		"flags:ball:0xa5b6c7ff",
		"store-flags:ball:0xa5b6c7bf",
		"force:player:ball:100:0xa5b6c7bf",
		"store-obj130:ball:nil",
		"clear-owner:ball:nil",
		"audio:926:player:0:0",
		"status:1:0",
	}
}

func TestWinkGameBallRelease4F7DF0ExactSuccessTraceAndFaultPrefixes(t *testing.T) {
	w := newWinkGameBallReleaseTestWorld4F7DF0()
	if got := winkGameBallRelease4F7DF0(w.player, w.hooks()); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	want := winkGameBallReleaseSuccessEvents4F7DF0()
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %#v, want %#v", w.events, want)
	}
	if w.cache != 0x2468 {
		t.Fatalf("cache = %#x, want 0x2468", w.cache)
	}
	if w.ball.flags != 0xa5b6c7bf {
		t.Fatalf("ball flags = %#x, want 0xa5b6c7bf", w.ball.flags)
	}
	if w.ball.obj130 != nil || w.ball.owner != nil {
		t.Fatalf("released pointers = (%p, %p), want nil", w.ball.obj130, w.ball.owner)
	}
	if w.wrong.next != w.ball || w.player.first != w.wrong || w.marker.name != "marker" {
		t.Fatal("release mutated unrelated owned-list or marker state")
	}

	for faultAt := 1; faultAt <= len(want); faultAt++ {
		t.Run(fmt.Sprintf("fault-%d", faultAt), func(t *testing.T) {
			w := newWinkGameBallReleaseTestWorld4F7DF0()
			w.faultAt = faultAt
			defer func() {
				if recover() == nil {
					t.Fatal("hook fault did not propagate")
				}
				if !reflect.DeepEqual(w.events, want[:faultAt]) {
					t.Fatalf("events = %#v, want prefix %#v", w.events, want[:faultAt])
				}
			}()
			_ = winkGameBallRelease4F7DF0(w.player, w.hooks())
		})
	}
}

func TestWinkGameBallRelease4F7DF0CacheHitAndMissingBallShortCircuit(t *testing.T) {
	w := newWinkGameBallReleaseTestWorld4F7DF0()
	w.cache = 0x9999
	hooks := w.hooks()
	hooks.lookupType = func(string) uint32 {
		t.Fatal("cache hit performed a lookup")
		return 0
	}
	hooks.storeTypeCache = func(uint32) {
		t.Fatal("cache hit performed a store")
	}
	hooks.loadFlags = func(*winkGameBallReleaseTestObject4F7DF0) uint32 {
		t.Fatal("missing ball loaded flags")
		return 0
	}

	if got := winkGameBallRelease4F7DF0(w.player, hooks); got != 0 {
		t.Fatalf("result = %d, want 0", got)
	}
	want := []string{
		"cache:0x9999", "first:player", "type:wrong:0x7", "next:wrong:ball",
		"type:ball:0x2468", "next:ball:nil",
	}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %#v, want %#v", w.events, want)
	}
	if w.ball.flags != 0xa5b6c7ff || w.ball.obj130 != w.marker || w.ball.owner != w.player {
		t.Fatal("missing-ball path changed release state")
	}
}

func TestWinkGameBallRelease4F7DF0CacheMissPrecedesNilPlayerFault(t *testing.T) {
	w := newWinkGameBallReleaseTestWorld4F7DF0()
	defer func() {
		if recover() == nil {
			t.Fatal("nil player did not fault")
		}
		want := []string{"cache:0x0", "lookup:GameBall", "store-cache:0x2468", "first:nil"}
		if !reflect.DeepEqual(w.events, want) {
			t.Fatalf("events = %#v, want %#v", w.events, want)
		}
		if w.cache != 0x2468 {
			t.Fatalf("cache = %#x, want initialized value", w.cache)
		}
	}()
	_ = winkGameBallRelease4F7DF0[*winkGameBallReleaseTestObject4F7DF0](nil, w.hooks())
}

func TestWinkGameBallRelease4F7DF0ZeroLookupAndZeroExtendedType(t *testing.T) {
	t.Run("zero lookup can match and repeats", func(t *testing.T) {
		w := newWinkGameBallReleaseTestWorld4F7DF0()
		w.lookupValue = 0
		w.player.first = w.ball
		w.ball.typeInd = 0
		for i := 0; i < 2; i++ {
			if got := winkGameBallRelease4F7DF0(w.player, w.hooks()); got != 1 {
				t.Fatalf("call %d result = %d, want 1", i+1, got)
			}
		}
		lookups := 0
		for _, event := range w.events {
			if event == "lookup:GameBall" {
				lookups++
			}
		}
		if lookups != 2 || w.cache != 0 {
			t.Fatalf("lookups/cache = %d/%#x, want 2/0", lookups, w.cache)
		}
	})

	t.Run("type index is zero extended", func(t *testing.T) {
		w := newWinkGameBallReleaseTestWorld4F7DF0()
		w.cache = 0xffff2468
		w.player.first = w.ball
		if got := winkGameBallRelease4F7DF0(w.player, w.hooks()); got != 0 {
			t.Fatalf("result = %d, want 0", got)
		}
		want := []string{"cache:0xffff2468", "first:player", "type:ball:0x2468", "next:ball:nil"}
		if !reflect.DeepEqual(w.events, want) {
			t.Fatalf("events = %#v, want %#v", w.events, want)
		}
	})
}
