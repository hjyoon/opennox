package server

import (
	"fmt"
	"math"
	"reflect"
	"testing"
)

type warpReadUseTestObject53F830 struct {
	name   string
	class  uint8
	update *warpReadUseTestUpdate53F830
	data   *warpReadUseTestData53F830
}

type warpReadUseTestUpdate53F830 struct {
	name   string
	player *warpReadUseTestPlayer53F830
}

type warpReadUseTestData53F830 struct {
	name  string
	state uint32
}

type warpReadUseTestPlayer53F830 struct {
	name  string
	index uint8
}

type warpReadUseTestWorld53F830 struct {
	owner       *warpReadUseTestObject53F830
	readable    *warpReadUseTestObject53F830
	fps         uint32
	frame       uint32
	mapResult   int32
	warpOpen    int32
	questStage  int32
	nextStage   int32
	events      []string
	faultAt     int
	frameLoads  int
	messageData *warpReadUseTestData53F830
}

func (w *warpReadUseTestWorld53F830) event(value string) {
	w.events = append(w.events, value)
	if w.faultAt != 0 && len(w.events) == w.faultAt {
		panic(value)
	}
}

func warpReadUseTestObjectName53F830(obj *warpReadUseTestObject53F830) string {
	if obj == nil {
		return "nil"
	}
	return obj.name
}

func warpReadUseTestUpdateName53F830(update *warpReadUseTestUpdate53F830) string {
	if update == nil {
		return "nil"
	}
	return update.name
}

func warpReadUseTestDataName53F830(data *warpReadUseTestData53F830) string {
	if data == nil {
		return "nil"
	}
	return data.name
}

func warpReadUseTestPlayerName53F830(player *warpReadUseTestPlayer53F830) string {
	if player == nil {
		return "nil"
	}
	return player.name
}

func (w *warpReadUseTestWorld53F830) hooks() warpReadUseHooks53F830[
	*warpReadUseTestObject53F830,
	*warpReadUseTestUpdate53F830,
	*warpReadUseTestData53F830,
	*warpReadUseTestPlayer53F830,
] {
	return warpReadUseHooks53F830[
		*warpReadUseTestObject53F830,
		*warpReadUseTestUpdate53F830,
		*warpReadUseTestData53F830,
		*warpReadUseTestPlayer53F830,
	]{
		loadOwnerArg: func() *warpReadUseTestObject53F830 {
			w.event("owner:" + warpReadUseTestObjectName53F830(w.owner))
			return w.owner
		},
		loadClassLow: func(owner *warpReadUseTestObject53F830) uint8 {
			w.event(fmt.Sprintf("class:%s=%02x", warpReadUseTestObjectName53F830(owner), owner.class))
			return owner.class
		},
		loadReadableArg: func() *warpReadUseTestObject53F830 {
			w.event("readable:" + warpReadUseTestObjectName53F830(w.readable))
			return w.readable
		},
		loadFPS: func() uint32 {
			w.event(fmt.Sprintf("fps:%08x", w.fps))
			return w.fps
		},
		loadUpdateData: func(owner *warpReadUseTestObject53F830) *warpReadUseTestUpdate53F830 {
			w.event("update:" + warpReadUseTestObjectName53F830(owner))
			return owner.update
		},
		loadUseData: func(readable *warpReadUseTestObject53F830) *warpReadUseTestData53F830 {
			w.event("data:" + warpReadUseTestObjectName53F830(readable))
			return readable.data
		},
		loadFrame: func() uint32 {
			w.frameLoads++
			w.event(fmt.Sprintf("frame:%d=%08x", w.frameLoads, w.frame))
			return w.frame
		},
		loadReadState: func(data *warpReadUseTestData53F830) uint32 {
			w.event(fmt.Sprintf("state:%s=%08x", warpReadUseTestDataName53F830(data), data.state))
			return data.state
		},
		mapCheck: func(owner, readable *warpReadUseTestObject53F830) int32 {
			w.event(fmt.Sprintf("map:%s:%s=%08x", warpReadUseTestObjectName53F830(owner), warpReadUseTestObjectName53F830(readable), uint32(w.mapResult)))
			return w.mapResult
		},
		warpEnabled: func() int32 {
			w.event(fmt.Sprintf("warp:%08x", uint32(w.warpOpen)))
			return w.warpOpen
		},
		currentQuestStage: func() int32 {
			w.event(fmt.Sprintf("stage:%08x", uint32(w.questStage)))
			return w.questStage
		},
		nextStageThreshold: func(stage int32) int32 {
			w.event(fmt.Sprintf("next:%08x=%08x", uint32(stage), uint32(w.nextStage)))
			return w.nextStage
		},
		loadPlayer: func(update *warpReadUseTestUpdate53F830) *warpReadUseTestPlayer53F830 {
			w.event("player:" + warpReadUseTestUpdateName53F830(update))
			return update.player
		},
		loadPlayerIndex: func(player *warpReadUseTestPlayer53F830) uint8 {
			w.event(fmt.Sprintf("index:%s=%02x", warpReadUseTestPlayerName53F830(player), player.index))
			return player.index
		},
		informText: func(index, code uint8, value int32) {
			w.event(fmt.Sprintf("inform:%02x:%02x=%08x", index, code, uint32(value)))
			w.frame = 0x89abcdef
		},
		priorityMessage: func(owner *warpReadUseTestObject53F830, key string, value uint8) {
			w.event(fmt.Sprintf("priority:%s:%s:%d", warpReadUseTestObjectName53F830(owner), key, value))
			w.frame = 0x76543210
		},
		storeReadState: func(data *warpReadUseTestData53F830, frame uint32) {
			w.messageData = data
			w.event(fmt.Sprintf("store:%s=%08x", warpReadUseTestDataName53F830(data), frame))
			data.state = frame
		},
	}
}

