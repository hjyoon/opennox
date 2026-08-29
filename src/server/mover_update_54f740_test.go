package server

import (
	"fmt"
	"math"
	"reflect"
	"testing"

	"github.com/opennox/libs/types"
)

type moverUpdateTestObject54F740 struct {
	name      string
	flags     uint32
	pos       types.Pointf
	velocity  types.Pointf
	speedCur  float32
	speedBase float32
	data      *moverUpdateTestData54F740
}

type moverUpdateTestWaypoint54F740 struct {
	name   string
	pos    types.Pointf
	points []*moverUpdateTestWaypoint54F740
}

type moverUpdateTestData54F740 struct {
	state       uint8
	speedUnits  int32
	extent      uint32
	target      *moverUpdateTestObject54F740
	waypointIDs map[int]uint32
	waypoints   map[int]*moverUpdateTestWaypoint54F740
}

type moverUpdateTestFixture54F740 struct {
	events         []string
	objects        map[uint32]*moverUpdateTestObject54F740
	waypoints      map[uint32]*moverUpdateTestWaypoint54F740
	random         []int
	afterMove      func(int, *moverUpdateTestObject54F740, types.Pointf)
	moveCount      int
	removed        []*moverUpdateTestObject54F740
	speedCurReads  []float32
	waypointCounts []uint8
}

func moverUpdateTestObjectName54F740(obj *moverUpdateTestObject54F740) string {
	if obj == nil {
		return "nil"
	}
	return obj.name
}

func moverUpdateTestWaypointName54F740(wp *moverUpdateTestWaypoint54F740) string {
	if wp == nil {
		return "nil"
	}
	return wp.name
}

func (f *moverUpdateTestFixture54F740) event(format string, args ...any) {
	f.events = append(f.events, fmt.Sprintf(format, args...))
}

