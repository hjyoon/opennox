package server

import (
	"fmt"
	"math"
	"reflect"
	"testing"
)

type mimicMorphTestAction534950 struct {
	action uint32
	x, y   float32
}

type mimicMorphTestUpdate534950 struct {
	index         int8
	head          *mimicMorphTestAction534950
	status, timer uint32
}

type mimicMorphTestObject534950 struct {
	update *mimicMorphTestUpdate534950
	x, y   float32
}

type mimicMorphTestWorld534950 struct {
	t               *testing.T
	obj             *mimicMorphTestObject534950
	frame, tickRate uint32
	events          []string
	fault           string
	after           func(string)
	nilPushes       bool
}

func newMimicMorphTestWorld534950(t *testing.T) *mimicMorphTestWorld534950 {
	return &mimicMorphTestWorld534950{
		t: t,
		obj: &mimicMorphTestObject534950{update: &mimicMorphTestUpdate534950{
			index: 7, head: &mimicMorphTestAction534950{}, timer: 100,
		}},
		frame: 131, tickRate: 30,
	}
}

func (w *mimicMorphTestWorld534950) record(event string) {
	w.events = append(w.events, event)
	if w.fault == event {
		panic(event)
	}
	if w.after != nil {
		w.after(event)
	}
}

func (w *mimicMorphTestWorld534950) hooks() mimicCheckMorphHooks534950[
	*mimicMorphTestObject534950, *mimicMorphTestUpdate534950, *mimicMorphTestAction534950,
] {
	return mimicCheckMorphHooks534950[*mimicMorphTestObject534950, *mimicMorphTestUpdate534950, *mimicMorphTestAction534950]{
		update: func(obj *mimicMorphTestObject534950) *mimicMorphTestUpdate534950 {
			w.record("update")
			return obj.update
		},
		stackIndex: func(update *mimicMorphTestUpdate534950) int8 {
			w.record("index")
			return update.index
		},
		head: func(update *mimicMorphTestUpdate534950, index int8) *mimicMorphTestAction534950 {
			w.record(fmt.Sprintf("head:%d", index))
			return update.head
		},
		action: func(head *mimicMorphTestAction534950) uint32 {
			w.record("action")
			return head.action
		},
		targetX: func(head *mimicMorphTestAction534950) float32 {
			w.record("target-x")
			return head.x
		},
		posX: func(obj *mimicMorphTestObject534950) float32 {
			w.record("object-x")
			return obj.x
		},
		targetY: func(head *mimicMorphTestAction534950) float32 {
			w.record("target-y")
			return head.y
		},
		posY: func(obj *mimicMorphTestObject534950) float32 {
			w.record("object-y")
			return obj.y
		},
		status: func(update *mimicMorphTestUpdate534950) uint32 {
			w.record("status")
			return update.status
		},
		frame: func() uint32 {
			w.record("frame")
			return w.frame
		},
		timer: func(update *mimicMorphTestUpdate534950) uint32 {
			w.record("timer")
			return update.timer
		},
		tickRate: func() uint32 {
			w.record("tick-rate")
			return w.tickRate
		},
		pushAction: func(obj *mimicMorphTestObject534950, action uint32) *mimicMorphTestAction534950 {
			w.record(fmt.Sprintf("push:%d", action))
			if obj != w.obj {
				w.t.Fatal("push object identity changed")
			}
			if w.nilPushes {
				return nil
			}
			return &mimicMorphTestAction534950{action: action}
		},
		audio: func(id uint32, obj *mimicMorphTestObject534950, kind int32, code uint32) {
			w.record(fmt.Sprintf("audio:%d:%d:%d", id, kind, code))
			if obj != w.obj {
				w.t.Fatal("audio object identity changed")
			}
		},
	}
}

func (w *mimicMorphTestWorld534950) run() { mimicCheckMorph534950(w.obj, w.hooks()) }

