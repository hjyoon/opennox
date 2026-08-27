package server

import (
	"fmt"
	"math"
	"reflect"
	"testing"
)

type rewardReplenishTestData4F2210 struct {
	name string
	low  uint8
}

type rewardReplenishTestObject4F2210 struct {
	name    string
	typeInd uint16
	data    *rewardReplenishTestData4F2210
	next    *rewardReplenishTestObject4F2210
}

type rewardReplenishTestArray4F2210 struct {
	name    string
	objects []*rewardReplenishTestObject4F2210
}

type rewardReplenishTestState4F2210 struct {
	questStage int32
	players    int32

	markerCache uint32
	plusCache   uint32
	potionCache uint32
	lookups     map[string]uint32

	firsts     []*rewardReplenishTestObject4F2210
	firstCalls int
	allocCalls int
	random     []int32
	events     []string
	freed      []*rewardReplenishTestArray4F2210
	deleted    []*rewardReplenishTestObject4F2210

	onFirst func(int)
	onNext  func(*rewardReplenishTestObject4F2210)
	onAlloc func(int32, int) *rewardReplenishTestArray4F2210
}

func rewardReplenishTestHooks4F2210(state *rewardReplenishTestState4F2210) rewardReplenishHooks4F2210[
	*rewardReplenishTestObject4F2210,
	*rewardReplenishTestData4F2210,
	*rewardReplenishTestArray4F2210,
] {
	return rewardReplenishHooks4F2210[
		*rewardReplenishTestObject4F2210,
		*rewardReplenishTestData4F2210,
		*rewardReplenishTestArray4F2210,
	]{
		loadQuestStage: func() uint32 {
			state.events = append(state.events, "quest")
			return uint32(state.questStage)
		},
		loadPlayerCount: func() int32 {
			state.events = append(state.events, "players")
			return state.players
		},
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
		loadPotionCache: func() uint32 {
			state.events = append(state.events, "cache:potion")
			return state.potionCache
		},
		storePotionCache: func(value uint32) {
			state.events = append(state.events, fmt.Sprintf("store:potion:%d", value))
			state.potionCache = value
		},
		lookupType: func(name string) uint32 {
			state.events = append(state.events, "lookup:"+name)
			return state.lookups[name]
		},
		firstObject: func() *rewardReplenishTestObject4F2210 {
			state.firstCalls++
			state.events = append(state.events, fmt.Sprintf("first:%d", state.firstCalls))
			if state.onFirst != nil {
				state.onFirst(state.firstCalls)
			}
			if state.firstCalls > len(state.firsts) {
				return nil
			}
			return state.firsts[state.firstCalls-1]
		},
		nextObject: func(object *rewardReplenishTestObject4F2210) *rewardReplenishTestObject4F2210 {
			state.events = append(state.events, "next:"+object.name)
			if state.onNext != nil {
				state.onNext(object)
			}
			return object.next
		},
		loadTypeInd: func(object *rewardReplenishTestObject4F2210) uint16 {
			state.events = append(state.events, "type:"+object.name)
			return object.typeInd
		},
		loadInitData: func(object *rewardReplenishTestObject4F2210) *rewardReplenishTestData4F2210 {
			state.events = append(state.events, "data:"+object.name)
			return object.data
		},
		loadFieldLow: func(data *rewardReplenishTestData4F2210) uint8 {
			name := "nil"
			if data != nil {
				name = data.name
			}
			state.events = append(state.events, "low:"+name)
			return data.low
		},
		storeFieldLow: func(data *rewardReplenishTestData4F2210, value uint8) {
			name := "nil"
			if data != nil {
				name = data.name
			}
			state.events = append(state.events, fmt.Sprintf("set:%s:%02x", name, value))
			data.low = value
		},
		allocObjects: func(count int32) *rewardReplenishTestArray4F2210 {
			state.allocCalls++
			state.events = append(state.events, fmt.Sprintf("alloc:%d:%d", state.allocCalls, count))
			if state.onAlloc != nil {
				return state.onAlloc(count, state.allocCalls)
			}
			return &rewardReplenishTestArray4F2210{
				name:    fmt.Sprintf("a%d", state.allocCalls),
				objects: make([]*rewardReplenishTestObject4F2210, int(count)),
			}
		},
		storeObject: func(array *rewardReplenishTestArray4F2210, index int32, object *rewardReplenishTestObject4F2210) {
			name := "nil"
			if array != nil {
				name = array.name
			}
			state.events = append(state.events, fmt.Sprintf("array-set:%s:%d:%s", name, index, object.name))
			array.objects[index] = object
		},
		loadObject: func(array *rewardReplenishTestArray4F2210, index int32) *rewardReplenishTestObject4F2210 {
			name := "nil"
			if array != nil {
				name = array.name
			}
			state.events = append(state.events, fmt.Sprintf("array-get:%s:%d", name, index))
			return array.objects[index]
		},
		freeObjects: func(array *rewardReplenishTestArray4F2210) {
			name := "nil"
			if array != nil {
				name = array.name
			}
			state.events = append(state.events, "free:"+name)
			state.freed = append(state.freed, array)
		},
		randomInt: func(minimum, maximum int32, path string, line int32) int32 {
			state.events = append(state.events, fmt.Sprintf(
				"random:%d:%d:%d:%s", minimum, maximum, line, path,
			))
			if len(state.random) == 0 {
				return minimum
			}
			value := state.random[0]
			state.random = state.random[1:]
			return value
		},
		delayedDelete: func(object *rewardReplenishTestObject4F2210) {
			state.events = append(state.events, "delete:"+object.name)
			state.deleted = append(state.deleted, object)
		},
	}
}

