package legacy

import (
	"fmt"
	"reflect"
	"testing"
)

const (
	questJournalDeleteTestEntry500790 = uint64(0x100000101)
	questJournalDeleteTestPrevA500790 = uint64(0x200000202)
	questJournalDeleteTestNextA500790 = uint64(0x300000303)
	questJournalDeleteTestPrevB500790 = uint64(0x400000404)
	questJournalDeleteTestNextB500790 = uint64(0x500000505)
	questJournalDeleteTestNextC500790 = uint64(0x600000606)
	questJournalDeleteTestOther500790 = uint64(0x700000707)
)

type questJournalDeleteTestEntryState500790 struct {
	prev uint64
	next uint64
}

type questJournalDeleteTestWorld500790 struct {
	events  []string
	counts  map[string]int
	after   map[string]func()
	faultAt int
	head    uint64
	freed   map[uint64]int
	entries map[uint64]*questJournalDeleteTestEntryState500790
}

func newQuestJournalDeleteTestWorld500790() *questJournalDeleteTestWorld500790 {
	return &questJournalDeleteTestWorld500790{
		counts: make(map[string]int),
		after:  make(map[string]func()),
		head:   questJournalDeleteTestOther500790,
		freed:  make(map[uint64]int),
		entries: map[uint64]*questJournalDeleteTestEntryState500790{
			questJournalDeleteTestEntry500790: {
				prev: questJournalDeleteTestPrevA500790,
				next: questJournalDeleteTestNextA500790,
			},
			questJournalDeleteTestPrevA500790: {
				next: questJournalDeleteTestEntry500790,
			},
			questJournalDeleteTestNextA500790: {
				prev: questJournalDeleteTestEntry500790,
			},
			questJournalDeleteTestPrevB500790: {},
			questJournalDeleteTestNextB500790: {},
			questJournalDeleteTestNextC500790: {},
		},
	}
}

func (w *questJournalDeleteTestWorld500790) entry(id uint64) *questJournalDeleteTestEntryState500790 {
	if w.entries[id] == nil {
		w.entries[id] = &questJournalDeleteTestEntryState500790{}
	}
	return w.entries[id]
}

