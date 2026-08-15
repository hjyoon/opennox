package server

import (
	"fmt"
	"reflect"
	"testing"
)

type soulGateTestObject4EBE40 struct {
	name     string
	classLow uint8
	data     *soulGateTestData4EBE40
	update   *soulGateTestUpdate4EBE40
	next     *soulGateTestObject4EBE40
}

type soulGateTestData4EBE40 struct {
	name string
	last uint32
}

type soulGateTestUpdate4EBE40 struct {
	name   string
	player *soulGateTestPlayer4EBE40
	gate   *soulGateTestObject4EBE40
}

type soulGateTestPlayer4EBE40 struct {
	name  string
	state uint32
}

type soulGateTestState4EBE40 struct {
	events       []string
	quest        uint32
	first        *soulGateTestObject4EBE40
	frames       []uint32
	fps          uint32
	onQuestMode  func()
	onAudio      func()
	onPointFX    func()
	onMessage    func()
	onFinalFrame func()
}

func soulGateObjectName4EBE40(obj *soulGateTestObject4EBE40) string {
	if obj == nil {
		return "nil"
	}
	return obj.name
}

func soulGateDataName4EBE40(data *soulGateTestData4EBE40) string {
	if data == nil {
		return "nil"
	}
	return data.name
}

func (s *soulGateTestState4EBE40) event(format string, args ...any) {
	s.events = append(s.events, fmt.Sprintf(format, args...))
}

func (s *soulGateTestState4EBE40) hooks() soulGateCollideHooks4EBE40[
	*soulGateTestObject4EBE40,
	*soulGateTestData4EBE40,
	*soulGateTestUpdate4EBE40,
	*soulGateTestPlayer4EBE40,
] {
	return soulGateCollideHooks4EBE40[
		*soulGateTestObject4EBE40,
		*soulGateTestData4EBE40,
		*soulGateTestUpdate4EBE40,
		*soulGateTestPlayer4EBE40,
	]{
		loadSourceCollideData: func(obj *soulGateTestObject4EBE40) *soulGateTestData4EBE40 {
			s.event("data:%s", soulGateObjectName4EBE40(obj))
			if obj == nil {
				panic("nil source")
			}
			return obj.data
		},
		gameFlagsCheck: func(flag uint32) uint32 {
			s.event("game:%#x=%d", flag, s.quest)
			return s.quest
		},
		loadTargetClassLow: func(obj *soulGateTestObject4EBE40) uint8 {
			s.event("class:%s=%#x", soulGateObjectName4EBE40(obj), obj.classLow)
			return obj.classLow
		},
		setQuestMode: func(value int32) {
			s.event("quest-mode:%d", value)
			if s.onQuestMode != nil {
				s.onQuestMode()
			}
		},
		firstPlayerUnit: func() *soulGateTestObject4EBE40 {
			s.event("first:%s", soulGateObjectName4EBE40(s.first))
			return s.first
		},
		nextPlayerUnit: func(obj *soulGateTestObject4EBE40) *soulGateTestObject4EBE40 {
			s.event("next:%s=%s", obj.name, soulGateObjectName4EBE40(obj.next))
			return obj.next
		},
		loadPlayerUpdate: func(obj *soulGateTestObject4EBE40) *soulGateTestUpdate4EBE40 {
			s.event("update:%s", soulGateObjectName4EBE40(obj))
			if obj == nil || obj.update == nil {
				panic("nil update")
			}
			return obj.update
		},
		loadPlayer: func(update *soulGateTestUpdate4EBE40) *soulGateTestPlayer4EBE40 {
			s.event("player:%s", update.name)
			if update.player == nil {
				panic("nil player")
			}
			return update.player
		},
		loadQuestState: func(player *soulGateTestPlayer4EBE40) uint32 {
			s.event("state:%s=%d", player.name, player.state)
			return player.state
		},
		loadSoulGate: func(update *soulGateTestUpdate4EBE40) *soulGateTestObject4EBE40 {
			s.event("gate:%s=%s", update.name, soulGateObjectName4EBE40(update.gate))
			return update.gate
		},
		loadFrame: func() uint32 {
			if len(s.frames) == 0 {
				panic("unexpected frame read")
			}
			frame := s.frames[0]
			s.frames = s.frames[1:]
			s.event("frame:%d", frame)
			if len(s.frames) == 0 && s.onFinalFrame != nil {
				s.onFinalFrame()
			}
			return frame
		},
		setQuestTimer: func(frame uint32) {
			s.event("timer:%d", frame)
		},
		loadLastUsedFrame: func(data *soulGateTestData4EBE40) uint32 {
			if data == nil {
				s.event("last:nil")
				panic("nil collide data")
			}
			s.event("last:%s=%d", data.name, data.last)
			return data.last
		},
		loadFPS: func() uint32 {
			s.event("fps:%d", s.fps)
			return s.fps
		},
		audio: func(id uint32, obj *soulGateTestObject4EBE40, first, second int32) {
			s.event("audio:%d:%s:%d:%d", id, soulGateObjectName4EBE40(obj), first, second)
			if s.onAudio != nil {
				s.onAudio()
			}
		},
		pointFX: func(id uint32, obj *soulGateTestObject4EBE40) uint32 {
			s.event("fx:%d:%s", id, soulGateObjectName4EBE40(obj))
			if s.onPointFX != nil {
				s.onPointFX()
			}
			return 0xf1234567
		},
		priorityMessage: func(obj *soulGateTestObject4EBE40, message string, value int32) {
			s.event("message:%s:%s:%d", soulGateObjectName4EBE40(obj), message, value)
			if s.onMessage != nil {
				s.onMessage()
			}
		},
		storeSoulGate: func(update *soulGateTestUpdate4EBE40, gate *soulGateTestObject4EBE40) {
			s.event("store-gate:%s=%s", update.name, soulGateObjectName4EBE40(gate))
			update.gate = gate
		},
		storeLastUsedFrame: func(data *soulGateTestData4EBE40, frame uint32) {
			if data == nil {
				s.event("store-last:nil=%d", frame)
				panic("nil collide data")
			}
			s.event("store-last:%s=%d", data.name, frame)
			data.last = frame
		},
	}
}

