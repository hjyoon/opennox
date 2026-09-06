package server

import (
	"fmt"
	"reflect"
	"testing"
)

const (
	playerCancelRecordA4FEAE0       = uint64(0x100000101)
	playerCancelRecordB4FEAE0       = uint64(0x200000202)
	playerCancelRecordC4FEAE0       = uint64(0x300000303)
	playerCancelCaster4FEAE0        = uint64(0x400000404)
	playerCancelSameLowCaster4FEAE0 = uint64(0x500000404)
)

type playerCancelRecordState4FEAE0 struct {
	caster uint64
	next   uint64
}

type playerCancelWorld4FEAE0 struct {
	events    []string
	after     map[string]func()
	faultAt   int
	head      uint64
	records   map[uint64]*playerCancelRecordState4FEAE0
	cancelled []uint64
}

func newPlayerCancelWorld4FEAE0() *playerCancelWorld4FEAE0 {
	return &playerCancelWorld4FEAE0{
		after: make(map[string]func()),
		head:  playerCancelRecordA4FEAE0,
		records: map[uint64]*playerCancelRecordState4FEAE0{
			playerCancelRecordA4FEAE0: {
				caster: playerCancelCaster4FEAE0,
				next:   playerCancelRecordB4FEAE0,
			},
			playerCancelRecordB4FEAE0: {
				caster: playerCancelSameLowCaster4FEAE0,
				next:   playerCancelRecordC4FEAE0,
			},
			playerCancelRecordC4FEAE0: {
				caster: playerCancelCaster4FEAE0,
			},
		},
	}
}

func (w *playerCancelWorld4FEAE0) observe(event string) {
	w.events = append(w.events, event)
	if w.faultAt != 0 && len(w.events) == w.faultAt {
		panic(event)
	}
	if after := w.after[event]; after != nil {
		after()
	}
}

func (w *playerCancelWorld4FEAE0) hooks() PlayerCancelSpellsHooks4FEAE0[uint64, uint64] {
	return PlayerCancelSpellsHooks4FEAE0[uint64, uint64]{
		LoadFirst: func() uint64 {
			value := w.head
			w.observe(fmt.Sprintf("first:%x", value))
			return value
		},
		LoadCaster: func(record uint64) uint64 {
			value := w.records[record].caster
			w.observe(fmt.Sprintf("caster:%x:%x", record, value))
			return value
		},
		LoadNext: func(record uint64) uint64 {
			value := w.records[record].next
			w.observe(fmt.Sprintf("next:%x:%x", record, value))
			return value
		},
		Cancel: func(record uint64) {
			w.cancelled = append(w.cancelled, record)
			w.observe(fmt.Sprintf("cancel:%x", record))
		},
	}
}

func TestPlayerCancelSpells4FEAE0ExactIdentityAndTrace(t *testing.T) {
	w := newPlayerCancelWorld4FEAE0()

	got := PlayerCancelSpells4FEAE0(playerCancelCaster4FEAE0, w.hooks())

	wantEvents := []string{
		"first:100000101",
		"caster:100000101:400000404", "next:100000101:200000202", "cancel:100000101",
		"caster:200000202:500000404", "next:200000202:300000303",
		"caster:300000303:400000404", "next:300000303:0", "cancel:300000303",
	}
	if !reflect.DeepEqual(w.events, wantEvents) {
		t.Fatalf("trace mismatch\n got: %q\nwant: %q", w.events, wantEvents)
	}
	wantCancelled := []uint64{playerCancelRecordA4FEAE0, playerCancelRecordC4FEAE0}
	if !reflect.DeepEqual(w.cancelled, wantCancelled) {
		t.Fatalf("cancelled = %x, want %x", w.cancelled, wantCancelled)
	}
	if got != 0 {
		t.Fatalf("result = %d, want canonical zero", got)
	}
}

func TestPlayerCancelSpells4FEAE0CasterBeforeNextAndNextBeforeCancel(t *testing.T) {
	w := newPlayerCancelWorld4FEAE0()
	w.after["caster:100000101:400000404"] = func() {
		w.records[playerCancelRecordA4FEAE0].next = playerCancelRecordC4FEAE0
	}
	w.after["cancel:100000101"] = func() {
		w.records[playerCancelRecordA4FEAE0].next = playerCancelRecordB4FEAE0
	}

	PlayerCancelSpells4FEAE0(playerCancelCaster4FEAE0, w.hooks())

	want := []string{
		"first:100000101",
		"caster:100000101:400000404", "next:100000101:300000303", "cancel:100000101",
		"caster:300000303:400000404", "next:300000303:0", "cancel:300000303",
	}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("trace mismatch\n got: %q\nwant: %q", w.events, want)
	}
	for _, event := range w.events {
		if event == "caster:200000202:500000404" {
			t.Fatal("visited the successor installed after it had been snapshotted")
		}
	}
}

func TestPlayerCancelSpells4FEAE0NilCasterIsComparable(t *testing.T) {
	w := newPlayerCancelWorld4FEAE0()
	w.records[playerCancelRecordA4FEAE0].caster = 0
	w.records[playerCancelRecordA4FEAE0].next = 0

	got := PlayerCancelSpells4FEAE0(uint64(0), w.hooks())

	want := []string{
		"first:100000101",
		"caster:100000101:0", "next:100000101:0", "cancel:100000101",
	}
	if got != 0 || !reflect.DeepEqual(w.events, want) {
		t.Fatalf("result/events = %d/%q, want 0/%q", got, w.events, want)
	}
}

func TestPlayerCancelSpells4FEAE0EmptyStopsAfterFirstLoad(t *testing.T) {
	w := newPlayerCancelWorld4FEAE0()
	w.head = 0

	got := PlayerCancelSpells4FEAE0(playerCancelCaster4FEAE0, w.hooks())

	if got != 0 || !reflect.DeepEqual(w.events, []string{"first:0"}) {
		t.Fatalf("result/events = %d/%q, want 0/[first:0]", got, w.events)
	}
}

func TestPlayerCancelSpells4FEAE0FaultPrefixes(t *testing.T) {
	baseline := newPlayerCancelWorld4FEAE0()
	PlayerCancelSpells4FEAE0(playerCancelCaster4FEAE0, baseline.hooks())
	want := append([]string(nil), baseline.events...)

	for faultAt := 1; faultAt <= len(want); faultAt++ {
		t.Run(fmt.Sprintf("fault-%d", faultAt), func(t *testing.T) {
			w := newPlayerCancelWorld4FEAE0()
			w.faultAt = faultAt
			var recovered any
			func() {
				defer func() { recovered = recover() }()
				PlayerCancelSpells4FEAE0(playerCancelCaster4FEAE0, w.hooks())
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
