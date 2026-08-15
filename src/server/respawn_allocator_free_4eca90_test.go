package server

import (
	"reflect"
	"testing"
)

func TestRespawnAllocatorFree4ECA90OrderAndCachedAllocator(t *testing.T) {
	allocator := "original"
	loads := 0
	var events []string

	RespawnAllocatorFree4ECA90(RespawnAllocatorFreeHooks4ECA90[string]{
		LoadAllocator: func() string {
			loads++
			events = append(events, "load-allocator")
			return allocator
		},
		FreeClass: func(got string) {
			events = append(events, "free-class:"+got)
			allocator = "replacement"
			if got != "original" {
				t.Fatalf("free-class allocator = %q, want cached original", got)
			}
		},
	})

	if loads != 1 {
		t.Fatalf("allocator loads = %d, want 1", loads)
	}
	if allocator != "replacement" {
		t.Fatalf("allocator after free callback = %q, want replacement", allocator)
	}
	want := []string{"load-allocator", "free-class:original"}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %#v, want %#v", events, want)
	}
}

func TestRespawnAllocatorFree4ECA90ForwardsNil(t *testing.T) {
	called := false
	RespawnAllocatorFree4ECA90(RespawnAllocatorFreeHooks4ECA90[*int]{
		LoadAllocator: func() *int { return nil },
		FreeClass: func(allocator *int) {
			called = true
			if allocator != nil {
				t.Fatalf("free-class allocator = %p, want nil", allocator)
			}
		},
	})
	if !called {
		t.Fatal("nil allocator did not reach free-class boundary")
	}
}

func TestRespawnAllocatorFree4ECA90FaultBoundaries(t *testing.T) {
	stop := &struct{}{}
	tests := []struct {
		name       string
		load       func() int
		free       func(int)
		wantEvents []string
	}{
		{
			name: "load",
			load: func() int {
				panic(stop)
			},
			free:       func(int) { t.Fatal("free ran after load fault") },
			wantEvents: nil,
		},
		{
			name: "free",
			load: func() int {
				return 7
			},
			free: func(int) {
				panic(stop)
			},
			wantEvents: []string{"load-allocator"},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var events []string
			var recovered any
			func() {
				defer func() { recovered = recover() }()
				RespawnAllocatorFree4ECA90(RespawnAllocatorFreeHooks4ECA90[int]{
					LoadAllocator: func() int {
						allocator := tc.load()
						events = append(events, "load-allocator")
						return allocator
					},
					FreeClass: tc.free,
				})
			}()
			if recovered != stop {
				t.Fatalf("recovered = %#v, want sentinel", recovered)
			}
			if !reflect.DeepEqual(events, tc.wantEvents) {
				t.Fatalf("events = %#v, want %#v", events, tc.wantEvents)
			}
		})
	}
}
