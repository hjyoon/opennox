package legacy

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"
)

const questJournalReadBuffer500B70 = uint64(0x100000123)

var errQuestJournalReadTest500B70 = errors.New("quest journal reader test transfer")

type questJournalReadTestEntry500B70 struct {
	name  []byte
	kind  uint32
	value uint32
}

type questJournalReadTestSetCall500B70 struct {
	buffer uint64
	name   string
	value  uint32
}

type questJournalReadTestWorld500B70 struct {
	version uint16
	count   uint32
	entries []questJournalReadTestEntry500B70
	buffer  [256]byte

	nextEntry  int
	current    int
	events     []string
	numeric    []questJournalReadTestSetCall500B70
	boolean    []questJournalReadTestSetCall500B70
	valueReads int
	faultAt    int
	errorAt    int
}

func newQuestJournalReadTestWorld500B70() *questJournalReadTestWorld500B70 {
	return &questJournalReadTestWorld500B70{
		version: questJournalReadVersion500B70,
		current: -1,
	}
}

func newQuestJournalReadFullTestWorld500B70() *questJournalReadTestWorld500B70 {
	w := newQuestJournalReadTestWorld500B70()
	w.entries = []questJournalReadTestEntry500B70{
		{name: []byte("Num"), kind: 0, value: 0x89abcdef},
		{name: []byte("Bool"), kind: 1, value: 1},
		{name: []byte("Skip"), kind: 0xffffffff, value: 0x11223344},
	}
	w.count = uint32(len(w.entries))
	return w
}

func (w *questJournalReadTestWorld500B70) observe(event string) error {
	w.events = append(w.events, event)
	index := len(w.events)
	if w.faultAt == index {
		panic(event)
	}
	if w.errorAt == index {
		return errQuestJournalReadTest500B70
	}
	return nil
}

func (w *questJournalReadTestWorld500B70) currentEntry() *questJournalReadTestEntry500B70 {
	if w.current < 0 || w.current >= len(w.entries) {
		panic("quest journal reader test entry underrun")
	}
	return &w.entries[w.current]
}

func (w *questJournalReadTestWorld500B70) bufferName() string {
	for i, value := range w.buffer {
		if value == 0 {
			return string(w.buffer[:i])
		}
	}
	return string(w.buffer[:])
}

func (w *questJournalReadTestWorld500B70) hooks() questJournalReadHooks500B70[uint64] {
	return questJournalReadHooks500B70[uint64]{
		deletePattern: func(pattern string) {
			_ = w.observe("delete:" + pattern)
		},
		readWriteVersion: func(input uint16) (uint16, error) {
			err := w.observe(fmt.Sprintf("rw-version:%d=%d", input, w.version))
			return w.version, err
		},
		readCount: func() (uint32, error) {
			err := w.observe(fmt.Sprintf("read-count=%d", w.count))
			return w.count, err
		},
		readNameLength: func() (uint8, error) {
			w.current = w.nextEntry
			w.nextEntry++
			length := uint8(len(w.currentEntry().name))
			err := w.observe(fmt.Sprintf("read-name-length=%d", length))
			return length, err
		},
		readName: func(buffer uint64, size uint8) error {
			err := w.observe(fmt.Sprintf("read-name:%x:%d", buffer, size))
			if err != nil {
				return err
			}
			name := w.currentEntry().name
			if len(name) < int(size) {
				panic("quest journal reader test name underrun")
			}
			copy(w.buffer[:size], name[:size])
			return nil
		},
		storeNameTerminator: func(buffer uint64, index uint8) {
			_ = w.observe(fmt.Sprintf("store-name-terminator:%x:%d", buffer, index))
			w.buffer[index] = 0
		},
		readKind: func() (uint32, error) {
			kind := w.currentEntry().kind
			err := w.observe(fmt.Sprintf("read-kind=%d", kind))
			return kind, err
		},
		readValue: func() (uint32, error) {
			w.valueReads++
			value := w.currentEntry().value
			err := w.observe(fmt.Sprintf("read-value=%d", value))
			return value, err
		},
		setNumeric: func(buffer uint64, value uint32) {
			name := w.bufferName()
			w.numeric = append(w.numeric, questJournalReadTestSetCall500B70{
				buffer: buffer,
				name:   name,
				value:  value,
			})
			_ = w.observe(fmt.Sprintf("set-numeric:%x:%s=%d", buffer, name, value))
		},
		setBoolean: func(buffer uint64, value uint32) {
			name := w.bufferName()
			w.boolean = append(w.boolean, questJournalReadTestSetCall500B70{
				buffer: buffer,
				name:   name,
				value:  value,
			})
			_ = w.observe(fmt.Sprintf("set-boolean:%x:%s=%d", buffer, name, value))
		},
	}
}

