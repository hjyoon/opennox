package server

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/opennox/libs/types"
)

type rewardAnkhTestData4F2110 struct {
	categoryLow uint8
}

type rewardAnkhTestObject4F2110 struct {
	name     string
	typeInd  uint16
	data     *rewardAnkhTestData4F2110
	next     *rewardAnkhTestObject4F2110
	position types.Pointf
}

type rewardAnkhTestState4F2110 struct {
	markerCache uint32
	plusCache   uint32
	lookups     map[string]uint32
	firsts      []*rewardAnkhTestObject4F2110
	firstCalls  int
	randomValue int32
	newObjects  []*rewardAnkhTestObject4F2110
	events      []string
	deleted     []*rewardAnkhTestObject4F2110
	onRandom    func(int32, int32, string, int32)
	onNext      func(*rewardAnkhTestObject4F2110)
}

func rewardAnkhTestHooks4F2110(state *rewardAnkhTestState4F2110) rewardAnkhReplaceHooks4F2110[
	*rewardAnkhTestObject4F2110,
	*rewardAnkhTestData4F2110,
] {
	return rewardAnkhReplaceHooks4F2110[*rewardAnkhTestObject4F2110, *rewardAnkhTestData4F2110]{
		loadMarkerCache: func() uint32 {
			state.events = append(state.events, "cache:marker")
			return state.markerCache
		},
		storeMarkerCache: func(value uint32) {
			state.events = append(state.events, fmt.Sprintf("store:marker:%d", value))
			state.markerCache = value
		},
		loadPlusCache: func() uint32 {
			state.events = append(state.events, "cache:plus")
			return state.plusCache
		},
		storePlusCache: func(value uint32) {
			state.events = append(state.events, fmt.Sprintf("store:plus:%d", value))
			state.plusCache = value
		},
		lookupType: func(name string) uint32 {
			state.events = append(state.events, "lookup:"+name)
			return state.lookups[name]
		},
		firstObject: func() *rewardAnkhTestObject4F2110 {
			state.firstCalls++
			state.events = append(state.events, fmt.Sprintf("first:%d", state.firstCalls))
			if state.firstCalls > len(state.firsts) {
				return nil
			}
			return state.firsts[state.firstCalls-1]
		},
		nextObject: func(object *rewardAnkhTestObject4F2110) *rewardAnkhTestObject4F2110 {
			state.events = append(state.events, "next:"+object.name)
			if state.onNext != nil {
				state.onNext(object)
			}
			return object.next
		},
		loadTypeInd: func(object *rewardAnkhTestObject4F2110) uint16 {
			state.events = append(state.events, "type:"+object.name)
			return object.typeInd
		},
		loadInitData: func(object *rewardAnkhTestObject4F2110) *rewardAnkhTestData4F2110 {
			state.events = append(state.events, "data:"+object.name)
			return object.data
		},
		loadCategoryLow: func(data *rewardAnkhTestData4F2110) uint8 {
			state.events = append(state.events, "category")
			return data.categoryLow
		},
		randomInt: func(minimum, maximum int32, path string, line int32) int32 {
			state.events = append(state.events, fmt.Sprintf("random:%d:%d", minimum, maximum))
			if state.onRandom != nil {
				state.onRandom(minimum, maximum, path, line)
			}
			return state.randomValue
		},
		newObject: func(name string) *rewardAnkhTestObject4F2110 {
			state.events = append(state.events, "new:"+name)
			if len(state.newObjects) == 0 {
				return nil
			}
			object := state.newObjects[0]
			state.newObjects = state.newObjects[1:]
			return object
		},
		loadPosX: func(object *rewardAnkhTestObject4F2110) float32 {
			state.events = append(state.events, "x:"+object.name)
			return object.position.X
		},
		loadPosY: func(object *rewardAnkhTestObject4F2110) float32 {
			state.events = append(state.events, "y:"+object.name)
			return object.position.Y
		},
		createAt: func(object, owner *rewardAnkhTestObject4F2110, point types.Pointf) {
			ownerName := "nil"
			if owner != nil {
				ownerName = owner.name
			}
			state.events = append(state.events, fmt.Sprintf(
				"create:%s:%s:%g:%g", object.name, ownerName, point.X, point.Y,
			))
			object.position = point
		},
		delayedDelete: func(object *rewardAnkhTestObject4F2110) {
			state.events = append(state.events, "delete:"+object.name)
			state.deleted = append(state.deleted, object)
		},
	}
}

