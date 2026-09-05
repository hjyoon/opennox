package server

import (
	"fmt"
	"reflect"
	"testing"
)

func TestSpellDurationCleanup4FE880OrderAndAllocatorSnapshot(t *testing.T) {
	const original = uint64(0xfedcba9876543210)
	const replacement = uint64(0x123456789abcdef0)
	live := original
	var events []string

	SpellDurationCleanup4FE880(SpellDurationCleanupHooks4FE880[uint64]{
		LoadAllocator: func() uint64 {
			events = append(events, "load-allocator")
			return live
		},
		FreeAllocator: func(value uint64) {
			events = append(events, "free-allocator")
			if value != original {
				t.Fatalf("freed allocator = %#x, want exact snapshot %#x", value, original)
			}
			live = replacement
		},
		ClearList: func() {
			events = append(events, "clear-list")
			if live != replacement {
				t.Fatalf("live allocator = %#x, want callback mutation %#x", live, replacement)
			}
		},
	})

	want := []string{"load-allocator", "free-allocator", "clear-list"}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %#v, want %#v", events, want)
	}
}

func TestSpellDurationCleanup4FE880ForwardsNilAllocator(t *testing.T) {
	freed := false
	cleared := false
	SpellDurationCleanup4FE880(SpellDurationCleanupHooks4FE880[*int]{
		LoadAllocator: func() *int { return nil },
		FreeAllocator: func(value *int) {
			freed = true
			if value != nil {
				t.Fatalf("freed allocator = %p, want nil", value)
			}
		},
		ClearList: func() {
			if !freed {
				t.Fatal("list cleared before nil allocator was forwarded")
			}
			cleared = true
		},
	})
	if !freed || !cleared {
		t.Fatalf("free/clear = (%t, %t), want (true, true)", freed, cleared)
	}
}

func TestSpellDurationCleanup4FE880FaultPrefixes(t *testing.T) {
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
				SpellDurationCleanup4FE880(SpellDurationCleanupHooks4FE880[uint64]{
					LoadAllocator: func() uint64 {
						step("load-allocator")
						return 7
					},
					FreeAllocator: func(uint64) {
						step("free-allocator")
					},
					ClearList: func() {
						step("clear-list")
					},
				})
			}()
			if recovered != stop {
				t.Fatalf("recovered = %#v, want sentinel", recovered)
			}
			want := []string{"load-allocator", "free-allocator", "clear-list"}[:faultAt]
			if !reflect.DeepEqual(events, want) {
				t.Fatalf("events = %#v, want fault prefix %#v", events, want)
			}
		})
	}
}
