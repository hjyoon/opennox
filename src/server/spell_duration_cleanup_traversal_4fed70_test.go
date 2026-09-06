package server

import (
	"fmt"
	"reflect"
	"testing"
)

const (
	spellDurationCleanupRecordA4FED70 = uint64(0x100000101)
	spellDurationCleanupRecordB4FED70 = uint64(0x200000202)
	spellDurationCleanupRecordC4FED70 = uint64(0x300000303)
)

type spellDurationCleanupRecordState4FED70 struct {
	flags byte
	next  uint64
}

type spellDurationCleanupWorld4FED70 struct {
	events    []string
	after     map[string]func()
	faultAt   int
	head      uint64
	records   map[uint64]*spellDurationCleanupRecordState4FED70
	destroyed []uint64
}

func newSpellDurationCleanupWorld4FED70() *spellDurationCleanupWorld4FED70 {
	return &spellDurationCleanupWorld4FED70{
		after: make(map[string]func()),
		head:  spellDurationCleanupRecordA4FED70,
		records: map[uint64]*spellDurationCleanupRecordState4FED70{
			spellDurationCleanupRecordA4FED70: {
				flags: 0x81,
				next:  spellDurationCleanupRecordB4FED70,
			},
			spellDurationCleanupRecordB4FED70: {
				flags: 0x80,
				next:  spellDurationCleanupRecordC4FED70,
			},
			spellDurationCleanupRecordC4FED70: {
				flags: 0x03,
			},
		},
	}
}

func (w *spellDurationCleanupWorld4FED70) observe(event string) {
	w.events = append(w.events, event)
	if w.faultAt != 0 && len(w.events) == w.faultAt {
		panic(event)
	}
	if after := w.after[event]; after != nil {
		after()
	}
}

func (w *spellDurationCleanupWorld4FED70) hooks() SpellDurationCleanupTraversalHooks4FED70[uint64] {
	return SpellDurationCleanupTraversalHooks4FED70[uint64]{
		LoadFirst: func() uint64 {
			value := w.head
			w.observe(fmt.Sprintf("first:%x", value))
			return value
		},
		LoadFlagsLowByte: func(record uint64) byte {
			value := w.records[record].flags
			w.observe(fmt.Sprintf("flags:%x:%02x", record, value))
			return value
		},
		LoadNext: func(record uint64) uint64 {
			value := w.records[record].next
			w.observe(fmt.Sprintf("next:%x:%x", record, value))
			return value
		},
		Destroy: func(record uint64) {
			w.destroyed = append(w.destroyed, record)
			w.observe(fmt.Sprintf("destroy:%x", record))
		},
	}
}

func TestSpellDurationCleanupTraversal4FED70ExactTraceAndLowBit(t *testing.T) {
	w := newSpellDurationCleanupWorld4FED70()

	SpellDurationCleanupTraversal4FED70(w.hooks())

	wantEvents := []string{
		"first:100000101",
		"flags:100000101:81", "next:100000101:200000202", "destroy:100000101",
		"flags:200000202:80", "next:200000202:300000303",
		"flags:300000303:03", "next:300000303:0", "destroy:300000303",
	}
	if !reflect.DeepEqual(w.events, wantEvents) {
		t.Fatalf("trace mismatch\n got: %q\nwant: %q", w.events, wantEvents)
	}
	wantDestroyed := []uint64{spellDurationCleanupRecordA4FED70, spellDurationCleanupRecordC4FED70}
	if !reflect.DeepEqual(w.destroyed, wantDestroyed) {
		t.Fatalf("destroyed = %x, want %x", w.destroyed, wantDestroyed)
	}
}

func TestSpellDurationCleanupTraversal4FED70FlagsBeforeNextAndSavedNextBeforeDestroy(t *testing.T) {
	w := newSpellDurationCleanupWorld4FED70()
	w.after["flags:100000101:81"] = func() {
		w.records[spellDurationCleanupRecordA4FED70].flags = 0x80
		w.records[spellDurationCleanupRecordA4FED70].next = spellDurationCleanupRecordC4FED70
	}
	w.after["next:100000101:300000303"] = func() {
		w.records[spellDurationCleanupRecordA4FED70].next = spellDurationCleanupRecordB4FED70
	}
	w.after["destroy:100000101"] = func() {
		w.head = spellDurationCleanupRecordB4FED70
		w.records[spellDurationCleanupRecordA4FED70].next = 0
		delete(w.records, spellDurationCleanupRecordA4FED70)
	}

	SpellDurationCleanupTraversal4FED70(w.hooks())

	want := []string{
		"first:100000101",
		"flags:100000101:81", "next:100000101:300000303", "destroy:100000101",
		"flags:300000303:03", "next:300000303:0", "destroy:300000303",
	}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("trace mismatch\n got: %q\nwant: %q", w.events, want)
	}
	if !reflect.DeepEqual(w.destroyed, []uint64{spellDurationCleanupRecordA4FED70, spellDurationCleanupRecordC4FED70}) {
		t.Fatalf("destroyed = %x, want cached-flag A then C", w.destroyed)
	}
}

func TestSpellDurationCleanupTraversal4FED70EmptyStopsAfterFirstLoad(t *testing.T) {
	w := newSpellDurationCleanupWorld4FED70()
	w.head = 0

	SpellDurationCleanupTraversal4FED70(w.hooks())

	if !reflect.DeepEqual(w.events, []string{"first:0"}) {
		t.Fatalf("events = %q, want [first:0]", w.events)
	}
}

func TestSpellDurationCleanupTraversal4FED70DoesNotAddCycleGuard(t *testing.T) {
	w := newSpellDurationCleanupWorld4FED70()
	w.records[spellDurationCleanupRecordA4FED70].flags = 0
	w.records[spellDurationCleanupRecordA4FED70].next = spellDurationCleanupRecordA4FED70
	w.after["next:100000101:100000101"] = func() {
		if len(w.events) >= 7 {
			panic(w)
		}
	}

	var recovered any
	func() {
		defer func() { recovered = recover() }()
		SpellDurationCleanupTraversal4FED70(w.hooks())
	}()
	if recovered != w {
		t.Fatalf("recovered = %#v, want third saved-successor load sentinel", recovered)
	}
	want := []string{
		"first:100000101",
		"flags:100000101:00", "next:100000101:100000101",
		"flags:100000101:00", "next:100000101:100000101",
		"flags:100000101:00", "next:100000101:100000101",
	}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %q, want repeated cycle prefix %q", w.events, want)
	}
}

func TestSpellDurationCleanupTraversal4FED70FaultPrefixes(t *testing.T) {
	baseline := newSpellDurationCleanupWorld4FED70()
	SpellDurationCleanupTraversal4FED70(baseline.hooks())
	want := append([]string(nil), baseline.events...)

	for faultAt := 1; faultAt <= len(want); faultAt++ {
		t.Run(fmt.Sprintf("fault-%d", faultAt), func(t *testing.T) {
			w := newSpellDurationCleanupWorld4FED70()
			w.faultAt = faultAt
			var recovered any
			func() {
				defer func() { recovered = recover() }()
				SpellDurationCleanupTraversal4FED70(w.hooks())
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
