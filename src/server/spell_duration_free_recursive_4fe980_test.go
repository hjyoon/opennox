package server

import (
	"fmt"
	"reflect"
	"testing"
)

const (
	spellDurationFreeRoot4FE980       = uint64(0x100000101)
	spellDurationFreeSub108A4FE980    = uint64(0x200000202)
	spellDurationFreeSub108B4FE980    = uint64(0x300000303)
	spellDurationFreeGrandchild4FE980 = uint64(0x400000404)
	spellDurationFreeOldSub1044FE980  = uint64(0x500000505)
	spellDurationFreeLiveSub1044FE980 = uint64(0x600000606)
	spellDurationFreeDecoyNext4FE980  = uint64(0x700000707)
	spellDurationFreeAllocatorA4FE980 = uint64(0x800000808)
	spellDurationFreeAllocatorB4FE980 = uint64(0x900000909)
)

type spellDurationFreeRecordState4FE980 struct {
	sub108 uint64
	sub104 uint64
	next   uint64
}

type spellDurationFreeWorld4FE980 struct {
	events        []string
	faultAt       int
	allocator     uint64
	records       map[uint64]*spellDurationFreeRecordState4FE980
	freed         []uint64
	mutateOnFirst bool
}

func newSpellDurationFreeWorld4FE980() *spellDurationFreeWorld4FE980 {
	return &spellDurationFreeWorld4FE980{
		allocator: spellDurationFreeAllocatorA4FE980,
		records: map[uint64]*spellDurationFreeRecordState4FE980{
			spellDurationFreeRoot4FE980: {
				sub108: spellDurationFreeSub108A4FE980,
				sub104: spellDurationFreeOldSub1044FE980,
			},
			spellDurationFreeSub108A4FE980: {
				sub108: spellDurationFreeGrandchild4FE980,
				next:   spellDurationFreeSub108B4FE980,
			},
			spellDurationFreeSub108B4FE980:    {},
			spellDurationFreeGrandchild4FE980: {},
			spellDurationFreeOldSub1044FE980:  {},
			spellDurationFreeLiveSub1044FE980: {},
			spellDurationFreeDecoyNext4FE980:  {},
		},
	}
}

func (w *spellDurationFreeWorld4FE980) record(id uint64) *spellDurationFreeRecordState4FE980 {
	if w.records[id] == nil {
		w.records[id] = &spellDurationFreeRecordState4FE980{}
	}
	return w.records[id]
}

func (w *spellDurationFreeWorld4FE980) observe(event string) {
	w.events = append(w.events, event)
	if w.faultAt != 0 && len(w.events) == w.faultAt {
		panic(event)
	}
}

func (w *spellDurationFreeWorld4FE980) hooks() SpellDurationFreeRecursiveHooks4FE980[uint64, uint64] {
	return SpellDurationFreeRecursiveHooks4FE980[uint64, uint64]{
		LoadSub108: func(record uint64) uint64 {
			w.observe(fmt.Sprintf("sub108:%x", record))
			return w.record(record).sub108
		},
		LoadSub104: func(record uint64) uint64 {
			w.observe(fmt.Sprintf("sub104:%x", record))
			return w.record(record).sub104
		},
		LoadNext: func(record uint64) uint64 {
			w.observe(fmt.Sprintf("next:%x", record))
			return w.record(record).next
		},
		LoadAllocator: func() uint64 {
			w.observe(fmt.Sprintf("allocator:%x", w.allocator))
			return w.allocator
		},
		FreeObjectFirst: func(allocator, record uint64) {
			w.observe(fmt.Sprintf("free:%x:%x", allocator, record))
			w.freed = append(w.freed, record)
			if w.mutateOnFirst && record == spellDurationFreeGrandchild4FE980 {
				w.record(spellDurationFreeSub108A4FE980).next = spellDurationFreeDecoyNext4FE980
				w.record(spellDurationFreeRoot4FE980).sub104 = spellDurationFreeLiveSub1044FE980
				w.allocator = spellDurationFreeAllocatorB4FE980
			}
		},
	}
}

