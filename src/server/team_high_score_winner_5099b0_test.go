package server

import (
	"encoding/binary"
	"fmt"
	"reflect"
	"testing"

	"github.com/opennox/libs/noxnet/netmsg"
)

type teamScoreTest5099B0 struct {
	name  string
	score int32
}

func TestTeamHighScoreWinner5099B0CandidateRulesAndModeOrder(t *testing.T) {
	tests := []struct {
		name     string
		scores   []int32
		flagBall bool
		winner   string
	}{
		{name: "none flag", winner: "flag:nil:1"},
		{name: "below threshold ignored", scores: []int32{-2}, winner: "flag:nil:1"},
		{name: "one team", scores: []int32{3}, winner: "flag:t0:1"},
		{name: "tie", scores: []int32{3, 3}, winner: "flag:nil:1"},
		{name: "higher clears tie", scores: []int32{3, 3, 4}, winner: "flag:t2:1"},
		{name: "flagball", scores: []int32{4}, flagBall: true, winner: "ball:t0"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var teams []*teamScoreTest5099B0
			for i, score := range tc.scores {
				teams = append(teams, &teamScoreTest5099B0{name: fmt.Sprintf("t%d", i), score: score})
			}
			var events []string
			hooks := teamHighScoreWinnerHooks5099B0[*teamScoreTest5099B0]{
				firstTeam: func() *teamScoreTest5099B0 {
					if len(teams) == 0 {
						return nil
					}
					return teams[0]
				},
				nextTeam: func(team *teamScoreTest5099B0) *teamScoreTest5099B0 {
					i := indexHighScore5098A0(teams, team) + 1
					if i <= 0 || i >= len(teams) {
						return nil
					}
					return teams[i]
				},
				loadTeamScore: func(team *teamScoreTest5099B0) int32 { return team.score },
				setGameFlags:  func(flags uint32) { events = append(events, fmt.Sprintf("set:%x", flags)) },
				hasGameFlags: func(mask uint32) bool {
					events = append(events, fmt.Sprintf("mode:%x", mask))
					return tc.flagBall
				},
				sendFlagBall: func(team *teamScoreTest5099B0) int32 {
					name := "nil"
					if team != nil {
						name = team.name
					}
					events = append(events, "ball:"+name)
					return 17
				},
				sendFlagWinner: func(team *teamScoreTest5099B0, arg uint8) int32 {
					name := "nil"
					if team != nil {
						name = team.name
					}
					events = append(events, fmt.Sprintf("flag:%s:%d", name, arg))
					return 23
				},
			}
			got := teamHighScoreWinner5099B0(hooks)
			wantResult := int32(23)
			if tc.flagBall {
				wantResult = 17
			}
			if got != wantResult || events[len(events)-1] != tc.winner {
				t.Fatalf("result/events = %d/%q, want %d/%q", got, events, wantResult, tc.winner)
			}
			if len(events) < 3 || events[len(events)-3] != "set:8" || events[len(events)-2] != "mode:40" {
				t.Fatalf("flag/mode order = %q", events)
			}
		})
	}
}

func TestTeamWinnerPacket5099B0(t *testing.T) {
	packet := teamWinnerPacket5099B0(netmsg.MSG_REPORT_FLAG_WINNER, 0x12, 1, 0x78563412)
	want := [8]byte{0x57, 0x12, 0, 1, 0x12, 0x34, 0x56, 0x78}
	if !reflect.DeepEqual(packet, want) {
		t.Fatalf("packet = %x, want %x", packet, want)
	}
	if got := binary.LittleEndian.Uint32(packet[4:]); got != 0x78563412 {
		t.Fatalf("frame = %#x", got)
	}
	team := &Team{IDVal: TeamID(0xfe)}
	if got := teamWinnerID5099B0(team, 0); got != 0xfe {
		t.Fatalf("team ID = %#x", got)
	}
	if got := teamWinnerID5099B0(nil, ^uint16(0)); got != 0xffff {
		t.Fatalf("nil team ID = %#x", got)
	}
}
