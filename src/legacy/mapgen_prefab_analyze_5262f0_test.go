package legacy

import (
	"math"
	"reflect"
	"testing"
	"unsafe"
)

type mapgenPrefabTestObject5262F0 struct {
	position mapgenPrefabFloatPoint5262F0
	extent   uint32
}

type mapgenPrefabTestState5262F0 struct {
	index  int32
	width  float32
	height float32
	exits  [mapgenPrefabDirections5262F0]mapgenPrefabExit5262F0
	max    int32
}

func mapgenPrefabTestDeps5262F0(
	state *mapgenPrefabTestState5262F0,
	objects []mapgenPrefabTestObject5262F0,
	trace *[]string,
) mapgenPrefabAnalyzeDeps5262F0[int] {
	markerIDs := [mapgenPrefabDirections5262F0]uint32{101, 202, 303, 404}
	return mapgenPrefabAnalyzeDeps5262F0[int]{
		lookup: func() int32 {
			*trace = append(*trace, "lookup")
			return 7
		},
		storeIndex: func(index int32) {
			state.index = index
			*trace = append(*trace, "index")
		},
		loadIndex: func() int32 {
			*trace = append(*trace, "load-index")
			return state.index
		},
		load: func(index int32) {
			*trace = append(*trace, "load")
		},
		width: func(index int32) float32 {
			*trace = append(*trace, "width")
			return 64.5
		},
		storeWidth: func(width float32) {
			state.width = width
			*trace = append(*trace, "store-width")
		},
		height: func(index int32) float32 {
			*trace = append(*trace, "height")
			return 96.25
		},
		storeHeight: func(height float32) {
			state.height = height
			*trace = append(*trace, "store-height")
		},
		firstObject: func() int {
			*trace = append(*trace, "first")
			if len(objects) == 0 {
				return 0
			}
			return 1
		},
		nextObject: func(object int) int {
			*trace = append(*trace, "next")
			if object >= len(objects) {
				return 0
			}
			return object + 1
		},
		position: func(object int) mapgenPrefabFloatPoint5262F0 {
			*trace = append(*trace, "position")
			return objects[object-1].position
		},
		round: func(point mapgenPrefabFloatPoint5262F0) mapgenPrefabIntPoint5262F0 {
			*trace = append(*trace, "round")
			return mapgenPrefabIntPoint5262F0{x: int32(point.x), y: int32(point.y)}
		},
		extent: func(object int) uint32 {
			*trace = append(*trace, "extent")
			return objects[object-1].extent
		},
		markerID: [mapgenPrefabDirections5262F0]func() uint32{
			func() uint32 { *trace = append(*trace, "north-id"); return markerIDs[0] },
			func() uint32 { *trace = append(*trace, "south-id"); return markerIDs[1] },
			func() uint32 { *trace = append(*trace, "east-id"); return markerIDs[2] },
			func() uint32 { *trace = append(*trace, "west-id"); return markerIDs[3] },
		},
		loadExit: func(direction int) mapgenPrefabExit5262F0 {
			return state.exits[direction]
		},
		storeExit: func(direction int, exit mapgenPrefabExit5262F0) {
			state.exits[direction] = exit
		},
		capacity: func() (int32, int32) {
			*trace = append(*trace, "capacity")
			return 1, 3
		},
		loadMax: func() int32 {
			return state.max
		},
		storeMax: func(maximum int32) {
			state.max = maximum
		},
	}
}