func TestSpellDurationFreeRecursive4FE980OrderSnapshotsAndLiveLoads(t *testing.T) {
	w := newSpellDurationFreeWorld4FE980()
	w.mutateOnFirst = true

	SpellDurationFreeRecursive4FE980(w.hooks(), spellDurationFreeRoot4FE980)

	wantEvents := []string{
		"sub108:100000101",
		"next:200000202",
		"sub108:200000202",
		"next:400000404",
		"sub108:400000404",
		"sub104:400000404",
		"allocator:800000808",
		"free:800000808:400000404",
		"sub104:200000202",
		"allocator:900000909",
		"free:900000909:200000202",
		"next:300000303",
		"sub108:300000303",
		"sub104:300000303",
		"allocator:900000909",
		"free:900000909:300000303",
		"sub104:100000101",
		"next:600000606",
		"sub108:600000606",
		"sub104:600000606",
		"allocator:900000909",
		"free:900000909:600000606",
		"allocator:900000909",
		"free:900000909:100000101",
	}
	if !reflect.DeepEqual(w.events, wantEvents) {
		t.Fatalf("trace mismatch\n got: %q\nwant: %q", w.events, wantEvents)
	}
	wantFreed := []uint64{
		spellDurationFreeGrandchild4FE980,
		spellDurationFreeSub108A4FE980,
		spellDurationFreeSub108B4FE980,
		spellDurationFreeLiveSub1044FE980,
		spellDurationFreeRoot4FE980,
	}
	if !reflect.DeepEqual(w.freed, wantFreed) {
		t.Fatalf("freed = %#x, want depth-first order %#x", w.freed, wantFreed)
	}
	for _, forbidden := range []uint64{spellDurationFreeOldSub1044FE980, spellDurationFreeDecoyNext4FE980} {
		for _, record := range w.freed {
			if record == forbidden {
				t.Fatalf("freed stale or post-snapshot record %#x", forbidden)
			}
		}
	}
}

func TestSpellDurationFreeRecursive4FE980LeafAndNilRecord(t *testing.T) {
	w := newSpellDurationFreeWorld4FE980()
	const leaf = uint64(0xa0000000a)
	w.records[leaf] = &spellDurationFreeRecordState4FE980{}
	SpellDurationFreeRecursive4FE980(w.hooks(), leaf)
	want := []string{
		"sub108:a0000000a",
		"sub104:a0000000a",
		"allocator:800000808",
		"free:800000808:a0000000a",
	}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("leaf events = %q, want %q", w.events, want)
	}

	stop := &struct{}{}
	var events []string
	var recovered any
	func() {
		defer func() { recovered = recover() }()
		SpellDurationFreeRecursive4FE980(SpellDurationFreeRecursiveHooks4FE980[uint64, uint64]{
			LoadSub108: func(record uint64) uint64 {
				events = append(events, fmt.Sprintf("sub108:%x", record))
				panic(stop)
			},
		}, 0)
	}()
	if recovered != stop {
		t.Fatalf("nil-record recovery = %#v, want sentinel", recovered)
	}
	if want := []string{"sub108:0"}; !reflect.DeepEqual(events, want) {
		t.Fatalf("nil-record events = %q, want first dereference %q", events, want)
	}
}

func TestSpellDurationFreeRecursive4FE980FaultPrefixes(t *testing.T) {
	baseline := newSpellDurationFreeWorld4FE980()
	baseline.mutateOnFirst = true
	SpellDurationFreeRecursive4FE980(baseline.hooks(), spellDurationFreeRoot4FE980)
	want := append([]string(nil), baseline.events...)

	for faultAt := 1; faultAt <= len(want); faultAt++ {
		t.Run(fmt.Sprintf("fault-%d", faultAt), func(t *testing.T) {
			w := newSpellDurationFreeWorld4FE980()
			w.mutateOnFirst = true
			w.faultAt = faultAt
			var recovered any
			func() {
				defer func() { recovered = recover() }()
				SpellDurationFreeRecursive4FE980(w.hooks(), spellDurationFreeRoot4FE980)
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