func (w *questJournalDeleteTestWorld500790) observe(base string) {
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

func (w *questJournalDeleteTestWorld500790) hooks() questJournalDeleteEntryHooks500790[uint64] {
	return questJournalDeleteEntryHooks500790[uint64]{
		loadPrev: func(entry uint64) uint64 {
			value := w.entry(entry).prev
			w.observe(fmt.Sprintf("load-prev:%x=%x", entry, value))
			return value
		},
		loadNext: func(entry uint64) uint64 {
			value := w.entry(entry).next
			w.observe(fmt.Sprintf("load-next:%x=%x", entry, value))
			return value
		},
		storeNext: func(entry, next uint64) {
			w.entry(entry).next = next
			w.observe(fmt.Sprintf("store-next:%x=%x", entry, next))
		},
		storePrev: func(entry, prev uint64) {
			w.entry(entry).prev = prev
			w.observe(fmt.Sprintf("store-prev:%x=%x", entry, prev))
		},
		loadHead: func() uint64 {
			value := w.head
			w.observe(fmt.Sprintf("load-head=%x", value))
			return value
		},
		storeHead: func(entry uint64) {
			w.head = entry
			w.observe(fmt.Sprintf("store-head:%x", entry))
		},
		freeEntry: func(entry uint64) {
			w.freed[entry]++
			w.observe(fmt.Sprintf("free:%x", entry))
		},
	}
}

func TestQuestJournalDeleteEntry500790ExactTraceAndLiveReloads(t *testing.T) {
	w := newQuestJournalDeleteTestWorld500790()
	w.head = questJournalDeleteTestEntry500790
	w.after["store-next:200000202=300000303#1"] = func() {
		w.entry(questJournalDeleteTestEntry500790).next = questJournalDeleteTestNextB500790
	}
	w.after["load-next:100000101=500000505#1"] = func() {
		w.entry(questJournalDeleteTestEntry500790).prev = questJournalDeleteTestPrevB500790
	}
	w.after["load-head=100000101#1"] = func() {
		w.entry(questJournalDeleteTestEntry500790).next = questJournalDeleteTestNextC500790
	}

	questJournalDeleteEntryContract500790(questJournalDeleteTestEntry500790, w.hooks())

	want := []string{
		"load-prev:100000101=200000202#1",
		"load-next:100000101=300000303#1",
		"store-next:200000202=300000303#1",
		"load-next:100000101=500000505#1",
		"load-prev:100000101=400000404#1",
		"store-prev:500000505=400000404#1",
		"load-head=100000101#1",
		"load-next:100000101=600000606#1",
		"store-head:600000606#1",
		"free:100000101#1",
	}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("trace mismatch\n got: %q\nwant: %q", w.events, want)
	}
	if got := w.entry(questJournalDeleteTestPrevA500790).next; got != questJournalDeleteTestNextA500790 {
		t.Fatalf("original predecessor next = %#x, want %#x", got, questJournalDeleteTestNextA500790)
	}
	if got := w.entry(questJournalDeleteTestNextB500790).prev; got != questJournalDeleteTestPrevB500790 {
		t.Fatalf("live successor previous = %#x, want %#x", got, questJournalDeleteTestPrevB500790)
	}
	if got := w.entry(questJournalDeleteTestNextA500790).prev; got != questJournalDeleteTestEntry500790 {
		t.Fatalf("original successor previous = %#x, want unchanged %#x", got, questJournalDeleteTestEntry500790)
	}
	if w.head != questJournalDeleteTestNextC500790 {
		t.Fatalf("head = %#x, want final live next %#x", w.head, questJournalDeleteTestNextC500790)
	}
	removed := w.entry(questJournalDeleteTestEntry500790)
	if removed.prev != questJournalDeleteTestPrevB500790 || removed.next != questJournalDeleteTestNextC500790 {
		t.Fatalf("removed links = %#x/%#x, want untouched %#x/%#x", removed.prev, removed.next, questJournalDeleteTestPrevB500790, questJournalDeleteTestNextC500790)
	}
	if w.freed[questJournalDeleteTestEntry500790] != 1 {
		t.Fatalf("free count = %d, want 1", w.freed[questJournalDeleteTestEntry500790])
	}
}

func TestQuestJournalDeleteEntry500790HeadIdentityIsIndependentOfPrev(t *testing.T) {
	t.Run("nil-prev-but-not-head", func(t *testing.T) {
		w := newQuestJournalDeleteTestWorld500790()
		w.entry(questJournalDeleteTestEntry500790).prev = 0

		questJournalDeleteEntryContract500790(questJournalDeleteTestEntry500790, w.hooks())

		want := []string{
			"load-prev:100000101=0#1",
			"load-next:100000101=300000303#1",
			"load-prev:100000101=0#2",
			"store-prev:300000303=0#1",
			"load-head=700000707#1",
			"free:100000101#1",
		}
		if !reflect.DeepEqual(w.events, want) {
			t.Fatalf("events = %q, want %q", w.events, want)
		}
		if w.head != questJournalDeleteTestOther500790 {
			t.Fatalf("head = %#x, want unchanged %#x", w.head, questJournalDeleteTestOther500790)
		}
	})

	t.Run("non-nil-prev-but-is-head", func(t *testing.T) {
		w := newQuestJournalDeleteTestWorld500790()
		w.head = questJournalDeleteTestEntry500790

		questJournalDeleteEntryContract500790(questJournalDeleteTestEntry500790, w.hooks())

		if w.head != questJournalDeleteTestNextA500790 {
			t.Fatalf("head = %#x, want %#x", w.head, questJournalDeleteTestNextA500790)
		}
		wantTail := []string{
			"load-head=100000101#1",
			"load-next:100000101=300000303#3",
			"store-head:300000303#1",
			"free:100000101#1",
		}
		if got := w.events[len(w.events)-len(wantTail):]; !reflect.DeepEqual(got, wantTail) {
			t.Fatalf("tail events = %q, want %q", got, wantTail)
		}
	})
}

