package server

import (
	"fmt"
	"math"
	"reflect"
	"testing"
)

const (
	spellCancelRecordA4FEB10       = uint64(0x100000101)
	spellCancelRecordB4FEB10       = uint64(0x200000202)
	spellCancelRecordC4FEB10       = uint64(0x300000303)
	spellCancelCaster4FEB10        = uint64(0x400000404)
	spellCancelSameLowCaster4FEB10 = uint64(0x500000404)
)

type spellCancelRecordState4FEB10 struct {
	spell  int32
	caster uint64
	next   uint64
}

type spellCancelWorld4FEB10 struct {
	events    []string
	after     map[string]func()
	faultAt   int
	head      uint64
	records   map[uint64]*spellCancelRecordState4FEB10
	cancelled []uint64
}

func newSpellCancelWorld4FEB10() *spellCancelWorld4FEB10 {
	return &spellCancelWorld4FEB10{
		after: make(map[string]func()),
		head:  spellCancelRecordA4FEB10,
		records: map[uint64]*spellCancelRecordState4FEB10{
			spellCancelRecordA4FEB10: {
				spell:  67,
				caster: spellCancelCaster4FEB10,
				next:   spellCancelRecordB4FEB10,
			},
			spellCancelRecordB4FEB10: {
				spell:  67,
				caster: spellCancelSameLowCaster4FEB10,
				next:   spellCancelRecordC4FEB10,
			},
			spellCancelRecordC4FEB10: {
				spell:  114,
				caster: spellCancelCaster4FEB10,
			},
		},
	}
}

func (w *spellCancelWorld4FEB10) observe(event string) {
	w.events = append(w.events, event)
	if w.faultAt != 0 && len(w.events) == w.faultAt {
		panic(event)
	}
	if after := w.after[event]; after != nil {
		after()
	}
}

func (w *spellCancelWorld4FEB10) hooks() SpellCancelDurSpellHooks4FEB10[uint64, uint64] {
	return SpellCancelDurSpellHooks4FEB10[uint64, uint64]{
		LoadFirst: func() uint64 {
			value := w.head
			w.observe(fmt.Sprintf("first:%x", value))
			return value
		},
		LoadSpell: func(record uint64) int32 {
			value := w.records[record].spell
			w.observe(fmt.Sprintf("spell:%x:%d", record, value))
			return value
		},
		LoadNext: func(record uint64) uint64 {
			value := w.records[record].next
			w.observe(fmt.Sprintf("next:%x:%x", record, value))
			return value
		},
		LoadCaster: func(record uint64) uint64 {
			value := w.records[record].caster
			w.observe(fmt.Sprintf("caster:%x:%x", record, value))
			return value
		},
		Cancel: func(record uint64) {
			w.cancelled = append(w.cancelled, record)
			w.observe(fmt.Sprintf("cancel:%x", record))
		},
	}
}

func TestSpellCancelDurSpell4FEB10ExactIdentityAndSummonReplacement(t *testing.T) {
	w := newSpellCancelWorld4FEB10()
	w.records[spellCancelRecordB4FEB10].spell = 75

	got := SpellCancelDurSpell4FEB10(int32(67), spellCancelCaster4FEB10, w.hooks())

	wantEvents := []string{
		"first:100000101",
		"spell:100000101:67", "next:100000101:200000202", "caster:100000101:400000404", "cancel:100000101",
		"spell:200000202:75", "next:200000202:300000303",
		"spell:300000303:114", "next:300000303:0",
	}
	if !reflect.DeepEqual(w.events, wantEvents) {
		t.Fatalf("trace mismatch\n got: %q\nwant: %q", w.events, wantEvents)
	}
	if !reflect.DeepEqual(w.cancelled, []uint64{spellCancelRecordA4FEB10}) {
		t.Fatalf("cancelled = %x, want [%x]", w.cancelled, spellCancelRecordA4FEB10)
	}
	if got != 0 {
		t.Fatalf("result = %d, want canonical zero", got)
	}

	w = newSpellCancelWorld4FEB10()
	w.records[spellCancelRecordA4FEB10].spell = 75
	w.records[spellCancelRecordA4FEB10].next = spellCancelRecordB4FEB10
	w.records[spellCancelRecordB4FEB10].spell = 114
	w.records[spellCancelRecordB4FEB10].caster = spellCancelCaster4FEB10
	w.records[spellCancelRecordB4FEB10].next = 0
	SpellCancelDurSpell4FEB10(int32(75), spellCancelCaster4FEB10, w.hooks())
	if !reflect.DeepEqual(w.cancelled, []uint64{spellCancelRecordA4FEB10, spellCancelRecordB4FEB10}) {
		t.Fatalf("summon replacement cancelled = %x, want [%x %x]",
			w.cancelled, spellCancelRecordA4FEB10, spellCancelRecordB4FEB10)
	}
}

