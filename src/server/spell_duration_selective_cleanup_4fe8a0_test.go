package server

import (
	"fmt"
	"reflect"
	"testing"
)

const (
	spellDurationTestAllocator4FE8A0 = uint64(0x100000101)
	spellDurationTestRecordA4FE8A0   = uint64(0x200000202)
	spellDurationTestRecordB4FE8A0   = uint64(0x300000303)
	spellDurationTestRecordC4FE8A0   = uint64(0x400000404)
	spellDurationTestTargetB4FE8A0   = uint64(0x500000505)
	spellDurationTestTargetC4FE8A0   = uint64(0x600000606)
)

type spellDurationSelectiveRecordState4FE8A0 struct {
	target uint64
	next   uint64
}

type spellDurationSelectiveTestWorld4FE8A0 struct {
	events    []string
	after     map[string]func()
	faultAt   int
	head      uint64
	allocator uint64
	records   map[uint64]*spellDurationSelectiveRecordState4FE8A0
	classes   map[uint64]uint8
	unlinked  []uint64
	freed     []uint64
}

func newSpellDurationSelectiveTestWorld4FE8A0() *spellDurationSelectiveTestWorld4FE8A0 {
	return &spellDurationSelectiveTestWorld4FE8A0{
		after:     make(map[string]func()),
		head:      spellDurationTestRecordA4FE8A0,
		allocator: spellDurationTestAllocator4FE8A0,
		records: map[uint64]*spellDurationSelectiveRecordState4FE8A0{
			spellDurationTestRecordA4FE8A0: {next: spellDurationTestRecordB4FE8A0},
			spellDurationTestRecordB4FE8A0: {target: spellDurationTestTargetB4FE8A0, next: spellDurationTestRecordC4FE8A0},
			spellDurationTestRecordC4FE8A0: {target: spellDurationTestTargetC4FE8A0},
		},
		classes: map[uint64]uint8{
			spellDurationTestTargetB4FE8A0: spellDurationPlayerClassLowByte4FE8A0 | 0x80,
			spellDurationTestTargetC4FE8A0: 0x40,
		},
	}
}

func (w *spellDurationSelectiveTestWorld4FE8A0) observe(event string) {
	w.events = append(w.events, event)
	if w.faultAt != 0 && len(w.events) == w.faultAt {
		panic(event)
	}
	if after := w.after[event]; after != nil {
		after()
	}
}

func (w *spellDurationSelectiveTestWorld4FE8A0) hooks() spellDurationSelectiveCleanupHooks4FE8A0[uint64, uint64, uint64] {
	return spellDurationSelectiveCleanupHooks4FE8A0[uint64, uint64, uint64]{
		loadAllocator: func() uint64 {
			value := w.allocator
			w.observe("load-allocator")
			return value
		},
		freeAllObjects: func(allocator uint64) {
			w.observe(fmt.Sprintf("free-all:%x", allocator))
		},
		clearList: func() {
			w.head = 0
			w.observe("clear-list")
		},
		loadList: func() uint64 {
			value := w.head
			w.observe("load-list")
			return value
		},
		loadTarget: func(record uint64) uint64 {
			value := w.records[record].target
			w.observe(fmt.Sprintf("target:%x", record))
			return value
		},
		loadNext: func(record uint64) uint64 {
			value := w.records[record].next
			w.observe(fmt.Sprintf("next:%x", record))
			return value
		},
		loadClassLowByte: func(object uint64) uint8 {
			value := w.classes[object]
			w.observe(fmt.Sprintf("class:%x", object))
			return value
		},
		unlink: func(record uint64) {
			w.unlinked = append(w.unlinked, record)
			w.observe(fmt.Sprintf("unlink:%x", record))
		},
		freeRecursive: func(record uint64) {
			w.freed = append(w.freed, record)
			w.observe(fmt.Sprintf("free:%x", record))
		},
	}
}

func TestSpellDurationSelectiveCleanup4FE8A0ZeroModeOrderAndSnapshot(t *testing.T) {
	w := newSpellDurationSelectiveTestWorld4FE8A0()
	w.after["load-allocator"] = func() { w.allocator = 0xabcdef0123456789 }
	w.after["free-all:100000101"] = func() {
		if w.head != spellDurationTestRecordA4FE8A0 {
			t.Fatalf("head cleared before allocator callback: %#x", w.head)
		}
	}

	spellDurationSelectiveCleanup4FE8A0(int32(0), w.hooks())

	want := []string{"load-allocator", "free-all:100000101", "clear-list"}
	if !reflect.DeepEqual(w.events, want) || w.head != 0 {
		t.Fatalf("events/head = %q/%#x, want %q/0", w.events, w.head, want)
	}
}

