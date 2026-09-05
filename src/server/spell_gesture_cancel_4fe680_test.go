package server

import (
	"fmt"
	"math"
	"reflect"
	"testing"
)

const (
	spellGestureCancelTestEntity4FE680     = uint64(0x100000101)
	spellGestureCancelTestSource4FE680     = uint64(0x200000202)
	spellGestureCancelTestTarget4FE680     = uint64(0x300000303)
	spellGestureCancelTestUpdate4FE680     = uint64(0x400000404)
	spellGestureCancelTestPlayer4FE680     = uint64(0x500000505)
	spellGestureCancelTestSourceTeam4FE680 = uint64(0x600000606)
	spellGestureCancelTestTargetTeam4FE680 = uint64(0x700000707)
	spellGestureCancelTestAllocator4FE680  = uint64(0x800000808)
)

type spellGestureCancelTestEntityState4FE680 struct {
	object uint64
	next   uint64
	prev   uint64
}

type spellGestureCancelTestObjectState4FE680 struct {
	class  uint32
	team   uint64
	x, y   float32
	update uint64
}

type spellGestureCancelTestUpdateState4FE680 struct {
	castStart uint32
	casting   uint8
	player    uint64
}

type spellGestureCancelTestWorld4FE680 struct {
	events  []string
	counts  map[string]int
	after   map[string]func()
	faultAt int

	head, source, allocator uint64
	radius                  float32
	entities                map[uint64]*spellGestureCancelTestEntityState4FE680
	objects                 map[uint64]*spellGestureCancelTestObjectState4FE680
	updates                 map[uint64]*spellGestureCancelTestUpdateState4FE680
	playerIndex             map[uint64]uint8
	teamResult              int32
	mapResult               int32

	objectLoads []uint64
	mapTargets  []uint64
	informs     [][3]int32
	audio       []uint64
	states      []uint64
	freed       []uint64
}

func newSpellGestureCancelTestWorld4FE680() *spellGestureCancelTestWorld4FE680 {
	w := &spellGestureCancelTestWorld4FE680{
		counts:      make(map[string]int),
		after:       make(map[string]func()),
		head:        spellGestureCancelTestEntity4FE680,
		source:      spellGestureCancelTestSource4FE680,
		allocator:   spellGestureCancelTestAllocator4FE680,
		radius:      6,
		entities:    make(map[uint64]*spellGestureCancelTestEntityState4FE680),
		objects:     make(map[uint64]*spellGestureCancelTestObjectState4FE680),
		updates:     make(map[uint64]*spellGestureCancelTestUpdateState4FE680),
		playerIndex: make(map[uint64]uint8),
		mapResult:   1,
	}
	w.entities[w.head] = &spellGestureCancelTestEntityState4FE680{object: spellGestureCancelTestTarget4FE680}
	w.objects[w.source] = &spellGestureCancelTestObjectState4FE680{team: spellGestureCancelTestSourceTeam4FE680}
	w.objects[spellGestureCancelTestTarget4FE680] = &spellGestureCancelTestObjectState4FE680{
		class:  uint32(spellGestureCancelPlayerClass4FE680),
		team:   spellGestureCancelTestTargetTeam4FE680,
		x:      3,
		y:      4,
		update: spellGestureCancelTestUpdate4FE680,
	}
	w.updates[spellGestureCancelTestUpdate4FE680] = &spellGestureCancelTestUpdateState4FE680{
		castStart: 0x89abcdef,
		casting:   0xe1,
		player:    spellGestureCancelTestPlayer4FE680,
	}
	w.playerIndex[spellGestureCancelTestPlayer4FE680] = 0xd2
	return w
}

func (w *spellGestureCancelTestWorld4FE680) entity(id uint64) *spellGestureCancelTestEntityState4FE680 {
	if w.entities[id] == nil {
		w.entities[id] = &spellGestureCancelTestEntityState4FE680{}
	}
	return w.entities[id]
}

func (w *spellGestureCancelTestWorld4FE680) object(id uint64) *spellGestureCancelTestObjectState4FE680 {
	if w.objects[id] == nil {
		w.objects[id] = &spellGestureCancelTestObjectState4FE680{}
	}
	return w.objects[id]
}