func newWarpReadUseTestWorld53F830() *warpReadUseTestWorld53F830 {
	player := &warpReadUseTestPlayer53F830{name: "player", index: 0xe1}
	update := &warpReadUseTestUpdate53F830{name: "update", player: player}
	data := &warpReadUseTestData53F830{name: "data"}
	return &warpReadUseTestWorld53F830{
		owner:      &warpReadUseTestObject53F830{name: "owner", class: 0xf4, update: update},
		readable:   &warpReadUseTestObject53F830{name: "readable", data: data},
		fps:        20,
		frame:      100,
		mapResult:  1,
		warpOpen:   1,
		questStage: 7,
		nextStage:  10,
	}
}

func warpReadUseOpenTrace53F830() []string {
	return []string{
		"owner:owner",
		"class:owner=f4",
		"readable:readable",
		"fps:00000014",
		"update:owner",
		"data:readable",
		"frame:1=00000064",
		"state:data=00000000",
		"map:owner:readable=00000001",
		"warp:00000001",
		"stage:00000007",
		"next:00000007=0000000a",
		"player:update",
		"index:player=e1",
		"inform:e1:15=0000000a",
		"frame:2=89abcdef",
		"store:data=89abcdef",
	}
}

func TestWarpReadUse53F830ExactOpenTraceAndFaultPrefixes(t *testing.T) {
	want := warpReadUseOpenTrace53F830()
	w := newWarpReadUseTestWorld53F830()
	if got := warpReadUse53F830(w.hooks()); got != 1 {
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
			w := newWarpReadUseTestWorld53F830()
			w.faultAt = faultAt
			defer func() {
				if got := recover(); got != want[faultAt-1] {
					t.Fatalf("panic = %v, want %q", got, want[faultAt-1])
				}
				if !reflect.DeepEqual(w.events, want[:faultAt]) {
					t.Fatalf("events = %v, want %v", w.events, want[:faultAt])
				}
			}()
			warpReadUse53F830(w.hooks())
		})
	}
}