func TestSpellDurationSelectiveCleanup4FE8A0NonzeroExactTraceAndCachedNext(t *testing.T) {
	w := newSpellDurationSelectiveTestWorld4FE8A0()
	w.after["next:200000202"] = func() { w.records[spellDurationTestRecordA4FE8A0].next = 0 }
	w.after["unlink:200000202"] = func() { w.head = 0 }
	w.after["next:300000303"] = func() { w.records[spellDurationTestRecordB4FE8A0].next = 0 }

	spellDurationSelectiveCleanup4FE8A0(int32(-1), w.hooks())

	wantEvents := []string{
		"load-list",
		"target:200000202", "next:200000202", "unlink:200000202", "free:200000202",
		"target:300000303", "next:300000303", "class:500000505",
		"target:400000404", "next:400000404", "class:600000606", "unlink:400000404", "free:400000404",
	}
	if !reflect.DeepEqual(w.events, wantEvents) {
		t.Fatalf("events mismatch\n got: %q\nwant: %q", w.events, wantEvents)
	}
	wantRemoved := []uint64{spellDurationTestRecordA4FE8A0, spellDurationTestRecordC4FE8A0}
	if !reflect.DeepEqual(w.unlinked, wantRemoved) || !reflect.DeepEqual(w.freed, wantRemoved) {
		t.Fatalf("unlinked/freed = %x/%x, want %x", w.unlinked, w.freed, wantRemoved)
	}
}

func TestSpellDurationSelectiveCleanup4FE8A0EmptyListStopsBeforeRecordReads(t *testing.T) {
	w := newSpellDurationSelectiveTestWorld4FE8A0()
	w.head = 0
	spellDurationSelectiveCleanup4FE8A0(int32(1), w.hooks())
	if !reflect.DeepEqual(w.events, []string{"load-list"}) {
		t.Fatalf("events = %q, want only list load", w.events)
	}
}

func TestSpellDurationSelectiveCleanup4FE8A0PlayerUsesOnlyClassLowByte(t *testing.T) {
	for _, tc := range []struct {
		name    string
		class   uint8
		removed bool
	}{
		{name: "player-bit", class: 0x04},
		{name: "player-bit-with-other-bits", class: 0xfc},
		{name: "different-bit", class: 0x40, removed: true},
		{name: "zero", removed: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			w := newSpellDurationSelectiveTestWorld4FE8A0()
			w.head = spellDurationTestRecordB4FE8A0
			w.records[spellDurationTestRecordB4FE8A0].next = 0
			w.classes[spellDurationTestTargetB4FE8A0] = tc.class
			spellDurationSelectiveCleanup4FE8A0(int32(0x7fffffff), w.hooks())
			if got := len(w.freed) != 0; got != tc.removed {
				t.Fatalf("class low byte %#02x removed = %t, want %t", tc.class, got, tc.removed)
			}
		})
	}
}

func TestSpellDurationSelectiveCleanup4FE8A0FaultPrefixes(t *testing.T) {
	for _, mode := range []int32{0, 1} {
		baseline := newSpellDurationSelectiveTestWorld4FE8A0()
		spellDurationSelectiveCleanup4FE8A0(mode, baseline.hooks())
		want := append([]string(nil), baseline.events...)
		for faultAt := 1; faultAt <= len(want); faultAt++ {
			t.Run(fmt.Sprintf("mode-%d/fault-%d", mode, faultAt), func(t *testing.T) {
				w := newSpellDurationSelectiveTestWorld4FE8A0()
				w.faultAt = faultAt
				var recovered any
				func() {
					defer func() { recovered = recover() }()
					spellDurationSelectiveCleanup4FE8A0(mode, w.hooks())
				}()
				if recovered == nil {
					t.Fatal("fault sentinel was not recovered")
				}
				if prefix := want[:faultAt]; !reflect.DeepEqual(w.events, prefix) {
					t.Fatalf("events = %q, want fault prefix %q", w.events, prefix)
				}
			})
		}
	}
}
