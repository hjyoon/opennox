package server

import (
	"testing"
	"unsafe"
)

func TestQuestMaybeWarpNative4E8F60UsesNamedPointersAndProgress(t *testing.T) {
	gate := &Object{}
	player1 := &Player{PlayerInd: 4, Field4792: 1, QuestStage: 7}
	player2 := &Player{PlayerInd: 5, Field4792: 1, QuestStage: 12}
	update1 := &PlayerUpdateData{Player: player1, QuestWarpGate: gate}
	update2 := &PlayerUpdateData{Player: player2, QuestWarpGate: gate}
	unit1 := &Object{UpdateData: unsafe.Pointer(update1)}
	unit2 := &Object{UpdateData: unsafe.Pointer(update2)}

	got := questMaybeWarpNative4E8F60(questMaybeWarpNativeDeps4E8F60{
		currentQuestStage: func() uint32 { return 6 },
		nextStageThreshold: func(stage uint32) uint32 {
			if stage != 6 {
				t.Fatalf("stage = %d, want 6", stage)
			}
			return 10
		},
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
		update1.QuestWarpGate != gate || update2.QuestWarpGate != gate {
		t.Fatal("native player or warp-gate pointers changed")
	}
}

func TestQuestMaybeWarpNative4E8F60MissingWarpGateFails(t *testing.T) {
	player := &Player{PlayerInd: 3, Field4792: 2, QuestStage: ^uint32(0)}
	update := &PlayerUpdateData{Player: player}
	unit := &Object{UpdateData: unsafe.Pointer(update)}
	got := questMaybeWarpNative4E8F60(questMaybeWarpNativeDeps4E8F60{
		currentQuestStage:  func() uint32 { return 0 },
		nextStageThreshold: func(uint32) uint32 { return 0 },
		firstUnit:          func() *Object { return unit },
		nextUnit:           func(*Object) *Object { t.Fatal("missing gate read successor"); return nil },
		gameHost:           func() int32 { return 0 },
		noRendering:        func() int32 { t.Fatal("non-host read rendering flag"); return 0 },
	})
	if got != 0 {
		t.Fatalf("result = %d, want 0", got)
	}
}

func TestQuestMaybeWarpNativeLayout4E8F60(t *testing.T) {
	wantPlayer, wantWarp := uintptr(276), uintptr(316)
	wantIndex, wantQuestStage, wantState := uintptr(2064), uintptr(4696), uintptr(4792)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantPlayer, wantWarp = 320, 392
		wantIndex, wantQuestStage, wantState = 2068, 6000, 6096
	}
	checks := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"PlayerUpdateData.Player", unsafe.Offsetof(PlayerUpdateData{}.Player), wantPlayer},
		{"PlayerUpdateData.QuestWarpGate", unsafe.Offsetof(PlayerUpdateData{}.QuestWarpGate), wantWarp},
		{"Player.PlayerInd", unsafe.Offsetof(Player{}.PlayerInd), wantIndex},
		{"Player.QuestStage", unsafe.Offsetof(Player{}.QuestStage), wantQuestStage},
		{"Player.Field4792", unsafe.Offsetof(Player{}.Field4792), wantState},
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Fatalf("%s offset = %d, want %d", check.name, check.got, check.want)
		}
	}
}
