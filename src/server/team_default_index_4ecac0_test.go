package server

import (
	"fmt"
	"reflect"
	"testing"
)

const (
	teamDefaultNilPointer4ECAC0   = 0
	teamDefaultInputPointer4ECAC0 = 100
)

type teamDefaultMemory4ECAC0 struct {
	input  string
	names  [17]teamDefaultNameEntry4ECAA0
	events []string
}

func newTeamDefaultMemory4ECAC0(input string) *teamDefaultMemory4ECAC0 {
	return &teamDefaultMemory4ECAC0{
		input: input,
		names: teamDefaultNameTable4ECAA0,
	}
}

func (m *teamDefaultMemory4ECAC0) loadName(index uint8) int {
	m.events = append(m.events, fmt.Sprintf("name:%d", index))
	entry := m.names[index]
	if !entry.present {
		return teamDefaultNilPointer4ECAC0
	}
	return int(index) + 1
}

func (m *teamDefaultMemory4ECAC0) loadByte(pointer int, offset uint32) byte {
	m.events = append(m.events, fmt.Sprintf("byte:%d:%d", pointer, offset))
	if pointer == teamDefaultNilPointer4ECAC0 {
		panic(teamDefaultNilPointer4ECAC0)
	}
	if pointer == teamDefaultInputPointer4ECAC0 {
		return teamDefaultCStringByte4ECAC0(m.input, offset)
	}
	return teamDefaultCStringByte4ECAC0(m.names[pointer-1].text, offset)
}

func (m *teamDefaultMemory4ECAC0) hooks() teamDefaultIndexHooks4ECAC0[int] {
	return teamDefaultIndexHooks4ECAC0[int]{
		loadName: m.loadName,
		loadByte: m.loadByte,
	}
}

func teamDefaultCStringByte4ECAC0(text string, offset uint32) byte {
	if uint64(offset) > uint64(len(text)) {
		panic("read beyond terminating NUL")
	}
	if int(offset) == len(text) {
		return 0
	}
	return text[offset]
}

func TestTeamDefaultIndex4ECAC0DefaultNames(t *testing.T) {
	for index := uint8(0); index < 16; index++ {
		entry := teamDefaultNameTable4ECAA0[index]
		t.Run(entry.text, func(t *testing.T) {
			memory := newTeamDefaultMemory4ECAC0(entry.text)
			got := teamDefaultIndex4ECAC0(teamDefaultInputPointer4ECAC0, memory.hooks())
			if got != index {
				t.Fatalf("index = %d, want %d", got, index)
			}
		})
	}
}

func TestTeamDefaultIndex4ECAC0TwoByteCompareAndRestart(t *testing.T) {
	memory := newTeamDefaultMemory4ECAC0("AB")
	memory.names[0] = teamDefaultNameEntry4ECAA0{text: "AX", present: true}
	memory.names[1] = teamDefaultNameEntry4ECAA0{text: "AB", present: true}

	got := teamDefaultIndex4ECAC0(teamDefaultInputPointer4ECAC0, memory.hooks())
	if got != 1 {
		t.Fatalf("index = %d, want 1", got)
	}
	want := []string{
		"name:0",
		"byte:100:0", "byte:1:0",
		"byte:100:1", "byte:1:1",
		"name:1",
		"byte:100:0", "byte:2:0",
		"byte:100:1", "byte:2:1",
		"byte:100:2", "byte:2:2",
	}
	if !reflect.DeepEqual(memory.events, want) {
		t.Fatalf("events = %#v, want %#v", memory.events, want)
	}
}

func TestTeamDefaultIndex4ECAC0FirstByteNULSkipsSecondLoad(t *testing.T) {
	memory := newTeamDefaultMemory4ECAC0("")
	memory.names[0] = teamDefaultNameEntry4ECAA0{text: "", present: true}

	got := teamDefaultIndex4ECAC0(teamDefaultInputPointer4ECAC0, memory.hooks())
	if got != 0 {
		t.Fatalf("index = %d, want 0", got)
	}
	want := []string{"name:0", "byte:100:0", "byte:1:0"}
	if !reflect.DeepEqual(memory.events, want) {
		t.Fatalf("events = %#v, want %#v", memory.events, want)
	}
}

