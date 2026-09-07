package server

import (
	"fmt"
	"math"
	"reflect"
	"testing"
)

const (
	spellDurationFindRecordA4FF2D0 = uint64(0x100000101)
	spellDurationFindRecordB4FF2D0 = uint64(0x200000202)
	spellDurationFindRecordC4FF2D0 = uint64(0x300000303)
	spellDurationFindRecordD4FF2D0 = uint64(0x400000404)
	spellDurationFindRecordE4FF2D0 = uint64(0x500000505)

	spellDurationFindTarget4FF2D0        = uint64(0x600000606)
	spellDurationFindSameLowTarget4FF2D0 = uint64(0x700000606)
)

type spellDurationFindRecordState4FF2D0 struct {
	flags  uint32
	spell  int32
	target uint64
	next   uint64
}

type spellDurationFindWorld4FF2D0 struct {
	events  []string
	after   map[string]func()
	faultAt int
	head    uint64
	records map[uint64]*spellDurationFindRecordState4FF2D0
}

func newSpellDurationFindWorld4FF2D0() *spellDurationFindWorld4FF2D0 {
	return &spellDurationFindWorld4FF2D0{
		after: make(map[string]func()),
		head:  spellDurationFindRecordA4FF2D0,
		records: map[uint64]*spellDurationFindRecordState4FF2D0{
			spellDurationFindRecordA4FF2D0: {
				flags: 1, spell: math.MinInt32, target: spellDurationFindTarget4FF2D0,
				next: spellDurationFindRecordB4FF2D0,
			},
			spellDurationFindRecordB4FF2D0: {
				spell: math.MaxInt32, target: spellDurationFindTarget4FF2D0,
				next: spellDurationFindRecordC4FF2D0,
			},
			spellDurationFindRecordC4FF2D0: {
				spell: math.MinInt32, next: spellDurationFindRecordD4FF2D0,
			},
			spellDurationFindRecordD4FF2D0: {
				spell: math.MinInt32, target: spellDurationFindSameLowTarget4FF2D0,
				next: spellDurationFindRecordE4FF2D0,
			},
			spellDurationFindRecordE4FF2D0: {
				spell: math.MinInt32, target: spellDurationFindTarget4FF2D0,
			},
		},
	}
}

func (w *spellDurationFindWorld4FF2D0) observe(event string) {
	w.events = append(w.events, event)
	if w.faultAt != 0 && len(w.events) == w.faultAt {
		panic(event)
	}
	if after := w.after[event]; after != nil {
		after()
	}
}

func (w *spellDurationFindWorld4FF2D0) hooks() SpellDurationFindActiveTargetHooks4FF2D0[uint64, uint64] {
	return SpellDurationFindActiveTargetHooks4FF2D0[uint64, uint64]{
		LoadFirst: func() uint64 {
			value := w.head
			w.observe(fmt.Sprintf("first:%x", value))
			return value
		},
		LoadFlagsLow: func(record uint64) byte {
			value := byte(w.records[record].flags)
			w.observe(fmt.Sprintf("flags:%x:%02x", record, value))
			return value
		},
		LoadSpell: func(record uint64) int32 {
			value := w.records[record].spell
			w.observe(fmt.Sprintf("spell:%x:%08x", record, uint32(value)))
			return value
		},
		LoadTarget: func(record uint64) uint64 {
			value := w.records[record].target
			w.observe(fmt.Sprintf("target:%x:%x", record, value))
			return value
		},
		LoadNext: func(record uint64) uint64 {
			value := w.records[record].next
			w.observe(fmt.Sprintf("next:%x:%x", record, value))
			return value
		},
	}
}

func TestSpellDurationFindActiveTarget4FF2D0OrderGatesAndNativeIdentity(t *testing.T) {
	w := newSpellDurationFindWorld4FF2D0()

	got := SpellDurationFindActiveTarget4FF2D0(
		int32(math.MinInt32), spellDurationFindTarget4FF2D0, w.hooks(),
	)

	wantEvents := []string{
		"first:100000101",
		"flags:100000101:01", "next:100000101:200000202",
		"flags:200000202:00", "spell:200000202:7fffffff", "next:200000202:300000303",
		"flags:300000303:00", "spell:300000303:80000000", "target:300000303:0", "next:300000303:400000404",
		"flags:400000404:00", "spell:400000404:80000000", "target:400000404:700000606", "next:400000404:500000505",
		"flags:500000505:00", "spell:500000505:80000000", "target:500000505:600000606",
	}
	if got != spellDurationFindRecordE4FF2D0 {
		t.Fatalf("result = %x, want exact native-width record %x", got, spellDurationFindRecordE4FF2D0)
	}
	if !reflect.DeepEqual(w.events, wantEvents) {
		t.Fatalf("trace mismatch\n got: %q\nwant: %q", w.events, wantEvents)
	}
}

func TestSpellDurationFindActiveTarget4FF2D0UsesFlagsLowByteAndFullSpellDword(t *testing.T) {
	tests := []struct {
		name      string
		flags     uint32
		requested int32
		spell     int32
		want      uint64
	}{
		{"high-flag-bits-do-not-disable", 0xffffff00, math.MinInt32, math.MinInt32, spellDurationFindRecordA4FF2D0},
		{"low-bit-disables", 0xffffff01, math.MinInt32, math.MinInt32, 0},
		{"signed-min-exact", 0, math.MinInt32, math.MinInt32, spellDurationFindRecordA4FF2D0},
		{"high-bit-dword-mismatch", 0, math.MaxInt32, math.MinInt32, 0},
		{"signed-max-exact", 0, math.MaxInt32, math.MaxInt32, spellDurationFindRecordA4FF2D0},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			w := newSpellDurationFindWorld4FF2D0()
			w.records[spellDurationFindRecordA4FF2D0] = &spellDurationFindRecordState4FF2D0{
				flags: tc.flags, spell: tc.spell, target: spellDurationFindTarget4FF2D0,
			}
			got := SpellDurationFindActiveTarget4FF2D0(tc.requested, spellDurationFindTarget4FF2D0, w.hooks())
			if got != tc.want {
				t.Fatalf("result = %x, want %x; events = %q", got, tc.want, w.events)
			}
		})
	}
}

