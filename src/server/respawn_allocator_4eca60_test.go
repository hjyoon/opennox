package server

import (
	"reflect"
	"testing"
)

func TestRespawnAllocator4ECA60OrderAndCachedResult(t *testing.T) {
	tests := []struct {
		name      string
		allocator int
		want      bool
	}{
		{name: "success", allocator: 0x1234, want: true},
		{name: "failure", allocator: 0, want: false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var events []string
			tested := false
			predicate := tc.allocator != 0
			stored := -1

			got := RespawnAllocator4ECA60(60, RespawnAllocatorHooks4ECA60[int]{
				NewClass: func(name string, recordSize uintptr, capacity int) int {
					events = append(events, "new-class")
					if name != "Respawn" || recordSize != 60 || capacity != 384 {
						t.Fatalf("allocation request = (%q, %d, %d), want (Respawn, 60, 384)", name, recordSize, capacity)
					}
					return tc.allocator
				},
				NonZero: func(allocator int) bool {
					events = append(events, "test-nonzero")
					if allocator != tc.allocator {
						t.Fatalf("tested allocator = %#x, want %#x", allocator, tc.allocator)
					}
					tested = true
					return predicate
				},
				StoreAllocator: func(allocator int) {
					events = append(events, "store-allocator")
					if !tested {
						t.Fatal("allocator was stored before its result was tested")
					}
					stored = allocator
					predicate = !predicate
				},
			})

			if got != tc.want {
				t.Fatalf("result = %v, want cached %v", got, tc.want)
			}
			if stored != tc.allocator {
				t.Fatalf("stored allocator = %#x, want %#x", stored, tc.allocator)
			}
			wantEvents := []string{"new-class", "test-nonzero", "store-allocator"}
			if !reflect.DeepEqual(events, wantEvents) {
				t.Fatalf("events = %#v, want %#v", events, wantEvents)
			}
		})
	}
}

func TestRespawnAllocator4ECA60FaultBoundaries(t *testing.T) {
	stop := &struct{}{}
	tests := []struct {
		name       string
		newClass   func(string, uintptr, int) int
		nonZero    func(int) bool
		wantEvents []string
	}{
		{
			name: "new-class",
			newClass: func(string, uintptr, int) int {
				panic(stop)
			},
			nonZero:    func(int) bool { t.Fatal("tested after allocation fault"); return false },
			wantEvents: nil,
		},
		{
			name: "nonzero-test",
			newClass: func(string, uintptr, int) int {
				return 7
			},
			nonZero: func(int) bool {
				panic(stop)
			},
			wantEvents: []string{"new-class"},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var events []string
			var recovered any
			func() {
				defer func() { recovered = recover() }()
				RespawnAllocator4ECA60(104, RespawnAllocatorHooks4ECA60[int]{
					NewClass: func(name string, size uintptr, capacity int) int {
						result := tc.newClass(name, size, capacity)
						events = append(events, "new-class")
						return result
					},
					NonZero: tc.nonZero,
					StoreAllocator: func(int) {
						events = append(events, "store-allocator")
					},
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