func TestTeamDefaultIndex4ECAC0UnmatchedFaultsAtNilSentinel(t *testing.T) {
	memory := newTeamDefaultMemory4ECAC0("Team 16")
	var recovered any
	func() {
		defer func() { recovered = recover() }()
		teamDefaultIndex4ECAC0(teamDefaultInputPointer4ECAC0, memory.hooks())
	}()
	if recovered != teamDefaultNilPointer4ECAC0 {
		t.Fatalf("panic = %#v, want nil-pointer sentinel", recovered)
	}
	wantTail := []string{"name:16", "byte:100:0", "byte:0:0"}
	if len(memory.events) < len(wantTail) || !reflect.DeepEqual(memory.events[len(memory.events)-len(wantTail):], wantTail) {
		t.Fatalf("event tail = %#v, want %#v", memory.events, wantTail)
	}
}

func TestTeamDefaultIndex4ECAC0InputFaultFollowsFirstTableLoad(t *testing.T) {
	stop := &struct{}{}
	var events []string
	var recovered any
	func() {
		defer func() { recovered = recover() }()
		teamDefaultIndex4ECAC0(0, teamDefaultIndexHooks4ECAC0[int]{
			loadName: func(index uint8) int {
				events = append(events, fmt.Sprintf("name:%d", index))
				return 1
			},
			loadByte: func(pointer int, offset uint32) byte {
				events = append(events, fmt.Sprintf("byte:%d:%d", pointer, offset))
				if pointer == 0 {
					panic(stop)
				}
				t.Fatal("candidate byte loaded after input fault")
				return 0
			},
		})
	}()
	if recovered != stop {
		t.Fatalf("panic = %#v, want sentinel", recovered)
	}
	want := []string{"name:0", "byte:0:0"}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %#v, want %#v", events, want)
	}
}

func TestTeamDefaultIndex4ECAC0SecondInputBytePrecedesCandidate(t *testing.T) {
	stop := &struct{}{}
	var events []string
	var recovered any
	func() {
		defer func() { recovered = recover() }()
		teamDefaultIndex4ECAC0(1, teamDefaultIndexHooks4ECAC0[int]{
			loadName: func(uint8) int {
				events = append(events, "name:0")
				return 2
			},
			loadByte: func(pointer int, offset uint32) byte {
				events = append(events, fmt.Sprintf("byte:%d:%d", pointer, offset))
				if pointer == 1 && offset == 1 {
					panic(stop)
				}
				if pointer == 2 && offset == 1 {
					t.Fatal("candidate second byte loaded after input fault")
				}
				return 'A'
			},
		})
	}()
	if recovered != stop {
		t.Fatalf("panic = %#v, want sentinel", recovered)
	}
	want := []string{"name:0", "byte:1:0", "byte:2:0", "byte:1:1"}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %#v, want %#v", events, want)
	}
}

func TestTeamDefaultIndex4ECAC0CandidateSixteenAndExhaustion(t *testing.T) {
	tests := []struct {
		name      string
		last      string
		want      uint8
		wantLoads int
	}{
		{name: "match", last: "Z", want: 16, wantLoads: 17},
		{name: "mismatch", last: "Y", want: 0, wantLoads: 17},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			memory := newTeamDefaultMemory4ECAC0("Z")
			for i := range memory.names {
				memory.names[i] = teamDefaultNameEntry4ECAA0{text: "A", present: true}
			}
			memory.names[16].text = tc.last

			got := teamDefaultIndex4ECAC0(teamDefaultInputPointer4ECAC0, memory.hooks())
			if got != tc.want {
				t.Fatalf("index = %d, want %d", got, tc.want)
			}
			loads := 0
			lastName := ""
			for _, event := range memory.events {
				if len(event) >= 5 && event[:5] == "name:" {
					loads++
					lastName = event
				}
			}
			if loads != tc.wantLoads {
				t.Fatalf("table loads = %d, want %d", loads, tc.wantLoads)
			}
			if lastName != "name:16" {
				t.Fatalf("last table event = %q, want name:16", lastName)
			}
		})
	}
}

func TestTeamDefaultIndex4ECAC0TableLoadFaultPrecedesInput(t *testing.T) {
	stop := &struct{}{}
	inputLoads := 0
	var recovered any
	func() {
		defer func() { recovered = recover() }()
		teamDefaultIndex4ECAC0(1, teamDefaultIndexHooks4ECAC0[int]{
			loadName: func(uint8) int { panic(stop) },
			loadByte: func(int, uint32) byte {
				inputLoads++
				return 0
			},
		})
	}()
	if recovered != stop {
		t.Fatalf("panic = %#v, want sentinel", recovered)
	}
	if inputLoads != 0 {
		t.Fatalf("byte loads = %d, want 0", inputLoads)
	}
}