func assertSoulGateEvents4EBE40(t *testing.T, got, want []string) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("events:\n got %#v\nwant %#v", got, want)
	}
}

func TestSoulGateCollide4EBE40CachesSourceDataBeforeAllGuards(t *testing.T) {
	source := &soulGateTestObject4EBE40{name: "source", data: &soulGateTestData4EBE40{name: "entry"}}
	target := &soulGateTestObject4EBE40{name: "target", classLow: 0xfb}

	t.Run("non-Quest reads only source data and flags", func(t *testing.T) {
		state := &soulGateTestState4EBE40{}
		soulGateCollide4EBE40(source, target, struct{ unread uint32 }{0xdeadbeef}, state.hooks())
		assertSoulGateEvents4EBE40(t, state.events, []string{"data:source", "game:0x1000=0"})
	})

	t.Run("nil target stops before class", func(t *testing.T) {
		state := &soulGateTestState4EBE40{quest: 1}
		soulGateCollide4EBE40(source, (*soulGateTestObject4EBE40)(nil), 7, state.hooks())
		assertSoulGateEvents4EBE40(t, state.events, []string{"data:source", "game:0x1000=1"})
	})

	t.Run("non-Player stops before Quest mutation", func(t *testing.T) {
		state := &soulGateTestState4EBE40{quest: 1}
		soulGateCollide4EBE40(source, target, 9, state.hooks())
		assertSoulGateEvents4EBE40(t, state.events, []string{
			"data:source", "game:0x1000=1", "class:target=0xfb",
		})
	})

	t.Run("nil source faults before flags", func(t *testing.T) {
		state := &soulGateTestState4EBE40{}
		defer func() {
			if recover() == nil {
				t.Fatal("nil source did not fault")
			}
			assertSoulGateEvents4EBE40(t, state.events, []string{"data:nil"})
		}()
		soulGateCollide4EBE40((*soulGateTestObject4EBE40)(nil), target, 0, state.hooks())
	})
}

