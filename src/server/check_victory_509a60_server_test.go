package server

import (
	"math"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

func TestCheckVictory509A60NativeLayouts(t *testing.T) {
	wantObjectUpdate := uintptr(748)
	wantObjectTeam := uintptr(48)
	wantUpdatePlayer := uintptr(276)
	wantPlayerLessons := uintptr(2136)
	wantPlayerDeaths := uintptr(2140)
	wantPlayerFlags := uintptr(3680)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantObjectUpdate = 872
		wantObjectTeam = 52
		wantUpdatePlayer = 336
		wantPlayerLessons = 2140
		wantPlayerDeaths = 2144
		wantPlayerFlags = 4976
	}
	checks := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"Object.TeamVal", unsafe.Offsetof(Object{}.TeamVal), wantObjectTeam},
		{"Object.UpdateData", unsafe.Offsetof(Object{}.UpdateData), wantObjectUpdate},
		{"PlayerUpdateData.Player", unsafe.Offsetof(PlayerUpdateData{}.Player), wantUpdatePlayer},
		{"Player.Lessons", unsafe.Offsetof(Player{}.Lessons), wantPlayerLessons},
		{"Player.Field2140", unsafe.Offsetof(Player{}.Field2140), wantPlayerDeaths},
		{"Player.Field3680", unsafe.Offsetof(Player{}.Field3680), wantPlayerFlags},
		{"Team.Lessons", unsafe.Offsetof(Team{}.Lessons), 52},
		{"Team.IDVal", unsafe.Offsetof(Team{}.IDVal), 57},
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Errorf("%s = %d, want %d", check.name, check.got, check.want)
		}
	}
}

func TestCheckVictory509A60PreservesNativePointers(t *testing.T) {
	teams, freeTeams := alloc.Make([]Team(nil), 2)
	defer freeTeams()
	players, freePlayers := alloc.Make([]Player(nil), 1)
	defer freePlayers()
	unit, freeUnit := alloc.New(Object{})
	defer freeUnit()
	update, freeUpdate := alloc.New(PlayerUpdateData{})
	defer freeUpdate()

	if unsafe.Sizeof(uintptr(0)) == 8 {
		for name, address := range map[string]uintptr{
			"team":   uintptr(unsafe.Pointer(&teams[1])),
			"player": uintptr(unsafe.Pointer(&players[0])),
			"unit":   uintptr(unsafe.Pointer(unit)),
			"update": uintptr(unsafe.Pointer(update)),
		} {
			if address <= math.MaxUint32 {
				t.Fatalf("%s address = %#x, want above 4 GiB", name, address)
			}
		}
	}

	teams[1].active = 1
	teams[1].ind = 1
	teams[1].IDVal = 7
	players[0].Active = 1
	players[0].PlayerInd = 0
	players[0].PlayerUnit = unit
	unit.ObjClass = object.ClassPlayer
	unit.TeamVal.ID = 7
	unit.UpdateData = unsafe.Pointer(update)
	update.Player = &players[0]
	players[0].Field2140 = 2

	var srv Server
	srv.Teams.Arr = teams
	srv.Players.list = players
	oldFlags := noxflags.GetGame()
	noxflags.ResetGame()
	t.Cleanup(func() {
		noxflags.ResetGame()
		noxflags.SetGame(oldFlags)
	})

	noxflags.SetGame(noxflags.GameModeElimination)
	var teamWinner *Team
	srv.CheckVictory509A60(CheckVictoryRuntime509A60{
		ScoreLimit: func(flags uint16) uint16 {
			if flags != uint16(noxflags.GameModeElimination) {
				t.Fatalf("score flags = %#x", flags)
			}
			return 3
		},
		GameplayHasRivals: func() bool { return true },
		SendTeamWinner: func(team *Team, arg uint8) int32 {
			teamWinner = team
			if arg != checkVictoryWinnerArg509A60 {
				t.Fatalf("team winner arg = %d", arg)
			}
			return math.MinInt32
		},
		SendPlayerWinner: func(*Object, uint8) int32 {
			t.Fatal("unexpected player winner")
			return 0
		},
	})
	if teamWinner != &teams[1] {
		t.Fatalf("team winner = %p, want %p", teamWinner, &teams[1])
	}
	if !noxflags.HasGame(noxflags.GameFlag4) {
		t.Fatal("completion flag was not set")
	}

	noxflags.ResetGame()
	teams[1].Lessons = 4
	players[0].Lessons = 5
	var playerWinner *Object
	srv.CheckVictory509A60(CheckVictoryRuntime509A60{
		ScoreLimit: func(flags uint16) uint16 {
			if flags != 0 {
				t.Fatalf("score flags = %#x, want zero", flags)
			}
			return 5
		},
		GameplayHasRivals: func() bool {
			t.Fatal("normal score mode queried elimination rivals")
			return false
		},
		SendTeamWinner: func(*Team, uint8) int32 {
			t.Fatal("unexpected team winner")
			return 0
		},
		SendPlayerWinner: func(unit *Object, arg uint8) int32 {
			playerWinner = unit
			if arg != checkVictoryWinnerArg509A60 {
				t.Fatalf("player winner arg = %d", arg)
			}
			return math.MaxInt32
		},
	})
	if playerWinner != unit {
		t.Fatalf("player winner = %p, want %p", playerWinner, unit)
	}
	if !noxflags.HasGame(noxflags.GameFlag4) {
		t.Fatal("completion flag was not set")
	}

	runtime.KeepAlive(teams)
	runtime.KeepAlive(players)
	runtime.KeepAlive(unit)
	runtime.KeepAlive(update)
}