func TestSpellDurationFindActiveTarget4FF2D0NilTargetNeverMatches(t *testing.T) {
	w := newSpellDurationFindWorld4FF2D0()
	w.records[spellDurationFindRecordA4FF2D0] = &spellDurationFindRecordState4FF2D0{
		spell: math.MinInt32,
	}

	got := SpellDurationFindActiveTarget4FF2D0(int32(math.MinInt32), uint64(0), w.hooks())

	want := []string{
		"first:100000101", "flags:100000101:00", "spell:100000101:80000000",
		"target:100000101:0", "next:100000101:0",
	}
	if got != 0 || !reflect.DeepEqual(w.events, want) {
		t.Fatalf("result/events = %x/%q, want canonical nil/%q", got, w.events, want)
	}
}

func TestSpellDurationFindActiveTarget4FF2D0ReadsLiveFieldsAndSuccessor(t *testing.T) {
	w := newSpellDurationFindWorld4FF2D0()
	w.records[spellDurationFindRecordA4FF2D0] = &spellDurationFindRecordState4FF2D0{
		flags: 0x100, spell: 12,
		next: spellDurationFindRecordB4FF2D0,
	}
	w.records[spellDurationFindRecordC4FF2D0] = &spellDurationFindRecordState4FF2D0{
		spell: 13, target: spellDurationFindTarget4FF2D0,
	}
	w.after["flags:100000101:00"] = func() {
		w.records[spellDurationFindRecordA4FF2D0].spell = 13
	}
	w.after["spell:100000101:0000000d"] = func() {
		w.records[spellDurationFindRecordA4FF2D0].target = spellDurationFindSameLowTarget4FF2D0
	}
	w.after["target:100000101:700000606"] = func() {
		w.records[spellDurationFindRecordA4FF2D0].next = spellDurationFindRecordC4FF2D0
	}

	got := SpellDurationFindActiveTarget4FF2D0(int32(13), spellDurationFindTarget4FF2D0, w.hooks())

	want := []string{
		"first:100000101", "flags:100000101:00", "spell:100000101:0000000d",
		"target:100000101:700000606", "next:100000101:300000303",
		"flags:300000303:00", "spell:300000303:0000000d", "target:300000303:600000606",
	}
	if got != spellDurationFindRecordC4FF2D0 || !reflect.DeepEqual(w.events, want) {
		t.Fatalf("result/events = %x/%q, want %x/%q", got, w.events, spellDurationFindRecordC4FF2D0, want)
	}
}

func TestSpellDurationFindActiveTarget4FF2D0EmptyStopsAfterFirst(t *testing.T) {
	w := newSpellDurationFindWorld4FF2D0()
	w.head = 0

	got := SpellDurationFindActiveTarget4FF2D0(int32(51), spellDurationFindTarget4FF2D0, w.hooks())

	if got != 0 || !reflect.DeepEqual(w.events, []string{"first:0"}) {
		t.Fatalf("result/events = %x/%q, want canonical nil/[first:0]", got, w.events)
	}
}

func TestSpellDurationFindActiveTarget4FF2D0HasNoCycleGuard(t *testing.T) {
	w := newSpellDurationFindWorld4FF2D0()
	w.records[spellDurationFindRecordA4FF2D0] = &spellDurationFindRecordState4FF2D0{
		spell: 1, next: spellDurationFindRecordA4FF2D0,
	}
	w.faultAt = 8

	var recovered any
	func() {
		defer func() { recovered = recover() }()
		SpellDurationFindActiveTarget4FF2D0(int32(2), spellDurationFindTarget4FF2D0, w.hooks())
	}()

	want := []string{
		"first:100000101",
		"flags:100000101:00", "spell:100000101:00000001", "next:100000101:100000101",
		"flags:100000101:00", "spell:100000101:00000001", "next:100000101:100000101",
		"flags:100000101:00",
	}
	if recovered == nil || !reflect.DeepEqual(w.events, want) {
		t.Fatalf("recovered/events = %v/%q, want repeated unguarded prefix %q", recovered, w.events, want)
	}
}

func TestSpellDurationFindActiveTarget4FF2D0FaultPrefixes(t *testing.T) {
	baseline := newSpellDurationFindWorld4FF2D0()
	SpellDurationFindActiveTarget4FF2D0(
		int32(math.MinInt32), spellDurationFindTarget4FF2D0, baseline.hooks(),
	)
	want := append([]string(nil), baseline.events...)

	for faultAt := 1; faultAt <= len(want); faultAt++ {
		t.Run(fmt.Sprintf("fault-%d", faultAt), func(t *testing.T) {
			w := newSpellDurationFindWorld4FF2D0()
			w.faultAt = faultAt
			var recovered any
			func() {
				defer func() { recovered = recover() }()
				SpellDurationFindActiveTarget4FF2D0(
					int32(math.MinInt32), spellDurationFindTarget4FF2D0, w.hooks(),
				)
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