func (w *spellGestureCancelTestWorld4FE680) update(id uint64) *spellGestureCancelTestUpdateState4FE680 {
	if w.updates[id] == nil {
		w.updates[id] = &spellGestureCancelTestUpdateState4FE680{}
	}
	return w.updates[id]
}

func (w *spellGestureCancelTestWorld4FE680) observe(base string) {
	w.counts[base]++
	event := fmt.Sprintf("%s#%d", base, w.counts[base])
	w.events = append(w.events, event)
	if w.faultAt != 0 && len(w.events) == w.faultAt {
		panic(event)
	}
	if after := w.after[event]; after != nil {
		after()
	}
}

func (w *spellGestureCancelTestWorld4FE680) hooks() spellGestureCancelHooks4FE680[
	uint64, uint64, uint64, uint64, uint64, uint64,
] {
	return spellGestureCancelHooks4FE680[
		uint64, uint64, uint64, uint64, uint64, uint64,
	]{
		loadHead: func() uint64 {
			value := w.head
			w.observe("head")
			return value
		},
		loadSourceArg: func() uint64 {
			value := w.source
			w.observe("source")
			return value
		},
		loadObject: func(entity uint64) uint64 {
			value := w.entity(entity).object
			w.objectLoads = append(w.objectLoads, value)
			w.observe(fmt.Sprintf("object:%x", entity))
			return value
		},
		loadClass: func(object uint64) uint32 {
			value := w.object(object).class
			w.observe(fmt.Sprintf("class:%x", object))
			return value
		},
		loadTeam: func(object uint64) uint64 {
			value := w.object(object).team
			w.observe(fmt.Sprintf("team:%x", object))
			return value
		},
		compareTeams: func(first, second uint64) int32 {
			value := w.teamResult
			w.observe(fmt.Sprintf("compare:%x:%x", first, second))
			return value
		},
		loadPosX: func(object uint64) float32 {
			value := w.object(object).x
			w.observe(fmt.Sprintf("x:%x", object))
			return value
		},
		loadPosY: func(object uint64) float32 {
			value := w.object(object).y
			w.observe(fmt.Sprintf("y:%x", object))
			return value
		},
		loadRadiusArg: func() float32 {
			value := w.radius
			w.observe("radius")
			return value
		},
		mapCheck: func(source, target uint64) int32 {
			value := w.mapResult
			w.mapTargets = append(w.mapTargets, target)
			w.observe(fmt.Sprintf("map:%x:%x", source, target))
			return value
		},
		loadUpdate: func(object uint64) uint64 {
			value := w.object(object).update
			w.observe(fmt.Sprintf("update:%x", object))
			return value
		},
		storeSpellCastStart: func(update uint64, value uint32) {
			w.update(update).castStart = value
			w.observe(fmt.Sprintf("store-cast:%x:%x", update, value))
		},
		storeCasting: func(update uint64, value uint8) {
			w.update(update).casting = value
			w.observe(fmt.Sprintf("store-casting:%x:%x", update, value))
		},
		loadPlayer: func(update uint64) uint64 {
			value := w.update(update).player
			w.observe(fmt.Sprintf("player:%x", update))
			return value
		},
		loadPlayerIndex: func(player uint64) uint8 {
			value := w.playerIndex[player]
			w.observe(fmt.Sprintf("index:%x", player))
			return value
		},
		informResult: func(index, code uint8, result int32) {
			w.informs = append(w.informs, [3]int32{int32(index), int32(code), result})
			w.observe(fmt.Sprintf("inform:%x:%x:%d", index, code, result))
		},
		audioEvent: func(id int32, object uint64, kind int32, code uint32) {
			w.audio = append(w.audio, object)
			w.observe(fmt.Sprintf("audio:%d:%x:%d:%x", id, object, kind, code))
		},
		setPlayerState: func(object uint64, state int32) {
			w.states = append(w.states, object)
			w.observe(fmt.Sprintf("state:%x:%d", object, state))
		},
		loadNext: func(entity uint64) uint64 {
			value := w.entity(entity).next
			w.observe(fmt.Sprintf("next:%x", entity))
			return value
		},
		loadPrev: func(entity uint64) uint64 {
			value := w.entity(entity).prev
			w.observe(fmt.Sprintf("prev:%x", entity))
			return value
		},
		storePrev: func(entity, prev uint64) {
			w.entity(entity).prev = prev
			w.observe(fmt.Sprintf("store-prev:%x:%x", entity, prev))
		},
		storeNext: func(entity, next uint64) {
			w.entity(entity).next = next
			w.observe(fmt.Sprintf("store-next:%x:%x", entity, next))
		},
		storeHead: func(entity uint64) {
			w.head = entity
			w.observe(fmt.Sprintf("store-head:%x", entity))
		},
		loadAllocator: func() uint64 {
			value := w.allocator
			w.observe("allocator")
			return value
		},
		free: func(allocator, entity uint64) {
			w.freed = append(w.freed, entity)
			w.observe(fmt.Sprintf("free:%x:%x", allocator, entity))
		},
	}
}