func TestRewardAnkhInitializesCachesAndDrawsEmptyRange4F2110(t *testing.T) {
	state := &rewardAnkhTestState4F2110{
		lookups: map[string]uint32{
			rewardAnkhMarkerTypeName4F2110:     17,
			rewardAnkhMarkerPlusTypeName4F2110: 19,
		},
		firsts:      []*rewardAnkhTestObject4F2110{nil, nil},
		randomValue: -1,
	}
	state.onRandom = func(minimum, maximum int32, path string, line int32) {
		if minimum != 0 || maximum != -1 || path != rewardAnkhRandomPath4F2110 || line != 2460 {
			t.Fatalf("random args = %d/%d/%q/%d", minimum, maximum, path, line)
		}
	}
	rewardAnkhReplace4F2110(rewardAnkhTestHooks4F2110(state))
	want := []string{
		"cache:marker", "lookup:RewardMarker", "store:marker:17",
		"lookup:RewardMarkerPlus", "store:plus:19", "first:1",
		"random:0:-1", "first:2",
	}
	if state.markerCache != 17 || state.plusCache != 19 || !reflect.DeepEqual(state.events, want) {
		t.Fatalf("caches/events = %d/%d/%v, want 17/19/%v", state.markerCache, state.plusCache, state.events, want)
	}

	state.events = nil
	state.firstCalls = 0
	state.firsts = []*rewardAnkhTestObject4F2110{nil, nil}
	state.markerCache = 17
	state.plusCache = 0
	state.lookups = nil
	rewardAnkhReplace4F2110(rewardAnkhTestHooks4F2110(state))
	want = []string{"cache:marker", "first:1", "random:0:-1", "first:2"}
	if state.plusCache != 0 || !reflect.DeepEqual(state.events, want) {
		t.Fatalf("cached plus/events = %d/%v, want 0/%v", state.plusCache, state.events, want)
	}
}

func TestRewardAnkhSelectsActiveOrdinalAndPreservesOrder4F2110(t *testing.T) {
	ordinary := &rewardAnkhTestObject4F2110{name: "ordinary", typeInd: 1}
	inactive := &rewardAnkhTestObject4F2110{
		name: "inactive", typeInd: 10, data: &rewardAnkhTestData4F2110{},
	}
	plus := &rewardAnkhTestObject4F2110{
		name: "plus", typeInd: 20, data: &rewardAnkhTestData4F2110{categoryLow: 0x80},
	}
	marker := &rewardAnkhTestObject4F2110{
		name: "marker", typeInd: 10, data: &rewardAnkhTestData4F2110{categoryLow: 0x81},
		position: types.Pointf{X: 7, Y: 9},
	}
	ordinary.next = inactive
	inactive.next = plus
	plus.next = marker
	ankh := &rewardAnkhTestObject4F2110{name: "ankh"}
	state := &rewardAnkhTestState4F2110{
		markerCache: 10,
		plusCache:   20,
		firsts:      []*rewardAnkhTestObject4F2110{ordinary, ordinary},
		randomValue: 1,
		newObjects:  []*rewardAnkhTestObject4F2110{ankh},
	}
	state.onRandom = func(minimum, maximum int32, _ string, _ int32) {
		if minimum != 0 || maximum != 1 {
			t.Fatalf("random range = %d..%d, want 0..1", minimum, maximum)
		}
	}
	rewardAnkhReplace4F2110(rewardAnkhTestHooks4F2110(state))
	wantSuffix := []string{
		"first:2",
		"cache:marker", "type:ordinary", "cache:plus", "next:ordinary",
		"cache:marker", "type:inactive", "data:inactive", "category", "next:inactive",
		"cache:marker", "type:plus", "cache:plus", "data:plus", "category", "next:plus",
		"cache:marker", "type:marker", "data:marker", "category", "new:Ankh",
		"y:marker", "x:marker", "create:ankh:nil:7:9", "delete:marker",
	}
	if got := state.events[len(state.events)-len(wantSuffix):]; !reflect.DeepEqual(got, wantSuffix) {
		t.Fatalf("selection suffix =\n%v\nwant\n%v", got, wantSuffix)
	}
	if !reflect.DeepEqual(state.deleted, []*rewardAnkhTestObject4F2110{marker}) ||
		ankh.position != marker.position {
		t.Fatalf("deleted/position = %v/%+v, want marker/%+v", state.deleted, ankh.position, marker.position)
	}
	for _, event := range state.events[len(state.events)-len(wantSuffix):] {
		if event == "next:marker" {
			t.Fatal("successful selected marker requested a successor")
		}
	}
}

func TestRewardAnkhRetriesNilFactoryAtSameOrdinal4F2110(t *testing.T) {
	a := &rewardAnkhTestObject4F2110{
		name: "a", typeInd: 10, data: &rewardAnkhTestData4F2110{categoryLow: 0x80},
	}
	b := &rewardAnkhTestObject4F2110{
		name: "b", typeInd: 10, data: &rewardAnkhTestData4F2110{categoryLow: 0x80},
	}
	c := &rewardAnkhTestObject4F2110{
		name: "c", typeInd: 10, data: &rewardAnkhTestData4F2110{categoryLow: 0x80},
		position: types.Pointf{X: 31, Y: 32},
	}
	a.next = b
	b.next = c
	ankh := &rewardAnkhTestObject4F2110{name: "ankh"}
	state := &rewardAnkhTestState4F2110{
		markerCache: 10,
		plusCache:   20,
		firsts:      []*rewardAnkhTestObject4F2110{a, a},
		randomValue: 1,
		newObjects:  []*rewardAnkhTestObject4F2110{nil, ankh},
	}
	rewardAnkhReplace4F2110(rewardAnkhTestHooks4F2110(state))
	var factories int
	for _, event := range state.events {
		if event == "new:Ankh" {
			factories++
		}
	}
	if factories != 2 || !reflect.DeepEqual(state.deleted, []*rewardAnkhTestObject4F2110{c}) ||
		ankh.position != c.position {
		t.Fatalf("factory/deleted/position = %d/%v/%+v, want 2/c/%+v", factories, state.deleted, ankh.position, c.position)
	}
	wantOrder := []string{"new:Ankh", "next:b", "cache:marker", "type:c", "data:c", "category", "new:Ankh"}
	for start := 0; start+len(wantOrder) <= len(state.events); start++ {
		if reflect.DeepEqual(state.events[start:start+len(wantOrder)], wantOrder) {
			return
		}
	}
	t.Fatalf("nil retry order %v not found in %v", wantOrder, state.events)
}

