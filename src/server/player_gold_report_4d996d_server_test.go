package server

import (
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
)

func TestPlayerGoldReportNative4D996DReloadsPlayerAfterReport(t *testing.T) {
	first := &Player{GoldVal: 37, Field2168: 12, PlayerInd: 7}
	second := &Player{GoldVal: 91, Field2168: 3, PlayerInd: 9}
	update := &PlayerUpdateData{Player: first}
	unit := &Object{ObjClass: object.ClassPlayer, UpdateData: unsafe.Pointer(update)}
	var calls int
	playerGoldReportNative4D996D(unit, update, playerGoldReportNativeDeps4D996D{
		sendReliable: func(recipient int32, packet [5]byte, related *Object, remove int32) int32 {
			calls++
			if recipient != 7 || packet != [5]byte{74, 37, 0, 0, 0} || related != nil || remove != 1 {
				t.Fatalf("send = recipient:%d packet:%v related:%p remove:%d", recipient, packet, related, remove)
			}
			update.Player = second
			return math.MinInt32
		},
	})
	if calls != 1 || first.Field2168 != 12 || second.Field2168 != 91 {
		t.Fatalf("final state = calls:%d first:%d second:%d, want 1/12/91", calls, first.Field2168, second.Field2168)
	}
}

func TestGoldReportNative4D8870ReloadsLiveGold(t *testing.T) {
	player := &Player{GoldVal: 0x78563412}
	update := &PlayerUpdateData{Player: player}
	unit := &Object{ObjClass: object.ClassPlayer, UpdateData: unsafe.Pointer(update)}
	var calls int
	goldReportNative4D8870(-17, unit, playerGoldReportNativeDeps4D996D{
		sendReliable: func(recipient int32, packet [5]byte, related *Object, remove int32) int32 {
			calls++
			if recipient != -17 || packet != [5]byte{74, 0x12, 0x34, 0x56, 0x78} || related != nil || remove != 1 {
				t.Fatalf("send = recipient:%d packet:%v related:%p remove:%d", recipient, packet, related, remove)
			}
			return -1
		},
	})
	if calls != 1 {
		t.Fatalf("send calls = %d, want 1", calls)
	}

	unit.ObjClass = object.ClassMonster
	goldReportNative4D8870(1, unit, playerGoldReportNativeDeps4D996D{
		sendReliable: func(int32, [5]byte, *Object, int32) int32 {
			t.Fatal("non-Player report sent a packet")
			return 0
		},
	})
}

func TestPlayerGoldReportSync4D9900NilAndNonPlayerReturn(t *testing.T) {
	s := &Server{}
	s.PlayerGoldReportSync4D9900(nil)
	s.PlayerGoldReportSync4D9900(&Object{ObjClass: object.ClassMonster})
}
