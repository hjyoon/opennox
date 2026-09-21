package opennox

import (
	"testing"

	"github.com/opennox/libs/noxnet/netmsg"
	"github.com/opennox/opennox/v1/server"
)

func TestSendSecretWallState519660(t *testing.T) {
	players := []*server.Player{
		{Active: 1, PlayerInd: 2, PlayerUnit: &server.Object{}},
		nil,
		{Active: 1, PlayerInd: 7},
		{PlayerInd: 9, PlayerUnit: &server.Object{}},
	}
	wl := &server.Wall{Field10: 0x1234}

	tests := []struct {
		name string
		open bool
		op   netmsg.Op
	}{
		{name: "open", open: true, op: netmsg.MSG_OPEN_WALL},
		{name: "close", open: false, op: netmsg.MSG_CLOSE_WALL},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var recipients []int
			var packets [][]byte
			ok := sendSecretWallState519660(players, wl, tc.open, func(recipient int, packet []byte) {
				recipients = append(recipients, recipient)
				packets = append(packets, append([]byte(nil), packet...))
			})
			if !ok {
				t.Fatal("valid wall was rejected")
			}
			if len(recipients) != 1 || recipients[0] != 2 {
				t.Fatalf("recipients = %v, want [2]", recipients)
			}
			want := []byte{byte(tc.op), 0x34, 0x12}
			if len(packets) != 1 || string(packets[0]) != string(want) {
				t.Fatalf("packets = %v, want [%v]", packets, want)
			}
		})
	}
}

func TestSendSecretWallState519660MissingWall(t *testing.T) {
	called := false
	ok := sendSecretWallState519660([]*server.Player{{Active: 1, PlayerUnit: &server.Object{}}}, nil, true, func(int, []byte) {
		called = true
	})
	if ok {
		t.Fatal("nil wall was accepted")
	}
	if called {
		t.Fatal("nil wall emitted a packet")
	}
}