func TestMapgenAnalyzePrefab5262F0VerticalAggregation(t *testing.T) {
	state := mapgenPrefabTestState5262F0{}
	trace := []string{}
	objects := []mapgenPrefabTestObject5262F0{
		{position: mapgenPrefabFloatPoint5262F0{x: 11.75, y: 21.5}, extent: 0xCAFE0065},
		{position: mapgenPrefabFloatPoint5262F0{x: 7.25, y: 31.5}, extent: 101},
		{position: mapgenPrefabFloatPoint5262F0{x: 51.5, y: 61.5}, extent: 202},
		{position: mapgenPrefabFloatPoint5262F0{x: 99.5, y: 101.5}, extent: 999},
	}
	if !mapgenAnalyzePrefabWithDeps5262F0(mapgenPrefabTestDeps5262F0(&state, objects, &trace)) {
		t.Fatal("vertical north/south prefab should be accepted")
	}
	if state.index != 7 || state.width != 64.5 || state.height != 96.25 {
		t.Fatalf("metadata mismatch: %+v", state)
	}
	if got, want := state.exits[mapgenPrefabNorth5262F0], (mapgenPrefabExit5262F0{x: 6, y: 30, count: 2, present: 1}); got != want {
		t.Fatalf("north aggregation = %+v, want %+v", got, want)
	}
	if got, want := state.exits[mapgenPrefabSouth5262F0], (mapgenPrefabExit5262F0{x: 50, y: 60, count: 1, present: 1}); got != want {
		t.Fatalf("south aggregation = %+v, want %+v", got, want)
	}
	if state.max != 2 {
		t.Fatalf("maximum marker count = %d, want 2", state.max)
	}
	if got := len(trace); got == 0 || trace[0] != "lookup" || trace[1] != "index" {
		t.Fatalf("unexpected leading call order: %v", trace)
	}
	if rounded := countStrings5262F0(trace, "round"); rounded != len(objects) {
		t.Fatalf("round calls = %d, want %d including unknown markers", rounded, len(objects))
	}
	firstExtent := indexString5262F0(trace, "extent")
	firstNext := indexString5262F0(trace[firstExtent+1:], "next")
	if firstExtent < 0 || firstNext < 0 {
		t.Fatalf("missing first-object trace boundary: %v", trace)
	}
	firstObjectTrace := trace[firstExtent : firstExtent+firstNext+2]
	if want := []string{"extent", "north-id", "next"}; !reflect.DeepEqual(firstObjectTrace, want) {
		t.Fatalf("north match trace = %v, want %v", firstObjectTrace, want)
	}
}

func TestMapgenAnalyzePrefab5262F0HorizontalAndMixedAxes(t *testing.T) {
	tests := []struct {
		name    string
		objects []mapgenPrefabTestObject5262F0
		want    bool
	}{
		{
			name: "east and west",
			objects: []mapgenPrefabTestObject5262F0{
				{position: mapgenPrefabFloatPoint5262F0{x: 15, y: 25}, extent: 303},
				{position: mapgenPrefabFloatPoint5262F0{x: 35, y: 20}, extent: 303},
				{position: mapgenPrefabFloatPoint5262F0{x: 45, y: 55}, extent: 404},
			},
			want: true,
		},
		{
			name: "mixed vertical and horizontal",
			objects: []mapgenPrefabTestObject5262F0{
				{position: mapgenPrefabFloatPoint5262F0{x: 15, y: 25}, extent: 101},
				{position: mapgenPrefabFloatPoint5262F0{x: 35, y: 20}, extent: 303},
			},
			want: false,
		},
		{
			name: "three populated directions",
			objects: []mapgenPrefabTestObject5262F0{
				{position: mapgenPrefabFloatPoint5262F0{x: 15, y: 25}, extent: 101},
				{position: mapgenPrefabFloatPoint5262F0{x: 35, y: 20}, extent: 202},
				{position: mapgenPrefabFloatPoint5262F0{x: 45, y: 55}, extent: 404},
			},
			want: false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			state := mapgenPrefabTestState5262F0{}
			trace := []string{}
			got := mapgenAnalyzePrefabWithDeps5262F0(mapgenPrefabTestDeps5262F0(&state, test.objects, &trace))
			if got != test.want {
				t.Fatalf("result = %v, want %v; state=%+v", got, test.want, state)
			}
			if test.name == "east and west" {
				if got, want := state.exits[mapgenPrefabEast5262F0], (mapgenPrefabExit5262F0{x: 34, y: 19, count: 2, present: 1}); got != want {
					t.Fatalf("east aggregation = %+v, want %+v", got, want)
				}
			}
		})
	}
}

