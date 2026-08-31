package server

import (
	"fmt"
	"math"
	"reflect"
	"testing"
)

type readUseTestObject53F7C0 struct {
	name  string
	class uint8
	data  *readUseTestData53F7C0
}

type readUseTestData53F7C0 struct {
	name  string
	state uint32
}

type readUseTestWorld53F7C0 struct {
	owner       *readUseTestObject53F7C0
	readable    *readUseTestObject53F7C0
	fps         uint32
	frame       uint32
	mapResult   int32
	events      []string
	faultAt     int
	frameLoads  int
	messageData *readUseTestData53F7C0
}

func (w *readUseTestWorld53F7C0) event(value string) {
	w.events = append(w.events, value)
	if w.faultAt != 0 && len(w.events) == w.faultAt {
		panic(value)
	}
}

func readUseTestObjectName53F7C0(obj *readUseTestObject53F7C0) string {
	if obj == nil {
		return "nil"
	}
	return obj.name
}

func readUseTestDataName53F7C0(data *readUseTestData53F7C0) string {
	if data == nil {
		return "nil"
	}
	return data.name
}

func (w *readUseTestWorld53F7C0) hooks() readUseHooks53F7C0[
	*readUseTestObject53F7C0,
	*readUseTestData53F7C0,
] {
	return readUseHooks53F7C0[*readUseTestObject53F7C0, *readUseTestData53F7C0]{
		loadOwnerArg: func() *readUseTestObject53F7C0 {
			w.event("owner:" + readUseTestObjectName53F7C0(w.owner))
			return w.owner
		},
		loadClassLow: func(owner *readUseTestObject53F7C0) uint8 {
			w.event(fmt.Sprintf("class:%s=%02x", readUseTestObjectName53F7C0(owner), owner.class))
			return owner.class
		},
		loadReadableArg: func() *readUseTestObject53F7C0 {
			w.event("readable:" + readUseTestObjectName53F7C0(w.readable))
			return w.readable
		},
		loadFPS: func() uint32 {
			w.event(fmt.Sprintf("fps:%08x", w.fps))
			return w.fps
		},
		loadFrame: func() uint32 {
			w.frameLoads++
			w.event(fmt.Sprintf("frame:%d=%08x", w.frameLoads, w.frame))
			return w.frame
		},
		loadUseData: func(readable *readUseTestObject53F7C0) *readUseTestData53F7C0 {
			w.event("data:" + readUseTestObjectName53F7C0(readable))
			return readable.data
		},
		loadReadState: func(data *readUseTestData53F7C0) uint32 {
			w.event(fmt.Sprintf("state:%s=%08x", readUseTestDataName53F7C0(data), data.state))
			return data.state
		},
		mapCheck: func(owner, readable *readUseTestObject53F7C0) int32 {
			w.event(fmt.Sprintf("map:%s:%s=%08x", readUseTestObjectName53F7C0(owner), readUseTestObjectName53F7C0(readable), uint32(w.mapResult)))
			return w.mapResult
		},
		primaryMessage: func(owner *readUseTestObject53F7C0, data *readUseTestData53F7C0, value uint8) {
			w.messageData = data
			w.event(fmt.Sprintf("message:%s:%s:%d", readUseTestObjectName53F7C0(owner), readUseTestDataName53F7C0(data), value))
			w.frame = 0x89abcdef
		},
		storeReadState: func(data *readUseTestData53F7C0, frame uint32) {
			w.event(fmt.Sprintf("store:%s=%08x", readUseTestDataName53F7C0(data), frame))
			data.state = frame
		},
	}
}

func newReadUseTestWorld53F7C0() *readUseTestWorld53F7C0 {
	data := &readUseTestData53F7C0{name: "data"}
	return &readUseTestWorld53F7C0{
		owner:     &readUseTestObject53F7C0{name: "owner", class: 0xf4},
		readable:  &readUseTestObject53F7C0{name: "readable", data: data},
		fps:       20,
		frame:     100,
		mapResult: 1,
	}
}

func readUseSuccessTrace53F7C0() []string {
	return []string{
		"owner:owner",
		"class:owner=f4",
		"readable:readable",
		"fps:00000014",
		"frame:1=00000064",
		"data:readable",
		"state:data=00000000",
		"map:owner:readable=00000001",
		"message:owner:data:1",
		"frame:2=89abcdef",
		"store:data=89abcdef",
	}
}

func TestReadUse53F7C0ExactSuccessTraceAndFaultPrefixes(t *testing.T) {
	want := readUseSuccessTrace53F7C0()
	build := func() *readUseTestWorld53F7C0 {
		return newReadUseTestWorld53F7C0()
	}

	w := build()
	if got := readUse53F7C0(w.hooks()); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %v, want %v", w.events, want)
	}
	if w.messageData != w.readable.data || w.readable.data.state != 0x89abcdef {
		t.Fatalf("message data/state = %p/%#x", w.messageData, w.readable.data.state)
	}

	for faultAt := 1; faultAt <= len(want); faultAt++ {
		t.Run(fmt.Sprintf("fault-%02d", faultAt), func(t *testing.T) {
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
			readUse53F7C0(w.hooks())
		})
	}
}

func TestReadUse53F7C0NonPlayerReturnsBeforeReadable(t *testing.T) {
	w := newReadUseTestWorld53F7C0()
	w.owner.class = 0xf0
	w.readable = nil
	if got := readUse53F7C0(w.hooks()); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	want := []string{"owner:owner", "class:owner=f0"}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %v, want %v", w.events, want)
	}
}

func TestReadUse53F7C0UnsignedCooldown(t *testing.T) {
	tests := []struct {
		name    string
		fps     uint32
		frame   uint32
		state   uint32
		wantMap bool
	}{
		{name: "zero state bypasses cooldown", fps: 20, frame: 1, state: 0, wantMap: true},
		{name: "equal threshold is blocked", fps: 20, frame: 61, state: 1},
		{name: "above threshold is stale", fps: 20, frame: 62, state: 1, wantMap: true},
		{name: "subtraction wraps unsigned", fps: 1, frame: 1, state: math.MaxUint32},
		{name: "three-times-fps wraps uint32", fps: 0x80000000, frame: 0x80000002, state: 1, wantMap: true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			w := newReadUseTestWorld53F7C0()
			w.fps = tc.fps
			w.frame = tc.frame
			w.readable.data.state = tc.state
			if got := readUse53F7C0(w.hooks()); got != 1 {
				t.Fatalf("result = %d, want 1", got)
			}
			mapped := false
			for _, event := range w.events {
				if len(event) >= 4 && event[:4] == "map:" {
					mapped = true
				}
			}
			if mapped != tc.wantMap {
				t.Fatalf("mapped = %t, want %t; events = %v", mapped, tc.wantMap, w.events)
			}
		})
	}
}

func TestReadUse53F7C0RequiresExactOneMapResult(t *testing.T) {
	for _, result := range []int32{0, -1, 2, math.MinInt32, math.MaxInt32} {
		t.Run(fmt.Sprintf("%08x", uint32(result)), func(t *testing.T) {
			w := newReadUseTestWorld53F7C0()
			w.mapResult = result
			if got := readUse53F7C0(w.hooks()); got != 1 {
				t.Fatalf("result = %d, want 1", got)
			}
			if w.messageData != nil || w.readable.data.state != 0 {
				t.Fatalf("message/state = %p/%#x", w.messageData, w.readable.data.state)
			}
		})
	}
}