func TestSoulGateCollide4EBE40ScansEveryPlayerAndUsesCachedPointers(t *testing.T) {
	entryData := &soulGateTestData4EBE40{name: "entry", last: 17}
	liveData := &soulGateTestData4EBE40{name: "live", last: 99}
	source := &soulGateTestObject4EBE40{name: "source", data: entryData}
	targetUpdate := &soulGateTestUpdate4EBE40{name: "target-update"}
	replacementUpdate := &soulGateTestUpdate4EBE40{name: "replacement-update"}
	target := &soulGateTestObject4EBE40{
		name: "target", classLow: 0x84, update: targetUpdate,
	}
	unit1 := &soulGateTestObject4EBE40{name: "unit-1", update: &soulGateTestUpdate4EBE40{
		name: "update-1", player: &soulGateTestPlayer4EBE40{name: "player-1", state: 2}, gate: source,
	}}
	unit2 := &soulGateTestObject4EBE40{name: "unit-2", update: &soulGateTestUpdate4EBE40{
		name: "update-2", player: &soulGateTestPlayer4EBE40{name: "player-2", state: 1},
	}}
	unit3 := &soulGateTestObject4EBE40{name: "unit-3", update: &soulGateTestUpdate4EBE40{
		name: "update-3", player: &soulGateTestPlayer4EBE40{name: "player-3", state: 1}, gate: source,
	}}
	unit1.next = unit2
	unit2.next = unit3
	state := &soulGateTestState4EBE40{quest: 7, first: unit1, frames: []uint32{1234}}
	state.onQuestMode = func() { source.data = liveData }
	state.onAudio = func() { target.update = replacementUpdate }
	state.onFinalFrame = func() {
		if targetUpdate.gate != source {
			panic("final frame was read before target SoulGate store")
		}
	}

	soulGateCollide4EBE40(source, target, &struct{ unread uint32 }{0xcafebabe}, state.hooks())
	if entryData.last != 1234 || liveData.last != 99 {
		t.Fatalf("entry/live last = %d/%d, want 1234/99", entryData.last, liveData.last)
	}
	if targetUpdate.gate != source || replacementUpdate.gate != nil {
		t.Fatalf("cached/replacement target gates = %p/%p, want source/nil", targetUpdate.gate, replacementUpdate.gate)
	}
	assertSoulGateEvents4EBE40(t, state.events, []string{
		"data:source", "game:0x1000=7", "class:target=0x84", "quest-mode:0",
		"first:unit-1",
		"update:unit-1", "player:update-1", "state:player-1=2", "next:unit-1=unit-2",
		"update:unit-2", "player:update-2", "state:player-2=1", "gate:update-2=nil", "next:unit-2=unit-3",
		"update:unit-3", "player:update-3", "state:player-3=1", "gate:update-3=source", "next:unit-3=nil",
		"update:target", "gate:target-update=nil",
		"audio:1005:source:0:0", "fx:130:source",
		"message:target:objcoll.c:SoulGateCollide:0",
		"store-gate:target-update=source", "frame:1234", "store-last:entry=1234",
	})
}

func TestSoulGateCollide4EBE40NoReadyGateRefreshesTimerBeforeTargetUpdate(t *testing.T) {
	source := &soulGateTestObject4EBE40{name: "source", data: &soulGateTestData4EBE40{name: "data"}}
	target := &soulGateTestObject4EBE40{
		name: "target", classLow: 4,
		update: &soulGateTestUpdate4EBE40{name: "target-update", gate: source},
	}
	state := &soulGateTestState4EBE40{quest: 1, frames: []uint32{41, 50, 77}, fps: 9}
	soulGateCollide4EBE40(source, target, 0, state.hooks())
	assertSoulGateEvents4EBE40(t, state.events, []string{
		"data:source", "game:0x1000=1", "class:target=0x4", "quest-mode:0",
		"first:nil", "frame:41", "timer:41", "update:target", "gate:target-update=source",
		"frame:50", "last:data=0", "fps:9",
		"audio:1005:source:0:0", "fx:130:source", "message:target:objcoll.c:SoulGateCollide:0",
		"store-gate:target-update=source", "frame:77", "store-last:data=77",
	})
}