func rewardReplenishLink4F2210(objects ...*rewardReplenishTestObject4F2210) *rewardReplenishTestObject4F2210 {
	if len(objects) == 0 {
		return nil
	}
	for index := range objects[:len(objects)-1] {
		objects[index].next = objects[index+1]
	}
	objects[len(objects)-1].next = nil
	return objects[0]
}

func rewardReplenishContainsEvents4F2210(events, want []string) bool {
	for start := 0; start+len(want) <= len(events); start++ {
		if reflect.DeepEqual(events[start:start+len(want)], want) {
			return true
		}
	}
	return false
}

func TestRewardReplenishFractionAndRounding4F2210(t *testing.T) {
	tests := []struct {
		stage   uint32
		players int32
		bits    uint32
	}{
		{0, 0, 0x00000000},
		{0, 1, 0x3ecccccd},
		{0, 2, 0x3ecccccd},
		{0, 3, 0x3f333333},
		{0, 4, 0x3f333333},
		{0, 5, 0x3f800000},
		{0, 6, 0x3f800000},
		{0, 7, 0x00000000},
		{1, -17, 0x3f000000},
		{1, 99, 0x3f000000},
		{2, 2, 0x3ecccccd},
	}
	for _, test := range tests {
		got := rewardReplenishFraction4F2210(test.stage, test.players)
		if bits := math.Float32bits(got); bits != test.bits {
			t.Errorf("fraction(%d,%d) bits = %#08x, want %#08x", test.stage, test.players, bits, test.bits)
		}
	}
	if got := rewardReplenishRounded4F2210(1, 0.4); got != 0 {
		t.Fatalf("round 1*0.4 = %d, want 0", got)
	}
	if got := rewardReplenishRounded4F2210(2, 0.4); got != 1 {
		t.Fatalf("round 2*0.4 = %d, want 1", got)
	}
	if got := rewardReplenishRounded4F2210(5, 0.7); got != 3 {
		t.Fatalf("round 5*0.7 = %d, want 3", got)
	}
}

