package opennox

import (
	"testing"
	"unsafe"

	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/server"
)

func TestPlayerCameraUnlock4E6040ClearsOnlyCameraFollow(t *testing.T) {
	target := &server.Object{}
	other := &server.Object{}
	pl := &server.Player{
		Frame3596:       0x12345678,
		CameraFollowObj: target,
		Pos3632Vec:      types.Pointf{X: 12.5, Y: -3.25},
		Obj3640:         other,
	}
	ud := &server.PlayerUpdateData{Player: pl}
	obj := &server.Object{UpdateData: unsafe.Pointer(ud)}

	playerCameraUnlock_4E6040(obj)
	if pl.CameraFollowObj != nil {
		t.Fatalf("camera follow = %p, want nil", pl.CameraFollowObj)
	}
	if pl.Frame3596 != 0x12345678 || pl.Pos3632Vec != (types.Pointf{X: 12.5, Y: -3.25}) || pl.Obj3640 != other {
		t.Fatal("camera unlock changed an adjacent player field")
	}
}

func TestPlayerCameraUnlock4E6040DoesNotAddGuards(t *testing.T) {
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
			playerCameraUnlock_4E6040(tc.obj())
		})
	}
}
