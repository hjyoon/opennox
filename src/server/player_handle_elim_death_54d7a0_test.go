package server

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"math"
	"os"
	goruntime "runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
)

const elimScoreNil54D7A0 = math.MaxUint32

type elimScoreSnapshot54D7A0 struct {
	Parameters []uint32     `json:"parameters"`
	Fault      bool         `json:"fault"`
	Events     [][]uint32   `json:"events"`
	Objects    [3][3]uint32 `json:"objects"`
	Updates    [6]uint32    `json:"updates"`
	Players    [6][2]uint32 `json:"players"`
	Teams      [2]uint32    `json:"teams"`
	Mode       uint32       `json:"mode"`
}

func elimScoreRun54D7A0(parameters []uint32) elimScoreSnapshot54D7A0 {
	var units [3]*Object
	var updates [6]*PlayerUpdateData
	var players [6]*Player
	teams := [3]*Team{nil, {}, {}}
	mode := parameters[4]
	snapshot := elimScoreSnapshot54D7A0{Parameters: parameters, Events: make([][]uint32, 0)}
	for i := range players {
		score := uint32(math.MaxUint32)
		if parameters[5] != 0 {
			score = math.MaxInt32
		}
		players[i] = &Player{Lessons: int32(score ^ uint32(i*0x10203)), Field2140: 0x81234560 + uint32(i)}
		updates[i] = &PlayerUpdateData{Player: players[i]}
	}
	players[0].Field2140 = math.MaxUint32
	for i := range units {
		units[i] = &Object{ObjClass: 0x80000004, UpdateData: unsafe.Pointer(updates[i])}
	}
	units[0].TeamVal.ID = TeamID(parameters[1])
	units[1].TeamVal.ID = TeamID(parameters[2])
	units[1].ObjClass = object.Class(parameters[3])
	units[2].ObjClass = 0x80000002
	teams[1].Lessons, teams[2].Lessons = -1, math.MaxInt32
	if parameters[5] != 0 {
		teams[1].Lessons, teams[2].Lessons = math.MaxInt32, 0
	}
	switch parameters[7] {
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
			return elimScoreNil54D7A0
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
		if parameters[6] == 0 || uint32(len(snapshot.Events)) != parameters[6] {
			return
		}
		for i := range units {
			units[i].UpdateData = unsafe.Pointer(updates[i+3])
			updates[i].Player = players[(i+1)%3]
			units[i].TeamVal.ID = (units[i].TeamVal.ID + 1) % 3
		}
		units[1].ObjClass ^= 6
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
			if id == 2 && parameters[8] != 0 {
				return teams[1]
			}
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
	var killer *Object
	if parameters[0] != elimScoreNil54D7A0 {
		killer = units[parameters[0]]
	}
	victim := units[0]
	if parameters[7] == 5 {
		victim = nil
	}
	func() {
		defer func() {
			if failure := recover(); failure != nil {
				// Only an actual runtime nil-chain fault is an expected fault.
				// Fixture assertions and unknown-identity panics must escape.
				if _, ok := failure.(goruntime.Error); !ok {
					panic(failure)
				}
				snapshot.Fault = true
			}
		}()
		PlayerHandleElimDeath54D7A0(victim, killer, rt)
	}()
	for i, unit := range units {
		updateID := uint32(elimScoreNil54D7A0)
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

func TestPlayerHandleElimDeath54D7A0OriginalTranscript(t *testing.T) {
	var transcript []byte
	cases, faults := 0, 0
	run := func(parameters []uint32) {
		snapshot := elimScoreRun54D7A0(parameters)
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
	for _, killer := range []uint32{elimScoreNil54D7A0, 0, 1, 2} {
		for tv := uint32(0); tv < 3; tv++ {
			for tk := uint32(0); tk < 3; tk++ {
				for _, class := range []uint32{0, 2, 4, 6, 0x80000000, 0x80000002, 0x80000004, math.MaxUint32} {
					for _, mode := range []uint32{0, 0x80000000} {
						for variant := uint32(0); variant < 2; variant++ {
							for alias := uint32(0); alias < 2; alias++ {
								run([]uint32{killer, tv, tk, class, mode, variant, 0, 0, alias})
							}
						}
					}
				}
			}
		}
	}
	for _, parameters := range [][]uint32{
		{1, 0, 0, 0x80000004, 0, 0},
		{1, 1, 2, 0x80000004, 0x80000000, 1},
		{1, 1, 1, 0x80000004, 0x80000000, 0},
		{0, 1, 1, 0x80000004, 0x80000000, 1},
		{elimScoreNil54D7A0, 1, 2, 0x80000004, 0x80000000, 0},
		{2, 1, 1, 0x80000004, 0, 1},
	} {
		for mutation := uint32(1); mutation < 13; mutation++ {
			input := append(append([]uint32(nil), parameters...), mutation, 0, 0)
			run(input)
		}
	}
	for invalid := uint32(1); invalid < 6; invalid++ {
		for _, killer := range []uint32{elimScoreNil54D7A0, 0, 1, 2} {
			for _, mode := range []uint32{0, 0x80000000} {
				for _, team := range [][2]uint32{{0, 0}, {1, 1}, {1, 2}} {
					run([]uint32{killer, team[0], team[1], 0x80000004, mode, 0, 0, invalid, 0})
				}
			}
		}
	}
	if path := os.Getenv("NOX_ELIM_SCORE_NATIVE_SNAPSHOT"); path != "" {
		if err := os.WriteFile(path, transcript, 0600); err != nil {
			t.Fatal(err)
		}
	}
	const originalSHA = "9d1b4f195e611b9e4fd23983e116802a3b9fad0f04d2acac137b8a1ac506a9bb"
	if cases != 2496 || faults != 84 || fmt.Sprintf("%x", sha256.Sum256(transcript)) != originalSHA {
		t.Fatalf("original Elimination transcript cases=%d faults=%d SHA=%x", cases, faults, sha256.Sum256(transcript))
	}
	t.Logf("unchanged PE32 root/scalar transcript: cases=%d faults=%d SHA=%s", cases, faults, originalSHA)
}
