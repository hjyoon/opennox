package server

import (
	"math"
	"reflect"
	"testing"

	"github.com/opennox/libs/types"
)

type mapFindPlayerStartTestObject4F7AB0 struct {
	typeIndex uint16
	flags     uint32
	teamID    uint8
	x         float32
	y         float32
}

type mapFindPlayerStartTestWorld4F7AB0 struct {
	cache      uint32
	lookup     uint32
	objects    []int
	players    []int
	values     map[int]*mapFindPlayerStartTestObject4F7AB0
	output     types.Pointf
	random     int32
	enemy      func(int, int) bool
	randomCall func(int32, int32, string, int32) int32
}

func mapFindPlayerStartTestNext4F7AB0(list []int, current int) int {
	for i, value := range list {
		if value == current && i+1 < len(list) {
			return list[i+1]
		}
	}
	return 0
}

func (w *mapFindPlayerStartTestWorld4F7AB0) hooks() mapFindPlayerStartHooks4F7AB0[int] {
	return mapFindPlayerStartHooks4F7AB0[int]{
		loadCachedType: func() uint32 { return w.cache },
		lookupType: func(string) uint32 {
			return w.lookup
		},
		storeCachedType: func(value uint32) { w.cache = value },
		hasTeam: func(object int) bool {
			return w.values[object].teamID != 0
		},
		loadTeamID: func(object int) uint8 {
			return w.values[object].teamID
		},
		touchTeam: func(uint8) {},
		firstObject: func() int {
			if len(w.objects) == 0 {
				return 0
			}
			return w.objects[0]
		},
		nextObject: func(object int) int {
			return mapFindPlayerStartTestNext4F7AB0(w.objects, object)
		},
		loadTypeIndex: func(object int) uint16 {
			return w.values[object].typeIndex
		},
		loadObjectFlags: func(object int) uint32 {
			return w.values[object].flags
		},
		teamContains: func(object int, id uint8) bool {
			return w.values[object].teamID == id
		},
		firstPlayer: func() int {
			if len(w.players) == 0 {
				return 0
			}
			return w.players[0]
		},
		nextPlayer: func(player int) int {
			return mapFindPlayerStartTestNext4F7AB0(w.players, player)
		},
		isEnemyTo: func(player, other int) bool {
			return w.enemy != nil && w.enemy(player, other)
		},
		loadPosX: func(object int) float32 {
			return w.values[object].x
		},
		loadPosY: func(object int) float32 {
			return w.values[object].y
		},
		randomInt: func(minimum, maximum int32, source string, line int32) int32 {
			if w.randomCall != nil {
				return w.randomCall(minimum, maximum, source, line)
			}
			return w.random
		},
		storeOutputX: func(value float32) { w.output.X = value },
		storeOutputY: func(value float32) { w.output.Y = value },
	}
}

func TestMapFindPlayerStartEligible4F7CE0ShortCircuitOrder(t *testing.T) {
	tests := []struct {
		name       string
		flags      uint32
		teamID     uint8
		hasTeam    bool
		contains   bool
		want       bool
		wantEvents []string
	}{
		{name: "disabled", teamID: 7, hasTeam: true, contains: true, wantEvents: []string{"flags"}},
		{name: "unscoped", flags: mapFindPlayerStartEnabled4F7CE0, want: true, wantEvents: []string{"flags"}},
		{name: "unteamed", flags: mapFindPlayerStartEnabled4F7CE0, teamID: 7, want: true, wantEvents: []string{"flags", "has-team"}},
		{name: "matching active team", flags: mapFindPlayerStartEnabled4F7CE0, teamID: 7, hasTeam: true, contains: true, want: true, wantEvents: []string{"flags", "has-team", "contains"}},
		{name: "inactive or different team", flags: mapFindPlayerStartEnabled4F7CE0, teamID: 7, hasTeam: true, wantEvents: []string{"flags", "has-team", "contains"}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var events []string
			hooks := mapFindPlayerStartHooks4F7AB0[int]{
				loadObjectFlags: func(int) uint32 {
					events = append(events, "flags")
					return test.flags
				},
				hasTeam: func(int) bool {
					events = append(events, "has-team")
					return test.hasTeam
				},
				teamContains: func(int, uint8) bool {
					events = append(events, "contains")
					return test.contains
				},
			}
			if got := mapFindPlayerStartEligible4F7CE0(1, test.teamID, hooks); got != test.want {
				t.Fatalf("eligible = %t, want %t", got, test.want)
			}
			if !reflect.DeepEqual(events, test.wantEvents) {
				t.Fatalf("events = %v, want %v", events, test.wantEvents)
			}
		})
	}
}

