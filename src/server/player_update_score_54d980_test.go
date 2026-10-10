package server

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
)

const arenaScoreNil54D980 = math.MaxUint32

type arenaScoreSnapshot54D980 struct {
	Parameters []uint32     `json:"parameters"`
	Fault      bool         `json:"fault"`
	Events     [][]uint32   `json:"events"`
	Objects    [3][3]uint32 `json:"objects"`
	Updates    [6]uint32    `json:"updates"`
	Players    [6][2]uint32 `json:"players"`
	Teams      [2]uint32    `json:"teams"`
	Mode       uint32       `json:"mode"`
}

func arenaScoreRun54D980(parameters []uint32) arenaScoreSnapshot54D980 {
	var units [3]*Object
	var updates [6]*PlayerUpdateData
	var players [6]*Player
	teams := [3]*Team{nil, {}, {}}
	mode := parameters[6]
	snapshot := arenaScoreSnapshot54D980{Parameters: parameters, Events: make([][]uint32, 0)}
	for i := range players {
		players[i] = &Player{Lessons: -1, Field2140: 0x81234560 + uint32(i)}
		if parameters[7] != 0 {
			players[i].Lessons = math.MaxInt32
		}
		players[i].Lessons = int32(uint32(players[i].Lessons) ^ uint32(i*0x10203))
		updates[i] = &PlayerUpdateData{Player: players[i]}
	}
	players[0].Field2140 = math.MaxUint32
	for i := range units {
		units[i] = &Object{ObjClass: 0x80000004, UpdateData: unsafe.Pointer(updates[i]), TeamVal: ObjectTeam{ID: TeamID(parameters[2+i])}}
	}
	units[2].ObjClass = 0x80000002
	teams[1].Lessons, teams[2].Lessons = -1, math.MaxInt32
	if parameters[7] != 0 {
		teams[1].Lessons, teams[2].Lessons = math.MaxInt32, 0
	}
	switch parameters[9] {
	case 1:
		units[0].UpdateData = nil
	case 2:
		units[1].UpdateData = nil
	case 3:
		updates[1].Player = nil
	case 4:
		updates[0].Player = nil
	}
	unitID := func(unit *Object) uint32 {
		for i, value := range units {
			if unit == value {
				return uint32(i)
			}
		}
		panic("unknown unit")
	}
	playerID := func(player *Player) uint32 {
		if player == nil {
			return arenaScoreNil54D980
		}
		for i, value := range players {
			if player == value {
				return uint32(i)
			}
		}
		panic("unknown player")
	}
	event := func(value []uint32) {
		snapshot.Events = append(snapshot.Events, value)
		if parameters[8] == 0 || uint32(len(snapshot.Events)) != parameters[8] {
			return
		}
		for i := range units {
			units[i].UpdateData = unsafe.Pointer(updates[i+3])
			updates[i].Player = players[(i+1)%3]
		}
		mode, teams[1].Lessons, teams[2].Lessons = 0x80000000, math.MinInt32, -1
	}
	rt := PlayerUpdateScoreRuntime54D980{
		HasTeam: func(team *ObjectTeam) bool {
			for i, unit := range units {
				if team == &unit.TeamVal {
					value := team.ID != 0
					event([]uint32{1, uint32(i)})
					return value
				}
			}
			panic("unknown object team")
		},
		TeamByID: func(id uint8) *Team {
			event([]uint32{2, uint32(id)})
			return teams[id]
		},
		AddScore: func(unit *Object, value uint32) {
			event([]uint32{3, unitID(unit), value})
			PlayerScoreAdd4D8E90(unit, value)
		},
		SubtractScore: func(unit *Object, value uint32) {
			event([]uint32{4, unitID(unit), value})
			PlayerScoreSubtract4D8EC0(unit, value)
		},
		ReportLesson: func(unit *Object) {
			player := (*PlayerUpdateData)(unit.UpdateData).Player
			event([]uint32{5, unitID(unit), uint32(player.Lessons), player.Field2140})
		},
		IncrementElimDeath: func(unit *Object) {
			event([]uint32{6, unitID(unit)})
			if unit.ObjClass.Has(object.ClassPlayer) {
				(*PlayerUpdateData)(unit.UpdateData).Player.Field2140++
			}
		},
		TeamChangeLessons: func(team *Team, value int32) {
			for i, current := range teams {
				if team == current {
					event([]uint32{7, uint32(i), uint32(value)})
					team.Lessons = value
					return
				}
			}
			panic("unknown team")
		},
		ObserverMode: func() uint32 {
			event([]uint32{9, mode})
			return mode
		},
		ObserverUpdate: func(first, second *Player) {
			event([]uint32{8, playerID(first), playerID(second)})
		},
	}
	arg := func(index uint32) *Object {
		if index == arenaScoreNil54D980 {
			return nil
		}
		return units[index]
	}
	func() {
		defer func() { snapshot.Fault = recover() != nil }()
		PlayerUpdateScore54D980(units[0], arg(parameters[0]), arg(parameters[1]), parameters[5], rt)
	}()
	for i, unit := range units {
		updateID := uint32(arenaScoreNil54D980)
		for j, update := range updates {
			if unit.UpdateData == unsafe.Pointer(update) {
				updateID = uint32(j)
			}
		}
		snapshot.Objects[i] = [3]uint32{uint32(unit.ObjClass), uint32(unit.TeamVal.ID), updateID}
	}
	for i := range updates {
		snapshot.Updates[i] = playerID(updates[i].Player)
		snapshot.Players[i] = [2]uint32{uint32(players[i].Lessons), players[i].Field2140}
	}
	snapshot.Teams = [2]uint32{uint32(teams[1].Lessons), uint32(teams[2].Lessons)}
	snapshot.Mode = mode
	return snapshot
}

