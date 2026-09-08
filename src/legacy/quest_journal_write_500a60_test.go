package legacy

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"
)

const (
	questJournalWriteEntryA500A60 = uint64(0x100000101)
	questJournalWriteEntryB500A60 = uint64(0x200000202)
	questJournalWriteEntryC500A60 = uint64(0x300000303)
	questJournalWriteEntryD500A60 = uint64(0x400000404)
	questJournalWriteEntryE500A60 = uint64(0x500000505)
)

var errQuestJournalWriteTest500A60 = errors.New("quest journal writer test transfer")

type questJournalWriteTestNode500A60 struct {
	nameLength uint32
	kind       uint32
	next       uint64
}

type questJournalWriteTestNameCall500A60 struct {
	entry uint64
	size  uint8
}

type questJournalWriteTestWorld500A60 struct {
	version          uint16
	coopResult       int32
	head             uint64
	nodes            map[uint64]*questJournalWriteTestNode500A60
	nameLengthResult map[uint64]uint8
	currentNameEntry uint64
	events           []string
	countInputs      []uint32
	nameCalls        []questJournalWriteTestNameCall500A60
	kindCalls        []uint64
	valueCalls       []uint64
	faultAt          int
	errorAt          int
	after            map[int]func()
}

func newQuestJournalWriteTestWorld500A60() *questJournalWriteTestWorld500A60 {
	return &questJournalWriteTestWorld500A60{
		version:          questJournalWriteVersion500A60,
		nodes:            make(map[uint64]*questJournalWriteTestNode500A60),
		nameLengthResult: make(map[uint64]uint8),
		after:            make(map[int]func()),
	}
}

func (w *questJournalWriteTestWorld500A60) observe(event string) error {
	w.events = append(w.events, event)
	index := len(w.events)
	if w.faultAt == index {
		panic(event)
	}
	if after := w.after[index]; after != nil {
		after()
	}
	if w.errorAt == index {
		return errQuestJournalWriteTest500A60
	}
	return nil
}

func (w *questJournalWriteTestWorld500A60) hooks() questJournalWriteHooks500A60[uint64] {
	return questJournalWriteHooks500A60[uint64]{
		readWriteVersion: func(input uint16) (uint16, error) {
			err := w.observe(fmt.Sprintf("rw-version:%d=%d", input, w.version))
			return w.version, err
		},
		loadHead: func() uint64 {
			head := w.head
			_ = w.observe(fmt.Sprintf("load-head=%x", head))
			return head
		},
		loadNext: func(entry uint64) uint64 {
			next := w.nodes[entry].next
			_ = w.observe(fmt.Sprintf("load-next:%x=%x", entry, next))
			return next
		},
		checkGameFlags: func(mask uint32) int32 {
			result := w.coopResult
			_ = w.observe(fmt.Sprintf("check-game-flags:%x=%d", mask, result))
			return result
		},
		readWriteCount: func(count uint32) error {
			w.countInputs = append(w.countInputs, count)
			return w.observe(fmt.Sprintf("rw-count:%d", count))
		},
		scanNameLength: func(entry uint64) uint32 {
			length := w.nodes[entry].nameLength
			w.currentNameEntry = entry
			_ = w.observe(fmt.Sprintf("scan-name-length:%x=%d", entry, length))
			return length
		},
		readWriteNameLength: func(input uint8) (uint8, error) {
			result := input
			if configured, ok := w.nameLengthResult[w.currentNameEntry]; ok {
				result = configured
			}
			err := w.observe(fmt.Sprintf("rw-name-length:%d=%d", input, result))
			return result, err
		},
		readWriteName: func(entry uint64, size uint8) error {
			w.nameCalls = append(w.nameCalls, questJournalWriteTestNameCall500A60{entry: entry, size: size})
			return w.observe(fmt.Sprintf("rw-name:%x:%d", entry, size))
		},
		readWriteKind: func(entry uint64) error {
			w.kindCalls = append(w.kindCalls, entry)
			return w.observe(fmt.Sprintf("rw-kind:%x", entry))
		},
		loadKind: func(entry uint64) uint32 {
			kind := w.nodes[entry].kind
			_ = w.observe(fmt.Sprintf("load-kind:%x=%d", entry, kind))
			return kind
		},
		readWriteValue: func(entry uint64) error {
			w.valueCalls = append(w.valueCalls, entry)
			return w.observe(fmt.Sprintf("rw-value:%x", entry))
		},
	}
}