func (f *moverUpdateTestFixture54F740) hooks() moverUpdateHooks54F740[
	*moverUpdateTestObject54F740,
	*moverUpdateTestData54F740,
	*moverUpdateTestWaypoint54F740,
] {
	return moverUpdateHooks54F740[
		*moverUpdateTestObject54F740,
		*moverUpdateTestData54F740,
		*moverUpdateTestWaypoint54F740,
	]{
		loadUpdateData: func(obj *moverUpdateTestObject54F740) *moverUpdateTestData54F740 {
			f.event("update:%s", moverUpdateTestObjectName54F740(obj))
			return obj.data
		},
		loadTargetExtent: func(data *moverUpdateTestData54F740) uint32 {
			f.event("extent:%d", data.extent)
			return data.extent
		},
		loadTarget: func(data *moverUpdateTestData54F740) *moverUpdateTestObject54F740 {
			f.event("target:%s", moverUpdateTestObjectName54F740(data.target))
			return data.target
		},
		storeTarget: func(data *moverUpdateTestData54F740, target *moverUpdateTestObject54F740) {
			f.event("store-target:%s", moverUpdateTestObjectName54F740(target))
			data.target = target
		},
		objectByExtent: func(extent uint32) *moverUpdateTestObject54F740 {
			f.event("object-by-extent:%d", extent)
			return f.objects[extent]
		},
		loadWaypointID: func(data *moverUpdateTestData54F740, slot int) uint32 {
			id := data.waypointIDs[slot]
			f.event("waypoint-id:%d:%d", slot, id)
			return id
		},
		loadWaypoint: func(data *moverUpdateTestData54F740, slot int) *moverUpdateTestWaypoint54F740 {
			wp := data.waypoints[slot]
			f.event("waypoint:%d:%s", slot, moverUpdateTestWaypointName54F740(wp))
			return wp
		},
		storeWaypoint: func(data *moverUpdateTestData54F740, slot int, wp *moverUpdateTestWaypoint54F740) {
			f.event("store-waypoint:%d:%s", slot, moverUpdateTestWaypointName54F740(wp))
			data.waypoints[slot] = wp
		},
		waypointByID: func(id uint32) *moverUpdateTestWaypoint54F740 {
			f.event("waypoint-by-id:%d", id)
			return f.waypoints[id]
		},
		loadFlags: func(obj *moverUpdateTestObject54F740) uint32 {
			f.event("flags:%s:%08x", moverUpdateTestObjectName54F740(obj), obj.flags)
			return obj.flags
		},
		loadState: func(data *moverUpdateTestData54F740) uint8 {
			f.event("state:%d", data.state)
			return data.state
		},
		storeState: func(data *moverUpdateTestData54F740, state uint8) {
			f.event("store-state:%d", state)
			data.state = state
		},
		loadSpeedUnits: func(data *moverUpdateTestData54F740) int32 {
			f.event("speed-units:%d", data.speedUnits)
			return data.speedUnits
		},
		storeSpeedBase: func(obj *moverUpdateTestObject54F740, speed float32) {
			f.event("store-speed-base:%g", speed)
			obj.speedBase = speed
		},
		storeSpeedCur: func(obj *moverUpdateTestObject54F740, speed float32) {
			f.event("store-speed-cur:%g", speed)
			obj.speedCur = speed
		},
		loadSpeedCur: func(obj *moverUpdateTestObject54F740) float32 {
			value := obj.speedCur
			if len(f.speedCurReads) != 0 {
				value = f.speedCurReads[0]
				f.speedCurReads = f.speedCurReads[1:]
			}
			f.event("speed-cur:%s:%g", obj.name, value)
			return value
		},
		loadPosition: func(obj *moverUpdateTestObject54F740) types.Pointf {
			f.event("position:%s", moverUpdateTestObjectName54F740(obj))
			return obj.pos
		},
		loadPosX: func(obj *moverUpdateTestObject54F740) float32 {
			f.event("pos-x:%s", obj.name)
			return obj.pos.X
		},
		loadPosY: func(obj *moverUpdateTestObject54F740) float32 {
			f.event("pos-y:%s", obj.name)
			return obj.pos.Y
		},
		loadVelocityX: func(obj *moverUpdateTestObject54F740) float32 {
			f.event("velocity-x:%s", obj.name)
			return obj.velocity.X
		},
		loadVelocityY: func(obj *moverUpdateTestObject54F740) float32 {
			f.event("velocity-y:%s", obj.name)
			return obj.velocity.Y
		},
		storeVelocityX: func(obj *moverUpdateTestObject54F740, value float32) {
			f.event("store-velocity-x:%g", value)
			obj.velocity.X = value
		},
		storeVelocityY: func(obj *moverUpdateTestObject54F740, value float32) {
			f.event("store-velocity-y:%g", value)
			obj.velocity.Y = value
		},
		loadWaypointPos: func(wp *moverUpdateTestWaypoint54F740) types.Pointf {
			f.event("waypoint-position:%s", moverUpdateTestWaypointName54F740(wp))
			return wp.pos
		},
		loadWaypointX: func(wp *moverUpdateTestWaypoint54F740) float32 {
			f.event("waypoint-x:%s", moverUpdateTestWaypointName54F740(wp))
			return wp.pos.X
		},
		loadWaypointY: func(wp *moverUpdateTestWaypoint54F740) float32 {
			f.event("waypoint-y:%s", moverUpdateTestWaypointName54F740(wp))
			return wp.pos.Y
		},
		waypointPointCnt: func(wp *moverUpdateTestWaypoint54F740) uint8 {
			count := uint8(len(wp.points))
			if len(f.waypointCounts) != 0 {
				count = f.waypointCounts[0]
				f.waypointCounts = f.waypointCounts[1:]
			}
			f.event("point-count:%s:%d", moverUpdateTestWaypointName54F740(wp), count)
			return count
		},
		waypointPoint: func(wp *moverUpdateTestWaypoint54F740, index int) *moverUpdateTestWaypoint54F740 {
			point := wp.points[index]
			f.event("point:%s:%d:%s", wp.name, index, moverUpdateTestWaypointName54F740(point))
			return point
		},
		randomInt: func(minimum, maximum int) int {
			value := f.random[0]
			f.random = f.random[1:]
			f.event("random:%d:%d:%d", minimum, maximum, value)
			return value
		},
		move: func(obj *moverUpdateTestObject54F740, position types.Pointf) {
			f.moveCount++
			f.event("move:%s:%g:%g", moverUpdateTestObjectName54F740(obj), position.X, position.Y)
			if f.afterMove != nil {
				f.afterMove(f.moveCount, obj, position)
			}
		},
		removeUpdatable: func(obj *moverUpdateTestObject54F740) {
			f.event("remove:%s", obj.name)
			f.removed = append(f.removed, obj)
		},
	}
}