func TestMapFindPlayerStart4F7AB0CachesBeforeNilAndRetriesZero(t *testing.T) {
	w := &mapFindPlayerStartTestWorld4F7AB0{
		values: map[int]*mapFindPlayerStartTestObject4F7AB0{},
		output: types.Pointf{X: 12.5, Y: -4.25},
	}
	hooks := w.hooks()
	lookups := 0
	var events []string
	hooks.loadCachedType = func() uint32 {
		events = append(events, "load")
		return w.cache
	}
	hooks.lookupType = func(name string) uint32 {
		events = append(events, "lookup:"+name)
		lookups++
		if lookups == 1 {
			return 0
		}
		return 9
	}
	hooks.storeCachedType = func(value uint32) {
		events = append(events, "store")
		w.cache = value
	}
	hooks.storeOutputX = func(float32) { t.Fatal("nil player wrote X") }
	hooks.storeOutputY = func(float32) { t.Fatal("nil player wrote Y") }

	mapFindPlayerStart4F7AB0(0, hooks)
	mapFindPlayerStart4F7AB0(0, hooks)
	mapFindPlayerStart4F7AB0(0, hooks)

	wantEvents := []string{
		"load", "lookup:" + mapFindPlayerStartType4F7AB0, "store",
		"load", "lookup:" + mapFindPlayerStartType4F7AB0, "store",
		"load",
	}
	if !reflect.DeepEqual(events, wantEvents) {
		t.Fatalf("events = %v, want %v", events, wantEvents)
	}
	if w.cache != 9 || w.output != (types.Pointf{X: 12.5, Y: -4.25}) {
		t.Fatalf("state = cache:%d output:%+v", w.cache, w.output)
	}
}

func TestMapFindPlayerStart4F7AB0CachesTeamBeforeTouchAndStoresEmptyInOrder(t *testing.T) {
	w := &mapFindPlayerStartTestWorld4F7AB0{
		cache:   9,
		values:  map[int]*mapFindPlayerStartTestObject4F7AB0{1: {teamID: 7}},
		objects: nil,
	}
	hooks := w.hooks()
	var events []string
	hooks.loadCachedType = func() uint32 {
		events = append(events, "cache")
		return w.cache
	}
	hooks.hasTeam = func(object int) bool {
		events = append(events, "has-team")
		return w.values[object].teamID != 0
	}
	hooks.loadTeamID = func(object int) uint8 {
		events = append(events, "team-id")
		return w.values[object].teamID
	}
	hooks.touchTeam = func(id uint8) {
		events = append(events, "touch")
		if id != 7 {
			t.Fatalf("touched team = %d, want 7", id)
		}
		w.values[1].teamID = 3
	}
	hooks.firstObject = func() int {
		events = append(events, "first")
		return 0
	}
	hooks.storeOutputX = func(value float32) {
		events = append(events, "store-x")
		w.output.X = value
	}
	hooks.storeOutputY = func(value float32) {
		events = append(events, "store-y")
		w.output.Y = value
	}

	mapFindPlayerStart4F7AB0(1, hooks)
	wantEvents := []string{"cache", "has-team", "team-id", "touch", "first", "store-x", "store-y"}
	if !reflect.DeepEqual(events, wantEvents) {
		t.Fatalf("events = %v, want %v", events, wantEvents)
	}
	if w.output != (types.Pointf{X: 2000, Y: 2000}) {
		t.Fatalf("output = %+v, want (2000,2000)", w.output)
	}
}

func TestMapFindPlayerStart4F7AB0UsesLastMatchingFallbackAndLiveY(t *testing.T) {
	w := &mapFindPlayerStartTestWorld4F7AB0{
		cache:   9,
		objects: []int{2, 4, 3},
		values: map[int]*mapFindPlayerStartTestObject4F7AB0{
			1: {},
			2: {typeIndex: 9, x: 10, y: 20},
			3: {typeIndex: 9, x: 30, y: 40},
			4: {typeIndex: 8, x: 50, y: 60},
		},
	}
	hooks := w.hooks()
	var events []string
	hooks.loadObjectFlags = func(object int) uint32 {
		events = append(events, "flags")
		return 0
	}
	hooks.loadPosX = func(object int) float32 {
		events = append(events, "load-x")
		if object != 3 {
			t.Fatalf("fallback = %d, want last matching object 3", object)
		}
		return w.values[object].x
	}
	hooks.storeOutputX = func(value float32) {
		events = append(events, "store-x")
		w.output.X = value
		w.values[3].y = 99
	}
	hooks.loadPosY = func(object int) float32 {
		events = append(events, "load-y")
		return w.values[object].y
	}
	hooks.storeOutputY = func(value float32) {
		events = append(events, "store-y")
		w.output.Y = value
	}

	mapFindPlayerStart4F7AB0(1, hooks)
	if w.output != (types.Pointf{X: 30, Y: 99}) {
		t.Fatalf("output = %+v, want live (30,99)", w.output)
	}
	wantSuffix := []string{"load-x", "store-x", "load-y", "store-y"}
	if got := events[len(events)-len(wantSuffix):]; !reflect.DeepEqual(got, wantSuffix) {
		t.Fatalf("terminal events = %v, want %v", got, wantSuffix)
	}
}