func newQuestJournalWriteFullTestWorld500A60() *questJournalWriteTestWorld500A60 {
	w := newQuestJournalWriteTestWorld500A60()
	w.coopResult = 1
	w.head = questJournalWriteEntryA500A60
	w.nodes[questJournalWriteEntryA500A60] = &questJournalWriteTestNode500A60{
		nameLength: 260,
		kind:       0,
		next:       questJournalWriteEntryB500A60,
	}
	w.nodes[questJournalWriteEntryB500A60] = &questJournalWriteTestNode500A60{
		nameLength: 3,
		kind:       2,
	}
	return w
}

func TestQuestJournalWrite500A60TraceAndLowByteNameLength(t *testing.T) {
	w := newQuestJournalWriteFullTestWorld500A60()
	got, err := questJournalWriteContract500A60(w.hooks())
	if err != nil || got != 1 {
		t.Fatalf("result = %d/%v, want 1/nil", got, err)
	}

	want := []string{
		"rw-version:1=1",
		"load-head=100000101",
		"load-next:100000101=200000202",
		"load-next:200000202=0",
		"check-game-flags:800=1",
		"rw-count:2",
		"load-head=100000101",
		"scan-name-length:100000101=260",
		"rw-name-length:4=4",
		"rw-name:100000101:4",
		"rw-kind:100000101",
		"load-kind:100000101=0",
		"rw-value:100000101",
		"load-next:100000101=200000202",
		"scan-name-length:200000202=3",
		"rw-name-length:3=3",
		"rw-name:200000202:3",
		"rw-kind:200000202",
		"load-kind:200000202=2",
		"load-next:200000202=0",
	}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %q, want %q", w.events, want)
	}
	if wantCounts := []uint32{2}; !reflect.DeepEqual(w.countInputs, wantCounts) {
		t.Fatalf("count inputs = %v, want %v", w.countInputs, wantCounts)
	}
}

func TestQuestJournalWrite500A60CountsBeforeNonCoopCheck(t *testing.T) {
	w := newQuestJournalWriteFullTestWorld500A60()
	w.coopResult = 0
	got, err := questJournalWriteContract500A60(w.hooks())
	if err != nil || got != 1 {
		t.Fatalf("result = %d/%v, want 1/nil", got, err)
	}
	want := []string{
		"rw-version:1=1",
		"load-head=100000101",
		"load-next:100000101=200000202",
		"load-next:200000202=0",
		"check-game-flags:800=0",
		"rw-count:0",
	}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %q, want %q", w.events, want)
	}
	if len(w.nameCalls) != 0 || len(w.kindCalls) != 0 || len(w.valueCalls) != 0 {
		t.Fatal("non-Coop path transferred journal entry fields")
	}
}

func TestQuestJournalWrite500A60SignedVersionGate(t *testing.T) {
	for _, tc := range []struct {
		version uint16
		want    int32
	}{
		{version: 0, want: 1},
		{version: 1, want: 1},
		{version: 2, want: 0},
		{version: 0x7fff, want: 0},
		{version: 0x8000, want: 1},
		{version: 0xffff, want: 1},
	} {
		t.Run(fmt.Sprintf("%04x", tc.version), func(t *testing.T) {
			w := newQuestJournalWriteTestWorld500A60()
			w.version = tc.version
			got, err := questJournalWriteContract500A60(w.hooks())
			if err != nil || got != tc.want {
				t.Fatalf("result = %d/%v, want %d/nil", got, err, tc.want)
			}
			if tc.want == 0 && len(w.events) != 1 {
				t.Fatalf("rejected-version events = %q, want only version transfer", w.events)
			}
		})
	}
}

