package legacy

import (
	"math"
	"testing"

	"github.com/opennox/opennox/v1/common/memmap"
)

func TestClientReconnectNativeFrameRollover(t *testing.T) {
	oldFrame, oldFPS, oldLog := gameFrameHook, gameFPSHook, NetworkLogPrint
	outcome := memmap.PtrInt32(0x5D4594, 528268)
	oldOutcome := *outcome
	t.Cleanup(func() {
		gameFrameHook, gameFPSHook, NetworkLogPrint = oldFrame, oldFPS, oldLog
		*outcome = oldOutcome
		sub_41E370()
	})
	sub_41E370()
	*outcome = 0
	var frame uint32 = math.MaxUint32 - 60
	const fps uint32 = 60
	gameFrameHook = func() uint32 { return frame }
	gameFPSHook = func() uint32 { return fps }
	NetworkLogPrint = func(string) {}

	nox_xxx_reconStart_41E400()
	want := frame + reconnectDelaySeconds*fps
	if started, delay := Get_dword_5d4594_528252(), uint32(Get_dword_5d4594_528260()); started != 1 || delay != want {
		t.Fatalf("reconnect started=%d deadline=%d, want active deadline %d", started, delay, want)
	}
	frame++
	nox_xxx_reconStart_41E400()
	if delay := uint32(Get_dword_5d4594_528260()); delay != want {
		t.Fatalf("repeated start moved the deadline to %d, want %d", delay, want)
	}

	frame += reconnectWindowSeconds * fps
	if got := nox_xxx_reconAttempt_41E390(); got != 1 || *outcome != 1 {
		t.Fatalf("expired reconnect returned %d, published %d", got, *outcome)
	}
	if started, delay := Get_dword_5d4594_528252(), Get_dword_5d4594_528260(); started != 0 || delay != 0 {
		t.Fatalf("expired reconnect retained active=%d deadline=%d", started, delay)
	}
}
