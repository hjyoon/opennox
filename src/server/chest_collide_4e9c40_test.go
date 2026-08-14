package server

import (
	"fmt"
	"math"
	"reflect"
	"testing"
)

func TestChestSilverKeyNameMatches4E9C40(t *testing.T) {
	tests := []struct {
		name string
		want bool
	}{
		{"SilverKey", true},
		{"SilverKey\x00trailing", true},
		{"SilverKeyX", false},
		{"SilverKe", false},
		{"silverKey", false},
	}
	for _, tc := range tests {
		if got := chestSilverKeyNameMatches4E9C40(tc.name); got != tc.want {
			t.Errorf("match(%q) = %v, want %v", tc.name, got, tc.want)
		}
	}
}

func TestChestCollide4E9C40SilverKeyAndLiveDeathOrder(t *testing.T) {
	classes := map[int]uint8{
		2: chestCollidePlayerClass4E9C40,
		3: 0,
		4: chestCollideKeyClass4E9C40,
		5: chestCollideKeyClass4E9C40,
	}
	names := map[int]string{4: "GoldKey", 5: "SilverKey\x00ignored"}
	next := map[int]int{3: 4, 4: 5}
	death := 77
	var events []string

	chestCollide4E9C40(1, 2, struct{ unread int }{unread: 99}, chestCollideHooks4E9C40[int, int]{
		loadClassLow: func(obj int) uint8 {
			events = append(events, fmt.Sprintf("class:%d", obj))
			return classes[obj]
		},
		loadFlags: func(obj int) uint32 {
			events = append(events, fmt.Sprintf("flags:%d", obj))
			return 0
		},
		gameFlagsCheck: func(flag uint32) int32 {
			events = append(events, fmt.Sprintf("flag:%#x", flag))
			return -1
		},
		loadSubclass: func(obj int) uint32 {
			events = append(events, fmt.Sprintf("subclass:%d", obj))
			return chestCollideLockedSubclass4E9C40
		},
		firstItem: func(obj int) int {
			events = append(events, fmt.Sprintf("first:%d", obj))
			return 3
		},
		loadTypeName: func(obj int) string {
			events = append(events, fmt.Sprintf("name:%d", obj))
			return names[obj]
		},
		nextItem: func(obj int) int {
			events = append(events, fmt.Sprintf("next:%d", obj))
			return next[obj]
		},
		delayedDelete: func(obj int) {
			events = append(events, fmt.Sprintf("delete:%d", obj))
			death = 88
		},
		audio: func(id uint32, obj int) {
			events = append(events, fmt.Sprintf("audio:%d:%d", id, obj))
		},
		loadDeath: func(obj int) int {
			events = append(events, fmt.Sprintf("death:%d", obj))
			return death
		},
		callDeath: func(callback, obj int) {
			events = append(events, fmt.Sprintf("call:%d:%d", callback, obj))
		},
		chestOpen: func(source, target int) {
			events = append(events, fmt.Sprintf("open:%d:%d", source, target))
		},
		dropAllItems: func(source int) {
			events = append(events, fmt.Sprintf("drop:%d", source))
		},
	})

	want := []string{
		"class:2", "flags:1", "flag:0x1000", "subclass:1", "first:2",
		"class:3", "next:3", "class:4", "name:4", "next:4",
		"class:5", "name:5", "delete:5", "audio:234:1", "death:1",
		"call:88:1", "open:1:2", "drop:1",
	}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %#v, want %#v", events, want)
	}
}