func TestRewardAnkhStartsFreshTraversalAndReloadsCaches4F2110(t *testing.T) {
	firstPass := &rewardAnkhTestObject4F2110{
		name: "first-pass", typeInd: 10, data: &rewardAnkhTestData4F2110{categoryLow: 0x80},
	}
	secondPass := &rewardAnkhTestObject4F2110{
		name: "second-pass", typeInd: 20, data: &rewardAnkhTestData4F2110{categoryLow: 0x80},
		position: types.Pointf{X: 41, Y: 42},
	}
	ankh := &rewardAnkhTestObject4F2110{name: "ankh"}
	state := &rewardAnkhTestState4F2110{
		markerCache: 10,
		plusCache:   11,
		firsts:      []*rewardAnkhTestObject4F2110{firstPass, secondPass},
		newObjects:  []*rewardAnkhTestObject4F2110{ankh},
	}
	state.onRandom = func(_, _ int32, _ string, _ int32) {
		state.markerCache = 99
		state.plusCache = 20
	}
	rewardAnkhReplace4F2110(rewardAnkhTestHooks4F2110(state))
	if !reflect.DeepEqual(state.deleted, []*rewardAnkhTestObject4F2110{secondPass}) || ankh.position != secondPass.position {
		t.Fatalf("deleted/position = %v/%+v, want second-pass/%+v", state.deleted, ankh.position, secondPass.position)
	}
	wantSuffix := []string{
		"random:0:0", "first:2", "cache:marker", "type:second-pass", "cache:plus",
		"data:second-pass", "category", "new:Ankh", "y:second-pass", "x:second-pass",
		"create:ankh:nil:41:42", "delete:second-pass",
	}
	if got := state.events[len(state.events)-len(wantSuffix):]; !reflect.DeepEqual(got, wantSuffix) {
		t.Fatalf("fresh traversal suffix = %v, want %v", got, wantSuffix)
	}
}

func TestRewardAnkhComparesZeroExtendedTypeWithFullCaches4F2110(t *testing.T) {
	object := &rewardAnkhTestObject4F2110{
		name: "object", typeInd: 10, data: &rewardAnkhTestData4F2110{categoryLow: 0x80},
	}
	state := &rewardAnkhTestState4F2110{
		markerCache: 0x0001000a,
		plusCache:   0x0002000a,
		firsts:      []*rewardAnkhTestObject4F2110{object, object},
		randomValue: -1,
	}
	rewardAnkhReplace4F2110(rewardAnkhTestHooks4F2110(state))
	for _, event := range state.events {
		if event == "data:object" || event == "new:Ankh" {
			t.Fatalf("full-width cache unexpectedly matched uint16 type: %v", state.events)
		}
	}
	want := []string{
		"cache:marker", "first:1", "cache:marker", "type:object", "cache:plus", "next:object",
		"random:0:-1", "first:2", "cache:marker", "type:object", "cache:plus", "next:object",
	}
	if !reflect.DeepEqual(state.events, want) {
		t.Fatalf("events = %v, want %v", state.events, want)
	}
}

func TestRewardAnkhNilInitDataFaultsBeforeSuccessorAndRNG4F2110(t *testing.T) {
	marker := &rewardAnkhTestObject4F2110{name: "marker", typeInd: 10}
	state := &rewardAnkhTestState4F2110{
		markerCache: 10,
		plusCache:   20,
		firsts:      []*rewardAnkhTestObject4F2110{marker, marker},
	}
	defer func() {
		if recover() == nil {
			t.Fatal("nil RewardMarker InitData did not fault at first-byte load")
		}
		want := []string{
			"cache:marker", "first:1", "cache:marker", "type:marker", "data:marker", "category",
		}
		if !reflect.DeepEqual(state.events, want) {
			t.Fatalf("fault prefix = %v, want %v", state.events, want)
		}
		if state.firstCalls != 1 || len(state.deleted) != 0 {
			t.Fatalf("fault continued into successor/RNG/second pass: first=%d deleted=%v", state.firstCalls, state.deleted)
		}
	}()
	rewardAnkhReplace4F2110(rewardAnkhTestHooks4F2110(state))
}
