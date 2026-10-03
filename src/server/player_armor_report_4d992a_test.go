package server

import (
	"math"
	"testing"
)

func TestPlayerArmorReport4D992AByteIndexAndFailedSendAcknowledgement(t *testing.T) {
	for index := range 256 {
		player := &Player{PlayerInd: byte(index)}
		update := &PlayerUpdateData{Player: player, Field57: 0x3f123456, Field58: 0x3e800000}
		calls := 0
		s := &Server{NetSendPacketXxx: func(recipient int, packet []byte, related *Object, remove, sequence int) int {
			calls++
			if recipient != index || len(packet) != 5 || packet[0] != 73 ||
				packet[1] != 0x56 || packet[2] != 0x34 || packet[3] != 0x12 || packet[4] != 0x3f ||
				related != nil || remove != 1 || sequence != 0 {
				t.Fatalf("index=%d transport=%d/%x/%p/%d/%d", index, recipient, packet, related, remove, sequence)
			}
			update.Field57 = 0x3f765432
			return math.MinInt32
		}}
		s.PlayerArmorReport4D992A(update)
		if calls != 1 || update.Field58 != 0x3f765432 {
			t.Fatalf("index=%d calls=%d acknowledgement=%08x", index, calls, update.Field58)
		}
		s.PlayerArmorReport4D992A(update)
		if calls != 1 {
			t.Fatalf("index=%d repeated unchanged report", index)
		}
	}
}

func TestPlayerArmorReport4D992ARetainsOriginalFaultPrefix(t *testing.T) {
	// No Player or transport is read before the ordered changed-value gate.
	for _, update := range []*PlayerUpdateData{
		{}, {Field57: 0x80000000}, {Field57: 0x7fc00000}, {Field58: 0x7f800001},
	} {
		var s *Server
		s.PlayerArmorReport4D992A(update)
	}
	for _, tc := range []struct {
		name   string
		update *PlayerUpdateData
	}{
		{"nil update", nil},
		{"nil player after changed-value gate", &PlayerUpdateData{Field57: 1}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			called := false
			s := &Server{NetSendPacketXxx: func(int, []byte, *Object, int, int) int { called = true; return 0 }}
			defer func() {
				if recover() == nil || called || tc.update != nil && tc.update.Field58 != 0 {
					t.Fatal("original fault prefix changed or transport/cache ran")
				}
			}()
			s.PlayerArmorReport4D992A(tc.update)
		})
	}
}
