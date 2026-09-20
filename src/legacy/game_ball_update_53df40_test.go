package legacy

import (
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

func TestGameBallUpdate53DF40RegisteredNativeSizeAndDispatch(t *testing.T) {
	callback, size, ok := server.ObjectUpdateHandler("BallUpdate")
	if !ok || callback == nil {
		t.Fatalf("BallUpdate registration = %p/%t", callback, ok)
	}
	if want := unsafe.Sizeof(server.GameBallUpdateData4EA800{}); size != want {
		t.Fatalf("BallUpdate data size = %d, want native size %d", size, want)
	}

	oldCall := gameBallUpdateCall53DF40
	t.Cleanup(func() { gameBallUpdateCall53DF40 = oldCall })
	ball := &server.Object{ObjOwner: &server.Object{}}
	var got *server.Object
	gameBallUpdateCall53DF40 = func(value *server.Object) { got = value }
	server.CallObjectUpdate(callback, ball)
	if got != ball {
		t.Fatalf("native update argument = %p, want %p", got, ball)
	}
}

type gameBallDeathLegacyServer417F50 struct {
	Server
	got *server.Object
}

func (s *gameBallDeathLegacyServer417F50) GameBallReset417F50(old *server.Object) int {
	s.got = old
	return 1
}

func TestGameBallDeath54E620DispatchesNativeReset(t *testing.T) {
	callback, size, ok := server.ObjectDeathHandler("GameBallDie")
	if !ok || callback == nil || size != 0 {
		t.Fatalf("GameBallDie registration = %p/%d/%t", callback, size, ok)
	}

	outer := &gameBallDeathLegacyServer417F50{}
	oldGetServer := GetServer
	GetServer = func() Server { return outer }
	t.Cleanup(func() { GetServer = oldGetServer })

	ball := &server.Object{ObjOwner: &server.Object{}}
	server.CallObjectDeath(callback, ball)
	if outer.got != ball {
		t.Fatalf("native death argument = %p, want %p", outer.got, ball)
	}
}
