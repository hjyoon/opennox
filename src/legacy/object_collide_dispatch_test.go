package legacy

import (
	"math"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

func TestCoreCollideDispatchStaysInGo(t *testing.T) {
	type installFunc func(func(*server.Object, *server.Object, unsafe.Pointer)) func()
	tests := []struct {
		name    string
		size    uintptr
		install installFunc
	}{
		{
			name: "PlayerCollide",
			install: func(call func(*server.Object, *server.Object, unsafe.Pointer)) func() {
				original := playerCollideCall4E8460
				playerCollideCall4E8460 = call
				return func() { playerCollideCall4E8460 = original }
			},
		},
		{
			name: "MonsterCollide",
			install: func(call func(*server.Object, *server.Object, unsafe.Pointer)) func() {
				original := monsterCollideCall4E83B0
				monsterCollideCall4E83B0 = func(first, second *server.Object, collision unsafe.Pointer) unsafe.Pointer {
					call(first, second, collision)
					return nil
				}
				return func() { monsterCollideCall4E83B0 = original }
			},
		},
		{
			name: "MimicCollide",
			install: func(call func(*server.Object, *server.Object, unsafe.Pointer)) func() {
				original := mimicCollideCall4E83D0
				mimicCollideCall4E83D0 = func(first, second *server.Object, collision unsafe.Pointer) unsafe.Pointer {
					call(first, second, collision)
					return nil
				}
				return func() { mimicCollideCall4E83D0 = original }
			},
		},
		{
			name: "ProjectileCollide",
			size: unsafe.Sizeof(server.ProjectileCollideData{}),
			install: func(call func(*server.Object, *server.Object, unsafe.Pointer)) func() {
				original := projectileCollideCall4E87B0
				projectileCollideCall4E87B0 = call
				return func() { projectileCollideCall4E87B0 = original }
			},
		},
		{
			name: "ProjectileSparkCollide",
			size: unsafe.Sizeof(server.ProjectileCollideData{}),
			install: func(call func(*server.Object, *server.Object, unsafe.Pointer)) func() {
				original := projectileSparkCollideCall4E8880
				projectileSparkCollideCall4E8880 = call
				return func() { projectileSparkCollideCall4E8880 = original }
			},
		},
		{
			name: "DoorCollide",
			install: func(call func(*server.Object, *server.Object, unsafe.Pointer)) func() {
				original := doorCollideCall4E8AC0
				doorCollideCall4E8AC0 = call
				return func() { doorCollideCall4E8AC0 = original }
			},
		},
		{
			name: "PickupCollide",
			install: func(call func(*server.Object, *server.Object, unsafe.Pointer)) func() {
				original := pickupCollideCall4E8DF0
				pickupCollideCall4E8DF0 = func(first, second *server.Object, collision unsafe.Pointer) uintptr {
					call(first, second, collision)
					return 0
				}
				return func() { pickupCollideCall4E8DF0 = original }
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			callback, size, ok := server.ObjectCollideHandler(tc.name)
			if !ok || callback == nil || size != tc.size {
				t.Fatalf("registration = %p/%d/%t, want non-nil/%d/true", callback, size, ok, tc.size)
			}

			owner := new(server.Object)
			first := &server.Object{ObjOwner: owner}
			second := new(server.Object)
			collision := new(byte)
			if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(first)) <= math.MaxUint32 {
				t.Fatalf("object pointer = %p, want address above the ABI32 range", first)
			}

			var calls int
			restore := tc.install(func(gotFirst, gotSecond *server.Object, gotCollision unsafe.Pointer) {
				calls++
				if gotFirst != first || gotSecond != second || gotCollision != unsafe.Pointer(collision) {
					t.Fatalf("collide args = (%p, %p, %p), want (%p, %p, %p)",
						gotFirst, gotSecond, gotCollision, first, second, collision)
				}
				if gotFirst.ObjOwner != owner {
					t.Fatalf("object owner = %p, want %p", gotFirst.ObjOwner, owner)
				}
			})
			defer restore()

			server.CallObjectCollide(callback, first, second, unsafe.Pointer(collision))
			if calls != 1 {
				t.Fatalf("calls = %d, want 1", calls)
			}
			runtime.KeepAlive(owner)
			runtime.KeepAlive(first)
			runtime.KeepAlive(second)
			runtime.KeepAlive(collision)
		})
	}
}

func TestNoopCollidesDispatchDirectly(t *testing.T) {
	for _, name := range []string{"DefaultCollide", "ElevatorCollide", "TelekinesisCollide"} {
		t.Run(name, func(t *testing.T) {
			callback, _, ok := server.ObjectCollideHandler(name)
			if !ok || callback == nil {
				t.Fatalf("registration = %p/%t", callback, ok)
			}
			server.CallObjectCollide(callback, new(server.Object), new(server.Object), unsafe.Pointer(new(byte)))
		})
	}
}
