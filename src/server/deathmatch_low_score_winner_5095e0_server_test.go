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

func TestDeathmatchLowScoreWinner5095E0NativeLayouts(t *testing.T) {
	ptrSize := unsafe.Sizeof(uintptr(0))
	wantObjectUpdate := uintptr(748)
	wantUpdatePlayer := uintptr(276)
	wantPlayerLessons := uintptr(2136)
	wantPlayerScore := uintptr(2140)
	wantPlayerFlags := uintptr(3680)
	wantTeamSize := uintptr(80)
	if ptrSize == 8 {
		wantObjectUpdate = 872
		wantUpdatePlayer = 336
		wantPlayerLessons = 2140
		wantPlayerScore = 2144
		wantPlayerFlags = 4976
		wantTeamSize = 88
	}
	checks := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"Object.UpdateData", unsafe.Offsetof(Object{}.UpdateData), wantObjectUpdate},
		{"PlayerUpdateData.Player", unsafe.Offsetof(PlayerUpdateData{}.Player), wantUpdatePlayer},
		{"Player.Lessons", unsafe.Offsetof(Player{}.Lessons), wantPlayerLessons},
		{"Player.Field2140", unsafe.Offsetof(Player{}.Field2140), wantPlayerScore},
		{"Player.Field3680", unsafe.Offsetof(Player{}.Field3680), wantPlayerFlags},
		{"Team.Lessons", unsafe.Offsetof(Team{}.Lessons), 52},
		{"Team size", unsafe.Sizeof(Team{}), wantTeamSize},
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Errorf("%s = %d, want %d", check.name, check.got, check.want)
		}
	}
}

func TestDeathmatchLowScoreWinner5095E0ServerPreservesNativePointers(t *testing.T) {
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
	teams[1].Lessons = 7
	players[0].Active = 1
	players[0].PlayerInd = 0
	players[0].PlayerUnit = unit
	players[0].Field2140 = uint32(6)
	unit.ObjClass = object.ClassPlayer
	unit.UpdateData = unsafe.Pointer(update)
	update.Player = &players[0]

	var srv Server
	srv.Teams.Arr = teams
	srv.Players.list = players
	oldFlags := noxflags.GetGame()
	noxflags.ResetGame()
	t.Cleanup(func() {
		noxflags.ResetGame()
		noxflags.SetGame(oldFlags)
	})
	var gotPlayer *Object
	result := srv.DeathmatchLowScoreWinner5095E0(DeathmatchLowScoreWinnerRuntime5095E0{
		SendTeamWinner: func(*Team, uint8) int32 {
			t.Fatal("unexpected team notification")
			return 0
		},
		SendPlayerWinner: func(got *Object, flag uint8) int32 {
			gotPlayer = got
			if flag != deathmatchLowScoreWinnerArg5095E0 {
				t.Fatalf("winner flag = %d, want 1", flag)
			}
			return math.MinInt32
		},
	})
	if result != math.MinInt32 || gotPlayer != unit {
		t.Fatalf("result/player = %d/%p, want %d/%p", result, gotPlayer, math.MinInt32, unit)
	}
	if !noxflags.HasGame(noxflags.GameFlag(deathmatchLowScoreCompleteFlag5095E0)) {
		t.Fatal("game flag 8 was not set")
	}
	runtime.KeepAlive(teams)
	runtime.KeepAlive(players)
	runtime.KeepAlive(unit)
	runtime.KeepAlive(update)
}
