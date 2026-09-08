package legacy

import (
	"bytes"
	"fmt"
	"reflect"
	"strings"
	"testing"
)

type questJournalQualifyTestWorld5009B0 struct {
	input   string
	mapName string
	scratch []byte
	events  []string
	faultAt int
}

func newQuestJournalQualifyTestWorld5009B0(input, mapName string) *questJournalQualifyTestWorld5009B0 {
	return &questJournalQualifyTestWorld5009B0{
		input:   input,
		mapName: mapName,
	}
}

func (w *questJournalQualifyTestWorld5009B0) observe(event string) {
	w.events = append(w.events, event)
	if w.faultAt != 0 && len(w.events) == w.faultAt {
		panic(event)
	}
}

func (w *questJournalQualifyTestWorld5009B0) writeCString(at int, value string, count int) {
	if count != len(value)+1 {
		panic("invalid test copy count")
	}
	end := at + count
	if len(w.scratch) < end {
		w.scratch = append(w.scratch, make([]byte, end-len(w.scratch))...)
	}
	copy(w.scratch[at:end], value)
	w.scratch[end-1] = 0
}

func (w *questJournalQualifyTestWorld5009B0) scratchString() string {
	at := bytes.IndexByte(w.scratch, 0)
	if at < 0 {
		panic("test scratch is not NUL-terminated")
	}
	return string(w.scratch[:at])
}

func (w *questJournalQualifyTestWorld5009B0) hooks() questJournalQualifyHooks5009B0 {
	return questJournalQualifyHooks5009B0{
		scanInputColon: func() bool {
			found := strings.IndexByte(w.input, ':') >= 0
			w.observe(fmt.Sprintf("scan-input-colon=%t", found))
			return found
		},
		scanInputLength: func() int {
			length := len(w.input)
			w.observe(fmt.Sprintf("scan-input-length=%d", length))
			return length
		},
		scanMapLength: func() int {
			length := len(w.mapName)
			w.observe(fmt.Sprintf("scan-map-length=%d", length))
			return length
		},
		copyInput: func(count int) {
			w.writeCString(0, w.input, count)
			w.observe(fmt.Sprintf("copy-input=%d", count))
		},
		copyMap: func(count int) {
			w.writeCString(0, w.mapName, count)
			w.observe(fmt.Sprintf("copy-map=%d", count))
		},
		scanScratchLength: func() int {
			length := len(w.scratchString())
			w.observe(fmt.Sprintf("scan-scratch-length=%d", length))
			return length
		},
		writeSeparator: func(at int) {
			w.writeCString(at, ":", 2)
			w.observe(fmt.Sprintf("write-separator=%d", at))
		},
		appendInput: func(at, count int) {
			w.writeCString(at, w.input, count)
			w.observe(fmt.Sprintf("append-input=%d/%d", at, count))
		},
	}
}

func TestQuestJournalQualify5009B0QualifiedTraceAndReturn(t *testing.T) {
	input := "War01a:Count:Extra"
	w := newQuestJournalQualifyTestWorld5009B0(input, "must-not-be-read")
	got := questJournalQualifyContract5009B0(w.hooks())

	wantEvents := []string{
		"scan-input-colon=true",
		fmt.Sprintf("scan-input-length=%d", len(input)),
		fmt.Sprintf("copy-input=%d", len(input)+1),
	}
	if !reflect.DeepEqual(w.events, wantEvents) {
		t.Fatalf("events = %q, want %q", w.events, wantEvents)
	}
	if got != uint32(len(input)+1) {
		t.Fatalf("return = %d, want qualified byte count %d", got, len(input)+1)
	}
	if qualified := w.scratchString(); qualified != input {
		t.Fatalf("scratch = %q, want %q", qualified, input)
	}
}

func TestQuestJournalQualify5009B0UnqualifiedTraceAndReturn(t *testing.T) {
	input, mapName := "Count", "War01a"
	w := newQuestJournalQualifyTestWorld5009B0(input, mapName)
	got := questJournalQualifyContract5009B0(w.hooks())

	wantEvents := []string{
		"scan-input-colon=false",
		fmt.Sprintf("scan-map-length=%d", len(mapName)),
		fmt.Sprintf("copy-map=%d", len(mapName)+1),
		fmt.Sprintf("scan-scratch-length=%d", len(mapName)),
		fmt.Sprintf("write-separator=%d", len(mapName)),
		fmt.Sprintf("scan-input-length=%d", len(input)),
		fmt.Sprintf("scan-scratch-length=%d", len(mapName)+1),
		fmt.Sprintf("append-input=%d/%d", len(mapName)+1, len(input)+1),
	}
	if !reflect.DeepEqual(w.events, wantEvents) {
		t.Fatalf("events = %q, want %q", w.events, wantEvents)
	}
	if got != 0 {
		t.Fatalf("return = %d, want zero for map qualification", got)
	}
	if qualified, want := w.scratchString(), mapName+":"+input; qualified != want {
		t.Fatalf("scratch = %q, want %q", qualified, want)
	}
}

func TestQuestJournalQualifyString5009B0ColonAndEmptyEdges(t *testing.T) {
	for _, tc := range []struct {
		name         string
		input        string
		mapName      string
		want         string
		wantResult   uint32
		wantMapLoads int
	}{
		{name: "leading-colon", input: ":Quest", mapName: "War01a", want: ":Quest", wantResult: 7},
		{name: "trailing-colon", input: "Quest:", mapName: "War01a", want: "Quest:", wantResult: 7},
		{name: "multiple-colons", input: "A:B:C", mapName: "War01a", want: "A:B:C", wantResult: 6},
		{name: "empty-input", input: "", mapName: "War01a", want: "War01a:", wantMapLoads: 1},
		{name: "empty-map", input: "Quest", mapName: "", want: ":Quest", wantMapLoads: 1},
		{name: "both-empty", input: "", mapName: "", want: ":", wantMapLoads: 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			loads := 0
			got, result := questJournalQualifyString5009B0(tc.input, func() string {
				loads++
				return tc.mapName
			})
			if got != tc.want || result != tc.wantResult {
				t.Fatalf("qualification = %q/%d, want %q/%d", got, result, tc.want, tc.wantResult)
			}
			if loads != tc.wantMapLoads {
				t.Fatalf("map loads = %d, want %d", loads, tc.wantMapLoads)
			}
		})
	}
}

func TestQuestJournalQualify5009B0FaultPrefixes(t *testing.T) {
	for _, tc := range []struct {
		name    string
		input   string
		mapName string
	}{
		{name: "qualified", input: "War01a:Count", mapName: "unused"},
		{name: "unqualified", input: "Count", mapName: "War01a"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			baseline := newQuestJournalQualifyTestWorld5009B0(tc.input, tc.mapName)
			questJournalQualifyContract5009B0(baseline.hooks())
			want := append([]string(nil), baseline.events...)

			for faultAt := 1; faultAt <= len(want); faultAt++ {
				t.Run(fmt.Sprintf("operation-%d", faultAt), func(t *testing.T) {
					w := newQuestJournalQualifyTestWorld5009B0(tc.input, tc.mapName)
					w.faultAt = faultAt
					var recovered any
					func() {
						defer func() { recovered = recover() }()
						questJournalQualifyContract5009B0(w.hooks())
					}()
					if recovered == nil {
						t.Fatal("operation fault did not propagate")
					}
					if prefix := want[:faultAt]; !reflect.DeepEqual(w.events, prefix) {
						t.Fatalf("events = %q, want prefix %q", w.events, prefix)
					}
				})
			}
		})
	}
}
