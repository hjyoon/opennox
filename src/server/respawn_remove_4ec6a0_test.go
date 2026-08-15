package server

import (
	"fmt"
	"reflect"
	"testing"
)

func TestRespawnRemove4EC6A0HeadFastPathUsesTwoLiveNextLoads(t *testing.T) {
	var events []string
	hooks := RespawnRemoveHooks4EC6A0[int, int, int]{
		LoadHead: func() int {
			events = append(events, "load-head")
			return 1
		},
		LoadObject: func(rec int) int {
			events = append(events, fmt.Sprintf("load-object:%d", rec))
			return 9
		},
		LoadNext: func(rec int) int {
			count := 0
			for _, event := range events {
				if event == "load-next:1" {
					count++
				}
			}
			events = append(events, "load-next:1")
			if count == 0 {
				return 2
			}
			return 3
		},
		LoadPrev: func(int) int {
			t.Fatal("head fast-path read prev")
			return 0
		},
		StoreHead: func(rec int) { events = append(events, fmt.Sprintf("store-head:%d", rec)) },
		StoreNext: func(int, int) { t.Fatal("head fast-path stored predecessor next") },
		StorePrev: func(rec, prev int) {
			events = append(events, fmt.Sprintf("store-prev:%d:%d", rec, prev))
		},
		LoadAllocator: func() int {
			events = append(events, "load-allocator")
			return 7
		},
		FreeFirst: func(allocator, rec int) {
			events = append(events, fmt.Sprintf("free:%d:%d", allocator, rec))
		},
	}

	RespawnRemove4EC6A0(9, hooks)

	want := []string{
		"load-head",
		"load-object:1",
		"load-next:1",
		"store-head:2",
		"load-next:1",
		"store-prev:3:0",
		"load-allocator",
		"free:7:1",
	}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
}

func TestRespawnRemove4EC6A0ReadsEmptyHeadBeforeZeroCheck(t *testing.T) {
	var events []string
	hooks := RespawnRemoveHooks4EC6A0[int, int, int]{
		LoadHead: func() int {
			events = append(events, "load-head")
			return 0
		},
		LoadObject: func(rec int) int {
			events = append(events, fmt.Sprintf("load-object:%d", rec))
			panic("zero record dereference")
		},
	}
	defer func() {
		if recover() == nil {
			t.Fatal("empty head did not fault")
		}
		want := []string{"load-head", "load-object:0"}
		if !reflect.DeepEqual(events, want) {
			t.Fatalf("events = %v, want %v", events, want)
		}
	}()
	RespawnRemove4EC6A0(9, hooks)
}

func TestRespawnRemove4EC6A0ReloadsHeadObjectInSearchPath(t *testing.T) {
	var events []string
	objectLoads := 0
	hooks := RespawnRemoveHooks4EC6A0[int, int, int]{
		LoadHead: func() int {
			events = append(events, "load-head")
			return 1
		},
		LoadObject: func(rec int) int {
			objectLoads++
			events = append(events, fmt.Sprintf("load-object:%d:%d", rec, objectLoads))
			if objectLoads == 1 {
				return 8
			}
			return 9
		},
		LoadNext: func(rec int) int {
			events = append(events, fmt.Sprintf("load-next:%d", rec))
			return 2
		},
		LoadPrev: func(rec int) int {
			events = append(events, fmt.Sprintf("load-prev:%d", rec))
			return 0
		},
		StoreHead: func(int) { t.Fatal("search path stored head") },
		StoreNext: func(int, int) { t.Fatal("nil predecessor stored next") },
		StorePrev: func(rec, prev int) {
			events = append(events, fmt.Sprintf("store-prev:%d:%d", rec, prev))
		},
		LoadAllocator: func() int {
			events = append(events, "load-allocator")
			return 7
		},
		FreeFirst: func(allocator, rec int) {
			events = append(events, fmt.Sprintf("free:%d:%d", allocator, rec))
		},
	}

	RespawnRemove4EC6A0(9, hooks)

	want := []string{
		"load-head",
		"load-object:1:1",
		"load-object:1:2",
		"load-prev:1",
		"load-next:1",
		"load-prev:1",
		"store-prev:2:0",
		"load-allocator",
		"free:7:1",
	}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
}

func TestRespawnRemove4EC6A0InteriorMatchUsesLiveLinks(t *testing.T) {
	var events []string
	objectLoads := map[int]int{}
	nextLoads := map[int]int{}
	prevLoads := map[int]int{}
	hooks := RespawnRemoveHooks4EC6A0[int, int, int]{
		LoadHead: func() int {
			events = append(events, "load-head")
			return 1
		},
		LoadObject: func(rec int) int {
			objectLoads[rec]++
			events = append(events, fmt.Sprintf("load-object:%d:%d", rec, objectLoads[rec]))
			if rec == 2 {
				return 9
			}
			return 8
		},
		LoadNext: func(rec int) int {
			nextLoads[rec]++
			events = append(events, fmt.Sprintf("load-next:%d:%d", rec, nextLoads[rec]))
			if rec == 1 {
				return 2
			}
			if nextLoads[rec] == 1 {
				return 4
			}
			return 5
		},
		LoadPrev: func(rec int) int {
			prevLoads[rec]++
			events = append(events, fmt.Sprintf("load-prev:%d:%d", rec, prevLoads[rec]))
			if prevLoads[rec] == 1 {
				return 3
			}
			return 6
		},
		StoreHead: func(int) { t.Fatal("interior path stored head") },
		StoreNext: func(rec, next int) {
			events = append(events, fmt.Sprintf("store-next:%d:%d", rec, next))
		},
		StorePrev: func(rec, prev int) {
			events = append(events, fmt.Sprintf("store-prev:%d:%d", rec, prev))
		},
		LoadAllocator: func() int {
			events = append(events, "load-allocator")
			return 7
		},
		FreeFirst: func(allocator, rec int) {
			events = append(events, fmt.Sprintf("free:%d:%d", allocator, rec))
		},
	}

	RespawnRemove4EC6A0(9, hooks)

	want := []string{
		"load-head",
		"load-object:1:1",
		"load-object:1:2",
		"load-next:1:1",
		"load-object:2:1",
		"load-prev:2:1",
		"load-next:2:1",
		"store-next:3:4",
		"load-next:2:2",
		"load-prev:2:2",
		"store-prev:5:6",
		"load-allocator",
		"free:7:2",
	}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
}

func TestRespawnRemove4EC6A0MissingObjectDoesNotLoadAllocator(t *testing.T) {
	var events []string
	objectLoads := map[int]int{}
	hooks := RespawnRemoveHooks4EC6A0[int, int, int]{
		LoadHead: func() int {
			events = append(events, "load-head")
			return 1
		},
		LoadObject: func(rec int) int {
			objectLoads[rec]++
			events = append(events, fmt.Sprintf("load-object:%d:%d", rec, objectLoads[rec]))
			return rec + 10
		},
		LoadNext: func(rec int) int {
			events = append(events, fmt.Sprintf("load-next:%d", rec))
			if rec == 1 {
				return 2
			}
			return 0
		},
		LoadAllocator: func() int {
			t.Fatal("missing path loaded allocator")
			return 0
		},
		FreeFirst: func(int, int) { t.Fatal("missing path freed a record") },
	}

	RespawnRemove4EC6A0(9, hooks)

	want := []string{
		"load-head",
		"load-object:1:1",
		"load-object:1:2",
		"load-next:1",
		"load-object:2:1",
		"load-next:2",
	}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
}
