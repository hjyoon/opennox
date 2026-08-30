package server

import (
	"testing"
	"unsafe"
)

func TestQuestAllPlayersExitedNative4E9010UsesNamedPointers(t *testing.T) {
	exit := &Object{}
	player1 := &Player{PlayerInd: 4, Field4792: 1}
	player2 := &Player{PlayerInd: 5, Field4792: 2}
	update1 := &PlayerUpdateData{Player: player1, QuestExit: exit}
	update2 := &PlayerUpdateData{Player: player2, QuestExit: exit}
	unit1 := &Object{UpdateData: unsafe.Pointer(update1)}
	unit2 := &Object{UpdateData: unsafe.Pointer(update2)}

	got := questAllPlayersExitedNative4E9010(questAllPlayersExitedNativeDeps4E9010{
		firstUnit: func() *Object { return unit1 },
		nextUnit: func(unit *Object) *Object {
			if unit == unit1 {
				return unit2
			}
			return nil
		},
		gameHost:    func() int32 { return 0 },
		noRendering: func() int32 { t.Fatal("non-host read rendering flag"); return 0 },
	})
	if got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	if update1.Player != player1 || update2.Player != player2 ||
		update1.QuestExit != exit || update2.QuestExit != exit {
		t.Fatal("native player or Quest-exit pointers changed")
	}
}

func TestQuestAllPlayersExitedNative4E9010MissingExitFails(t *testing.T) {
	player := &Player{PlayerInd: 3, Field4792: 0xffffffff}
	update := &PlayerUpdateData{Player: player}
	unit := &Object{UpdateData: unsafe.Pointer(update)}
	got := questAllPlayersExitedNative4E9010(questAllPlayersExitedNativeDeps4E9010{
		firstUnit:   func() *Object { return unit },
		nextUnit:    func(*Object) *Object { t.Fatal("missing exit read successor"); return nil },
		gameHost:    func() int32 { return 0 },
		noRendering: func() int32 { t.Fatal("non-host read rendering flag"); return 0 },
	})
	if got != 0 {
		t.Fatalf("result = %d, want 0", got)
	}
}

func TestQuestAllPlayersExitedNativeLayout4E9010(t *testing.T) {
	wantPlayer, wantExit := uintptr(276), uintptr(312)
	wantIndex, wantState := uintptr(2064), uintptr(4792)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantPlayer, wantExit = 336, 400
		wantIndex, wantState = 2068, 6096
	}
	checks := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"PlayerUpdateData.Player", unsafe.Offsetof(PlayerUpdateData{}.Player), wantPlayer},
		{"PlayerUpdateData.QuestExit", unsafe.Offsetof(PlayerUpdateData{}.QuestExit), wantExit},
		{"Player.PlayerInd", unsafe.Offsetof(Player{}.PlayerInd), wantIndex},
		{"Player.Field4792", unsafe.Offsetof(Player{}.Field4792), wantState},
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Fatalf("%s offset = %d, want %d", check.name, check.got, check.want)
		}
	}
}
