package server

import (
	"fmt"
	"math"
	"reflect"
	"testing"
)

type mapInitializeScriptTable4FC590 struct {
	id    string
	names []string
}

type mapInitializeTestWorld4FC590 struct {
	events      []string
	faultAt     int
	state       int32
	hasUnit     bool
	count       int32
	table       *mapInitializeScriptTable4FC590
	clearResult int32
	callResult  int32
	calls       []int32
	nameOffsets []uint32
	onLoadName  func(uint32)
	onCall      func(int32)
}

func (w *mapInitializeTestWorld4FC590) record(event string) {
	w.events = append(w.events, event)
	if w.faultAt != 0 && len(w.events) == w.faultAt {
		panic(event)
	}
}

func (w *mapInitializeTestWorld4FC590) hooks() mapInitializeHooks4FC590[*mapInitializeScriptTable4FC590] {
	return mapInitializeHooks4FC590[*mapInitializeScriptTable4FC590]{
		loadState: func() int32 {
			value := w.state
			w.record(fmt.Sprintf("state:%#08x", uint32(value)))
			return value
		},
		hasPlayerUnit: func() bool {
			w.record(fmt.Sprintf("unit:%t", w.hasUnit))
			return w.hasUnit
		},
		loadScriptCount: func() int32 {
			value := w.count
			w.record(fmt.Sprintf("count:%d", value))
			return value
		},
		loadScriptTable: func() *mapInitializeScriptTable4FC590 {
			table := w.table
			id := "<nil>"
			if table != nil {
				id = table.id
			}
			w.record("table:" + id)
			return table
		},
		loadScriptName: func(table *mapInitializeScriptTable4FC590, offset uint32) string {
			id := "<nil>"
			if table != nil {
				id = table.id
			}
			w.record(fmt.Sprintf("name:%s:%d", id, offset))
			w.nameOffsets = append(w.nameOffsets, offset)
			if w.onLoadName != nil {
				w.onLoadName(offset)
			}
			return table.names[offset/mapInitializeScriptStride4FC590]
		},
		callScriptByIndex: func(index int32, caller, trigger uintptr) int32 {
			w.record(fmt.Sprintf("call:%d:%d:%d", index, caller, trigger))
			w.calls = append(w.calls, index)
			if w.onCall != nil {
				w.onCall(index)
			}
			return w.callResult
		},
		clearState: func(value int32) int32 {
			w.record(fmt.Sprintf("clear:%#08x", uint32(value)))
			w.state = value
			return w.clearResult
		},
	}
}

func TestMapInitialize4FC590StateAndPlayerGates(t *testing.T) {
	for _, tc := range []struct {
		name       string
		state      int32
		hasUnit    bool
		wantState  int32
		wantEvents []string
	}{
		{
			name:       "zero state",
			state:      0,
			wantState:  0,
			wantEvents: []string{"state:0x00000000"},
		},
		{
			name:       "no player unit",
			state:      -1985229329, // 0x89abcdef
			wantState:  -1985229329,
			wantEvents: []string{"state:0x89abcdef", "unit:false"},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			w := &mapInitializeTestWorld4FC590{
				state:   tc.state,
				hasUnit: tc.hasUnit,
			}
			if got := mapInitialize4FC590(w.hooks()); got != 0 {
				t.Fatalf("result = %#08x, want zero", uint32(got))
			}
			if uint32(w.state) != uint32(tc.wantState) {
				t.Fatalf("state = %#08x, want %#08x", uint32(w.state), uint32(tc.wantState))
			}
			if !reflect.DeepEqual(w.events, tc.wantEvents) {
				t.Fatalf("events = %#v, want %#v", w.events, tc.wantEvents)
			}
		})
	}
}

func TestMapInitialize4FC590NonpositiveCountsStillClearState(t *testing.T) {
	for _, count := range []int32{math.MinInt32, -1, 0} {
		t.Run(fmt.Sprintf("%d", count), func(t *testing.T) {
			w := &mapInitializeTestWorld4FC590{
				state:       math.MinInt32,
				hasUnit:     true,
				count:       count,
				clearResult: -1985229329, // verify the clear setter's EAX is propagated
			}
			if got := mapInitialize4FC590(w.hooks()); uint32(got) != 0x89abcdef {
				t.Fatalf("result = %#08x, want clear result", uint32(got))
			}
			if w.state != 0 {
				t.Fatalf("state = %#08x, want zero", uint32(w.state))
			}
			want := []string{
				"state:0x80000000",
				"unit:true",
				fmt.Sprintf("count:%d", count),
				"clear:0x00000000",
			}
			if !reflect.DeepEqual(w.events, want) {
				t.Fatalf("events = %#v, want %#v", w.events, want)
			}
		})
	}
}