func TestRewardReplenishInitializesAllCachesOnlyFromZeroPrimary4F2210(t *testing.T) {
	state := &rewardReplenishTestState4F2210{
		questStage: 7,
		players:    4,
		lookups: map[string]uint32{
			rewardReplenishMarkerTypeName4F2210:     10,
			rewardReplenishMarkerPlusTypeName4F2210: 20,
			rewardReplenishPotionTypeName4F2210:     30,
		},
		firsts: []*rewardReplenishTestObject4F2210{nil},
	}
	rewardReplenish4F2210(rewardReplenishTestHooks4F2210(state))
	want := []string{
		"quest", "players", "cache:marker",
		"lookup:RewardMarker", "store:marker:10",
		"lookup:RewardMarkerPlus", "store:plus:20",
		"lookup:RedPotion", "store:potion:30", "first:1",
	}
	if !reflect.DeepEqual(state.events, want) || state.markerCache != 10 ||
		state.plusCache != 20 || state.potionCache != 30 {
		t.Fatalf("events/caches = %v / %d/%d/%d, want %v / 10/20/30",
			state.events, state.markerCache, state.plusCache, state.potionCache, want)
	}

	state.events = nil
	state.firstCalls = 0
	state.firsts = []*rewardReplenishTestObject4F2210{nil}
	state.markerCache = 10
	state.plusCache = 0
	state.potionCache = 0
	state.lookups = nil
	rewardReplenish4F2210(rewardReplenishTestHooks4F2210(state))
	want = []string{"quest", "players", "cache:marker", "first:1"}
	if !reflect.DeepEqual(state.events, want) || state.plusCache != 0 || state.potionCache != 0 {
		t.Fatalf("nonzero primary events/caches = %v / %d/%d, want %v / 0/0",
			state.events, state.plusCache, state.potionCache, want)
	}
}

func TestRewardReplenishTwoPassActivationShuffleAndDelete4F2210(t *testing.T) {
	inactiveData := &rewardReplenishTestData4F2210{name: "inactive", low: 0x20}
	fixedData := &rewardReplenishTestData4F2210{name: "fixed", low: 0x41}
	plusData := &rewardReplenishTestData4F2210{name: "plus", low: 0x02}
	inactive := &rewardReplenishTestObject4F2210{name: "inactive", typeInd: 10, data: inactiveData}
	fixed := &rewardReplenishTestObject4F2210{name: "fixed", typeInd: 10, data: fixedData}
	plus := &rewardReplenishTestObject4F2210{name: "plus", typeInd: 20, data: plusData}
	p1 := &rewardReplenishTestObject4F2210{name: "p1", typeInd: 30}
	p2 := &rewardReplenishTestObject4F2210{name: "p2", typeInd: 30}
	p3 := &rewardReplenishTestObject4F2210{name: "p3", typeInd: 30}
	first := rewardReplenishLink4F2210(inactive, fixed, plus, p1, p2, p3)
	state := &rewardReplenishTestState4F2210{
		questStage:  1,
		players:     99,
		markerCache: 10,
		plusCache:   20,
		potionCache: 30,
		firsts:      []*rewardReplenishTestObject4F2210{first, first},
		random:      []int32{0, 0},
	}
	rewardReplenish4F2210(rewardReplenishTestHooks4F2210(state))

	if inactiveData.low != 0xa0 || fixedData.low != 0xc1 || plusData.low != 0x82 {
		t.Fatalf("marker lows = %#02x/%#02x/%#02x, want 0xa0/0xc1/0x82",
			inactiveData.low, fixedData.low, plusData.low)
	}
	if !reflect.DeepEqual(state.deleted, []*rewardReplenishTestObject4F2210{p1}) {
		t.Fatalf("deleted = %v, want p1", state.deleted)
	}
	if len(state.freed) != 2 || state.freed[0].name != "a1" || state.freed[1].name != "a2" {
		t.Fatalf("freed = %v, want a1/a2", state.freed)
	}
	wantRandoms := []string{
		"random:0:2:2660:" + rewardReplenishRandomPath4F2210,
		"random:0:1:2660:" + rewardReplenishRandomPath4F2210,
	}
	var gotRandoms []string
	for _, event := range state.events {
		if len(event) >= len("random:") && event[:len("random:")] == "random:" {
			gotRandoms = append(gotRandoms, event)
		}
	}
	if !reflect.DeepEqual(gotRandoms, wantRandoms) {
		t.Fatalf("randoms = %v, want %v", gotRandoms, wantRandoms)
	}
	// First pass ignores RewardMarkerPlus and compares RedPotion only after a
	// primary mismatch. The second pass compares Plus before Potion.
	wantFirstPlus := []string{"cache:marker", "type:plus", "cache:potion", "next:plus"}
	if !rewardReplenishContainsEvents4F2210(state.events, wantFirstPlus) {
		t.Fatalf("first-pass plus comparison order %v not found in %v", wantFirstPlus, state.events)
	}
	wantSecondPlus := []string{
		"cache:marker", "type:plus", "cache:plus", "data:plus", "low:plus", "set:plus:82", "next:plus",
	}
	if !rewardReplenishContainsEvents4F2210(state.events, wantSecondPlus) {
		t.Fatalf("second-pass plus order %v not found in %v", wantSecondPlus, state.events)
	}
	// Retained potion slots are never dereferenced by the deletion loop.
	wantDelete := []string{"array-get:a2:2", "delete:p1", "free:a2"}
	if !rewardReplenishContainsEvents4F2210(state.events, wantDelete) {
		t.Fatalf("potion delete suffix %v not found in %v", wantDelete, state.events)
	}
}