func TestSpellGestureCancel4FE680ExactTrace(t *testing.T) {
	w := newSpellGestureCancelTestWorld4FE680()
	spellGestureCancel4FE680(w.hooks())

	want := []string{
		"head#1", "source#1",
		"object:100000101#1", "class:300000303#1",
		"team:300000303#1", "team:200000202#1", "compare:600000606:700000707#1",
		"object:100000101#2", "x:300000303#1", "x:200000202#1", "y:300000303#1", "y:200000202#1", "radius#1",
		"map:200000202:300000303#1",
		"object:100000101#3", "class:300000303#2", "update:300000303#1",
		"store-cast:400000404:0#1", "store-casting:400000404:0#1",
		"player:400000404#1", "index:500000505#1", "inform:d2:0:15#1",
		"object:100000101#4", "audio:231:300000303:0:0#1",
		"object:100000101#5", "state:300000303:13#1",
		"next:100000101#1", "prev:100000101#1", "next:100000101#2", "store-head:0#1",
		"allocator#1", "next:100000101#3", "free:800000808:100000101#1",
	}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("trace mismatch\n got: %q\nwant: %q", w.events, want)
	}
	update := w.update(spellGestureCancelTestUpdate4FE680)
	if update.castStart != 0 || update.casting != 0 {
		t.Fatalf("player update = (%#x, %#x), want zero", update.castStart, update.casting)
	}
	if !reflect.DeepEqual(w.informs, [][3]int32{{0xd2, 0, spellGestureCancelResult4FE680}}) ||
		!reflect.DeepEqual(w.audio, []uint64{spellGestureCancelTestTarget4FE680}) ||
		!reflect.DeepEqual(w.states, []uint64{spellGestureCancelTestTarget4FE680}) ||
		!reflect.DeepEqual(w.freed, []uint64{spellGestureCancelTestEntity4FE680}) || w.head != 0 {
		t.Fatalf("effects = informs %v audio %x states %x freed %x head %#x", w.informs, w.audio, w.states, w.freed, w.head)
	}
}

func TestSpellGestureCancel4FE680EarlyGates(t *testing.T) {
	t.Run("empty head does not observe arguments", func(t *testing.T) {
		w := newSpellGestureCancelTestWorld4FE680()
		w.head = 0
		spellGestureCancel4FE680(w.hooks())
		if !reflect.DeepEqual(w.events, []string{"head#1"}) {
			t.Fatalf("events = %q", w.events)
		}
	})

	t.Run("only canonical team equality skips", func(t *testing.T) {
		for _, result := range []int32{1, 2, -1} {
			w := newSpellGestureCancelTestWorld4FE680()
			w.teamResult = result
			spellGestureCancel4FE680(w.hooks())
			if result == 1 {
				if len(w.freed) != 0 || w.counts["radius"] != 0 || w.counts["next:100000101"] != 1 {
					t.Fatalf("canonical result effects = freed %x radius %d next %d", w.freed, w.counts["radius"], w.counts["next:100000101"])
				}
			} else if !reflect.DeepEqual(w.freed, []uint64{spellGestureCancelTestEntity4FE680}) {
				t.Fatalf("noncanonical result %d freed = %x", result, w.freed)
			}
		}
	})

	t.Run("distance rejects before map", func(t *testing.T) {
		w := newSpellGestureCancelTestWorld4FE680()
		w.radius = 5.1
		spellGestureCancel4FE680(w.hooks())
		if w.counts["map:200000202:300000303"] != 0 || len(w.freed) != 0 || w.counts["next:100000101"] != 1 {
			t.Fatalf("map/freed/next = %d/%x/%d", w.counts["map:200000202:300000303"], w.freed, w.counts["next:100000101"])
		}
	})

	t.Run("map zero rejects after cached target", func(t *testing.T) {
		w := newSpellGestureCancelTestWorld4FE680()
		w.mapResult = 0
		spellGestureCancel4FE680(w.hooks())
		if !reflect.DeepEqual(w.mapTargets, []uint64{spellGestureCancelTestTarget4FE680}) || len(w.freed) != 0 || len(w.informs) != 0 {
			t.Fatalf("map targets/freed/informs = %x/%x/%v", w.mapTargets, w.freed, w.informs)
		}
	})

	t.Run("non-player bypasses team and player effects", func(t *testing.T) {
		w := newSpellGestureCancelTestWorld4FE680()
		w.object(spellGestureCancelTestTarget4FE680).class = 0xabcdef00
		spellGestureCancel4FE680(w.hooks())
		if w.counts["team:300000303"] != 0 || w.counts["compare:600000606:700000707"] != 0 || len(w.informs) != 0 || len(w.audio) != 0 || len(w.states) != 0 || len(w.freed) != 1 {
			t.Fatalf("team/compare/inform/audio/state/freed = %d/%d/%v/%x/%x/%x", w.counts["team:300000303"], w.counts["compare:600000606:700000707"], w.informs, w.audio, w.states, w.freed)
		}
	})
}