func TestWarpReadUse53F830ClosedTraceAndCachedPointers(t *testing.T) {
	w := newWarpReadUseTestWorld53F830()
	w.warpOpen = 0
	cachedUpdate := w.owner.update
	cachedData := w.readable.data
	hooks := w.hooks()
	originalPriority := hooks.priorityMessage
	hooks.priorityMessage = func(owner *warpReadUseTestObject53F830, key string, value uint8) {
		w.owner.update = &warpReadUseTestUpdate53F830{name: "replacement-update"}
		w.readable.data = &warpReadUseTestData53F830{name: "replacement-data"}
		originalPriority(owner, key, value)
	}
	if got := warpReadUse53F830(hooks); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	want := []string{
		"owner:owner",
		"class:owner=f4",
		"readable:readable",
		"fps:00000014",
		"update:owner",
		"data:readable",
		"frame:1=00000064",
		"state:data=00000000",
		"map:owner:readable=00000001",
		"warp:00000000",
		"priority:owner:GeneralPrint:WarpClosed:1",
		"frame:2=76543210",
		"store:data=76543210",
	}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %v, want %v", w.events, want)
	}
	if w.messageData != cachedData || cachedData.state != 0x76543210 {
		t.Fatalf("cached data/state = %p/%#x, want %p/0x76543210", w.messageData, cachedData.state, cachedData)
	}
	if cachedUpdate == w.owner.update {
		t.Fatal("test did not replace live owner UpdateData")
	}
}

func TestWarpReadUse53F830OpenReloadsPlayerFromCachedUpdate(t *testing.T) {
	w := newWarpReadUseTestWorld53F830()
	cachedUpdate := w.owner.update
	replacement := &warpReadUseTestPlayer53F830{name: "replacement-player", index: 0xfe}
	hooks := w.hooks()
	originalNext := hooks.nextStageThreshold
	hooks.nextStageThreshold = func(stage int32) int32 {
		result := originalNext(stage)
		w.owner.update = &warpReadUseTestUpdate53F830{name: "new-update"}
		cachedUpdate.player = replacement
		return result
	}
	if got := warpReadUse53F830(hooks); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	wantSuffix := []string{
		"player:update",
		"index:replacement-player=fe",
		"inform:fe:15=0000000a",
		"frame:2=89abcdef",
		"store:data=89abcdef",
	}
	if got := w.events[len(w.events)-len(wantSuffix):]; !reflect.DeepEqual(got, wantSuffix) {
		t.Fatalf("suffix = %v, want %v", got, wantSuffix)
	}
}

func TestWarpReadUse53F830NonPlayerReturnsBeforeReadable(t *testing.T) {
	w := newWarpReadUseTestWorld53F830()
	w.owner.class = 0xf0
	w.readable = nil
	if got := warpReadUse53F830(w.hooks()); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	want := []string{"owner:owner", "class:owner=f0"}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %v, want %v", w.events, want)
	}
}

func TestWarpReadUse53F830UnsignedCooldown(t *testing.T) {
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
			w := newWarpReadUseTestWorld53F830()
			w.fps = tc.fps
			w.frame = tc.frame
			w.readable.data.state = tc.state
			if got := warpReadUse53F830(w.hooks()); got != 1 {
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

func TestWarpReadUse53F830RequiresExactOneMapAndAnyNonzeroWarp(t *testing.T) {
	for _, result := range []int32{0, -1, 2, math.MinInt32, math.MaxInt32} {
		t.Run(fmt.Sprintf("map-%08x", uint32(result)), func(t *testing.T) {
			w := newWarpReadUseTestWorld53F830()
			w.mapResult = result
			if got := warpReadUse53F830(w.hooks()); got != 1 {
				t.Fatalf("result = %d, want 1", got)
			}
			if w.messageData != nil || w.readable.data.state != 0 {
				t.Fatalf("message/state = %p/%#x", w.messageData, w.readable.data.state)
			}
		})
	}

	for _, enabled := range []int32{-1, 2, math.MinInt32, math.MaxInt32} {
		t.Run(fmt.Sprintf("warp-%08x", uint32(enabled)), func(t *testing.T) {
			w := newWarpReadUseTestWorld53F830()
			w.warpOpen = enabled
			if got := warpReadUse53F830(w.hooks()); got != 1 {
				t.Fatalf("result = %d, want 1", got)
			}
			if !containsWarpReadUseEvent53F830(w.events, "inform:") {
				t.Fatalf("nonzero warp did not take open path: %v", w.events)
			}
		})
	}
}

func containsWarpReadUseEvent53F830(events []string, prefix string) bool {
	for _, event := range events {
		if len(event) >= len(prefix) && event[:len(prefix)] == prefix {
			return true
		}
	}
	return false
}