func TestMimicCheckMorph534950BranchesAndReadOrder(t *testing.T) {
	tests := []struct {
		name    string
		action  uint32
		status  uint32
		targetX float32
		want    []string
	}{
		{"idle creature", 0, 0, 0, []string{"status", "frame", "timer", "tick-rate", "push:61", "push:33", "audio:460:0:0"}},
		{"idle chest", 0, 0x40000, 0, []string{"status"}},
		{"active chest", 15, 0x40000, 0, []string{"status", "push:61", "push:34", "audio:460:0:0"}},
		{"active creature", 15, 0, 0, []string{"status"}},
		{"morphing to creature", 34, 0x40000, 0, nil},
		{"morphing creature without flag", 34, 0, 0, nil},
		{"morphing to chest is still active", 33, 0x40000, 0, []string{"status", "push:61", "push:34", "audio:460:0:0"}},
		{"high action dword", math.MaxUint32, 0x40000, 0, []string{"status", "push:61", "push:34", "audio:460:0:0"}},
		{"only chest bit tests", 1, 0xfffbffff, 0, []string{"status"}},
		{"all status bits", 1, math.MaxUint32, 0, []string{"status", "push:61", "push:34", "audio:460:0:0"}},
		{"move far chest", 4, 0x40000, 9, []string{"target-x", "object-x", "target-y", "object-y", "status", "push:61", "push:34", "audio:460:0:0"}},
		{"move far creature", 4, 0, 9, []string{"target-x", "object-x", "target-y", "object-y", "status"}},
		{"move at limit chest", 4, 0x40000, 8, []string{"target-x", "object-x", "target-y", "object-y", "status"}},
		{"move at limit creature", 4, 0, 8, []string{"target-x", "object-x", "target-y", "object-y", "status", "frame", "timer", "tick-rate", "push:61", "push:33", "audio:460:0:0"}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			w := newMimicMorphTestWorld534950(t)
			w.obj.update.head.action, w.obj.update.head.x, w.obj.update.status = test.action, test.targetX, test.status
			w.run()
			want := append([]string{"update", "index", "head:7", "action"}, test.want...)
			if !reflect.DeepEqual(w.events, want) {
				t.Fatalf("events=%v want=%v", w.events, want)
			}
		})
	}
}

func TestMimicCheckMorph534950UnspilledDistanceAndUnordered(t *testing.T) {
	tests := []struct {
		name                   string
		targetX, x, targetY, y float32
		active                 bool
	}{
		{"equal", 8, 0, 0, 0, false},
		{"next binary32 above", math.Nextafter32(8, 9), 0, 0, 0, true},
		{"next binary32 below", math.Nextafter32(8, 7), 0, 0, 0, false},
		{"negative delta", -8, 0, 0, 0, false},
		{"unspilled subtraction above", 8, -0x1p-22, 0, 0, true},
		{"unspilled subtraction below", 8, 0x1p-22, 0, 0, false},
		{"unspilled sum above", 8, 0, 0x1p-11, 0, true},
		{"53-bit sum rounds to boundary", 8, 0, 0x1p-26, 0, false},
		{"signed zero", math.Float32frombits(0x80000000), 0, 0, 0, false},
		{"NaN target X", math.Float32frombits(0x7fa12345), 0, 0, 0, false},
		{"NaN object Y", 9, 0, 0, math.Float32frombits(0xffcabcde), false},
		{"positive infinity", float32(math.Inf(1)), 0, 0, 0, true},
		{"negative infinity", float32(math.Inf(-1)), 0, 0, 0, true},
		{"infinity minus infinity unordered", float32(math.Inf(1)), float32(math.Inf(1)), 0, 0, false},
		{"square beyond binary32 range", math.MaxFloat32, 0, 0, 0, true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			w := newMimicMorphTestWorld534950(t)
			w.obj.update.head.action = 4
			w.obj.update.head.x, w.obj.x, w.obj.update.head.y, w.obj.y = test.targetX, test.x, test.targetY, test.y
			w.run()
			want := []string{"update", "index", "head:7", "action", "target-x", "object-x", "target-y", "object-y", "status"}
			if !test.active {
				want = append(want, "frame", "timer", "tick-rate", "push:61", "push:33", "audio:460:0:0")
			}
			if !reflect.DeepEqual(w.events, want) {
				t.Fatalf("distance branch events=%v want=%v", w.events, want)
			}
		})
	}
}

func TestMimicCheckMorph534950UnsignedTimerBoundaryAndWrap(t *testing.T) {
	tests := []struct {
		name                   string
		frame, timer, tickRate uint32
		push                   bool
	}{
		{"equal", 130, 100, 30, false},
		{"one above", 131, 100, 30, true},
		{"zero elapsed", 100, 100, 0, false},
		{"zero tick rate", 101, 100, 0, true},
		{"timer ahead wraps", 100, 101, 30, true},
		{"frame rollover equal", 2, 0xffffffe4, 30, false},
		{"frame rollover above", 2, 0xffffffe3, 30, true},
		{"unsigned high tick rate equal", 0x80000000, 0, 0x80000000, false},
		{"unsigned high tick rate above", 0x80000001, 0, 0x80000000, true},
		{"largest tick rate", 0xffffffff, 0, 0xffffffff, false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			w := newMimicMorphTestWorld534950(t)
			w.frame, w.obj.update.timer, w.tickRate = test.frame, test.timer, test.tickRate
			w.run()
			want := []string{"update", "index", "head:7", "action", "status", "frame", "timer", "tick-rate"}
			if test.push {
				want = append(want, "push:61", "push:33", "audio:460:0:0")
			}
			if !reflect.DeepEqual(w.events, want) {
				t.Fatalf("unsigned timer events=%v want=%v", w.events, want)
			}
		})
	}
}

