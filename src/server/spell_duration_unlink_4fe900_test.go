package server

import (
	"fmt"
	"reflect"
	"testing"
)

const (
	spellDurationUnlinkTestRecord4FE900 = uint64(0x100000101)
	spellDurationUnlinkTestPrevA4FE900  = uint64(0x200000202)
	spellDurationUnlinkTestNextA4FE900  = uint64(0x300000303)
	spellDurationUnlinkTestPrevB4FE900  = uint64(0x400000404)
	spellDurationUnlinkTestNextB4FE900  = uint64(0x500000505)
)

type spellDurationUnlinkTestRecordState4FE900 struct {
	prev uint64
	next uint64
}

type spellDurationUnlinkTestWorld4FE900 struct {
	events  []string
	counts  map[string]int
	after   map[string]func()
	faultAt int
	head    uint64
	records map[uint64]*spellDurationUnlinkTestRecordState4FE900
}

func newSpellDurationUnlinkTestWorld4FE900() *spellDurationUnlinkTestWorld4FE900 {
	w := &spellDurationUnlinkTestWorld4FE900{
		counts: make(map[string]int),
		after:  make(map[string]func()),
		head:   spellDurationUnlinkTestRecord4FE900,
		records: map[uint64]*spellDurationUnlinkTestRecordState4FE900{
			spellDurationUnlinkTestRecord4FE900: {
				prev: spellDurationUnlinkTestPrevA4FE900,
				next: spellDurationUnlinkTestNextA4FE900,
			},
			spellDurationUnlinkTestPrevA4FE900: {next: spellDurationUnlinkTestRecord4FE900},
			spellDurationUnlinkTestNextA4FE900: {prev: spellDurationUnlinkTestRecord4FE900},
			spellDurationUnlinkTestPrevB4FE900: {},
			spellDurationUnlinkTestNextB4FE900: {},
		},
	}
	return w
}

func (w *spellDurationUnlinkTestWorld4FE900) record(id uint64) *spellDurationUnlinkTestRecordState4FE900 {
	if w.records[id] == nil {
		w.records[id] = &spellDurationUnlinkTestRecordState4FE900{}
	}
	return w.records[id]
}

func (w *spellDurationUnlinkTestWorld4FE900) observe(base string) {
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

func (w *spellDurationUnlinkTestWorld4FE900) hooks() spellDurationUnlinkHooks4FE900[uint64] {
	return spellDurationUnlinkHooks4FE900[uint64]{
		loadPrev: func(record uint64) uint64 {
			value := w.record(record).prev
			w.observe(fmt.Sprintf("prev:%x", record))
			return value
		},
		loadNext: func(record uint64) uint64 {
			value := w.record(record).next
			w.observe(fmt.Sprintf("next:%x", record))
			return value
		},
		storeNext: func(record, next uint64) {
			w.record(record).next = next
			w.observe(fmt.Sprintf("store-next:%x:%x", record, next))
		},
		storeHead: func(record uint64) {
			w.head = record
			w.observe(fmt.Sprintf("store-head:%x", record))
		},
		storePrev: func(record, prev uint64) {
			w.record(record).prev = prev
			w.observe(fmt.Sprintf("store-prev:%x:%x", record, prev))
		},
	}
}

func TestSpellDurationUnlink4FE900InteriorExactTraceAndLiveReloads(t *testing.T) {
	w := newSpellDurationUnlinkTestWorld4FE900()
	w.after["store-next:200000202:300000303#1"] = func() {
		w.record(spellDurationUnlinkTestRecord4FE900).next = spellDurationUnlinkTestNextB4FE900
	}
	w.after["next:100000101#2"] = func() {
		w.record(spellDurationUnlinkTestRecord4FE900).prev = spellDurationUnlinkTestPrevB4FE900
	}

	spellDurationUnlink4FE900(spellDurationUnlinkTestRecord4FE900, w.hooks())

	want := []string{
		"prev:100000101#1",
		"next:100000101#1",
		"store-next:200000202:300000303#1",
		"next:100000101#2",
		"prev:100000101#2",
		"store-prev:500000505:400000404#1",
	}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("trace mismatch\n got: %q\nwant: %q", w.events, want)
	}
	if got := w.record(spellDurationUnlinkTestPrevA4FE900).next; got != spellDurationUnlinkTestNextA4FE900 {
		t.Fatalf("original predecessor next = %#x, want %#x", got, spellDurationUnlinkTestNextA4FE900)
	}
	if got := w.record(spellDurationUnlinkTestNextB4FE900).prev; got != spellDurationUnlinkTestPrevB4FE900 {
		t.Fatalf("live successor prev = %#x, want %#x", got, spellDurationUnlinkTestPrevB4FE900)
	}
	if got := w.record(spellDurationUnlinkTestNextA4FE900).prev; got != spellDurationUnlinkTestRecord4FE900 {
		t.Fatalf("original successor prev = %#x, want unchanged %#x", got, spellDurationUnlinkTestRecord4FE900)
	}
}