func newMoverUpdateFixture54F740() (
	*moverUpdateTestFixture54F740,
	*moverUpdateTestObject54F740,
	*moverUpdateTestObject54F740,
) {
	target := &moverUpdateTestObject54F740{
		name: "target", flags: moverUpdateTargetRequiredFlag54F740, pos: types.Pointf{X: 30, Y: 40},
	}
	data := &moverUpdateTestData54F740{
		extent:      77,
		target:      target,
		waypointIDs: make(map[int]uint32),
		waypoints:   make(map[int]*moverUpdateTestWaypoint54F740),
	}
	source := &moverUpdateTestObject54F740{name: "source", data: data}
	fixture := &moverUpdateTestFixture54F740{
		objects:   make(map[uint32]*moverUpdateTestObject54F740),
		waypoints: make(map[uint32]*moverUpdateTestWaypoint54F740),
	}
	return fixture, source, target
}

func assertMoverUpdateEvents54F740(t *testing.T, got, want []string) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("events = %#v, want %#v", got, want)
	}
}

func TestMoverUpdate54F740TargetExtentAndResolutionFailures(t *testing.T) {
	t.Run("missing extent", func(t *testing.T) {
		f, source, _ := newMoverUpdateFixture54F740()
		source.data.extent = 0
		moverUpdate54F740(source, f.hooks())
		assertMoverUpdateEvents54F740(t, f.events, []string{
			"update:source", "extent:0", "remove:source",
		})
	})

	t.Run("unresolved object", func(t *testing.T) {
		f, source, _ := newMoverUpdateFixture54F740()
		source.data.target = nil
		moverUpdate54F740(source, f.hooks())
		assertMoverUpdateEvents54F740(t, f.events, []string{
			"update:source", "extent:77", "target:nil", "object-by-extent:77",
			"store-target:nil", "remove:source",
		})
	})
}

func TestMoverUpdate54F740ResolvesTransientWaypointsBeforeInvalidTarget(t *testing.T) {
	f, source, target := newMoverUpdateFixture54F740()
	target.flags = moverUpdateTargetBlockedFlag54F740
	wp3 := &moverUpdateTestWaypoint54F740{name: "wp3"}
	wp5 := &moverUpdateTestWaypoint54F740{name: "wp5"}
	source.data.waypointIDs[3] = 13
	source.data.waypointIDs[5] = 15
	f.waypoints[13] = wp3
	f.waypoints[15] = wp5

	moverUpdate54F740(source, f.hooks())
	if source.data.target != nil || source.data.waypoints[3] != wp3 || source.data.waypoints[5] != wp5 {
		t.Fatalf("resolved data = %+v", source.data)
	}
	assertMoverUpdateEvents54F740(t, f.events, []string{
		"update:source", "extent:77", "target:target",
		"waypoint-id:3:13", "waypoint:3:nil", "waypoint-by-id:13", "store-waypoint:3:wp3",
		"waypoint-id:5:15", "waypoint:5:nil", "waypoint-by-id:15", "store-waypoint:5:wp5",
		"flags:target:00000020", "store-target:nil", "remove:source",
	})
}

func TestMoverUpdate54F740StateZeroSuccessUsesSignedQuarterUnitsAndLiveTarget(t *testing.T) {
	f, source, target := newMoverUpdateFixture54F740()
	source.flags = moverUpdateActiveFlag54F740
	source.data.speedUnits = -7
	source.data.waypointIDs[2] = 12
	start := &moverUpdateTestWaypoint54F740{name: "start"}
	f.waypoints[12] = start
	liveTarget := &moverUpdateTestObject54F740{name: "live-target", pos: types.Pointf{X: 8, Y: 9}}
	f.afterMove = func(_ int, _ *moverUpdateTestObject54F740, _ types.Pointf) {}
	loadTargetCalls := 0
	hooks := f.hooks()
	originalLoadTarget := hooks.loadTarget
	hooks.loadTarget = func(data *moverUpdateTestData54F740) *moverUpdateTestObject54F740 {
		loadTargetCalls++
		if loadTargetCalls == 2 {
			data.target = liveTarget
		}
		return originalLoadTarget(data)
	}

	moverUpdate54F740(source, hooks)
	if source.data.state != 1 || source.data.waypoints[3] != start || source.data.waypoints[5] != nil {
		t.Fatalf("state/waypoints = %d/%p/%p", source.data.state, source.data.waypoints[3], source.data.waypoints[5])
	}
	if source.speedBase != -1.75 || source.speedCur != -1.75 {
		t.Fatalf("speeds = %v/%v, want -1.75", source.speedBase, source.speedCur)
	}
	if target == liveTarget {
		t.Fatal("fixture target identity unexpectedly changed")
	}
	assertMoverUpdateEvents54F740(t, f.events, []string{
		"update:source", "extent:77", "target:target", "waypoint-id:3:0", "waypoint-id:5:0",
		"flags:target:00000004", "state:0", "flags:source:01000000", "waypoint-id:2:12",
		"waypoint-by-id:12", "target:live-target", "position:live-target", "move:source:8:9",
		"speed-units:-7", "store-state:1", "store-waypoint:3:start", "store-waypoint:5:nil",
		"store-speed-base:-1.75", "store-speed-cur:-1.75",
	})
}

