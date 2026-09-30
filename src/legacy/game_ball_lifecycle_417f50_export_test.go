package legacy

import (
	"math"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

type gameBallLifecycleLegacyServer417F50 struct {
	Server
	want   *server.Object
	result int
	calls  int
}

func (s *gameBallLifecycleLegacyServer417F50) GameBallReset417F50(old *server.Object) int {
	s.calls++
	if old != s.want {
		panic("game ball reset received a truncated object pointer")
	}
	return s.result
}

func TestGameBallLifecycleExportsPreserveNativePointer417F50(t *testing.T) {
	ball := new(server.Object)
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(ball)) <= math.MaxUint32 {
		t.Fatalf("ball pointer = %p, want native address above 4 GiB", ball)
	}
	var pin runtime.Pinner
	pin.Pin(ball)
	defer pin.Unpin()

	outer := &gameBallLifecycleLegacyServer417F50{want: ball, result: 73}
	oldGetServer := GetServer
	GetServer = func() Server { return outer }
	t.Cleanup(func() { GetServer = oldGetServer })

	if got := gameBallResetExportCall417F50(ball); got != 73 {
		t.Fatalf("reset result = %d, want 73", got)
	}
	if got := gameBallDeathExportCall54E620(ball); got != 73 {
		t.Fatalf("death result = %d, want 73", got)
	}
	if outer.calls != 2 {
		t.Fatalf("reset calls = %d, want 2", outer.calls)
	}
	runtime.KeepAlive(ball)
}

func TestGameBallUpdateExportPreservesNativePointer53DF40(t *testing.T) {
	ball := new(server.Object)
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(ball)) <= math.MaxUint32 {
		t.Fatalf("ball pointer = %p, want native address above 4 GiB", ball)
	}
	var pin runtime.Pinner
	pin.Pin(ball)
	defer pin.Unpin()

	oldCall := gameBallUpdateCall53DF40
	t.Cleanup(func() { gameBallUpdateCall53DF40 = oldCall })
	var calls int
	gameBallUpdateCall53DF40 = func(got *server.Object) {
		calls++
		if got != ball {
			t.Fatalf("update ball = %p, want %p", got, ball)
		}
	}

	gameBallUpdateExportCall53DF40(ball)
	if calls != 1 {
		t.Fatalf("update calls = %d, want 1", calls)
	}
	runtime.KeepAlive(ball)
}
