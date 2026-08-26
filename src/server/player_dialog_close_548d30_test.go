package server

import (
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
)

func newPlayerDialogCloseFixture548D30() (*Server, *Object, *Object, *PlayerUpdateData, *MonsterUpdateData) {
	playerData := &PlayerUpdateData{Player: &Player{PlayerInd: 7}}
	npcData := &MonsterUpdateData{
		DialogStartFunc: 11,
		DialogEndFunc:   22,
		DialogFlags:     1,
		DialogResult:    0x7f,
	}
	player := &Object{ObjClass: object.ClassPlayer, UpdateData: unsafe.Pointer(playerData)}
	npc := &Object{ObjClass: object.ClassMonster, UpdateData: unsafe.Pointer(npcData)}
	playerData.DialogWith = npc
	return new(Server), player, npc, playerData, npcData
}

func TestPlayerDialogClose548D30OriginalOrder(t *testing.T) {
	s, player, npc, playerData, npcData := newPlayerDialogCloseFixture548D30()
	replacement := &PlayerUpdateData{}
	var events []string
	s.PlayerDialogClose548D30(player, 2, PlayerDialogCloseRuntime548D30{
		Unfreeze: func(got *Object, force uint32) {
			events = append(events, "unfreeze")
			if got != player || force != 0 {
				t.Fatalf("unfreeze = (%p, %d), want (%p, 0)", got, force, player)
			}
			// GAME.EXE cached the original PlayerUpdateData before this call.
			player.UpdateData = unsafe.Pointer(replacement)
		},
		Send: func(recipient byte, packet [2]byte) {
			events = append(events, "send")
			if recipient != 7 || packet != [2]byte{0xd0, 0x04} {
				t.Fatalf("send = (%d, % x), want (7, d0 04)", recipient, packet)
			}
			if playerData.DialogWith != npc {
				t.Fatal("DialogWith was cleared before the close packet")
			}
		},
		CallEnd: func(index int32, caller, trigger *Object) {
			events = append(events, "call")
			if index != 22 || caller != player || trigger != npc {
				t.Fatalf("end call = (%d, %p, %p), want (22, %p, %p)", index, caller, trigger, player, npc)
			}
			if playerData.DialogWith != nil || npcData.DialogResult != 2 {
				t.Fatalf("state at end call = (%p, %d), want (nil, 2)", playerData.DialogWith, npcData.DialogResult)
			}
		},
	})
	if !reflect.DeepEqual(events, []string{"unfreeze", "send", "call"}) {
		t.Fatalf("events = %v, want [unfreeze send call]", events)
	}
	if player.UpdateData != unsafe.Pointer(replacement) {
		t.Fatal("test did not retain the replacement live update-data pointer")
	}
}

func TestPlayerDialogClose548D30UnfreezesBeforeEarlyReturns(t *testing.T) {
	t.Run("no dialog", func(t *testing.T) {
		s, player, _, playerData, _ := newPlayerDialogCloseFixture548D30()
		playerData.DialogWith = nil
		calls := 0
		s.PlayerDialogClose548D30(player, 1, PlayerDialogCloseRuntime548D30{
			Unfreeze: func(*Object, uint32) { calls++ },
			Send:     func(byte, [2]byte) { t.Fatal("unexpected send") },
			CallEnd:  func(int32, *Object, *Object) { t.Fatal("unexpected end call") },
		})
		if calls != 1 {
			t.Fatalf("unfreeze calls = %d, want 1", calls)
		}
	})

	for _, tc := range []struct {
		name       string
		start, end int32
	}{
		{name: "missing start", start: -1, end: 22},
		{name: "missing end", start: 11, end: -1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s, player, npc, playerData, npcData := newPlayerDialogCloseFixture548D30()
			npcData.DialogStartFunc, npcData.DialogEndFunc = tc.start, tc.end
			calls := 0
			s.PlayerDialogClose548D30(player, 1, PlayerDialogCloseRuntime548D30{
				Unfreeze: func(*Object, uint32) { calls++ },
				Send:     func(byte, [2]byte) { t.Fatal("unexpected send") },
				CallEnd:  func(int32, *Object, *Object) { t.Fatal("unexpected end call") },
			})
			if calls != 1 || playerData.DialogWith != npc {
				t.Fatalf("early-return state = (unfreeze %d, dialog %p), want (1, %p)", calls, playerData.DialogWith, npc)
			}
		})
	}
}

func TestPlayerDialogClose548D30ReloadsPostSendNPCFields(t *testing.T) {
	s, player, _, _, npcData := newPlayerDialogCloseFixture548D30()
	called := int32(-1)
	s.PlayerDialogClose548D30(player, 2, PlayerDialogCloseRuntime548D30{
		Unfreeze: func(*Object, uint32) {},
		Send: func(byte, [2]byte) {
			npcData.DialogFlags = 0
			npcData.DialogEndFunc = 33
		},
		CallEnd: func(index int32, _, _ *Object) { called = index },
	})
	if npcData.DialogResult != 0 || called != 33 {
		t.Fatalf("post-send state = (result %d, end %d), want (0, 33)", npcData.DialogResult, called)
	}
}
