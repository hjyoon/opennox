package opennox

import (
	"testing"
	"unsafe"

	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/server"
)

func TestPlayerCameraFollow4E6060TogglesExactTarget(t *testing.T) {
	first := &server.Object{}
	second := &server.Object{}
	pl := &server.Player{
		Frame3596:       0x12345678,
		CameraFollowObj: first,
		Pos3632Vec:      types.Pointf{X: 12.5, Y: -3.25},
		Obj3640:         second,
	}
	ud := &server.PlayerUpdateData{Player: pl}
	obj := &server.Object{UpdateData: unsafe.Pointer(ud)}

	playerCameraFollow_4E6060(obj, second)
	if pl.CameraFollowObj != second {
		t.Fatalf("camera follow = %p, want %p", pl.CameraFollowObj, second)
	}
	playerCameraFollow_4E6060(obj, second)
	if pl.CameraFollowObj != nil {
		t.Fatalf("second selection camera follow = %p, want nil", pl.CameraFollowObj)
	}
	if pl.Frame3596 != 0x12345678 || pl.Pos3632Vec != (types.Pointf{X: 12.5, Y: -3.25}) || pl.Obj3640 != second {
		t.Fatal("camera toggle changed an adjacent player field")
	}
}

func TestPlayerCameraFollow4E6060AcceptsNilTarget(t *testing.T) {
	pl := &server.Player{CameraFollowObj: &server.Object{}}
	ud := &server.PlayerUpdateData{Player: pl}
	obj := &server.Object{UpdateData: unsafe.Pointer(ud)}

	playerCameraFollow_4E6060(obj, nil)
	if pl.CameraFollowObj != nil {
		t.Fatalf("camera follow = %p, want nil", pl.CameraFollowObj)
	}
	playerCameraFollow_4E6060(obj, nil)
	if pl.CameraFollowObj != nil {
		t.Fatalf("second nil selection camera follow = %p, want nil", pl.CameraFollowObj)
	}
}

func TestPlayerCameraFollow4E6060DoesNotAddGuards(t *testing.T) {
	for _, tc := range []struct {
		name string
		obj  func() *server.Object
	}{
		{name: "nil object", obj: func() *server.Object { return nil }},
		{name: "nil update data", obj: func() *server.Object { return &server.Object{} }},
		{name: "nil player", obj: func() *server.Object {
			ud := &server.PlayerUpdateData{}
			return &server.Object{UpdateData: unsafe.Pointer(ud)}
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Fatal("invalid pointer chain returned without a panic")
				}
			}()
			playerCameraFollow_4E6060(tc.obj(), &server.Object{})
		})
	}
}