func TestQuestJournalRead500B70TraceAndKindDispatch(t *testing.T) {
	w := newQuestJournalReadFullTestWorld500B70()
	got, err := questJournalReadContract500B70(questJournalReadBuffer500B70, w.hooks())
	if err != nil || got != 1 {
		t.Fatalf("result = %d/%v, want 1/nil", got, err)
	}

	want := []string{
		"delete:*:*",
		"rw-version:1=1",
		"read-count=3",
		"read-name-length=3",
		"read-name:100000123:3",
		"store-name-terminator:100000123:3",
		"read-kind=0",
		"read-value=2309737967",
		"set-numeric:100000123:Num=2309737967",
		"read-name-length=4",
		"read-name:100000123:4",
		"store-name-terminator:100000123:4",
		"read-kind=1",
		"read-value=1",
		"set-boolean:100000123:Bool=1",
		"read-name-length=4",
		"read-name:100000123:4",
		"store-name-terminator:100000123:4",
		"read-kind=4294967295",
	}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %q, want %q", w.events, want)
	}
	if w.valueReads != 2 {
		t.Fatalf("value reads = %d, want 2; unsupported kind consumed a value", w.valueReads)
	}
	if wantNumeric := []questJournalReadTestSetCall500B70{{
		buffer: questJournalReadBuffer500B70,
		name:   "Num",
		value:  0x89abcdef,
	}}; !reflect.DeepEqual(w.numeric, wantNumeric) {
		t.Fatalf("numeric calls = %#v, want %#v", w.numeric, wantNumeric)
	}
	if wantBoolean := []questJournalReadTestSetCall500B70{{
		buffer: questJournalReadBuffer500B70,
		name:   "Bool",
		value:  1,
	}}; !reflect.DeepEqual(w.boolean, wantBoolean) {
		t.Fatalf("Boolean calls = %#v, want %#v", w.boolean, wantBoolean)
	}
}

func TestQuestJournalRead500B70SignedVersionGateAfterDelete(t *testing.T) {
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
			w := newQuestJournalReadTestWorld500B70()
			w.version = tc.version
			got, err := questJournalReadContract500B70(questJournalReadBuffer500B70, w.hooks())
			if err != nil || got != tc.want {
				t.Fatalf("result = %d/%v, want %d/nil", got, err, tc.want)
			}
			wantEvents := 3
			if tc.want == 0 {
				wantEvents = 2
			}
			if len(w.events) != wantEvents || w.events[0] != "delete:*:*" {
				t.Fatalf("events = %q, want delete first and %d operations", w.events, wantEvents)
			}
		})
	}
}

func TestQuestJournalRead500B70AllowsFullByteNameAndStoresTerminator(t *testing.T) {
	w := newQuestJournalReadTestWorld500B70()
	w.entries = []questJournalReadTestEntry500B70{{
		name:  bytes.Repeat([]byte{'x'}, 255),
		kind:  0,
		value: 7,
	}}
	w.count = 1
	got, err := questJournalReadContract500B70(questJournalReadBuffer500B70, w.hooks())
	if err != nil || got != 1 {
		t.Fatalf("result = %d/%v, want 1/nil", got, err)
	}
	if len(w.numeric) != 1 || len(w.numeric[0].name) != 255 {
		t.Fatalf("numeric name length = %d, want 255", len(w.numeric[0].name))
	}
	if w.buffer[255] != 0 {
		t.Fatalf("buffer[255] = %#x, want trailing NUL", w.buffer[255])
	}
}

func TestQuestJournalRead500B70FaultPrefixes(t *testing.T) {
	baseline := newQuestJournalReadFullTestWorld500B70()
	if _, err := questJournalReadContract500B70(questJournalReadBuffer500B70, baseline.hooks()); err != nil {
		t.Fatal(err)
	}
	want := append([]string(nil), baseline.events...)

	for faultAt := 1; faultAt <= len(want); faultAt++ {
		t.Run(fmt.Sprintf("operation-%d", faultAt), func(t *testing.T) {
			w := newQuestJournalReadFullTestWorld500B70()
			w.faultAt = faultAt
			var recovered any
			func() {
				defer func() { recovered = recover() }()
				_, _ = questJournalReadContract500B70(questJournalReadBuffer500B70, w.hooks())
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

func TestQuestJournalRead500B70TransferErrorPrefixes(t *testing.T) {
	baseline := newQuestJournalReadFullTestWorld500B70()
	if _, err := questJournalReadContract500B70(questJournalReadBuffer500B70, baseline.hooks()); err != nil {
		t.Fatal(err)
	}
	want := append([]string(nil), baseline.events...)

	for errorAt, event := range want {
		if !strings.HasPrefix(event, "rw-") && !strings.HasPrefix(event, "read-") {
			continue
		}
		t.Run(fmt.Sprintf("operation-%d", errorAt+1), func(t *testing.T) {
			w := newQuestJournalReadFullTestWorld500B70()
			w.errorAt = errorAt + 1
			got, err := questJournalReadContract500B70(questJournalReadBuffer500B70, w.hooks())
			if got != 0 || !errors.Is(err, errQuestJournalReadTest500B70) {
				t.Fatalf("result = %d/%v, want 0/test error", got, err)
			}
			if prefix := want[:errorAt+1]; !reflect.DeepEqual(w.events, prefix) {
				t.Fatalf("events = %q, want prefix %q", w.events, prefix)
			}
		})
	}
}