func TestMapgenAnalyzePrefab5262F0FailureAndSignedArithmetic(t *testing.T) {
	t.Run("dimension calls reload stored index", func(t *testing.T) {
		state := mapgenPrefabTestState5262F0{}
		trace := []string{}
		deps := mapgenPrefabTestDeps5262F0(&state, nil, &trace)
		var widthIndex int32
		var heightIndex int32
		deps.width = func(index int32) float32 {
			widthIndex = index
			state.index = 9
			return 64.5
		}
		deps.height = func(index int32) float32 {
			heightIndex = index
			return 96.25
		}
		if !mapgenAnalyzePrefabWithDeps5262F0(deps) {
			t.Fatal("empty prefab with sufficient capacity should pass")
		}
		if widthIndex != 7 || heightIndex != 9 {
			t.Fatalf("dimension indices = %d/%d, want 7/9", widthIndex, heightIndex)
		}
	})

	t.Run("lookup failure stores only index", func(t *testing.T) {
		state := mapgenPrefabTestState5262F0{index: 99, width: 12, height: 13, max: 14}
		trace := []string{}
		deps := mapgenPrefabTestDeps5262F0(&state, nil, &trace)
		deps.lookup = func() int32 { trace = append(trace, "lookup"); return -1 }
		if mapgenAnalyzePrefabWithDeps5262F0(deps) {
			t.Fatal("missing AreaMap record should fail")
		}
		if state.index != -1 || state.width != 12 || state.height != 13 || state.max != 14 {
			t.Fatalf("unexpected partial state: %+v", state)
		}
		if want := []string{"lookup", "index"}; !reflect.DeepEqual(trace, want) {
			t.Fatalf("trace = %v, want %v", trace, want)
		}
	})

	t.Run("existing maximum is not cleared", func(t *testing.T) {
		state := mapgenPrefabTestState5262F0{max: 7}
		trace := []string{}
		deps := mapgenPrefabTestDeps5262F0(&state, []mapgenPrefabTestObject5262F0{{extent: 101}}, &trace)
		deps.capacity = func() (int32, int32) { return 3, 3 }
		if mapgenAnalyzePrefabWithDeps5262F0(deps) {
			t.Fatal("preserved maximum 7 must exceed capacity 6")
		}
		if state.max != 7 {
			t.Fatalf("maximum = %d, want preserved 7", state.max)
		}
	})

	t.Run("signed capacity overflow", func(t *testing.T) {
		state := mapgenPrefabTestState5262F0{}
		trace := []string{}
		deps := mapgenPrefabTestDeps5262F0(&state, []mapgenPrefabTestObject5262F0{{extent: 101}}, &trace)
		deps.capacity = func() (int32, int32) { return math.MaxInt32, 1 }
		if mapgenAnalyzePrefabWithDeps5262F0(deps) {
			t.Fatal("count 1 must be greater than wrapped signed capacity MinInt32")
		}
	})

	t.Run("signed count overflow does not raise maximum", func(t *testing.T) {
		state := mapgenPrefabTestState5262F0{}
		state.exits[mapgenPrefabNorth5262F0] = mapgenPrefabExit5262F0{
			x: 1, y: 2, count: math.MaxInt32, present: 1,
		}
		trace := []string{}
		deps := mapgenPrefabTestDeps5262F0(&state, []mapgenPrefabTestObject5262F0{{extent: 101}}, &trace)
		deps.capacity = func() (int32, int32) { return 0, 0 }
		if !mapgenAnalyzePrefabWithDeps5262F0(deps) {
			t.Fatal("wrapped negative count with maximum 0 should pass signed comparison")
		}
		if state.exits[mapgenPrefabNorth5262F0].count != math.MinInt32 || state.max != 0 {
			t.Fatalf("signed wrap state = %+v, max=%d", state.exits[0], state.max)
		}
	})
}

func TestMapgenAnalyzePrefabEntry5262F0NativePointers(t *testing.T) {
	old := mapgenPrefabAnalyzeEntry5262F0
	t.Cleanup(func() { mapgenPrefabAnalyzeEntry5262F0 = old })

	var themeAddress uintptr
	var prefabAddress uintptr
	var calls int
	mapgenPrefabAnalyzeEntry5262F0 = func(theme, prefab unsafe.Pointer) bool {
		calls++
		themeAddress = uintptr(theme)
		prefabAddress = uintptr(prefab)
		return true
	}
	result := mapgenPrefabEntryFixture5262F0()
	if !result.result || result.nullThemeResult || result.nullPrefabResult {
		t.Fatalf("wrapper results = %+v", result)
	}
	if calls != 1 {
		t.Fatalf("native entry calls = %d, want 1", calls)
	}
	if themeAddress != result.themeAddress || prefabAddress != result.prefabAddress {
		t.Fatalf("pointer round trip = %#x/%#x, want %#x/%#x", themeAddress, prefabAddress, result.themeAddress, result.prefabAddress)
	}
	if unsafe.Sizeof(uintptr(0)) > 4 && (themeAddress <= math.MaxUint32 || prefabAddress <= math.MaxUint32) {
		t.Fatalf("fixture addresses must exercise >4GiB pointers: %#x/%#x", themeAddress, prefabAddress)
	}
}

func countStrings5262F0(values []string, want string) int {
	count := 0
	for _, value := range values {
		if value == want {
			count++
		}
	}
	return count
}

func indexString5262F0(values []string, want string) int {
	for index, value := range values {
		if value == want {
			return index
		}
	}
	return -1
}
