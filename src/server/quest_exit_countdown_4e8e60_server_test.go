package server

import (
	"math"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
)

func TestQuestExitCountdownNative4E8E60UsesNamedPointers(t *testing.T) {
	player1 := &Player{Field4792: 1}
	player2 := &Player{Field4792: 1}
	exit := &Object{}
	update1 := &PlayerUpdateData{Player: player1, QuestExit: exit}
	update2 := &PlayerUpdateData{Player: player2}
	unit1 := &Object{UpdateData: unsafe.Pointer(update1)}
	unit2 := &Object{UpdateData: unsafe.Pointer(update2)}
	events := make([]string, 0, 2)
	got := questExitCountdownNative4E8E60(questExitCountdownNativeDeps4E8E60{
		balanceFloat:         func(string) float64 { return 12 },
		timerActive:          func() int32 { return 0 },
		timerRemainingMillis: func() int32 { t.Fatal("inactive timer read remaining"); return 0 },
		firstUnit:            func() *Object { return unit1 },
		nextUnit: func(unit *Object) *Object {
			if unit == unit1 {
				return unit2
			}
			return nil
		},
		stopTimer:        func(int32) int32 { t.Fatal("eligible players stopped timer"); return 0 },
		countdownStarted: func() int32 { t.Fatal("positive portion queried timer flag"); return 0 },
		startCountdown: func(seconds int32, id string) {
			events = append(events, id)
			if seconds != 6 {
				t.Fatalf("seconds = %d, want 6", seconds)
			}
		},
		sendGauntlet: func(recipient int32) int32 {
			events = append(events, "send")
			if recipient != 255 {
				t.Fatalf("recipient = %d, want 255", recipient)
			}
			return math.MinInt32 + 41
		},
	})
	if got != math.MinInt32+41 {
		t.Fatalf("result = %d, want %d", got, int32(math.MinInt32+41))
	}
	if want := []string{"objcoll.c:ExitCountdown", "send"}; !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %q, want %q", events, want)
	}
}

func TestQuestNextPlayerUnit4DA7F0UsesLegacyFieldChain(t *testing.T) {
	s := &Server{}
	s.Players.list = make([]Player, 3)
	first := &Object{ObjClass: object.ClassPlayer}
	last := &Object{ObjClass: object.ClassPlayer}
	for i := range s.Players.list {
		s.Players.list[i].PlayerInd = byte(i)
		s.Players.list[i].Active = 1
	}
	s.Players.list[0].PlayerUnit = first
	s.Players.list[1].PlayerUnit = nil
	s.Players.list[2].PlayerUnit = last
	first.UpdateData = unsafe.Pointer(&PlayerUpdateData{Player: &s.Players.list[0]})
	if got := s.questNextPlayerUnit4DA7F0(first); got != last {
		t.Fatalf("next = %p, want %p", got, last)
	}
	first.ObjClass = object.ClassMonster
	first.UpdateData = nil
	if got := s.questNextPlayerUnit4DA7F0(first); got != nil {
		t.Fatalf("non-Player next = %p, want nil", got)
	}
}

func TestQuestExitSendGauntlet4E8E60PacketAndReturn(t *testing.T) {
	s := &Server{NetSendPacketXxx: func(recipient int, packet []byte, related *Object, remove, sequence int) int {
		if recipient != 255 || !reflect.DeepEqual(packet, []byte{0xf0, 0x14}) || related != nil || remove != 1 || sequence != 0 {
			t.Fatalf("send = (%d, % x, %p, %d, %d)", recipient, packet, related, remove, sequence)
		}
		return math.MinInt32 + 53
	}}
	if got := s.questExitSendGauntlet4E8E60(255); got != math.MinInt32+53 {
		t.Fatalf("result = %d, want %d", got, int32(math.MinInt32+53))
	}
}

func TestQuestExitCountdownNativeLayout4E8E60(t *testing.T) {
	wantSize := uintptr(556)
	wantExit := uintptr(312)
	wantWarp := uintptr(316)
	wantState := uintptr(4792)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantSize = 640
		wantExit = 384
		wantWarp = 392
		wantState = 6096
	}
	if got := unsafe.Sizeof(PlayerUpdateData{}); got != wantSize {
		t.Fatalf("PlayerUpdateData size = %d, want %d", got, wantSize)
	}
	if got := unsafe.Offsetof(PlayerUpdateData{}.QuestExit); got != wantExit {
		t.Fatalf("QuestExit offset = %d, want %d", got, wantExit)
	}
	if got := unsafe.Offsetof(PlayerUpdateData{}.QuestWarpGate); got != wantWarp {
		t.Fatalf("QuestWarpGate offset = %d, want %d", got, wantWarp)
	}
	if got := unsafe.Offsetof(Player{}.Field4792); got != wantState {
		t.Fatalf("Player.Field4792 offset = %d, want %d", got, wantState)
	}
}