func TestRewardReplenishReloadsCachesAndStartsFreshTraversal4F2210(t *testing.T) {
	firstMarker := &rewardReplenishTestObject4F2210{
		name: "first-marker", typeInd: 10,
		data: &rewardReplenishTestData4F2210{name: "first-marker"},
	}
	firstPotion := &rewardReplenishTestObject4F2210{name: "first-potion", typeInd: 30}
	rewardReplenishLink4F2210(firstMarker, firstPotion)
	secondPotion := &rewardReplenishTestObject4F2210{name: "second-potion", typeInd: 40}
	state := &rewardReplenishTestState4F2210{
		players:     1,
		markerCache: 10,
		plusCache:   20,
		potionCache: 30,
		firsts:      []*rewardReplenishTestObject4F2210{firstMarker, secondPotion},
	}
	state.onFirst = func(call int) {
		if call == 2 {
			state.markerCache = 99
			state.plusCache = 98
			state.potionCache = 40
		}
	}
	rewardReplenish4F2210(rewardReplenishTestHooks4F2210(state))
	want := []string{
		"first:2", "cache:marker", "type:second-potion", "cache:plus", "cache:potion",
		"array-set:a2:0:second-potion",
	}
	if !rewardReplenishContainsEvents4F2210(state.events, want) {
		t.Fatalf("fresh/live-cache sequence %v not found in %v", want, state.events)
	}
	if !reflect.DeepEqual(state.deleted, []*rewardReplenishTestObject4F2210{secondPotion}) ||
		len(state.freed) != 1 || state.freed[0].name != "a2" {
		t.Fatalf("fresh traversal deleted/freed = %v/%v, want second-potion/a2", state.deleted, state.freed)
	}
}

func TestRewardReplenishSecondPassGrowthFaultsBeforeSuccessorOrFree4F2210(t *testing.T) {
	firstMarker := &rewardReplenishTestObject4F2210{
		name: "first", typeInd: 10,
		data: &rewardReplenishTestData4F2210{name: "first"},
	}
	secondA := &rewardReplenishTestObject4F2210{
		name: "second-a", typeInd: 10,
		data: &rewardReplenishTestData4F2210{name: "second-a"},
	}
	secondB := &rewardReplenishTestObject4F2210{
		name: "second-b", typeInd: 10,
		data: &rewardReplenishTestData4F2210{name: "second-b"},
	}
	rewardReplenishLink4F2210(secondA, secondB)
	state := &rewardReplenishTestState4F2210{
		players:     6,
		markerCache: 10,
		plusCache:   20,
		potionCache: 30,
		firsts:      []*rewardReplenishTestObject4F2210{firstMarker, secondA},
	}
	defer func() {
		if recover() == nil {
			t.Fatal("second-pass growth did not fault at exact allocation bound")
		}
		want := []string{
			"cache:marker", "type:second-b", "data:second-b", "low:second-b",
			"array-set:a1:1:second-b",
		}
		if !rewardReplenishContainsEvents4F2210(state.events, want) {
			t.Fatalf("growth fault prefix %v not found in %v", want, state.events)
		}
		if len(state.freed) != 0 {
			t.Fatalf("growth fault freed arrays: %v", state.freed)
		}
		for _, event := range state.events {
			if event == "next:second-b" {
				t.Fatalf("growth fault requested successor: %v", state.events)
			}
		}
	}()
	rewardReplenish4F2210(rewardReplenishTestHooks4F2210(state))
}