func TestSpellCancelDurSpell4FEB10RepeatsLiveCasterForExactSummonMismatch(t *testing.T) {
	w := newSpellCancelWorld4FEB10()
	w.records[spellCancelRecordA4FEB10].spell = 75
	w.records[spellCancelRecordA4FEB10].caster = spellCancelSameLowCaster4FEB10
	w.records[spellCancelRecordA4FEB10].next = 0
	firstCaster := "caster:100000101:500000404"
	w.after[firstCaster] = func() {
		w.records[spellCancelRecordA4FEB10].caster = spellCancelCaster4FEB10
	}

	got := SpellCancelDurSpell4FEB10(int32(75), spellCancelCaster4FEB10, w.hooks())

	want := []string{
		"first:100000101", "spell:100000101:75", "next:100000101:0",
		firstCaster, "caster:100000101:400000404", "cancel:100000101",
	}
	if got != 0 || !reflect.DeepEqual(w.events, want) {
		t.Fatalf("result/events = %d/%q, want 0/%q", got, w.events, want)
	}
}

func TestSpellCancelDurSpell4FEB10SignedBoundariesAndCasterReadGates(t *testing.T) {
	tests := []struct {
		name         string
		requested    int32
		record       int32
		recordCaster uint64
		wantReads    int
		wantCancel   bool
	}{
		{"min-exact", math.MinInt32, math.MinInt32, spellCancelCaster4FEB10, 1, true},
		{"negative-different", -1, math.MinInt32, spellCancelCaster4FEB10, 0, false},
		{"requested-below", 74, 75, spellCancelCaster4FEB10, 0, false},
		{"record-below", 75, 74, spellCancelCaster4FEB10, 0, false},
		{"lower-to-upper", 75, 114, spellCancelCaster4FEB10, 1, true},
		{"upper-to-lower", 114, 75, spellCancelCaster4FEB10, 1, true},
		{"requested-above", 115, 75, spellCancelCaster4FEB10, 0, false},
		{"record-above", 75, 115, spellCancelCaster4FEB10, 0, false},
		{"max-exact", math.MaxInt32, math.MaxInt32, spellCancelCaster4FEB10, 1, true},
		{"exact-nonsummon-mismatch", 67, 67, spellCancelSameLowCaster4FEB10, 1, false},
		{"exact-summon-mismatch", 75, 75, spellCancelSameLowCaster4FEB10, 2, false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			w := newSpellCancelWorld4FEB10()
			w.records[spellCancelRecordA4FEB10].spell = tc.record
			w.records[spellCancelRecordA4FEB10].caster = tc.recordCaster
			w.records[spellCancelRecordA4FEB10].next = 0
			SpellCancelDurSpell4FEB10(tc.requested, spellCancelCaster4FEB10, w.hooks())

			reads := 0
			for _, event := range w.events {
				if len(event) >= len("caster:") && event[:len("caster:")] == "caster:" {
					reads++
				}
			}
			if reads != tc.wantReads {
				t.Errorf("caster reads = %d, want %d; events: %q", reads, tc.wantReads, w.events)
			}
			if got := len(w.cancelled) != 0; got != tc.wantCancel {
				t.Errorf("cancelled = %t, want %t; events: %q", got, tc.wantCancel, w.events)
			}
		})
	}
}

