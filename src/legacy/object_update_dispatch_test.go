package legacy

import (
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"

	"github.com/opennox/opennox/v1/server"
)

func TestMonsterUpdateDispatchStaysInGo(t *testing.T) {
	callback, size, ok := server.ObjectUpdateHandler("MonsterUpdate")
	if !ok || callback == nil || size != unsafe.Sizeof(server.MonsterUpdateData{}) {
		t.Fatalf("MonsterUpdate registration = %p/%d/%t", callback, size, ok)
	}

	original := Nox_xxx_unitUpdateMonster_50A5C0
	t.Cleanup(func() {
		Nox_xxx_unitUpdateMonster_50A5C0 = original
	})

	var called *server.Object
	Nox_xxx_unitUpdateMonster_50A5C0 = func(obj *server.Object) {
		called = obj
	}

	// Keep a live Go pointer in the object. A regression to the legacy C
	// trampoline would make cgo reject this pointer graph before the callback.
	owner := &server.Object{}
	obj := &server.Object{
		ObjClass: object.ClassMonster,
		ObjOwner: owner,
	}
	server.CallObjectUpdate(callback, obj)
	if called != obj {
		t.Fatalf("MonsterUpdate called with %p, want %p", called, obj)
	}
}

type pentagramUpdateLegacyServer53BEF0 struct {
	Server
	srv *server.Server
}

func (s *pentagramUpdateLegacyServer53BEF0) S() *server.Server { return s.srv }

func TestPentagramUpdateDispatchStaysInGo(t *testing.T) {
	original := GetServer
	GetServer = func() Server {
		return &pentagramUpdateLegacyServer53BEF0{srv: new(server.Server)}
	}
	t.Cleanup(func() {
		GetServer = original
	})

	for _, tc := range []struct {
		name  string
		data  server.PentagramUpdateData
		check func(*testing.T, *server.PentagramUpdateData)
	}{
		{
			name: "PentagramUpdate",
			data: server.PentagramUpdateData{State: 1, AnimationFrame: 1},
			check: func(t *testing.T, data *server.PentagramUpdateData) {
				if data.AnimationTick != 1 {
					t.Fatalf("animation tick = %d, want 1", data.AnimationTick)
				}
			},
		},
		{
			name: "InvisiblePentagramUpdate",
			data: server.PentagramUpdateData{Triggered: 1},
			check: func(t *testing.T, data *server.PentagramUpdateData) {
				if data.Triggered != 0 {
					t.Fatalf("triggered = %d, want 0", data.Triggered)
				}
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			callback, size, ok := server.ObjectUpdateHandler(tc.name)
			if !ok || callback == nil || size != unsafe.Sizeof(server.PentagramUpdateData{}) {
				t.Fatalf("registration = %p/%d/%t", callback, size, ok)
			}

			owner := &server.Object{}
			obj := &server.Object{
				ObjOwner:   owner,
				UpdateData: unsafe.Pointer(&tc.data),
			}
			server.CallObjectUpdate(callback, obj)
			tc.check(t, &tc.data)
		})
	}
}