func TestRewardReplenishPreservesOriginalConditionalLeaks4F2210(t *testing.T) {
	marker := &rewardReplenishTestObject4F2210{
		name: "marker", typeInd: 10,
		data: &rewardReplenishTestData4F2210{name: "marker"},
	}
	potion := &rewardReplenishTestObject4F2210{name: "potion", typeInd: 30}
	rewardReplenishLink4F2210(marker, potion)
	state := &rewardReplenishTestState4F2210{
		players:     6,
		markerCache: 10,
		plusCache:   20,
		potionCache: 30,
		firsts:      []*rewardReplenishTestObject4F2210{marker, nil},
	}
	rewardReplenish4F2210(rewardReplenishTestHooks4F2210(state))
	if state.allocCalls != 2 || len(state.freed) != 0 {
		t.Fatalf("empty second pass alloc/free = %d/%v, want 2/no frees", state.allocCalls, state.freed)
	}

	ordinary := &rewardReplenishTestObject4F2210{name: "ordinary", typeInd: 77}
	marker.next = nil
	state = &rewardReplenishTestState4F2210{
		players:     6,
		markerCache: 10,
		plusCache:   20,
		potionCache: 30,
		firsts:      []*rewardReplenishTestObject4F2210{marker, ordinary},
	}
	rewardReplenish4F2210(rewardReplenishTestHooks4F2210(state))
	if state.allocCalls != 1 || len(state.freed) != 0 {
		t.Fatalf("shrunk second pass alloc/free = %d/%v, want 1/no frees", state.allocCalls, state.freed)
	}
}

func TestRewardReplenishNilInitDataFaultsBeforeSuccessorOrAllocation4F2210(t *testing.T) {
	marker := &rewardReplenishTestObject4F2210{name: "marker", typeInd: 10}
	state := &rewardReplenishTestState4F2210{
		markerCache: 10,
		plusCache:   20,
		potionCache: 30,
		firsts:      []*rewardReplenishTestObject4F2210{marker},
	}
	defer func() {
		if recover() == nil {
			t.Fatal("nil RewardMarker InitData did not fault")
		}
		want := []string{
			"quest", "players", "cache:marker", "first:1", "cache:marker",
			"type:marker", "data:marker", "low:nil",
		}
		if !reflect.DeepEqual(state.events, want) {
			t.Fatalf("nil-data fault prefix = %v, want %v", state.events, want)
		}
		if state.allocCalls != 0 || state.firstCalls != 1 {
			t.Fatalf("nil-data fault continued: alloc=%d first=%d", state.allocCalls, state.firstCalls)
		}
	}()
	rewardReplenish4F2210(rewardReplenishTestHooks4F2210(state))
}

func TestRewardReplenishComparesZeroExtendedTypeAgainstFullCaches4F2210(t *testing.T) {
	object := &rewardReplenishTestObject4F2210{
		name: "object", typeInd: 10,
		data: &rewardReplenishTestData4F2210{name: "object"},
	}
	state := &rewardReplenishTestState4F2210{
		markerCache: 0x0001000a,
		plusCache:   0x0002000a,
		potionCache: 0x0003000a,
		firsts:      []*rewardReplenishTestObject4F2210{object},
	}
	rewardReplenish4F2210(rewardReplenishTestHooks4F2210(state))
	want := []string{
		"quest", "players", "cache:marker", "first:1", "cache:marker",
		"type:object", "cache:potion", "next:object",
	}
	if !reflect.DeepEqual(state.events, want) {
		t.Fatalf("full-cache events = %v, want %v", state.events, want)
	}
}
