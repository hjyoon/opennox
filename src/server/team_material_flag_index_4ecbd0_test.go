package server

import (
	"fmt"
	"reflect"
	"testing"
)

const (
	teamMaterialModifierPointer4ECC00 = 200
	teamMaterialInputPointer4ECC00    = 100
)

type teamMaterialModifierMemory4ECC00 struct {
	input  string
	names  map[uint32]string
	teams  map[uint32]uint32
	events []string
}

func newTeamMaterialModifierMemory4ECC00(input string) *teamMaterialModifierMemory4ECC00 {
	memory := &teamMaterialModifierMemory4ECC00{
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

func (m *teamMaterialModifierMemory4ECC00) loadName(index uint32) int {
	m.events = append(m.events, fmt.Sprintf("name:%d", index))
	if _, ok := m.names[index]; !ok {
		return 0
	}
	return int(index) + 1
}

func (m *teamMaterialModifierMemory4ECC00) loadInputName(modifier int) int {
	m.events = append(m.events, fmt.Sprintf("input-name:%d", modifier))
	if modifier != teamMaterialModifierPointer4ECC00 {
		panic("unexpected modifier pointer")
	}
	return teamMaterialInputPointer4ECC00
}

func (m *teamMaterialModifierMemory4ECC00) loadInputByte(pointer int, offset uint32) byte {
	m.events = append(m.events, fmt.Sprintf("input:%d:%d", pointer, offset))
	if pointer != teamMaterialInputPointer4ECC00 {
		panic("unexpected input pointer")
	}
	return teamMaterialCStringByte4ECC00(m.input, offset)
}

func (m *teamMaterialModifierMemory4ECC00) loadCandidateByte(pointer int, offset uint32) byte {
	m.events = append(m.events, fmt.Sprintf("candidate:%d:%d", pointer, offset))
	if pointer == 0 {
		panic("nil candidate name")
	}
	return teamMaterialCStringByte4ECC00(m.names[uint32(pointer-1)], offset)
}

func (m *teamMaterialModifierMemory4ECC00) loadTeam(index uint32) uint32 {
	m.events = append(m.events, fmt.Sprintf("team:%d", index))
	return m.teams[index]
}

func (m *teamMaterialModifierMemory4ECC00) hooks() teamMaterialModifierIndexHooks4ECC00[int, int, int, uint32] {
	return teamMaterialModifierIndexHooks4ECC00[int, int, int, uint32]{
		loadName:          m.loadName,
		loadInputName:     m.loadInputName,
		loadInputByte:     m.loadInputByte,
		loadCandidateByte: m.loadCandidateByte,
		loadTeam:          m.loadTeam,
	}
}

func teamMaterialCStringByte4ECC00(text string, offset uint32) byte {
	if uint64(offset) > uint64(len(text)) {
		panic("read beyond terminating NUL")
	}
	if uint64(offset) == uint64(len(text)) {
		return 0
	}
	return text[offset]
}

func TestTeamMaterialModifierIndex4ECC00NilModifierReturnsBeforeTable(t *testing.T) {
	got := teamMaterialModifierIndex4ECC00(0, teamMaterialModifierIndexHooks4ECC00[int, int, int, uint32]{
		loadName: func(uint32) int {
			t.Fatal("table read after nil modifier")
			return 0
		},
		loadInputName: func(int) int {
			t.Fatal("input-name read after nil modifier")
			return 0
		},
		loadInputByte: func(int, uint32) byte {
			t.Fatal("input byte read after nil modifier")
			return 0
		},
		loadCandidateByte: func(int, uint32) byte {
			t.Fatal("candidate byte read after nil modifier")
			return 0
		},
		loadTeam: func(uint32) uint32 {
			t.Fatal("team read after nil modifier")
			return 0
		},
	})
	if got != 0 {
		t.Fatalf("team = %d, want 0", got)
	}
}

func TestTeamMaterialModifierIndex4ECC00HeadGatePrecedesInputName(t *testing.T) {
	var events []string
	got := teamMaterialModifierIndex4ECC00(1, teamMaterialModifierIndexHooks4ECC00[int, int, int, uint32]{
		loadName: func(index uint32) int {
			events = append(events, fmt.Sprintf("name:%d", index))
			return 0
		},
		loadInputName: func(int) int {
			t.Fatal("input-name read after nil head")
			return 0
		},
		loadInputByte: func(int, uint32) byte {
			t.Fatal("input byte read after nil head")
			return 0
		},
		loadCandidateByte: func(int, uint32) byte {
			t.Fatal("candidate byte read after nil head")
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

func TestTeamMaterialModifierIndex4ECC00TwoByteCompareAndRestart(t *testing.T) {
	memory := newTeamMaterialModifierMemory4ECC00("AB")
	memory.names = map[uint32]string{0: "AX", 1: "AB"}
	memory.teams[1] = 3

	got := teamMaterialModifierIndex4ECC00(teamMaterialModifierPointer4ECC00, memory.hooks())
	if got != 3 {
		t.Fatalf("team = %d, want 3", got)
	}
	want := []string{
		"name:0", "input-name:200",
		"input:100:0", "candidate:1:0",
		"input:100:1", "candidate:1:1",
		"name:1",
		"input:100:0", "candidate:2:0",
		"input:100:1", "candidate:2:1",
		"input:100:2", "candidate:2:2",
		"team:1",
	}
	if !reflect.DeepEqual(memory.events, want) {
		t.Fatalf("events = %#v, want %#v", memory.events, want)
	}
}

func TestTeamMaterialModifierIndex4ECC00FirstByteNULSkipsSecondPair(t *testing.T) {
	memory := newTeamMaterialModifierMemory4ECC00("")
	memory.names = map[uint32]string{0: ""}
	memory.teams[0] = 7
	hooks := memory.hooks()
	hooks.loadName = func(index uint32) int {
		memory.events = append(memory.events, fmt.Sprintf("name:%d", index))
		if index == 0 {
			return 1
		}
		return 0
	}

	got := teamMaterialModifierIndex4ECC00(teamMaterialModifierPointer4ECC00, hooks)
	if got != 7 {
		t.Fatalf("team = %d, want 7", got)
	}
	want := []string{
		"name:0", "input-name:200",
		"input:100:0", "candidate:1:0", "team:0",
	}
	if !reflect.DeepEqual(memory.events, want) {
		t.Fatalf("events = %#v, want %#v", memory.events, want)
	}
}

func TestTeamMaterialModifierIndex4ECC00SentinelStopsBeforeBytesAndTeam(t *testing.T) {
	memory := newTeamMaterialModifierMemory4ECC00("B")
	memory.names = map[uint32]string{0: "A"}

	got := teamMaterialModifierIndex4ECC00(teamMaterialModifierPointer4ECC00, memory.hooks())
	if got != 0 {
		t.Fatalf("team = %d, want 0", got)
	}
	want := []string{
		"name:0", "input-name:200",
		"input:100:0", "candidate:1:0", "name:1",
	}
	if !reflect.DeepEqual(memory.events, want) {
		t.Fatalf("events = %#v, want %#v", memory.events, want)
	}
}

func TestTeamMaterialModifierIndex4ECC00InputFaultPrecedesCandidate(t *testing.T) {
	stop := &struct{}{}
	var events []string
	var recovered any
	func() {
		defer func() { recovered = recover() }()
		teamMaterialModifierIndex4ECC00(2, teamMaterialModifierIndexHooks4ECC00[int, int, int, uint32]{
			loadName: func(uint32) int {
				events = append(events, "name:0")
				return 1
			},
			loadInputName: func(int) int {
				events = append(events, "input-name")
				return 3
			},
			loadInputByte: func(int, uint32) byte {
				events = append(events, "input:0")
				panic(stop)
			},
			loadCandidateByte: func(int, uint32) byte {
				t.Fatal("candidate read after input fault")
				return 0
			},
			loadTeam: func(uint32) uint32 { return 0 },
		})
	}()
	if recovered != stop {
		t.Fatalf("panic = %#v, want sentinel", recovered)
	}
	want := []string{"name:0", "input-name", "input:0"}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %#v, want %#v", events, want)
	}
}

func TestTeamMaterialModifierIndex4ECC00CandidateFaultFollowsInput(t *testing.T) {
	stop := &struct{}{}
	var events []string
	var recovered any
	func() {
		defer func() { recovered = recover() }()
		teamMaterialModifierIndex4ECC00(2, teamMaterialModifierIndexHooks4ECC00[int, int, int, uint32]{
			loadName:      func(uint32) int { return 1 },
			loadInputName: func(int) int { return 3 },
			loadInputByte: func(int, uint32) byte {
				events = append(events, "input:0")
				return 'A'
			},
			loadCandidateByte: func(int, uint32) byte {
				events = append(events, "candidate:0")
				panic(stop)
			},
			loadTeam: func(uint32) uint32 { return 0 },
		})
	}()
	if recovered != stop {
		t.Fatalf("panic = %#v, want sentinel", recovered)
	}
	want := []string{"input:0", "candidate:0"}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %#v, want %#v", events, want)
	}
}

func TestTeamMaterialModifierIndex4ECC00SecondInputBytePrecedesCandidate(t *testing.T) {
	stop := &struct{}{}
	var events []string
	var recovered any
	func() {
		defer func() { recovered = recover() }()
		teamMaterialModifierIndex4ECC00(2, teamMaterialModifierIndexHooks4ECC00[int, int, int, uint32]{
			loadName:      func(uint32) int { return 1 },
			loadInputName: func(int) int { return 3 },
			loadInputByte: func(_ int, offset uint32) byte {
				events = append(events, fmt.Sprintf("input:%d", offset))
				if offset == 1 {
					panic(stop)
				}
				return 'A'
			},
			loadCandidateByte: func(_ int, offset uint32) byte {
				events = append(events, fmt.Sprintf("candidate:%d", offset))
				if offset == 1 {
					t.Fatal("second candidate byte read after input fault")
				}
				return 'A'
			},
			loadTeam: func(uint32) uint32 { return 0 },
		})
	}()
	if recovered != stop {
		t.Fatalf("panic = %#v, want sentinel", recovered)
	}
	want := []string{"input:0", "candidate:0", "input:1"}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %#v, want %#v", events, want)
	}
}

func TestTeamMaterialModifierIndex4ECC00MatchLoadsLiveTeam(t *testing.T) {
	memory := newTeamMaterialModifierMemory4ECC00("A")
	memory.names = map[uint32]string{0: "A"}
	memory.teams[0] = 1
	hooks := memory.hooks()
	hooks.loadTeam = func(index uint32) uint32 {
		memory.events = append(memory.events, fmt.Sprintf("team:%d", index))
		memory.teams[index] = 9
		return memory.teams[index]
	}

	if got := teamMaterialModifierIndex4ECC00(teamMaterialModifierPointer4ECC00, hooks); got != 9 {
		t.Fatalf("team = %d, want live value 9", got)
	}
	if memory.events[len(memory.events)-1] != "team:0" {
		t.Fatalf("last event = %q, want team:0", memory.events[len(memory.events)-1])
	}
}

func TestTeamMaterialModifierIndex4ECC00CanContinuePastShippedSentinel(t *testing.T) {
	memory := newTeamMaterialModifierMemory4ECC00("Z")
	memory.names = make(map[uint32]string)
	for index := uint32(0); index <= 9; index++ {
		memory.names[index] = "A"
	}
	memory.names[10] = "Z"
	memory.teams[10] = 42

	if got := teamMaterialModifierIndex4ECC00(teamMaterialModifierPointer4ECC00, memory.hooks()); got != 42 {
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

func TestTeamMaterialObjectIndex4ECBD0ClassGateAndLoadOrder(t *testing.T) {
	var events []string
	hooks := teamMaterialObjectIndexHooks4ECBD0[int, int, int, uint32]{
		loadClass: func(obj int) uint32 {
			events = append(events, fmt.Sprintf("class:%d", obj))
			return teamMaterialFlagClass4ECBD0
		},
		loadInitData: func(obj int) int {
			events = append(events, fmt.Sprintf("init:%d", obj))
			return 20
		},
		loadSecondModifier: func(data int) int {
			events = append(events, fmt.Sprintf("modifier:%d", data))
			return 30
		},
		lookup: func(modifier int) uint32 {
			events = append(events, fmt.Sprintf("lookup:%d", modifier))
			return 9
		},
	}
	if got := teamMaterialObjectIndex4ECBD0(10, hooks); got != 9 {
		t.Fatalf("team = %d, want 9", got)
	}
	want := []string{"class:10", "init:10", "modifier:20", "lookup:30"}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %#v, want %#v", events, want)
	}

	events = nil
	hooks.loadClass = func(obj int) uint32 {
		events = append(events, fmt.Sprintf("class:%d", obj))
		return 0
	}
	if got := teamMaterialObjectIndex4ECBD0(10, hooks); got != 0 {
		t.Fatalf("non-flag team = %d, want 0", got)
	}
	if want := []string{"class:10"}; !reflect.DeepEqual(events, want) {
		t.Fatalf("non-flag events = %#v, want %#v", events, want)
	}
}

func TestTeamMaterialObjectIndex4ECBD0PassesNilModifierToLookup(t *testing.T) {
	var gotModifier int
	got := teamMaterialObjectIndex4ECBD0(1, teamMaterialObjectIndexHooks4ECBD0[int, int, int, uint32]{
		loadClass:          func(int) uint32 { return teamMaterialFlagClass4ECBD0 },
		loadInitData:       func(int) int { return 2 },
		loadSecondModifier: func(int) int { return 0 },
		lookup: func(modifier int) uint32 {
			gotModifier = modifier
			return 4
		},
	})
	if got != 4 || gotModifier != 0 {
		t.Fatalf("team/modifier = %d/%d, want 4/0", got, gotModifier)
	}
}