func TestPlayerUpdateScore54D980OriginalTranscript(t *testing.T) {
	var transcript []byte
	cases, faults := 0, 0
	run := func(parameters []uint32) {
		snapshot := arenaScoreRun54D980(parameters)
		data, err := json.Marshal(snapshot)
		if err != nil {
			t.Fatal(err)
		}
		transcript = append(transcript, data...)
		transcript = append(transcript, '\n')
		cases++
		if snapshot.Fault {
			faults++
		}
	}
	for _, killer := range []uint32{arenaScoreNil54D980, 0, 1, 2} {
		for _, assist := range []uint32{arenaScoreNil54D980, 0, 1, 2} {
			for tv := uint32(0); tv < 3; tv++ {
				for tk := uint32(0); tk < 3; tk++ {
					for ta := uint32(0); ta < 3; ta++ {
						for _, tracking := range []uint32{0, 0x80000000} {
							for _, mode := range []uint32{0, 0x80000000} {
								for variant := uint32(0); variant < 2; variant++ {
									run([]uint32{killer, assist, tv, tk, ta, tracking, mode, variant, 0, 0})
								}
							}
						}
					}
				}
			}
		}
	}
	for _, parameters := range [][]uint32{
		{1, 2, 0, 0, 0, 0x80000000, 0x80000000, 0},
		{1, 2, 1, 2, 1, 0x80000000, 0x80000000, 1},
		{0, 1, 1, 1, 2, 0x80000000, 0x80000000, 0},
		{arenaScoreNil54D980, 1, 1, 2, 0, 0, 0, 1},
	} {
		for mutation := uint32(1); mutation < 12; mutation++ {
			run(append(append([]uint32{}, parameters...), mutation, 0))
		}
	}
	for invalid := uint32(1); invalid < 5; invalid++ {
		for _, pair := range [][2]uint32{{arenaScoreNil54D980, arenaScoreNil54D980}, {0, 1}, {1, 2}, {2, 1}} {
			run([]uint32{pair[0], pair[1], 0, 0, 0, 0x80000000, 0x80000000, 0, 0, invalid})
		}
	}
	if path := os.Getenv("NOX_ARENA_SCORE_NATIVE_JSON"); path != "" {
		if err := os.WriteFile(path, transcript, 0600); err != nil {
			t.Fatal(err)
		}
	}
	// Frozen unchanged GAME.EXE 0054D980, with the real 004D8E90/004D8EC0
	// instructions. This covers full snapshots, service input/order, cached vs
	// live pointers, overflow, any-bit class tests, and declared service faults.
	if got := fmt.Sprintf("%x", sha256.Sum256(transcript)); cases != 3516 || got != "48d01a9df52239fe8a022642c7709ade96e2d18dd15bfb72553a926b38cd47ba" {
		t.Fatalf("original Arena transcript: cases=%d faults=%d sha256=%s", cases, faults, got)
	}
	t.Logf("unchanged PE32 whole snapshots: %d cases, %d fault boundaries", cases, faults)
}