func TestSoulGateCollide4EBE40ThrottleIsStrictAndWrapping(t *testing.T) {
	for _, tc := range []struct {
		name       string
		frame      uint32
		last       uint32
		fps        uint32
		wantNotify bool
	}{
		{name: "equal boundary", frame: 110, last: 100, fps: 10},
		{name: "strictly greater", frame: 111, last: 100, fps: 10, wantNotify: true},
		{name: "wraparound", frame: 2, last: ^uint32(0) - 5, fps: 7, wantNotify: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			source := &soulGateTestObject4EBE40{name: "source", data: &soulGateTestData4EBE40{name: "data", last: tc.last}}
			target := &soulGateTestObject4EBE40{name: "target", classLow: 4, update: &soulGateTestUpdate4EBE40{name: "target-update", gate: source}}
			ready := &soulGateTestObject4EBE40{name: "ready", update: &soulGateTestUpdate4EBE40{
				name: "ready-update", player: &soulGateTestPlayer4EBE40{name: "ready-player", state: 1}, gate: source,
			}}
			state := &soulGateTestState4EBE40{quest: 1, first: ready, frames: []uint32{tc.frame, 999}, fps: tc.fps}
			soulGateCollide4EBE40(source, target, 0, state.hooks())
			var notifications int
			for _, event := range state.events {
				if event == "audio:1005:source:0:0" {
					notifications++
				}
			}
			if got := notifications == 1; got != tc.wantNotify {
				t.Fatalf("notification = %v, want %v; events=%v", got, tc.wantNotify, state.events)
			}
		})
	}
}

func TestSoulGateCollide4EBE40NilDataFaultTimingDependsOnGateIdentity(t *testing.T) {
	newObjects := func(sameGate bool) (*soulGateTestObject4EBE40, *soulGateTestObject4EBE40, *soulGateTestState4EBE40) {
		source := &soulGateTestObject4EBE40{name: "source"}
		gate := (*soulGateTestObject4EBE40)(nil)
		if sameGate {
			gate = source
		}
		target := &soulGateTestObject4EBE40{name: "target", classLow: 4, update: &soulGateTestUpdate4EBE40{name: "target-update", gate: gate}}
		ready := &soulGateTestObject4EBE40{name: "ready", update: &soulGateTestUpdate4EBE40{
			name: "ready-update", player: &soulGateTestPlayer4EBE40{name: "player", state: 1}, gate: source,
		}}
		return source, target, &soulGateTestState4EBE40{quest: 1, first: ready, frames: []uint32{10, 20}, fps: 30}
	}

	t.Run("same gate faults at throttle read before feedback and stores", func(t *testing.T) {
		source, target, state := newObjects(true)
		defer func() {
			if recover() == nil {
				t.Fatal("nil collide data did not fault")
			}
			for _, event := range state.events {
				if event == "fps:30" || event == "audio:1005:source:0:0" || event == "store-gate:target-update=source" {
					t.Fatalf("unexpected post-fault event %q in %v", event, state.events)
				}
			}
		}()
		soulGateCollide4EBE40(source, target, 0, state.hooks())
	})

	t.Run("different gate performs feedback and gate store before final fault", func(t *testing.T) {
		source, target, state := newObjects(false)
		defer func() {
			if recover() == nil {
				t.Fatal("nil collide data did not fault")
			}
			if target.update.gate != source {
				t.Fatal("target gate was not stored before final nil-data fault")
			}
			wantTail := []string{
				"audio:1005:source:0:0", "fx:130:source", "message:target:objcoll.c:SoulGateCollide:0",
				"store-gate:target-update=source", "frame:10", "store-last:nil=10",
			}
			if len(state.events) < len(wantTail) || !reflect.DeepEqual(state.events[len(state.events)-len(wantTail):], wantTail) {
				t.Fatalf("event tail = %#v, want %#v", state.events, wantTail)
			}
		}()
		soulGateCollide4EBE40(source, target, 0, state.hooks())
	})
}