func TestMoverUpdate54F740StateZeroMissingWaypointDefersRemoval(t *testing.T) {
	f, source, _ := newMoverUpdateFixture54F740()
	source.flags = moverUpdateActiveFlag54F740
	source.data.waypointIDs[2] = 99
	moverUpdate54F740(source, f.hooks())
	if source.data.state != 3 || len(f.removed) != 0 {
		t.Fatalf("state/removals = %d/%d, want 3/0", source.data.state, len(f.removed))
	}
}

func TestMoverUpdate54F740InactiveStateTransitionsAndTerminalStates(t *testing.T) {
	tests := []struct {
		name       string
		state      uint8
		active     bool
		wantState  uint8
		wantMove   int
		wantRemove int
	}{
		{name: "state one sleeps", state: 1, wantState: 2},
		{name: "state two stays asleep", state: 2, wantState: 2},
		{name: "state two wakes", state: 2, active: true, wantState: 1, wantMove: 1},
		{name: "state three removes", state: 3, wantState: 3, wantRemove: 1},
		{name: "unknown returns", state: 9, wantState: 9},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			f, source, _ := newMoverUpdateFixture54F740()
			source.data.state = tc.state
			if tc.active {
				source.flags = moverUpdateActiveFlag54F740
			}
			moverUpdate54F740(source, f.hooks())
			if source.data.state != tc.wantState || f.moveCount != tc.wantMove || len(f.removed) != tc.wantRemove {
				t.Fatalf("state/moves/removes = %d/%d/%d, want %d/%d/%d", source.data.state, f.moveCount, len(f.removed), tc.wantState, tc.wantMove, tc.wantRemove)
			}
		})
	}
}

func TestMoverUpdate54F740StateOneDeadEndMovesCachedTargetThenStops(t *testing.T) {
	f, source, target := newMoverUpdateFixture54F740()
	source.data.state = 1
	source.flags = moverUpdateActiveFlag54F740
	source.velocity = types.Pointf{X: 1, Y: 1}
	source.pos = types.Pointf{X: 5, Y: 5}
	current := &moverUpdateTestWaypoint54F740{name: "dead-end", pos: types.Pointf{X: 4, Y: 4}}
	source.data.waypoints[3] = current

	moverUpdate54F740(source, f.hooks())
	if source.data.state != 3 || f.moveCount != 1 {
		t.Fatalf("state/moves = %d/%d, want 3/1", source.data.state, f.moveCount)
	}
	if f.events[len(f.events)-1] != "state:3" {
		t.Fatalf("tail = %q, want state recheck", f.events[len(f.events)-1])
	}
	wantMove := "move:" + target.name + ":4:4"
	found := false
	for _, event := range f.events {
		if event == wantMove {
			found = true
		}
		if event == "target:target" && found {
			t.Fatalf("target was reloaded after dead-end move: %q", f.events)
		}
	}
	if !found {
		t.Fatalf("events = %q, missing %q", f.events, wantMove)
	}
}