func TestSpellDurationUnlink4FE900HeadExactTraceAndLiveReloads(t *testing.T) {
	w := newSpellDurationUnlinkTestWorld4FE900()
	w.record(spellDurationUnlinkTestRecord4FE900).prev = 0
	w.after["store-head:300000303#1"] = func() {
		w.record(spellDurationUnlinkTestRecord4FE900).next = spellDurationUnlinkTestNextB4FE900
	}
	w.after["next:100000101#2"] = func() {
		w.record(spellDurationUnlinkTestRecord4FE900).prev = spellDurationUnlinkTestPrevB4FE900
	}

	spellDurationUnlink4FE900(spellDurationUnlinkTestRecord4FE900, w.hooks())

	want := []string{
		"prev:100000101#1",
		"next:100000101#1",
		"store-head:300000303#1",
		"next:100000101#2",
		"prev:100000101#2",
		"store-prev:500000505:400000404#1",
	}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("trace mismatch\n got: %q\nwant: %q", w.events, want)
	}
	if w.head != spellDurationUnlinkTestNextA4FE900 {
		t.Fatalf("head = %#x, want first loaded successor %#x", w.head, spellDurationUnlinkTestNextA4FE900)
	}
	if got := w.record(spellDurationUnlinkTestNextB4FE900).prev; got != spellDurationUnlinkTestPrevB4FE900 {
		t.Fatalf("live successor prev = %#x, want live predecessor %#x", got, spellDurationUnlinkTestPrevB4FE900)
	}
}

func TestSpellDurationUnlink4FE900TailAndSingletonStopAtLiveNilNext(t *testing.T) {
	for _, tc := range []struct {
		name string
		prev uint64
		want []string
	}{
		{
			name: "tail",
			prev: spellDurationUnlinkTestPrevA4FE900,
			want: []string{
				"prev:100000101#1", "next:100000101#1",
				"store-next:200000202:0#1", "next:100000101#2",
			},
		},
		{
			name: "singleton",
			want: []string{
				"prev:100000101#1", "next:100000101#1",
				"store-head:0#1", "next:100000101#2",
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			w := newSpellDurationUnlinkTestWorld4FE900()
			record := w.record(spellDurationUnlinkTestRecord4FE900)
			record.prev = tc.prev
			record.next = 0

			spellDurationUnlink4FE900(spellDurationUnlinkTestRecord4FE900, w.hooks())

			if !reflect.DeepEqual(w.events, tc.want) {
				t.Fatalf("events = %q, want %q", w.events, tc.want)
			}
		})
	}
}

func TestSpellDurationUnlink4FE900DoesNotGuardNilRecord(t *testing.T) {
	w := newSpellDurationUnlinkTestWorld4FE900()
	spellDurationUnlink4FE900(uint64(0), w.hooks())
	if len(w.events) == 0 || w.events[0] != "prev:0#1" {
		t.Fatalf("events = %q, want nil record dereferenced first", w.events)
	}
}

func TestSpellDurationUnlink4FE900FaultPrefixes(t *testing.T) {
	for _, head := range []bool{false, true} {
		baseline := newSpellDurationUnlinkTestWorld4FE900()
		if head {
			baseline.record(spellDurationUnlinkTestRecord4FE900).prev = 0
		}
		spellDurationUnlink4FE900(spellDurationUnlinkTestRecord4FE900, baseline.hooks())
		want := append([]string(nil), baseline.events...)

		for faultAt := 1; faultAt <= len(want); faultAt++ {
			t.Run(fmt.Sprintf("head-%t/fault-%d", head, faultAt), func(t *testing.T) {
				w := newSpellDurationUnlinkTestWorld4FE900()
				if head {
					w.record(spellDurationUnlinkTestRecord4FE900).prev = 0
				}
				w.faultAt = faultAt
				var recovered any
				func() {
					defer func() { recovered = recover() }()
					spellDurationUnlink4FE900(spellDurationUnlinkTestRecord4FE900, w.hooks())
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