func TestSpellCancelDurSpell4FEB10SpellBeforeNextAndSnapshotBeforeCaster(t *testing.T) {
	w := newSpellCancelWorld4FEB10()
	w.records[spellCancelRecordA4FEB10].next = spellCancelRecordB4FEB10
	w.records[spellCancelRecordC4FEB10].spell = 67
	w.after["spell:100000101:67"] = func() {
		w.records[spellCancelRecordA4FEB10].next = spellCancelRecordC4FEB10
	}
	w.after["caster:100000101:400000404"] = func() {
		w.records[spellCancelRecordA4FEB10].next = spellCancelRecordB4FEB10
	}
	w.after["cancel:100000101"] = func() {
		w.records[spellCancelRecordA4FEB10].next = spellCancelRecordB4FEB10
	}

	SpellCancelDurSpell4FEB10(int32(67), spellCancelCaster4FEB10, w.hooks())

	want := []string{
		"first:100000101",
		"spell:100000101:67", "next:100000101:300000303", "caster:100000101:400000404", "cancel:100000101",
		"spell:300000303:67", "next:300000303:0", "caster:300000303:400000404", "cancel:300000303",
	}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("trace mismatch\n got: %q\nwant: %q", w.events, want)
	}
}

func TestSpellCancelDurSpell4FEB10NilCasterIsComparable(t *testing.T) {
	w := newSpellCancelWorld4FEB10()
	w.records[spellCancelRecordA4FEB10].caster = 0
	w.records[spellCancelRecordA4FEB10].next = 0

	got := SpellCancelDurSpell4FEB10(int32(67), uint64(0), w.hooks())

	want := []string{
		"first:100000101", "spell:100000101:67", "next:100000101:0",
		"caster:100000101:0", "cancel:100000101",
	}
	if got != 0 || !reflect.DeepEqual(w.events, want) {
		t.Fatalf("result/events = %d/%q, want 0/%q", got, w.events, want)
	}
}

func TestSpellCancelDurSpell4FEB10EmptyStopsAfterFirstLoad(t *testing.T) {
	w := newSpellCancelWorld4FEB10()
	w.head = 0

	got := SpellCancelDurSpell4FEB10(int32(75), spellCancelCaster4FEB10, w.hooks())

	if got != 0 || !reflect.DeepEqual(w.events, []string{"first:0"}) {
		t.Fatalf("result/events = %d/%q, want 0/[first:0]", got, w.events)
	}
}

func TestSpellCancelDurSpell4FEB10FaultPrefixes(t *testing.T) {
	baseline := newSpellCancelWorld4FEB10()
	baseline.records[spellCancelRecordA4FEB10].spell = 75
	baseline.records[spellCancelRecordA4FEB10].caster = spellCancelSameLowCaster4FEB10
	baseline.records[spellCancelRecordB4FEB10].spell = 114
	baseline.records[spellCancelRecordB4FEB10].caster = spellCancelCaster4FEB10
	baseline.records[spellCancelRecordC4FEB10].spell = 74
	baseline.after["caster:100000101:500000404"] = func() {
		baseline.records[spellCancelRecordA4FEB10].caster = spellCancelCaster4FEB10
	}
	SpellCancelDurSpell4FEB10(int32(75), spellCancelCaster4FEB10, baseline.hooks())
	want := append([]string(nil), baseline.events...)

	for faultAt := 1; faultAt <= len(want); faultAt++ {
		t.Run(fmt.Sprintf("fault-%d", faultAt), func(t *testing.T) {
			w := newSpellCancelWorld4FEB10()
			w.records[spellCancelRecordA4FEB10].spell = 75
			w.records[spellCancelRecordA4FEB10].caster = spellCancelSameLowCaster4FEB10
			w.records[spellCancelRecordB4FEB10].spell = 114
			w.records[spellCancelRecordB4FEB10].caster = spellCancelCaster4FEB10
			w.records[spellCancelRecordC4FEB10].spell = 74
			w.after["caster:100000101:500000404"] = func() {
				w.records[spellCancelRecordA4FEB10].caster = spellCancelCaster4FEB10
			}
			w.faultAt = faultAt
			var recovered any
			func() {
				defer func() { recovered = recover() }()
				SpellCancelDurSpell4FEB10(int32(75), spellCancelCaster4FEB10, w.hooks())
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