func TestMapInitialize4FC590ExactThirteenBytePrefixAndStride(t *testing.T) {
	w := &mapInitializeTestWorld4FC590{
		state:   2,
		hasUnit: true,
		count:   6,
		table: &mapInitializeScriptTable4FC590{
			id: "prefixes",
			names: []string{
				"MapInitialize",
				"MapInitializeSuffix",
				"mapInitialize",
				"MapInitializ",
				"MapInitializ\x00e",
				"MapInitialize\x00Tail",
			},
		},
		callResult: math.MinInt32,
	}
	if got := mapInitialize4FC590(w.hooks()); got != 0 {
		t.Fatalf("result = %#08x, want zero", uint32(got))
	}
	if want := []int32{0, 1, 5}; !reflect.DeepEqual(w.calls, want) {
		t.Fatalf("calls = %v, want %v", w.calls, want)
	}
	wantOffsets := []uint32{0, 48, 96, 144, 192, 240}
	if !reflect.DeepEqual(w.nameOffsets, wantOffsets) {
		t.Fatalf("name offsets = %v, want %v", w.nameOffsets, wantOffsets)
	}
	if w.state != 0 {
		t.Fatalf("state = %d, want zero", w.state)
	}
}

func TestMapInitialize4FC590ReloadsTableAndGrownCountAcrossCallback(t *testing.T) {
	tableA := &mapInitializeScriptTable4FC590{
		id:    "A",
		names: []string{"MapInitializeFirst"},
	}
	tableB := &mapInitializeScriptTable4FC590{
		id: "B",
		names: []string{
			"ignored",
			"Other",
			"MapInitializeAdded",
		},
	}
	w := &mapInitializeTestWorld4FC590{
		state:   1,
		hasUnit: true,
		count:   1,
		table:   tableA,
	}
	w.onCall = func(index int32) {
		if index == 0 {
			w.table = tableB
			w.count = 3
		}
	}
	mapInitialize4FC590(w.hooks())
	if want := []int32{0, 2}; !reflect.DeepEqual(w.calls, want) {
		t.Fatalf("calls = %v, want %v", w.calls, want)
	}
	wantEvents := []string{
		"state:0x00000001", "unit:true", "count:1",
		"table:A", "name:A:0", "call:0:0:0", "count:3",
		"table:B", "name:B:48", "count:3",
		"table:B", "name:B:96", "call:2:0:0", "count:3",
		"clear:0x00000000",
	}
	if !reflect.DeepEqual(w.events, wantEvents) {
		t.Fatalf("events = %#v, want %#v", w.events, wantEvents)
	}
}

func TestMapInitialize4FC590ReloadsShrunkCountAfterCallback(t *testing.T) {
	w := &mapInitializeTestWorld4FC590{
		state:   1,
		hasUnit: true,
		count:   3,
		table: &mapInitializeScriptTable4FC590{
			id:    "A",
			names: []string{"MapInitialize", "MapInitializeSecond", "MapInitializeThird"},
		},
	}
	w.onCall = func(index int32) {
		if index == 0 {
			w.count = 1
		}
	}
	mapInitialize4FC590(w.hooks())
	if want := []int32{0}; !reflect.DeepEqual(w.calls, want) {
		t.Fatalf("calls = %v, want %v", w.calls, want)
	}
	wantEvents := []string{
		"state:0x00000001", "unit:true", "count:3",
		"table:A", "name:A:0", "call:0:0:0", "count:1",
		"clear:0x00000000",
	}
	if !reflect.DeepEqual(w.events, wantEvents) {
		t.Fatalf("events = %#v, want %#v", w.events, wantEvents)
	}
}

func TestMapInitialize4FC590ReloadsCountAfterMismatch(t *testing.T) {
	w := &mapInitializeTestWorld4FC590{
		state:   1,
		hasUnit: true,
		count:   3,
		table: &mapInitializeScriptTable4FC590{
			id:    "A",
			names: []string{"Other", "MapInitializeSecond", "MapInitializeThird"},
		},
	}
	w.onLoadName = func(offset uint32) {
		if offset == 0 {
			w.count = 0
		}
	}
	mapInitialize4FC590(w.hooks())
	if len(w.calls) != 0 {
		t.Fatalf("calls = %v, want none", w.calls)
	}
	wantEvents := []string{
		"state:0x00000001", "unit:true", "count:3",
		"table:A", "name:A:0", "count:0", "clear:0x00000000",
	}
	if !reflect.DeepEqual(w.events, wantEvents) {
		t.Fatalf("events = %#v, want %#v", w.events, wantEvents)
	}
}

func TestMapInitialize4FC590FaultPrefixes(t *testing.T) {
	wantEvents := []string{
		"state:0x89abcdef",
		"unit:true",
		"count:2",
		"table:A",
		"name:A:0",
		"call:0:0:0",
		"count:2",
		"table:A",
		"name:A:48",
		"count:2",
		"clear:0x00000000",
	}
	for faultAt := 1; faultAt <= len(wantEvents); faultAt++ {
		t.Run(fmt.Sprintf("fault-%d", faultAt), func(t *testing.T) {
			w := &mapInitializeTestWorld4FC590{
				faultAt: faultAt,
				state:   -1985229329, // 0x89abcdef
				hasUnit: true,
				count:   2,
				table: &mapInitializeScriptTable4FC590{
					id:    "A",
					names: []string{"MapInitialize", "Other"},
				},
			}
			defer func() {
				if recover() == nil {
					t.Fatal("fault was not propagated")
				}
				if want := wantEvents[:faultAt]; !reflect.DeepEqual(w.events, want) {
					t.Fatalf("events = %#v, want prefix %#v", w.events, want)
				}
				if uint32(w.state) != 0x89abcdef {
					t.Fatalf("state = %#08x, want unchanged", uint32(w.state))
				}
			}()
			mapInitialize4FC590(w.hooks())
		})
	}
}
