package server

import (
	"fmt"
	"reflect"
	"testing"
)

const teamMaterialInputPointer4ECB60 = 100

type teamMaterialMemory4ECB60 struct {
	input  string
	names  map[uint32]string
	teams  map[uint32]uint32
	events []string
}

func newTeamMaterialMemory4ECB60(input string) *teamMaterialMemory4ECB60 {
	memory := &teamMaterialMemory4ECB60{
		input: input,
		names: make(map[uint32]string),
		teams: make(map[uint32]uint32),
	}
	for index, entry := range teamMaterialTable4ECB20 {
		if entry.name != "" {
			memory.names[uint32(index)] = entry.name
		}
		memory.teams[uint32(index)] = entry.team
	}
	return memory
}

func (m *teamMaterialMemory4ECB60) loadName(index uint32) int {
	m.events = append(m.events, fmt.Sprintf("name:%d", index))
	if _, ok := m.names[index]; !ok {
		return 0
	}
	return int(index) + 1
}

func (m *teamMaterialMemory4ECB60) loadByte(pointer int, offset uint32) byte {
	m.events = append(m.events, fmt.Sprintf("byte:%d:%d", pointer, offset))
	if pointer == 0 {
		panic("nil name")
	}
	if pointer == teamMaterialInputPointer4ECB60 {
		return teamMaterialCStringByte4ECB60(m.input, offset)
	}
	return teamMaterialCStringByte4ECB60(m.names[uint32(pointer-1)], offset)
}

func (m *teamMaterialMemory4ECB60) loadTeam(index uint32) uint32 {
	m.events = append(m.events, fmt.Sprintf("team:%d", index))
	return m.teams[index]
}

func (m *teamMaterialMemory4ECB60) hooks() teamMaterialIndexHooks4ECB60[int, uint32] {
	return teamMaterialIndexHooks4ECB60[int, uint32]{
		loadName: m.loadName,
		loadByte: m.loadByte,
		loadTeam: m.loadTeam,
	}
}

func teamMaterialCStringByte4ECB60(text string, offset uint32) byte {
	if uint64(offset) > uint64(len(text)) {
		panic("read beyond terminating NUL")
	}
	if uint64(offset) == uint64(len(text)) {
		return 0
	}
	return text[offset]
}

func TestTeamMaterialIndexValue4ECB60Table(t *testing.T) {
	for _, entry := range teamMaterialTable4ECB20[:len(teamMaterialTable4ECB20)-1] {
		t.Run(entry.name, func(t *testing.T) {
			if got := teamMaterialIndexValue4ECB60(entry.name); got != entry.team {
				t.Fatalf("team = %d, want %d", got, entry.team)
			}
		})
	}
	if got := teamMaterialIndexValue4ECB60("materialteamred"); got != 0 {
		t.Fatalf("case-mismatched team = %d, want 0", got)
	}
	if got := teamMaterialIndexValue4ECB60("MaterialTeamMissing"); got != 0 {
		t.Fatalf("missing team = %d, want 0", got)
	}
}

func TestTeamMaterialIndex4ECB60HeadGatePrecedesInput(t *testing.T) {
	var events []string
	got := teamMaterialIndex4ECB60(1, teamMaterialIndexHooks4ECB60[int, uint32]{
		loadName: func(index uint32) int {
			events = append(events, fmt.Sprintf("name:%d", index))
			return 0
		},
		loadByte: func(int, uint32) byte {
			t.Fatal("input or candidate read after nil head")
			return 0
		},
		loadTeam: func(uint32) uint32 {
			t.Fatal("team read after nil head")
			return 0
		},
	})
	if got != 0 {
		t.Fatalf("team = %d, want 0", got)
	}
	if want := []string{"name:0"}; !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %#v, want %#v", events, want)
	}
}

func TestTeamMaterialIndex4ECB60TwoByteCompareAndRestart(t *testing.T) {
	memory := newTeamMaterialMemory4ECB60("AB")
	memory.names = map[uint32]string{0: "AX", 1: "AB"}
	memory.teams[1] = 3

	got := teamMaterialIndex4ECB60(teamMaterialInputPointer4ECB60, memory.hooks())
	if got != 3 {
		t.Fatalf("team = %d, want 3", got)
	}
	want := []string{
		"name:0",
		"byte:1:0", "byte:100:0",
		"byte:1:1", "byte:100:1",
		"name:1",
		"byte:2:0", "byte:100:0",
		"byte:2:1", "byte:100:1",
		"byte:2:2", "byte:100:2",
		"team:1",
	}
	if !reflect.DeepEqual(memory.events, want) {
		t.Fatalf("events = %#v, want %#v", memory.events, want)
	}
}

func TestTeamMaterialIndex4ECB60FirstByteNULSkipsSecondPair(t *testing.T) {
	memory := newTeamMaterialMemory4ECB60("")
	memory.names = map[uint32]string{0: ""}
	memory.teams[0] = 7
	loadName := func(index uint32) int {
		memory.events = append(memory.events, fmt.Sprintf("name:%d", index))
		if index == 0 {
			return 1
		}
		return 0
	}
	hooks := memory.hooks()
	hooks.loadName = loadName

	got := teamMaterialIndex4ECB60(teamMaterialInputPointer4ECB60, hooks)
	if got != 7 {
		t.Fatalf("team = %d, want 7", got)
	}
	want := []string{"name:0", "byte:1:0", "byte:100:0", "team:0"}
	if !reflect.DeepEqual(memory.events, want) {
		t.Fatalf("events = %#v, want %#v", memory.events, want)
	}
}