func TestMimicCheckMorph534950CachedPointersAndLiveCoordinateReads(t *testing.T) {
	w := newMimicMorphTestWorld534950(t)
	entryUpdate, entryHead := w.obj.update, w.obj.update.head
	entryHead.action = 4
	w.after = func(event string) {
		switch event {
		case "index":
			// The already-loaded update remains the source of head, status and timer.
			w.obj.update = &mimicMorphTestUpdate534950{index: 1, status: 0x40000, timer: 9999, head: &mimicMorphTestAction534950{action: 34}}
		case "action":
			entryUpdate.index, entryUpdate.head = 23, &mimicMorphTestAction534950{action: 34}
		case "target-x":
			entryHead.action, entryHead.x, w.obj.x = 34, 8, -0x1p-22
		case "object-x":
			entryHead.y = 0x1p-11
		case "target-y":
			w.obj.y = 0x1p-11
		case "object-y":
			entryUpdate.status = 0x40000
		case "push:61":
			// No status/head/time reload and no result dereference after either push.
			entryUpdate.status = 0
			w.obj.update = nil
		}
	}
	w.nilPushes = true
	w.run()
	want := []string{"update", "index", "head:7", "action", "target-x", "object-x", "target-y", "object-y", "status", "push:61", "push:34", "audio:460:0:0"}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("cached/live reads=%v want=%v", w.events, want)
	}
}

func TestMimicCheckMorph534950FrameBeforeLiveTimerBeforeTickRate(t *testing.T) {
	w := newMimicMorphTestWorld534950(t)
	update := w.obj.update
	w.after = func(event string) {
		switch event {
		case "frame":
			update.timer = 101 // elapsed now equals 30, not the entry value 31
			w.tickRate = 29
			w.obj.update = nil
		case "timer":
			w.frame = 9999 // the frame read is cached
		case "tick-rate":
			update.timer = 9999 // the timer read is cached
		}
	}
	w.nilPushes = true
	w.run()
	want := []string{"update", "index", "head:7", "action", "status", "frame", "timer", "tick-rate", "push:61", "push:33", "audio:460:0:0"}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("timer read order=%v want=%v", w.events, want)
	}
}

func TestMimicCheckMorph534950SignedIndexAndFailedPushes(t *testing.T) {
	for _, index := range []int8{-128, -1, 0, 23, 127} {
		t.Run(fmt.Sprint(index), func(t *testing.T) {
			w := newMimicMorphTestWorld534950(t)
			w.obj.update.index, w.nilPushes = index, true
			w.run()
			want := []string{"update", "index", fmt.Sprintf("head:%d", index), "action", "status", "frame", "timer", "tick-rate", "push:61", "push:33", "audio:460:0:0"}
			if !reflect.DeepEqual(w.events, want) {
				t.Fatalf("signed index/order=%v want=%v", w.events, want)
			}
		})
	}
}

func TestMimicCheckMorph534950FaultPrefixes(t *testing.T) {
	for _, active := range []bool{false, true} {
		seed := newMimicMorphTestWorld534950(t)
		if active {
			seed.obj.update.head.action, seed.obj.update.head.x, seed.obj.update.status = 4, 9, 0x40000
		}
		seed.run()
		for i, event := range seed.events {
			t.Run(fmt.Sprintf("active=%v/%s", active, event), func(t *testing.T) {
				w := newMimicMorphTestWorld534950(t)
				if active {
					w.obj.update.head.action, w.obj.update.head.x, w.obj.update.status = 4, 9, 0x40000
				}
				w.fault = event
				var recovered any
				func() {
					defer func() { recovered = recover() }()
					w.run()
				}()
				if recovered != event || !reflect.DeepEqual(w.events, seed.events[:i+1]) {
					t.Fatalf("fault=%v events=%v want=%v/%v", recovered, w.events, event, seed.events[:i+1])
				}
			})
		}
	}
}
