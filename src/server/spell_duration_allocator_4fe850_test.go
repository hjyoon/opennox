package server

import (
	"fmt"
	"reflect"
	"testing"
)

func TestSpellDurationAllocator4FE850OrderConstantsAndCachedResult(t *testing.T) {
	const nativeRecordSize = uintptr(184)
	tests := []struct {
		name      string
		allocator uintptr
		want      int32
	}{
		{name: "success", allocator: uintptr(0x100001234), want: 1},
		{name: "failure", allocator: 0, want: 0},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var events []string
			tested := false
			predicate := tc.allocator != 0
			stored := ^uintptr(0)

			got := SpellDurationAllocator4FE850(
				nativeRecordSize,
				SpellDurationAllocatorHooks4FE850[uintptr]{
					NewClass: func(name string, recordSize uintptr, capacity int) uintptr {
						events = append(events, "new-class")
						if name != "spellDuration" || recordSize != nativeRecordSize || capacity != 512 {
							t.Fatalf(
								"allocation request = (%q, %d, %d), want (spellDuration, %d, 512)",
								name, recordSize, capacity, nativeRecordSize,
							)
						}
						return tc.allocator
					},
					NonZero: func(allocator uintptr) bool {
						events = append(events, "test-nonzero")
						if allocator != tc.allocator {
							t.Fatalf("tested allocator = %#x, want %#x", allocator, tc.allocator)
						}
						tested = true
						return predicate
					},
					StoreAllocator: func(allocator uintptr) {
						events = append(events, "store-allocator")
						if !tested {
							t.Fatal("allocator was stored before its result was tested")
						}
						stored = allocator
						predicate = !predicate
					},
				},
			)

			if got != tc.want {
				t.Fatalf("result = %d, want cached canonical %d", got, tc.want)
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

func TestSpellDurationAllocator4FE850FaultPrefixes(t *testing.T) {
	stop := &struct{}{}
	for faultAt := 1; faultAt <= 3; faultAt++ {
		t.Run(fmt.Sprintf("fault-%d", faultAt), func(t *testing.T) {
			var events []string
			step := func(event string) {
				events = append(events, event)
				if len(events) == faultAt {
					panic(stop)
				}
			}
			var recovered any
			func() {
				defer func() { recovered = recover() }()
				SpellDurationAllocator4FE850(
					120,
					SpellDurationAllocatorHooks4FE850[uintptr]{
						NewClass: func(string, uintptr, int) uintptr {
							step("new-class")
							return 7
						},
						NonZero: func(uintptr) bool {
							step("test-nonzero")
							return true
						},
						StoreAllocator: func(uintptr) {
							step("store-allocator")
						},
					},
				)
			}()
			if recovered != stop {
				t.Fatalf("recovered = %#v, want sentinel", recovered)
			}
			want := []string{"new-class", "test-nonzero", "store-allocator"}[:faultAt]
			if !reflect.DeepEqual(events, want) {
				t.Fatalf("events = %#v, want fault prefix %#v", events, want)
			}
		})
	}
}