func TestQuestJournalWrite500A60ReloadsMutableState(t *testing.T) {
	w := newQuestJournalWriteFullTestWorld500A60()
	w.nodes[questJournalWriteEntryC500A60] = &questJournalWriteTestNode500A60{
		nameLength: 260,
		kind:       2,
		next:       questJournalWriteEntryD500A60,
	}
	w.nodes[questJournalWriteEntryD500A60] = &questJournalWriteTestNode500A60{kind: 3}
	w.nodes[questJournalWriteEntryE500A60] = &questJournalWriteTestNode500A60{kind: 3}
	w.nameLengthResult[questJournalWriteEntryC500A60] = 7
	w.after[6] = func() {
		w.head = questJournalWriteEntryC500A60
	}
	w.after[11] = func() {
		w.nodes[questJournalWriteEntryC500A60].kind = 1
	}
	w.after[13] = func() {
		w.nodes[questJournalWriteEntryC500A60].next = questJournalWriteEntryE500A60
	}

	got, err := questJournalWriteContract500A60(w.hooks())
	if err != nil || got != 1 {
		t.Fatalf("result = %d/%v, want 1/nil", got, err)
	}
	if want := []uint32{2}; !reflect.DeepEqual(w.countInputs, want) {
		t.Fatalf("cached count = %v, want %v", w.countInputs, want)
	}
	if want := []questJournalWriteTestNameCall500A60{
		{entry: questJournalWriteEntryC500A60, size: 7},
		{entry: questJournalWriteEntryE500A60, size: 0},
	}; !reflect.DeepEqual(w.nameCalls, want) {
		t.Fatalf("name calls = %#v, want %#v", w.nameCalls, want)
	}
	if want := []uint64{questJournalWriteEntryC500A60}; !reflect.DeepEqual(w.valueCalls, want) {
		t.Fatalf("value calls = %x, want %x", w.valueCalls, want)
	}
	if strings.Contains(strings.Join(w.events, " "), fmt.Sprintf("%x", questJournalWriteEntryD500A60)) {
		t.Fatal("writer followed the stale pre-transfer next link")
	}
}

func TestQuestJournalWrite500A60ReloadsHeadAfterZeroCount(t *testing.T) {
	w := newQuestJournalWriteTestWorld500A60()
	w.coopResult = 1
	w.nodes[questJournalWriteEntryA500A60] = &questJournalWriteTestNode500A60{kind: 2}
	w.after[4] = func() {
		w.head = questJournalWriteEntryA500A60
	}

	got, err := questJournalWriteContract500A60(w.hooks())
	if err != nil || got != 1 {
		t.Fatalf("result = %d/%v, want 1/nil", got, err)
	}
	if want := []uint32{0}; !reflect.DeepEqual(w.countInputs, want) {
		t.Fatalf("count inputs = %v, want %v", w.countInputs, want)
	}
	if want := []uint64{questJournalWriteEntryA500A60}; !reflect.DeepEqual(w.kindCalls, want) {
		t.Fatalf("kind calls = %x, want post-count head %x", w.kindCalls, want)
	}
}

func TestQuestJournalWrite500A60FaultPrefixes(t *testing.T) {
	baseline := newQuestJournalWriteFullTestWorld500A60()
	if _, err := questJournalWriteContract500A60(baseline.hooks()); err != nil {
		t.Fatal(err)
	}
	want := append([]string(nil), baseline.events...)

	for faultAt := 1; faultAt <= len(want); faultAt++ {
		t.Run(fmt.Sprintf("operation-%d", faultAt), func(t *testing.T) {
			w := newQuestJournalWriteFullTestWorld500A60()
			w.faultAt = faultAt
			var recovered any
			func() {
				defer func() { recovered = recover() }()
				_, _ = questJournalWriteContract500A60(w.hooks())
			}()
			if recovered == nil {
				t.Fatal("operation fault did not propagate")
			}
			if prefix := want[:faultAt]; !reflect.DeepEqual(w.events, prefix) {
				t.Fatalf("events = %q, want prefix %q", w.events, prefix)
			}
		})
	}
}

func TestQuestJournalWrite500A60TransferErrorPrefixes(t *testing.T) {
	baseline := newQuestJournalWriteFullTestWorld500A60()
	if _, err := questJournalWriteContract500A60(baseline.hooks()); err != nil {
		t.Fatal(err)
	}
	want := append([]string(nil), baseline.events...)

	for errorAt, event := range want {
		if !strings.HasPrefix(event, "rw-") {
			continue
		}
		t.Run(fmt.Sprintf("operation-%d", errorAt+1), func(t *testing.T) {
			w := newQuestJournalWriteFullTestWorld500A60()
			w.errorAt = errorAt + 1
			got, err := questJournalWriteContract500A60(w.hooks())
			if got != 0 || !errors.Is(err, errQuestJournalWriteTest500A60) {
				t.Fatalf("result = %d/%v, want 0/test error", got, err)
			}
			if prefix := want[:errorAt+1]; !reflect.DeepEqual(w.events, prefix) {
				t.Fatalf("events = %q, want prefix %q", w.events, prefix)
			}
		})
	}
}
