package server

import (
	"reflect"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/noxnet/netmsg"
	"github.com/opennox/opennox/v1/common/ntype"
)

func TestLocalUnitOrderNativeLayout500C70(t *testing.T) {
	wantOffset := uintptr(3648)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantOffset = 4944
	}
	if got := unsafe.Offsetof(Player{}.SummonOrderAll); got != wantOffset {
		t.Fatalf("Player.SummonOrderAll offset on %s/%s = %d, want %d", runtime.GOOS, runtime.GOARCH, got, wantOffset)
	}
	if got := unsafe.Sizeof(Player{}.SummonOrderAll); got != 4 {
		t.Fatalf("Player.SummonOrderAll width = %d, want 4", got)
	}
}

func TestLocalUnitOrderServer500C70StoresBeforeSending(t *testing.T) {
	const (
		owner   = ntype.PlayerInd(2)
		wantRet = -0x2345678
	)
	orderType := uint32(0x89abcdef)
	s := &Server{}
	s.Players.list = make([]Player, 3)
	s.Players.list[owner].Active = 1
	var gotPacket []byte
	s.NetSendPacketXxx = func(recipient int, packet []byte, related *Object, removeIfDisconnected, sequenceEnabled int) int {
		if got := s.Players.list[owner].SummonOrderAll; got != orderType {
			t.Fatalf("stored order at send = %#x, want %#x", got, orderType)
		}
		if recipient != int(owner) || related != nil || removeIfDisconnected != 1 || sequenceEnabled != 1 {
			t.Fatalf("send args = (%d, %p, %d, %d), want (%d, nil, 1, 1)", recipient, related, removeIfDisconnected, sequenceEnabled, owner)
		}
		gotPacket = append([]byte(nil), packet...)
		return wantRet
	}

	if got := s.Nox_xxx_orderUnitLocal_500C70(owner, orderType); got != wantRet {
		t.Fatalf("result = %d, want %d", got, wantRet)
	}
	wantPacket := []byte{byte(netmsg.MSG_REPORT_CREATURE_CMD), byte(orderType)}
	if !reflect.DeepEqual(gotPacket, wantPacket) {
		t.Fatalf("packet = % x, want % x", gotPacket, wantPacket)
	}
}

func TestLocalUnitOrderServer500C70InvalidPlayerFaultsBeforeSending(t *testing.T) {
	tests := []struct {
		name  string
		owner ntype.PlayerInd
	}{
		{name: "negative", owner: -1},
		{name: "inactive", owner: 0},
		{name: "out of range", owner: 1},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			s := &Server{}
			s.Players.list = make([]Player, 1)
			s.NetSendPacketXxx = func(int, []byte, *Object, int, int) int {
				t.Fatal("invalid player path sent a packet")
				return 0
			}
			defer func() {
				if recover() == nil {
					t.Fatal("invalid player path did not fault")
				}
			}()
			s.Nox_xxx_orderUnitLocal_500C70(tc.owner, 0x12345678)
		})
	}
}
