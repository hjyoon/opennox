package legacy

import (
	"math"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/server"
)

type winkGameBallReleaseLegacyServer4F7DF0 struct {
	Server
	srv   *server.Server
	apply func(*server.Object, types.Pointf, float64)
}

func (s *winkGameBallReleaseLegacyServer4F7DF0) S() *server.Server {
	return s.srv
}

func (s *winkGameBallReleaseLegacyServer4F7DF0) ApplyForce(
	obj *server.Object,
	pos types.Pointf,
	force float64,
) {
	s.apply(obj, pos, force)
}

func TestWinkGameBallReleaseExport4F7DF0PreservesNativePointers(t *testing.T) {
	s := &server.Server{}
	marker := &server.Object{}
	ball := &server.Object{
		ObjFlags: object.Flags(0xffffffff),
		Obj130:   marker,
	}
	player := &server.Object{
		PosVec:   types.Pointf{X: -9.5, Y: 12.25},
		Field129: ball,
	}
	ball.ObjOwner = player

	forces := 0
	statuses := 0
	legacyServer := &winkGameBallReleaseLegacyServer4F7DF0{srv: s}
	legacyServer.apply = func(gotBall *server.Object, pos types.Pointf, force float64) {
		forces++
		if gotBall != ball || pos != player.PosVec || force != 100 {
			t.Fatalf("force = ball:%p pos:%v value:%g", gotBall, pos, force)
		}
		if gotBall.ObjFlags != object.Flags(0xffffffbf) || gotBall.Obj130 != marker {
			t.Fatalf("force-time ball = flags:%#x obj130:%p", gotBall.ObjFlags, gotBall.Obj130)
		}
	}
	oldGetServer := GetServer
	oldBallStatus := Sub_4E8290
	GetServer = func() Server { return legacyServer }
	Sub_4E8290 = func(state uint8, netCode uint16) int32 {
		statuses++
		if state != 1 || netCode != 0 {
			t.Fatalf("status = %d/%d", state, netCode)
		}
		return -1
	}
	t.Cleanup(func() {
		GetServer = oldGetServer
		Sub_4E8290 = oldBallStatus
	})

	var pin runtime.Pinner
	for _, obj := range []*server.Object{marker, ball, player} {
		pin.Pin(obj)
	}
	defer pin.Unpin()
	if unsafe.Sizeof(uintptr(0)) == 8 {
		for name, pointer := range map[string]uintptr{
			"marker": uintptr(unsafe.Pointer(marker)),
			"ball":   uintptr(unsafe.Pointer(ball)),
			"player": uintptr(unsafe.Pointer(player)),
		} {
			if pointer <= math.MaxUint32 {
				t.Fatalf("%s pointer = %#x, want native address above 4 GiB", name, pointer)
			}
		}
	}

	if got := winkGameBallReleaseExportCall4F7DF0(player); got != 1 {
		t.Fatalf("export result = %d, want 1", got)
	}
	if forces != 1 || statuses != 1 {
		t.Fatalf("forces/statuses = %d/%d, want 1/1", forces, statuses)
	}
	if ball.ObjFlags != object.Flags(0xffffffbf) || ball.Obj130 != nil ||
		ball.ObjOwner != nil || player.Field129 != nil {
		t.Fatalf("released ball = flags:%#x obj130:%p owner:%p head:%p",
			ball.ObjFlags, ball.Obj130, ball.ObjOwner, player.Field129)
	}
	runtime.KeepAlive(marker)
	runtime.KeepAlive(ball)
	runtime.KeepAlive(player)
}