func TestTeamMaterialIndex4ECB60SentinelStopsBeforeBytesAndTeam(t *testing.T) {
	memory := newTeamMaterialMemory4ECB60("B")
	memory.names = map[uint32]string{0: "A"}

	got := teamMaterialIndex4ECB60(teamMaterialInputPointer4ECB60, memory.hooks())
	if got != 0 {
		t.Fatalf("team = %d, want 0", got)
	}
	want := []string{"name:0", "byte:1:0", "byte:100:0", "name:1"}
	if !reflect.DeepEqual(memory.events, want) {
		t.Fatalf("events = %#v, want %#v", memory.events, want)
	}
}

func TestTeamMaterialIndex4ECB60CandidateBytePrecedesInputFault(t *testing.T) {
	stop := &struct{}{}
	var events []string
	var recovered any
	func() {
		defer func() { recovered = recover() }()
		teamMaterialIndex4ECB60(2, teamMaterialIndexHooks4ECB60[int, uint32]{
			loadName: func(uint32) int {
				events = append(events, "name:0")
				return 1
			},
			loadByte: func(pointer int, offset uint32) byte {
				events = append(events, fmt.Sprintf("byte:%d:%d", pointer, offset))
				if pointer == 2 {
					panic(stop)
				}
				return 'A'
			},
			loadTeam: func(uint32) uint32 { return 0 },
		})
	}()
	if recovered != stop {
		t.Fatalf("panic = %#v, want sentinel", recovered)
	}
	want := []string{"name:0", "byte:1:0", "byte:2:0"}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %#v, want %#v", events, want)
	}
}

func TestTeamMaterialIndex4ECB60CandidateFaultPrecedesInput(t *testing.T) {
	stop := &struct{}{}
	inputLoads := 0
	var recovered any
	func() {
		defer func() { recovered = recover() }()
		teamMaterialIndex4ECB60(2, teamMaterialIndexHooks4ECB60[int, uint32]{
			loadName: func(uint32) int { return 1 },
			loadByte: func(pointer int, _ uint32) byte {
				if pointer == 1 {
					panic(stop)
				}
				inputLoads++
				return 0
			},
			loadTeam: func(uint32) uint32 { return 0 },
		})
	}()
	if recovered != stop {
		t.Fatalf("panic = %#v, want sentinel", recovered)
	}
	if inputLoads != 0 {
		t.Fatalf("input loads = %d, want 0", inputLoads)
	}
}

func TestTeamMaterialIndex4ECB60SecondCandidateBytePrecedesInput(t *testing.T) {
	stop := &struct{}{}
	var events []string
	var recovered any
	func() {
		defer func() { recovered = recover() }()
		teamMaterialIndex4ECB60(2, teamMaterialIndexHooks4ECB60[int, uint32]{
			loadName: func(uint32) int { return 1 },
			loadByte: func(pointer int, offset uint32) byte {
				events = append(events, fmt.Sprintf("byte:%d:%d", pointer, offset))
				if pointer == 1 && offset == 1 {
					panic(stop)
				}
				if pointer == 2 && offset == 1 {
					t.Fatal("input second byte read after candidate fault")
				}
				return 'A'
			},
			loadTeam: func(uint32) uint32 { return 0 },
		})
	}()
	if recovered != stop {
		t.Fatalf("panic = %#v, want sentinel", recovered)
	}
	want := []string{"byte:1:0", "byte:2:0", "byte:1:1"}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %#v, want %#v", events, want)
	}
}

func TestTeamMaterialIndex4ECB60MatchLoadsLiveTeam(t *testing.T) {
	memory := newTeamMaterialMemory4ECB60("A")
	memory.names = map[uint32]string{0: "A"}
	memory.teams[0] = 1
	hooks := memory.hooks()
	hooks.loadTeam = func(index uint32) uint32 {
		memory.events = append(memory.events, fmt.Sprintf("team:%d", index))
		memory.teams[index] = 9
		return memory.teams[index]
	}

	if got := teamMaterialIndex4ECB60(teamMaterialInputPointer4ECB60, hooks); got != 9 {
		t.Fatalf("team = %d, want live value 9", got)
	}
	if memory.events[len(memory.events)-1] != "team:0" {
		t.Fatalf("last event = %q, want team:0", memory.events[len(memory.events)-1])
	}
}

func TestTeamMaterialIndex4ECB60CanContinuePastShippedSentinel(t *testing.T) {
	memory := newTeamMaterialMemory4ECB60("Z")
	memory.names = make(map[uint32]string)
	for index := uint32(0); index <= 9; index++ {
		memory.names[index] = "A"
	}
	memory.names[10] = "Z"
	memory.teams[10] = 42

	if got := teamMaterialIndex4ECB60(teamMaterialInputPointer4ECB60, memory.hooks()); got != 42 {
		t.Fatalf("team = %d, want 42", got)
	}
	wantNames := []string{
		"name:0", "name:1", "name:2", "name:3", "name:4", "name:5",
		"name:6", "name:7", "name:8", "name:9", "name:10",
	}
	var gotNames []string
	for _, event := range memory.events {
		if len(event) >= len("name:") && event[:len("name:")] == "name:" {
			gotNames = append(gotNames, event)
		}
	}
	if !reflect.DeepEqual(gotNames, wantNames) {
		t.Fatalf("name loads = %#v, want %#v", gotNames, wantNames)
	}
}