func TestChestCollide4E9C40UnlockedSkipsSubclassAndNilDeath(t *testing.T) {
	var events []string
	chestCollide4E9C40(1, 2, nil, chestCollideHooks4E9C40[int, int]{
		loadClassLow: func(obj int) uint8 {
			events = append(events, "class")
			return chestCollidePlayerClass4E9C40
		},
		loadFlags: func(int) uint32 {
			events = append(events, "flags")
			return 0
		},
		gameFlagsCheck: func(uint32) int32 {
			events = append(events, "flag")
			return 0
		},
		loadSubclass: func(int) uint32 {
			t.Fatal("subclass read outside Quest mode")
			return 0
		},
		loadDeath: func(int) int {
			events = append(events, "death")
			return 0
		},
		callDeath: func(int, int) {
			t.Fatal("nil Death callback invoked")
		},
		chestOpen: func(int, int) {
			events = append(events, "open")
		},
		dropAllItems: func(int) {
			events = append(events, "drop")
		},
	})
	want := []string{"class", "flags", "flag", "death", "open", "drop"}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %#v, want %#v", events, want)
	}
}

func TestChestCollide4E9C40LockedFeedbackBoundaryAndWrap(t *testing.T) {
	tests := []struct {
		name       string
		now        uint64
		last       uint64
		wantNotify bool
	}{
		{name: "exact boundary", now: 2500, last: 1000},
		{name: "one over", now: 2501, last: 1000, wantNotify: true},
		{name: "unsigned wrap", now: 100, last: math.MaxUint64 - 1400, wantNotify: true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var events []string
			tickCalls := 0
			chestCollide4E9C40(1, 2, nil, chestCollideHooks4E9C40[int, int]{
				loadClassLow: func(obj int) uint8 {
					if obj == 2 {
						return chestCollidePlayerClass4E9C40
					}
					return 0
				},
				loadFlags:      func(int) uint32 { return 0 },
				gameFlagsCheck: func(uint32) int32 { return 1 },
				loadSubclass:   func(int) uint32 { return chestCollideLockedSubclass4E9C40 },
				firstItem:      func(int) int { return 0 },
				ticks: func() uint64 {
					tickCalls++
					if tickCalls == 1 {
						events = append(events, fmt.Sprintf("ticks:%d", tc.now))
						return tc.now
					}
					events = append(events, "ticks:9000")
					return 9000
				},
				loadFeedbackTicks: func() uint64 {
					events = append(events, fmt.Sprintf("feedback:%d", tc.last))
					return tc.last
				},
				audio: func(id uint32, obj int) {
					events = append(events, fmt.Sprintf("audio:%d:%d", id, obj))
				},
				priorityMessage: func(obj int, message string) {
					events = append(events, fmt.Sprintf("message:%d:%s", obj, message))
				},
				storeFeedbackTicks: func(value uint64) {
					events = append(events, fmt.Sprintf("store:%d", value))
				},
			})

			want := []string{fmt.Sprintf("ticks:%d", tc.now), fmt.Sprintf("feedback:%d", tc.last)}
			if tc.wantNotify {
				want = append(want,
					"audio:1012:1",
					"message:2:objcoll.c:ChestLockedSilver",
					"ticks:9000",
					"store:9000",
				)
			}
			if !reflect.DeepEqual(events, want) {
				t.Fatalf("events = %#v, want %#v", events, want)
			}
		})
	}
}

func TestChestCollide4E9C40EarlyGatesDoNotReadSource(t *testing.T) {
	tests := []struct {
		name   string
		target int
		class  uint8
	}{
		{name: "nil target", target: 0},
		{name: "non player", target: 2, class: 0x80},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			chestCollide4E9C40(0, tc.target, nil, chestCollideHooks4E9C40[int, int]{
				loadClassLow: func(int) uint8 { return tc.class },
				loadFlags: func(int) uint32 {
					t.Fatal("source flags read after target rejection")
					return 0
				},
			})
		})
	}
}

func TestChestCollide4E9C40DestroyedStopsBeforeGameFlag(t *testing.T) {
	chestCollide4E9C40(1, 2, nil, chestCollideHooks4E9C40[int, int]{
		loadClassLow: func(int) uint8 { return chestCollidePlayerClass4E9C40 },
		loadFlags:    func(int) uint32 { return chestCollideDestroyedFlag4E9C40 },
		gameFlagsCheck: func(uint32) int32 {
			t.Fatal("game flag read for destroyed chest")
			return 0
		},
	})
}
