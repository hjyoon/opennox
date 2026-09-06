package server

import (
	"fmt"
	"reflect"
	"testing"
)

const (
	spellDurationInsertTestRecord4FED40 = uint64(0x100000101)
	spellDurationInsertTestHeadA4FED40  = uint64(0x200000202)
	spellDurationInsertTestHeadB4FED40  = uint64(0x300000303)
	spellDurationInsertTestPrev4FED40   = uint64(0x400000404)
)

type spellDurationInsertTestRecordState4FED40 struct {
	prev uint64
	next uint64
}

type spellDurationInsertTestWorld4FED40 struct {
	events    []string
	counts    map[string]int
	after     map[string]func()
	faultAt   int
	head      uint64
	recordArg uint64
	records   map[uint64]*spellDurationInsertTestRecordState4FED40
}

func newSpellDurationInsertTestWorld4FED40(head uint64) *spellDurationInsertTestWorld4FED40 {
	return &spellDurationInsertTestWorld4FED40{
		counts:    make(map[string]int),
		after:     make(map[string]func()),
		head:      head,
		recordArg: spellDurationInsertTestRecord4FED40,
		records: map[uint64]*spellDurationInsertTestRecordState4FED40{
			spellDurationInsertTestRecord4FED40: {
				prev: spellDurationInsertTestPrev4FED40,
				next: spellDurationInsertTestPrev4FED40,
			},
			spellDurationInsertTestHeadA4FED40: {},
			spellDurationInsertTestHeadB4FED40: {},
		},
	}
}

func (w *spellDurationInsertTestWorld4FED40) record(id uint64) *spellDurationInsertTestRecordState4FED40 {
	if w.records[id] == nil {
		w.records[id] = &spellDurationInsertTestRecordState4FED40{}
	}
	return w.records[id]
}

func (w *spellDurationInsertTestWorld4FED40) observe(base string) {
	w.counts[base]++
	event := fmt.Sprintf("%s#%d", base, w.counts[base])
	w.events = append(w.events, event)
	if w.faultAt != 0 && len(w.events) == w.faultAt {
		panic(event)
	}
	if after := w.after[event]; after != nil {
		after()
	}
}

func (w *spellDurationInsertTestWorld4FED40) hooks() spellDurationInsertHooks4FED40[uint64] {
	return spellDurationInsertHooks4FED40[uint64]{
		loadHead: func() uint64 {
			value := w.head
			w.observe("head")
			return value
		},
		loadRecordArg: func() uint64 {
			value := w.recordArg
			w.observe("record")
			return value
		},
		storePrev: func(record, prev uint64) {
			w.record(record).prev = prev
			w.observe(fmt.Sprintf("store-prev:%x:%x", record, prev))
		},
		storeNext: func(record, next uint64) {
			w.record(record).next = next
			w.observe(fmt.Sprintf("store-next:%x:%x", record, next))
		},
		storeHead: func(record uint64) {
			w.head = record
			w.observe(fmt.Sprintf("store-head:%x", record))
		},
	}
}

func TestSpellDurationInsert4FED40ExactTraceAndLiveHeadReload(t *testing.T) {
	w := newSpellDurationInsertTestWorld4FED40(spellDurationInsertTestHeadA4FED40)
	w.after["store-prev:200000202:100000101#1"] = func() {
		w.head = spellDurationInsertTestHeadB4FED40
	}

	spellDurationInsert4FED40(w.hooks())

	want := []string{
		"head#1",
		"record#1",
		"store-prev:200000202:100000101#1",
		"store-prev:100000101:0#1",
		"head#2",
		"store-next:100000101:300000303#1",
		"store-head:100000101#1",
	}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("trace mismatch\n got: %q\nwant: %q", w.events, want)
	}
	if got := w.record(spellDurationInsertTestHeadA4FED40).prev; got != spellDurationInsertTestRecord4FED40 {
		t.Fatalf("cached old head Prev = %#x, want record %#x", got, spellDurationInsertTestRecord4FED40)
	}
	record := w.record(spellDurationInsertTestRecord4FED40)
	if record.prev != 0 || record.next != spellDurationInsertTestHeadB4FED40 || w.head != spellDurationInsertTestRecord4FED40 {
		t.Fatalf("record/head = Prev %#x Next %#x head %#x, want 0/%#x/%#x",
			record.prev, record.next, w.head,
			spellDurationInsertTestHeadB4FED40, spellDurationInsertTestRecord4FED40)
	}
	if got := w.record(spellDurationInsertTestHeadB4FED40).prev; got != 0 {
		t.Fatalf("live replacement head Prev = %#x, want unchanged zero", got)
	}
}

func TestSpellDurationInsert4FED40EmptyExactTrace(t *testing.T) {
	w := newSpellDurationInsertTestWorld4FED40(0)
	spellDurationInsert4FED40(w.hooks())

	want := []string{
		"head#1",
		"record#1",
		"store-prev:100000101:0#1",
		"head#2",
		"store-next:100000101:0#1",
		"store-head:100000101#1",
	}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %q, want %q", w.events, want)
	}
	record := w.record(spellDurationInsertTestRecord4FED40)
	if record.prev != 0 || record.next != 0 || w.head != spellDurationInsertTestRecord4FED40 {
		t.Fatalf("record/head = Prev %#x Next %#x head %#x, want 0/0/%#x",
			record.prev, record.next, w.head, spellDurationInsertTestRecord4FED40)
	}
}

func TestSpellDurationInsert4FED40DoesNotGuardNilRecord(t *testing.T) {
	w := newSpellDurationInsertTestWorld4FED40(spellDurationInsertTestHeadA4FED40)
	w.recordArg = 0
	stop := &struct{}{}
	hooks := w.hooks()
	storePrev := hooks.storePrev
	hooks.storePrev = func(record, prev uint64) {
		storePrev(record, prev)
		if record == 0 {
			panic(stop)
		}
	}

	var recovered any
	func() {
		defer func() { recovered = recover() }()
		spellDurationInsert4FED40(hooks)
	}()
	if recovered != stop {
		t.Fatalf("recovered = %#v, want nil-record store sentinel", recovered)
	}
	want := []string{
		"head#1",
		"record#1",
		"store-prev:200000202:0#1",
		"store-prev:0:0#1",
	}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %q, want fault prefix %q", w.events, want)
	}
	if got := w.record(spellDurationInsertTestHeadA4FED40).prev; got != 0 {
		t.Fatalf("old head Prev = %#x, want nil record written before fault", got)
	}
}

func TestSpellDurationInsert4FED40FaultPrefixes(t *testing.T) {
	for _, head := range []uint64{0, spellDurationInsertTestHeadA4FED40} {
		baseline := newSpellDurationInsertTestWorld4FED40(head)
		spellDurationInsert4FED40(baseline.hooks())
		want := append([]string(nil), baseline.events...)

		for faultAt := 1; faultAt <= len(want); faultAt++ {
			t.Run(fmt.Sprintf("head-%x/fault-%d", head, faultAt), func(t *testing.T) {
				w := newSpellDurationInsertTestWorld4FED40(head)
				w.faultAt = faultAt
				var recovered any
				func() {
					defer func() { recovered = recover() }()
					spellDurationInsert4FED40(w.hooks())
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