func TestQuestJournalDeleteEntry500790TailAndSingleton(t *testing.T) {
	for _, tc := range []struct {
		name string
		prev uint64
		head uint64
		want []string
	}{
		{
			name: "tail",
			prev: questJournalDeleteTestPrevA500790,
			head: questJournalDeleteTestOther500790,
			want: []string{
				"load-prev:100000101=200000202#1",
				"load-next:100000101=0#1",
				"store-next:200000202=0#1",
				"load-next:100000101=0#2",
				"load-head=700000707#1",
				"free:100000101#1",
			},
		},
		{
			name: "singleton",
			head: questJournalDeleteTestEntry500790,
			want: []string{
				"load-prev:100000101=0#1",
				"load-next:100000101=0#1",
				"load-head=100000101#1",
				"load-next:100000101=0#2",
				"store-head:0#1",
				"free:100000101#1",
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			w := newQuestJournalDeleteTestWorld500790()
			entry := w.entry(questJournalDeleteTestEntry500790)
			entry.prev = tc.prev
			entry.next = 0
			w.head = tc.head

			questJournalDeleteEntryContract500790(questJournalDeleteTestEntry500790, w.hooks())

			if !reflect.DeepEqual(w.events, tc.want) {
				t.Fatalf("events = %q, want %q", w.events, tc.want)
			}
			if w.freed[questJournalDeleteTestEntry500790] != 1 {
				t.Fatalf("free count = %d, want 1", w.freed[questJournalDeleteTestEntry500790])
			}
		})
	}
}

func TestQuestJournalDeleteEntry500790DoesNotGuardNilEntry(t *testing.T) {
	w := newQuestJournalDeleteTestWorld500790()
	w.faultAt = 1
	defer func() {
		if got := recover(); got != "load-prev:0=0#1" {
			t.Fatalf("panic = %v, want first nil-entry previous load", got)
		}
	}()
	questJournalDeleteEntryContract500790(uint64(0), w.hooks())
}

func TestQuestJournalDeleteEntry500790FaultPrefixes(t *testing.T) {
	builds := map[string]func() *questJournalDeleteTestWorld500790{
		"interior": func() *questJournalDeleteTestWorld500790 {
			return newQuestJournalDeleteTestWorld500790()
		},
		"head-with-prev": func() *questJournalDeleteTestWorld500790 {
			w := newQuestJournalDeleteTestWorld500790()
			w.head = questJournalDeleteTestEntry500790
			return w
		},
		"detached-nil-prev": func() *questJournalDeleteTestWorld500790 {
			w := newQuestJournalDeleteTestWorld500790()
			w.entry(questJournalDeleteTestEntry500790).prev = 0
			return w
		},
		"singleton": func() *questJournalDeleteTestWorld500790 {
			w := newQuestJournalDeleteTestWorld500790()
			entry := w.entry(questJournalDeleteTestEntry500790)
			entry.prev = 0
			entry.next = 0
			w.head = questJournalDeleteTestEntry500790
			return w
		},
	}

	for name, build := range builds {
		t.Run(name, func(t *testing.T) {
			baseline := build()
			questJournalDeleteEntryContract500790(questJournalDeleteTestEntry500790, baseline.hooks())
			want := append([]string(nil), baseline.events...)

			for faultAt := 1; faultAt <= len(want); faultAt++ {
				t.Run(fmt.Sprintf("event-%d", faultAt), func(t *testing.T) {
					w := build()
					w.faultAt = faultAt
					var recovered any
					func() {
						defer func() { recovered = recover() }()
						questJournalDeleteEntryContract500790(questJournalDeleteTestEntry500790, w.hooks())
					}()
					if recovered == nil {
						t.Fatal("fault sentinel was not recovered")
					}
					if prefix := want[:faultAt]; !reflect.DeepEqual(w.events, prefix) {
						t.Fatalf("events = %q, want fault prefix %q", w.events, prefix)
					}
				})
			}
		})
	}
}