func TestPlayerScore54D980ScalarWrapAndReturn(t *testing.T) {
	for _, value := range []uint32{0, 1, 0x80000000, math.MaxUint32} {
		for _, class := range []object.Class{object.ClassMonster, object.ClassPlayer, 0x80000004} {
			player := &Player{Lessons: math.MaxInt32}
			update := &PlayerUpdateData{Player: player}
			unit := &Object{ObjClass: class, UpdateData: unsafe.Pointer(update)}
			wantReturn := unsafe.Pointer(unit)
			want := uint32(math.MaxInt32)
			if class.Has(object.ClassPlayer) {
				wantReturn, want = unsafe.Pointer(player), want+value
			}
			if result := PlayerScoreAdd4D8E90(unit, value); result != wantReturn || uint32(player.Lessons) != want {
				t.Fatalf("add class/value=%x/%x return=%p score=%x", class, value, result, player.Lessons)
			}
			if result := PlayerScoreSubtract4D8EC0(unit, value); result != wantReturn || player.Lessons != math.MaxInt32 {
				t.Fatalf("subtract class/value=%x/%x return=%p score=%x", class, value, result, player.Lessons)
			}
		}
	}
	unit := &Object{ObjClass: object.ClassMonster}
	if PlayerScoreAdd4D8E90(unit, 1) != unsafe.Pointer(unit) || PlayerScoreSubtract4D8EC0(unit, 1) != unsafe.Pointer(unit) {
		t.Fatal("non-player nil update should not be dereferenced")
	}
}

func TestPlayerScore54D980ScalarOriginalTranscript(t *testing.T) {
	var transcript []byte
	cases, faults := 0, 0
	run := func(operation uint32, class object.Class, value, invalid uint32) {
		player := &Player{Lessons: math.MaxInt32}
		update := &PlayerUpdateData{Player: player}
		unit := &Object{ObjClass: class, UpdateData: unsafe.Pointer(update)}
		switch invalid {
		case 1:
			unit.UpdateData = nil
		case 2:
			update.Player = nil
		case 3:
			unit = nil
		}
		snapshot := struct {
			Parameters []uint32 `json:"parameters"`
			Fault      bool     `json:"fault"`
			Score      uint32   `json:"score"`
			Return     uint32   `json:"return"`
		}{Parameters: []uint32{operation, uint32(class), value, invalid}, Return: arenaScoreNil54D980}
		var result unsafe.Pointer
		func() {
			defer func() { snapshot.Fault = recover() != nil }()
			if operation == 0 {
				result = PlayerScoreAdd4D8E90(unit, value)
			} else {
				result = PlayerScoreSubtract4D8EC0(unit, value)
			}
		}()
		snapshot.Score = uint32(player.Lessons)
		if snapshot.Fault {
			faults++
		} else if result == unsafe.Pointer(unit) {
			snapshot.Return = 0
		} else if result == unsafe.Pointer(player) {
			snapshot.Return = 1
		} else {
			t.Fatalf("unknown original scalar return %p", result)
		}
		data, err := json.Marshal(snapshot)
		if err != nil {
			t.Fatal(err)
		}
		transcript = append(transcript, data...)
		transcript = append(transcript, '\n')
		cases++
	}
	for _, value := range []uint32{0, 1, 0x80000000, math.MaxUint32} {
		for _, class := range []object.Class{object.ClassMonster, object.ClassPlayer, 0x80000004} {
			for operation := uint32(0); operation < 2; operation++ {
				run(operation, class, value, 0)
			}
		}
	}
	for invalid := uint32(1); invalid < 3; invalid++ {
		for _, class := range []object.Class{object.ClassMonster, object.ClassPlayer, 0x80000004} {
			for operation := uint32(0); operation < 2; operation++ {
				run(operation, class, 1, invalid)
			}
		}
	}
	for operation := uint32(0); operation < 2; operation++ {
		run(operation, object.ClassPlayer, 1, 3)
	}
	if path := os.Getenv("NOX_ARENA_SCORE_SCALAR_NATIVE_JSON"); path != "" {
		if err := os.WriteFile(path, transcript, 0600); err != nil {
			t.Fatal(err)
		}
	}
	// Frozen executed 004D8E90/004D8EC0 bytes. Invalid update/player and nil
	// unit faults are from these actual instructions, not intercepted services.
	if got := fmt.Sprintf("%x", sha256.Sum256(transcript)); cases != 38 || faults != 10 || got != "d59f6998a1c4fd9c22bae0f5a83228da7892fdca2780929013fc078cdf30794a" {
		t.Fatalf("original scalar transcript: cases=%d faults=%d sha256=%s", cases, faults, got)
	}
}