func TestMoverUpdate54F740StateOneSinglePointFollowsAndSteers(t *testing.T) {
	f, source, target := newMoverUpdateFixture54F740()
	source.data.state = 1
	source.flags = moverUpdateActiveFlag54F740
	source.pos = types.Pointf{X: 10, Y: 20}
	source.velocity = types.Pointf{X: 1, Y: 1}
	source.speedCur = 5
	next := &moverUpdateTestWaypoint54F740{name: "next", pos: types.Pointf{X: 13, Y: 24}}
	current := &moverUpdateTestWaypoint54F740{
		name: "current", pos: types.Pointf{X: 9, Y: 19}, points: []*moverUpdateTestWaypoint54F740{next},
	}
	source.data.waypoints[3] = current

	moverUpdate54F740(source, f.hooks())
	if source.data.waypoints[5] != current || source.data.waypoints[3] != next {
		t.Fatalf("waypoints = %p/%p, want current/next", source.data.waypoints[5], source.data.waypoints[3])
	}
	denominator := float32(math.Sqrt(4*4+3*3) + float64(float32(0.1)))
	wantX := float32(float64(3*source.speedCur) / float64(denominator))
	wantY := float32(float64(4*source.speedCur) / float64(denominator))
	if source.velocity.X != wantX || source.velocity.Y != wantY {
		t.Fatalf("velocity = %v/%v, want %v/%v", source.velocity.X, source.velocity.Y, wantX, wantY)
	}
	if f.moveCount != 1 || target == source {
		t.Fatalf("moves/identity = %d/%v", f.moveCount, target == source)
	}
	wantOrder := []string{
		"store-waypoint:5:current", "point:current:0:next", "store-waypoint:3:next",
		"state:1", "target:target", "position:source", "move:target:10:20", "waypoint:3:next",
		"waypoint-x:next", "pos-x:source", "waypoint-y:next", "pos-y:source",
		"speed-cur:source:5", "store-velocity-x:" + fmt.Sprint(wantX),
		"speed-cur:source:5", "store-velocity-y:" + fmt.Sprint(wantY),
	}
	start := -1
	for i, event := range f.events {
		if event == wantOrder[0] {
			start = i
			break
		}
	}
	if start < 0 || !reflect.DeepEqual(f.events[start:], wantOrder) {
		t.Fatalf("steering tail = %#v, want %#v", f.events[start:], wantOrder)
	}
}

func TestMoverUpdate54F740RandomBranchReloadsLiveSlotsAndCandidate(t *testing.T) {
	f, source, _ := newMoverUpdateFixture54F740()
	source.data.state = 1
	source.flags = moverUpdateActiveFlag54F740
	source.velocity = types.Pointf{X: 1, Y: 1}
	source.speedCur = 2
	previous := &moverUpdateTestWaypoint54F740{name: "previous"}
	candidate := &moverUpdateTestWaypoint54F740{name: "candidate", pos: types.Pointf{X: 3, Y: 4}}
	current := &moverUpdateTestWaypoint54F740{
		name: "current", pos: types.Pointf{X: -1, Y: -1},
		points: []*moverUpdateTestWaypoint54F740{previous, candidate},
	}
	source.data.waypoints[3] = current
	source.data.waypoints[5] = previous
	f.random = []int{0, 1}

	moverUpdate54F740(source, f.hooks())
	wantSequence := []string{
		"point-count:current:2",
		"point-count:current:2", "random:0:1:0", "waypoint:3:current", "waypoint:5:previous", "point:current:0:previous",
		"point-count:current:2", "random:0:1:1", "waypoint:3:current", "waypoint:5:previous", "point:current:1:candidate",
		"store-waypoint:5:current", "point:current:1:candidate", "store-waypoint:3:candidate",
	}
	start := -1
	for i, event := range f.events {
		if event == wantSequence[0] {
			start = i
			break
		}
	}
	if start < 0 || !reflect.DeepEqual(f.events[start:start+len(wantSequence)], wantSequence) {
		t.Fatalf("random sequence = %#v, want %#v", f.events, wantSequence)
	}
	if source.data.waypoints[5] != current || source.data.waypoints[3] != candidate {
		t.Fatalf("waypoints = %p/%p, want current/candidate", source.data.waypoints[5], source.data.waypoints[3])
	}
}

func TestMoverUpdate54F740NaNVelocitySkipsArrivalButStillSteers(t *testing.T) {
	f, source, _ := newMoverUpdateFixture54F740()
	source.data.state = 1
	source.flags = moverUpdateActiveFlag54F740
	source.velocity = types.Pointf{X: float32(math.NaN()), Y: 1}
	source.speedCur = 4
	current := &moverUpdateTestWaypoint54F740{name: "current", pos: types.Pointf{X: 3, Y: 4}}
	source.data.waypoints[3] = current

	moverUpdate54F740(source, f.hooks())
	for _, event := range f.events {
		if len(event) >= len("point-count:") && event[:len("point-count:")] == "point-count:" {
			t.Fatalf("NaN velocity entered arrival branch: %q", f.events)
		}
	}
	if math.IsNaN(float64(source.velocity.X)) || source.velocity.X == 0 || source.velocity.Y == 0 {
		t.Fatalf("velocity = %v/%v, want finite steering", source.velocity.X, source.velocity.Y)
	}
}