func TestSpellGestureCancelWithinRadius4FE680(t *testing.T) {
	nan := math.Float32frombits(0x7fa12345)
	epsilon := spellGestureCancelEpsilon4FE680
	for _, tc := range []struct {
		name                                       string
		targetX, sourceX, targetY, sourceY, radius float32
		want                                       bool
	}{
		{name: "strictly below", radius: math.Nextafter32(epsilon, float32(math.Inf(1))), want: true},
		{name: "equal rejects", radius: epsilon},
		{name: "above rejects", targetX: 3, targetY: 4, radius: 5.1},
		{name: "unordered coordinate passes", targetX: nan, radius: 0, want: true},
		{name: "unordered radius passes", radius: nan, want: true},
		{name: "infinity subtraction passes unordered", targetX: float32(math.Inf(1)), sourceX: float32(math.Inf(1)), radius: 0, want: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := spellGestureCancelWithinRadius4FE680(tc.targetX, tc.sourceX, tc.targetY, tc.sourceY, tc.radius); got != tc.want {
				t.Fatalf("result = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestSpellGestureCancel4FE680ReloadsTargetsAroundCallbacks(t *testing.T) {
	w := newSpellGestureCancelTestWorld4FE680()
	const (
		distanceTarget = uint64(0x900000909)
		effectTarget   = uint64(0xa00000a0a)
		audioTarget    = uint64(0xb00000b0b)
		stateTarget    = uint64(0xc00000c0c)
	)
	w.objects[distanceTarget] = &spellGestureCancelTestObjectState4FE680{x: 3, y: 4}
	w.objects[effectTarget] = &spellGestureCancelTestObjectState4FE680{
		class: uint32(spellGestureCancelPlayerClass4FE680), update: spellGestureCancelTestUpdate4FE680,
	}
	w.objects[audioTarget] = &spellGestureCancelTestObjectState4FE680{}
	w.objects[stateTarget] = &spellGestureCancelTestObjectState4FE680{}
	w.after["compare:600000606:700000707#1"] = func() {
		w.entity(w.head).object = distanceTarget
	}
	w.after["map:200000202:900000909#1"] = func() {
		w.entity(w.head).object = effectTarget
	}
	w.after["inform:d2:0:15#1"] = func() {
		w.entity(w.head).object = audioTarget
	}
	w.after["audio:231:b00000b0b:0:0#1"] = func() {
		w.entity(w.head).object = stateTarget
	}

	spellGestureCancel4FE680(w.hooks())
	if !reflect.DeepEqual(w.objectLoads, []uint64{
		spellGestureCancelTestTarget4FE680,
		distanceTarget,
		effectTarget,
		audioTarget,
		stateTarget,
	}) {
		t.Fatalf("object reloads = %x", w.objectLoads)
	}
	if !reflect.DeepEqual(w.mapTargets, []uint64{distanceTarget}) ||
		!reflect.DeepEqual(w.audio, []uint64{audioTarget}) ||
		!reflect.DeepEqual(w.states, []uint64{stateTarget}) {
		t.Fatalf("map/audio/state = %x/%x/%x", w.mapTargets, w.audio, w.states)
	}
}

func TestSpellGestureCancel4FE680ReloadsIntrusiveLinks(t *testing.T) {
	w := newSpellGestureCancelTestWorld4FE680()
	w.object(spellGestureCancelTestTarget4FE680).class = 0
	const (
		next1      = uint64(0x900000909)
		prev1      = uint64(0xa00000a0a)
		prev2      = uint64(0xb00000b0b)
		prev3      = uint64(0xc00000c0c)
		next2      = uint64(0xd00000d0d)
		next3      = uint64(0xe00000e0e)
		next4      = uint64(0xf00000f0f)
		next5      = uint64(0x110000111)
		tailObject = uint64(0x120000121)
	)
	w.entity(w.head).next = next1
	w.entity(w.head).prev = prev1
	w.entities[next1] = &spellGestureCancelTestEntityState4FE680{}
	w.entities[prev1] = &spellGestureCancelTestEntityState4FE680{}
	w.entities[prev2] = &spellGestureCancelTestEntityState4FE680{}
	w.entities[prev3] = &spellGestureCancelTestEntityState4FE680{}
	w.entities[next2] = &spellGestureCancelTestEntityState4FE680{}
	w.entities[next3] = &spellGestureCancelTestEntityState4FE680{}
	w.entities[next4] = &spellGestureCancelTestEntityState4FE680{object: tailObject}
	w.entities[next5] = &spellGestureCancelTestEntityState4FE680{}
	w.objects[tailObject] = &spellGestureCancelTestObjectState4FE680{
		class: uint32(spellGestureCancelPlayerClass4FE680),
		team:  spellGestureCancelTestSourceTeam4FE680,
	}
	w.after["next:100000101#1"] = func() { w.entity(w.head).prev = prev2 }
	w.after["prev:100000101#1"] = func() { w.entity(w.head).prev = prev3 }
	w.after["prev:100000101#2"] = func() { w.entity(w.head).next = next2 }
	w.after["next:100000101#2"] = func() { w.entity(w.head).next = next3 }
	w.after["allocator#1"] = func() { w.entity(w.head).next = next4 }
	w.after["free:800000808:100000101#1"] = func() {
		w.entity(w.head).next = next5
		w.teamResult = 1
	}

	spellGestureCancel4FE680(w.hooks())
	if w.entity(next1).prev != prev2 {
		t.Fatalf("first next prev = %#x, want %#x", w.entity(next1).prev, prev2)
	}
	if w.entity(prev3).next != next2 {
		t.Fatalf("reloaded prev next = %#x, want %#x", w.entity(prev3).next, next2)
	}
	if got := w.objectLoads[len(w.objectLoads)-1]; got != tailObject {
		t.Fatalf("post-free iteration object = %#x, want cached-next object %#x", got, tailObject)
	}
	if len(w.freed) != 1 || w.freed[0] != spellGestureCancelTestEntity4FE680 || w.counts["object:f00000f0f"] != 1 || w.counts["object:110000111"] != 0 {
		t.Fatalf("freed/tail/corrupt-tail = %x/%d/%d", w.freed, w.counts["object:f00000f0f"], w.counts["object:110000111"])
	}
}

func TestSpellGestureCancel4FE680FaultPrefixes(t *testing.T) {
	baseline := newSpellGestureCancelTestWorld4FE680()
	spellGestureCancel4FE680(baseline.hooks())
	for faultAt := 1; faultAt <= len(baseline.events); faultAt++ {
		w := newSpellGestureCancelTestWorld4FE680()
		w.faultAt = faultAt
		func() {
			defer func() {
				if recover() == nil {
					t.Fatalf("fault %d did not panic", faultAt)
				}
			}()
			spellGestureCancel4FE680(w.hooks())
		}()
		if !reflect.DeepEqual(w.events, baseline.events[:faultAt]) {
			t.Fatalf("fault %d prefix\n got: %q\nwant: %q", faultAt, w.events, baseline.events[:faultAt])
		}
	}
}
