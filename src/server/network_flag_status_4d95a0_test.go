package server

import (
	"math"
	"reflect"
	"testing"
)

func TestServerNetSendFlagStatus4D95A0ExactPacketAndReturn(t *testing.T) {
	s := &Server{}
	wantResult := int(math.MinInt32 + 29)
	called := 0
	s.NetSendPacketXxx = func(recipient int, payload []byte, related *Object, removeIfDisconnected, sequenceEnabled int) int {
		called++
		if recipient != 255 {
			t.Fatalf("recipient = %d, want 255", recipient)
		}
		if related != nil {
			t.Fatalf("related object = %p, want nil", related)
		}
		if removeIfDisconnected != 1 || sequenceEnabled != 1 {
			t.Fatalf("delivery flags = (%d, %d), want (1, 1)", removeIfDisconnected, sequenceEnabled)
		}
		wantPayload := []byte{0xd8, 0x81, 0xfe, 0xa7, 0xde, 0xbc}
		if !reflect.DeepEqual(payload, wantPayload) {
			t.Fatalf("payload = % x, want % x", payload, wantPayload)
		}
		return wantResult
	}
	if got := s.Nox_xxx_netSendFlagStatus_4D95A0(255, 0xfe, 0x81, 0xa7, 0xbcde); got != int32(wantResult) {
		t.Fatalf("result = %d, want %d", got, wantResult)
	}
	if called != 1 {
		t.Fatalf("send calls = %d, want 1", called)
	}
}