func TestMapFindPlayerStart4F7AB0ReloadsTypeCacheForEveryComparison(t *testing.T) {
	w := &mapFindPlayerStartTestWorld4F7AB0{
		cache:   9,
		objects: []int{2, 3},
		values: map[int]*mapFindPlayerStartTestObject4F7AB0{
			1: {},
			2: {typeIndex: 9, flags: mapFindPlayerStartEnabled4F7CE0, x: 2},
			3: {typeIndex: 8, flags: mapFindPlayerStartEnabled4F7CE0, x: 3},
		},
		random: 1,
	}
	hooks := w.hooks()
	cacheLoads := 0
	hooks.loadCachedType = func() uint32 {
		cacheLoads++
		return w.cache
	}
	hooks.loadTypeIndex = func(object int) uint16 {
		// Mutate the backing cache after its current value is loaded. Object 2
		// matches 9 and makes object 3 match 8; object 3 restores 9 for the
		// next traversal. This exercises the count, distance, and random passes.
		if object == 2 {
			w.cache = 8
		} else {
			w.cache = 9
		}
		return w.values[object].typeIndex
	}

	mapFindPlayerStart4F7AB0(1, hooks)
	if cacheLoads != 7 { // initial cache test plus two objects in each of three passes
		t.Fatalf("cache loads = %d, want 7", cacheLoads)
	}
	if w.output != (types.Pointf{X: 3, Y: 0}) {
		t.Fatalf("output = %+v, want live random-pass object 3", w.output)
	}
}

func TestMapFindPlayerStart4F7AB0SelectsStrictGreatestNearestEnemy(t *testing.T) {
	w := &mapFindPlayerStartTestWorld4F7AB0{
		cache:   9,
		objects: []int{2, 3},
		players: []int{1, 4, 5},
		values: map[int]*mapFindPlayerStartTestObject4F7AB0{
			1: {},
			2: {typeIndex: 9, flags: mapFindPlayerStartEnabled4F7CE0, x: 0, y: 0},
			3: {typeIndex: 9, flags: mapFindPlayerStartEnabled4F7CE0, x: 10, y: 0},
			4: {x: 1, y: 0},
			5: {x: 100, y: 100},
		},
		enemy: func(_, other int) bool { return other == 4 },
	}
	hooks := w.hooks()
	var enemyCalls [][2]int
	var positionLoads []string
	hooks.isEnemyTo = func(player, other int) bool {
		enemyCalls = append(enemyCalls, [2]int{player, other})
		return w.enemy(player, other)
	}
	hooks.loadPosX = func(object int) float32 {
		positionLoads = append(positionLoads, "x:"+string(rune('0'+object)))
		return w.values[object].x
	}
	hooks.loadPosY = func(object int) float32 {
		positionLoads = append(positionLoads, "y:"+string(rune('0'+object)))
		return w.values[object].y
	}

	mapFindPlayerStart4F7AB0(1, hooks)
	if w.output != (types.Pointf{X: 10, Y: 0}) {
		t.Fatalf("output = %+v, want second start (10,0)", w.output)
	}
	if want := [][2]int{{1, 4}, {1, 5}, {1, 4}, {1, 5}}; !reflect.DeepEqual(enemyCalls, want) {
		t.Fatalf("enemy calls = %v, want %v (identity must be skipped)", enemyCalls, want)
	}
	wantPrefix := []string{"x:2", "x:4", "y:2", "y:4", "x:3", "x:4", "y:3", "y:4"}
	if !reflect.DeepEqual(positionLoads[:len(wantPrefix)], wantPrefix) {
		t.Fatalf("distance loads = %v, want %v", positionLoads[:len(wantPrefix)], wantPrefix)
	}
}

func TestMapFindPlayerStart4F7AB0EqualDistanceKeepsFirst(t *testing.T) {
	w := &mapFindPlayerStartTestWorld4F7AB0{
		cache:   9,
		objects: []int{2, 3},
		players: []int{4},
		values: map[int]*mapFindPlayerStartTestObject4F7AB0{
			1: {},
			2: {typeIndex: 9, flags: mapFindPlayerStartEnabled4F7CE0, x: -1, y: 0},
			3: {typeIndex: 9, flags: mapFindPlayerStartEnabled4F7CE0, x: 1, y: 0},
			4: {x: 0, y: 0},
		},
		enemy: func(int, int) bool { return true },
	}

	mapFindPlayerStart4F7AB0(1, w.hooks())
	if w.output != (types.Pointf{X: -1, Y: 0}) {
		t.Fatalf("output = %+v, want first equal-distance start", w.output)
	}
}

