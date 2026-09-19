package server

import (
	"encoding/binary"
	"math"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/noxnet/netmsg"
	"github.com/opennox/libs/object"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

func TestMatchState5096F0NativeLayouts(t *testing.T) {
	ptrSize := unsafe.Sizeof(uintptr(0))
	wantObjectUpdate := uintptr(748)
	wantUpdatePlayer := uintptr(276)
	wantUpdateQuestExit := uintptr(312)
	wantPlayerIndex := uintptr(2064)
	wantPlayerPosition := uintptr(3632)
	wantPlayerQuestState := uintptr(4792)
	if ptrSize == 8 {
		wantObjectUpdate = 872
		wantUpdatePlayer = 336
		wantUpdateQuestExit = 400
		wantPlayerIndex = 2068
		wantPlayerPosition = 4920
		wantPlayerQuestState = 6096
	}
	checks := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"Object.UpdateData", unsafe.Offsetof(Object{}.UpdateData), wantObjectUpdate},
		{"PlayerUpdateData.Player", unsafe.Offsetof(PlayerUpdateData{}.Player), wantUpdatePlayer},
		{"PlayerUpdateData.QuestExit", unsafe.Offsetof(PlayerUpdateData{}.QuestExit), wantUpdateQuestExit},
		{"Player.PlayerInd", unsafe.Offsetof(Player{}.PlayerInd), wantPlayerIndex},
		{"Player.Pos3632Vec", unsafe.Offsetof(Player{}.Pos3632Vec), wantPlayerPosition},
		{"Player.Field4792", unsafe.Offsetof(Player{}.Field4792), wantPlayerQuestState},
		{"Team.IDVal", unsafe.Offsetof(Team{}.IDVal), 57},
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Errorf("%s = %d, want %d", check.name, check.got, check.want)
		}
	}
}

func TestMatchState509xServerPreservesNativePointers(t *testing.T) {
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
	teams[1].IDVal = 0x7e
	teams[1].Lessons = 10
	players[0].Active = 1
	players[0].PlayerInd = 0
	players[0].PlayerUnit = unit
	players[0].Lessons = 11
	players[0].Pos3632Vec.X = -2.9
	players[0].Pos3632Vec.Y = 65537.9
	unit.ObjClass = object.ClassPlayer
	unit.NetCode = 0x12345678
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

	var highWinner *Object
	if got := srv.DeathmatchHighScoreWinner5098A0(DeathmatchHighScoreWinnerRuntime5098A0{
		SendTeamWinner: func(*Team, uint8) int32 {
			t.Fatal("unexpected team winner")
			return 0
		},
		SendPlayerWinner: func(winner *Object, arg uint8) int32 {
			highWinner = winner
			if arg != deathmatchHighScoreWinnerArg5098A0 {
				t.Fatalf("high-score winner arg = %d, want 1", arg)
			}
			return math.MinInt32
		},
	}); got != math.MinInt32 || highWinner != unit {
		t.Fatalf("high-score result/winner = %d/%p, want %d/%p", got, highWinner, math.MinInt32, unit)
	}

	noxflags.ResetGame()
	teams[1].Lessons = 12
	srv.SetFrame(0x89abcdef)
	var recipient, remove, sequence int
	var sent []byte
	srv.NetSendPacketXxx = func(gotRecipient int, buf []byte, related *Object, gotRemove, gotSequence int) int {
		recipient, remove, sequence = gotRecipient, gotRemove, gotSequence
		if related != nil {
			t.Fatalf("related object = %p, want nil", related)
		}
		sent = append([]byte(nil), buf...)
		return -91
	}
	if got := srv.TeamHighScoreWinner5099B0(); got != -91 {
		t.Fatalf("team result = %d, want -91", got)
	}
	wantTeamPacket := [8]byte{byte(netmsg.MSG_REPORT_FLAG_WINNER), 0x7e, 0x00, 0x01}
	binary.LittleEndian.PutUint32(wantTeamPacket[4:], 0x89abcdef)
	if recipient != 255 || remove != 1 || sequence != 1 || string(sent) != string(wantTeamPacket[:]) {
		t.Fatalf("team send = recipient %d remove %d sequence %d packet %x, want 255/1/1/%x", recipient, remove, sequence, sent, wantTeamPacket)
	}

	noxflags.ResetGame()
	players[0].PlayerInd = 7
	var positionIndex uint8
	var positionPacket [5]byte
	var audioUnit *Object
	var audioID, audioCode uint32
	var audioKind int32
	got := srv.matchLimitStateNative5096F0(MatchLimitStateRuntime5096F0{
		LimitExpired:     func() int32 { return 1 },
		MatchStateActive: func() int32 { return 1 },
		ClearTimer:       func() int32 { return -123 },
	}, matchLimitStateNativeDeps5096F0{
		firstPlayer: func() *Object { return unit },
		nextPlayer:  func(*Object) *Object { return nil },
		sendPosition: func(index uint8, packet [5]byte) {
			positionIndex, positionPacket = index, packet
		},
		audioEvent: func(id uint32, gotUnit *Object, kind int32, code uint32) {
			audioID, audioUnit, audioKind, audioCode = id, gotUnit, kind, code
		},
	})
	if got != -123 {
		t.Fatalf("match-limit result = %d, want -123", got)
	}
	wantPosition := [5]byte{0x9a, 0xfe, 0xff, 0x01, 0x00}
	if positionIndex != 7 || positionPacket != wantPosition {
		t.Fatalf("position send = %d/%x, want 7/%x", positionIndex, positionPacket, wantPosition)
	}
	if audioID != 582 || audioUnit != unit || audioKind != 2 || audioCode != unit.NetCode {
		t.Fatalf("audio = %d/%p/%d/%#x, want 582/%p/2/%#x", audioID, audioUnit, audioKind, audioCode, unit, unit.NetCode)
	}
	if !noxflags.HasGame(noxflags.GameFlag(matchLimitSuddenDeathFlag5096F0)) {
		t.Fatal("sudden-death game flag was not set")
	}

	runtime.KeepAlive(teams)
	runtime.KeepAlive(players)
	runtime.KeepAlive(unit)
	runtime.KeepAlive(update)
}
