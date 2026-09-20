package legacy

import (
	"math"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"

	"github.com/opennox/opennox/v1/server"
)

func TestExportBackedObjectUpdatesDispatchDirectlyInGo(t *testing.T) {
	type installFunc func(func(*server.Object)) func()
	tests := []struct {
		name    string
		size    uintptr
		install installFunc
	}{
		{
			name: "PushUpdate",
			size: unsafe.Sizeof(server.PushUpdateData53B030{}),
			install: func(call func(*server.Object)) func() {
				old := pushUpdateCall53B030
				pushUpdateCall53B030 = call
				return func() { pushUpdateCall53B030 = old }
			},
		},
		{
			name: "PlayerUpdate",
			size: unsafe.Sizeof(server.PlayerUpdateData{}),
			install: func(call func(*server.Object)) func() {
				old := Nox_xxx_updatePlayer_4F8100
				Nox_xxx_updatePlayer_4F8100 = call
				return func() { Nox_xxx_updatePlayer_4F8100 = old }
			},
		},
		{
			name: "ProjectileUpdate",
			install: func(call func(*server.Object)) func() {
				old := Nox_xxx_updateProjectile_53AC10
				Nox_xxx_updateProjectile_53AC10 = call
				return func() { Nox_xxx_updateProjectile_53AC10 = old }
			},
		},
		{
			name: "PixieUpdate",
			size: unsafe.Sizeof(server.PixieUpdateData{}),
			install: func(call func(*server.Object)) func() {
				old := Nox_xxx_updatePixie_53CD20
				Nox_xxx_updatePixie_53CD20 = call
				return func() { Nox_xxx_updatePixie_53CD20 = old }
			},
		},
		{
			name: "LifetimeUpdate",
			size: unsafe.Sizeof(server.LifetimeUpdateData53B8F0{}),
			install: func(call func(*server.Object)) func() {
				old := lifetimeUpdateCall53B8F0
				lifetimeUpdateCall53B8F0 = call
				return func() { lifetimeUpdateCall53B8F0 = old }
			},
		},
		{
			name: "OneSecondDieUpdate",
			install: func(call func(*server.Object)) func() {
				old := oneSecondDieUpdateCall53CB60
				oneSecondDieUpdateCall53CB60 = call
				return func() { oneSecondDieUpdateCall53CB60 = old }
			},
		},
		{
			name: "SelfDestructUpdate",
			install: func(call func(*server.Object)) func() {
				old := selfDestructUpdateCall53CC90
				selfDestructUpdateCall53CC90 = call
				return func() { selfDestructUpdateCall53CC90 = old }
			},
		},
		{
			name: "BlackPowderBurnUpdate",
			install: func(call func(*server.Object)) func() {
				old := blackPowderBurnUpdateCall53CCB0
				blackPowderBurnUpdateCall53CCB0 = call
				return func() { blackPowderBurnUpdateCall53CCB0 = old }
			},
		},
		{
			name: "BlackPowderBarrelUpdate",
			install: func(call func(*server.Object)) func() {
				old := blackPowderBarrelUpdateCall53C9A0
				blackPowderBarrelUpdateCall53C9A0 = call
				return func() { blackPowderBarrelUpdateCall53C9A0 = old }
			},
		},
		{
			name: "FistUpdate",
			size: unsafe.Sizeof(server.FistUpdateData{}),
			install: func(call func(*server.Object)) func() {
				old := fistUpdateCall53D400
				fistUpdateCall53D400 = call
				return func() { fistUpdateCall53D400 = old }
			},
		},
		{
			name: "ArachnaphobiaUpdate",
			install: func(call func(*server.Object)) func() {
				old := arachnaphobiaUpdateCall53DA60
				arachnaphobiaUpdateCall53DA60 = call
				return func() { arachnaphobiaUpdateCall53DA60 = old }
			},
		},
		{
			name: "ExpireUpdate",
			install: func(call func(*server.Object)) func() {
				old := expireUpdateCall53DB00
				expireUpdateCall53DB00 = call
				return func() { expireUpdateCall53DB00 = old }
			},
		},
		{
			name: "BreakUpdate",
			install: func(call func(*server.Object)) func() {
				old := breakUpdateCall53DB30
				breakUpdateCall53DB30 = call
				return func() { breakUpdateCall53DB30 = old }
			},
		},
		{
			name: "BreakAndRemoveUpdate",
			install: func(call func(*server.Object)) func() {
				old := breakAndRemoveUpdateCall53DC30
				breakAndRemoveUpdateCall53DC30 = call
				return func() { breakAndRemoveUpdateCall53DC30 = old }
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			callback, size, ok := server.ObjectUpdateHandler(tc.name)
			if !ok || callback == nil || size != tc.size {
				t.Fatalf("registration = %p/%d/%t, want non-nil/%d/true", callback, size, ok, tc.size)
			}

			// A C trampoline would receive a Go object that contains another live
			// Go pointer. Direct dispatch must preserve both the object identity and
			// its native-width pointer graph without crossing the cgo boundary.
			owner := new(server.Object)
			obj := &server.Object{ObjOwner: owner}
			if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(obj)) <= math.MaxUint32 {
				t.Fatalf("object pointer = %p, want address above the ABI32 range", obj)
			}

			var calls int
			restore := tc.install(func(got *server.Object) {
				calls++
				if got != obj {
					t.Fatalf("object = %p, want %p", got, obj)
				}
				if got.ObjOwner != owner {
					t.Fatalf("owner = %p, want %p", got.ObjOwner, owner)
				}
			})
			defer restore()

			server.CallObjectUpdate(callback, obj)
			if calls != 1 {
				t.Fatalf("calls = %d, want 1", calls)
			}
			runtime.KeepAlive(owner)
			runtime.KeepAlive(obj)
		})
	}
}

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