func TestMapFindPlayerStart4F7AB0NaNComparisonSemantics(t *testing.T) {
	t.Run("later finite overwrites NaN nearest", func(t *testing.T) {
		w := &mapFindPlayerStartTestWorld4F7AB0{
			cache:   9,
			objects: []int{2},
			players: []int{4, 5},
			values: map[int]*mapFindPlayerStartTestObject4F7AB0{
				1: {},
				2: {typeIndex: 9, flags: mapFindPlayerStartEnabled4F7CE0, x: 3, y: 0},
				4: {},
				5: {},
			},
			enemy: func(int, int) bool { return true },
		}
		hooks := w.hooks()
		startXLoads := 0
		hooks.loadPosX = func(object int) float32 {
			if object == 2 {
				startXLoads++
				if startXLoads == 1 {
					return float32(math.NaN())
				}
			}
			return w.values[object].x
		}

		mapFindPlayerStart4F7AB0(1, hooks)
		if w.output != (types.Pointf{X: 3, Y: 0}) {
			t.Fatalf("output = %+v, want finite candidate after NaN nearest", w.output)
		}
	})

	t.Run("NaN candidate rejects and uses random fallback", func(t *testing.T) {
		w := &mapFindPlayerStartTestWorld4F7AB0{
			cache:   9,
			objects: []int{2},
			players: []int{4},
			values: map[int]*mapFindPlayerStartTestObject4F7AB0{
				1: {},
				2: {typeIndex: 9, flags: mapFindPlayerStartEnabled4F7CE0, x: float32(math.NaN()), y: 8},
				4: {},
			},
			enemy: func(int, int) bool { return true },
		}
		randomCalls := 0
		w.randomCall = func(minimum, maximum int32, source string, line int32) int32 {
			randomCalls++
			return 0
		}

		mapFindPlayerStart4F7AB0(1, w.hooks())
		if randomCalls != 1 || !math.IsNaN(float64(w.output.X)) || w.output.Y != 8 {
			t.Fatalf("fallback = calls:%d output:%+v", randomCalls, w.output)
		}
	})
}

func TestMapFindPlayerStart4F7AB0RandomFallbackUsesExactCallAndLiveEligibility(t *testing.T) {
	w := &mapFindPlayerStartTestWorld4F7AB0{
		cache:   9,
		objects: []int{2, 3, 4},
		values: map[int]*mapFindPlayerStartTestObject4F7AB0{
			1: {},
			2: {typeIndex: 9, flags: mapFindPlayerStartEnabled4F7CE0, x: 2},
			3: {typeIndex: 9, flags: mapFindPlayerStartEnabled4F7CE0, x: 3},
			4: {typeIndex: 9, flags: mapFindPlayerStartEnabled4F7CE0, x: 4},
		},
	}
	w.randomCall = func(minimum, maximum int32, source string, line int32) int32 {
		if minimum != 0 || maximum != 2 || source != mapFindPlayerStartSource4F7AB0 || line != 0x116 {
			t.Fatalf("random = (%d,%d,%q,%#x)", minimum, maximum, source, line)
		}
		w.values[2].flags = 0
		return 1
	}

	mapFindPlayerStart4F7AB0(1, w.hooks())
	if w.output != (types.Pointf{X: 4, Y: 0}) {
		t.Fatalf("output = %+v, want live eligible index 1 at object 4", w.output)
	}
}

func TestMapFindPlayerStart4F7AB0InvalidRandomIndexRetainsNilFault(t *testing.T) {
	w := &mapFindPlayerStartTestWorld4F7AB0{
		cache:   9,
		objects: []int{2},
		values: map[int]*mapFindPlayerStartTestObject4F7AB0{
			1: {},
			2: {typeIndex: 9, flags: mapFindPlayerStartEnabled4F7CE0},
		},
		random: 2,
	}
	hooks := w.hooks()
	stores := 0
	hooks.storeOutputX = func(float32) { stores++ }
	hooks.storeOutputY = func(float32) { stores++ }

	defer func() {
		if recover() == nil {
			t.Fatal("invalid random index did not fault on nil selected object")
		}
		if stores != 0 {
			t.Fatalf("stores before nil position fault = %d, want 0", stores)
		}
	}()
	mapFindPlayerStart4F7AB0(1, hooks)
}
