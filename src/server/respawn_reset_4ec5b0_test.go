package server

import (
	"reflect"
	"testing"
)

func TestRespawnReset4EC5B0OrderAndCachedAllocator(t *testing.T) {
	allocator := "original"
	head := "record"
	allow := uint32(0x11223344)
	var events []string

	RespawnReset4EC5B0(RespawnResetHooks4EC5B0[string]{
		LoadAllocator: func() string {
			events = append(events, "load-allocator")
			return allocator
		},
		ClearHead: func() {
			events = append(events, "clear-head")
			head = ""
			allocator = "replacement"
		},
		FreeAll: func(got string) {
			events = append(events, "free-all:"+got)
			if got != "original" {
				t.Fatalf("free-all allocator = %q, want cached original", got)
			}
			if head != "" {
				t.Fatalf("head during free-all = %q, want empty", head)
			}
			if allow != 0x11223344 {
				t.Fatalf("allow during free-all = %#x, want old value", allow)
			}
			allow = 0xaabbccdd
		},
		Enable: func() {
			events = append(events, "enable")
			allow = 1
		},
	})

	want := []string{"load-allocator", "clear-head", "free-all:original", "enable"}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %#v, want %#v", events, want)
	}
	if head != "" {
		t.Fatalf("final head = %q, want empty", head)
	}
	if allow != 1 {
		t.Fatalf("final allow = %d, want 1", allow)
	}
}

func TestRespawnReset4EC5B0ForwardsNilAllocator(t *testing.T) {
	var got *int
	cleared := false
	enabled := false
	RespawnReset4EC5B0(RespawnResetHooks4EC5B0[*int]{
		LoadAllocator: func() *int { return nil },
		ClearHead:     func() { cleared = true },
		FreeAll: func(allocator *int) {
			got = allocator
			if !cleared {
				t.Fatal("free-all ran before head clear")
			}
		},
		Enable: func() { enabled = true },
	})
	if got != nil {
		t.Fatalf("free-all allocator = %p, want nil", got)
	}
	if !enabled {
		t.Fatal("nil allocator did not reach enable")
	}
}

func TestRespawnReset4EC5B0FreeFaultSkipsEnable(t *testing.T) {
	stop := &struct{}{}
	var events []string
	var recovered any
	func() {
		defer func() { recovered = recover() }()
		RespawnReset4EC5B0(RespawnResetHooks4EC5B0[int]{
			LoadAllocator: func() int {
				events = append(events, "load-allocator")
				return 7
			},
			ClearHead: func() { events = append(events, "clear-head") },
			FreeAll: func(int) {
				events = append(events, "free-all")
				panic(stop)
			},
			Enable: func() { events = append(events, "enable") },
		})
	}()
	if recovered != stop {
		t.Fatalf("recovered = %#v, want sentinel", recovered)
	}
	if want := []string{"load-allocator", "clear-head", "free-all"}; !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %#v, want %#v", events, want)
	}
}
