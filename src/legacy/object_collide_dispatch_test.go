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
		{
			name: "SpiderSpitCollide",
			install: func(call func(*server.Object, *server.Object, unsafe.Pointer)) func() {
				original := webbingCollideCall4EA380
				webbingCollideCall4EA380 = call
				return func() { webbingCollideCall4EA380 = original }
			},
		},
		{
			name: "DeathBallCollide",
			install: func(call func(*server.Object, *server.Object, unsafe.Pointer)) func() {
				original := deathBallCollideCall4E9E90
				deathBallCollideCall4E9E90 = call
				return func() { deathBallCollideCall4E9E90 = original }
			},
		},
		{
			name: "DeathBallFragmentCollide",
			install: func(call func(*server.Object, *server.Object, unsafe.Pointer)) func() {
				original := deathBallFragmentCollideCall4E9FE0
				deathBallFragmentCollideCall4E9FE0 = call
				return func() { deathBallFragmentCollideCall4E9FE0 = original }
			},
		},
		{
			name: "FistCollide",
			install: func(call func(*server.Object, *server.Object, unsafe.Pointer)) func() {
				original := fistCollideCall4EADF0
				fistCollideCall4EADF0 = call
				return func() { fistCollideCall4EADF0 = original }
			},
		},
		{
			name: "DamageCollide",
			size: unsafe.Sizeof(server.DamageCollideData{}),
			install: func(call func(*server.Object, *server.Object, unsafe.Pointer)) func() {
				original := damageCollideCall4E9430
				damageCollideCall4E9430 = call
				return func() { damageCollideCall4E9430 = original }
			},
		},
		{
			name: "SparkExplosionCollide",
			size: unsafe.Sizeof(server.SparkExplosionCollideData{}),
			install: func(call func(*server.Object, *server.Object, unsafe.Pointer)) func() {
				original := sparkExplosionCollideCall4E9AC0
				sparkExplosionCollideCall4E9AC0 = call
				return func() { sparkExplosionCollideCall4E9AC0 = original }
			},
		},
		{
			name: "WallReflectCollide",
			size: unsafe.Sizeof(server.ProjectileCollideData{}),
			install: func(call func(*server.Object, *server.Object, unsafe.Pointer)) func() {
				original := wallReflectCollideCall4E9D80
				wallReflectCollideCall4E9D80 = call
				return func() { wallReflectCollideCall4E9D80 = original }
			},
		},
		{
			name: "WallReflectSparkCollide",
			size: unsafe.Sizeof(server.ProjectileCollideData{}),
			install: func(call func(*server.Object, *server.Object, unsafe.Pointer)) func() {
				original := wallReflectSparkCollideCall4EA200
				wallReflectSparkCollideCall4EA200 = call
				return func() { wallReflectSparkCollideCall4EA200 = original }
			},
		},
		{
			name: "PixieCollide",
			size: unsafe.Sizeof(server.ProjectileCollideData{}),
			install: func(call func(*server.Object, *server.Object, unsafe.Pointer)) func() {
				original := pixieCollideCall4EA080
				pixieCollideCall4EA080 = call
				return func() { pixieCollideCall4EA080 = original }
			},
		},
		{
			name: "YellowStarShotCollide",
			size: unsafe.Sizeof(server.ProjectileCollideData{}),
			install: func(call func(*server.Object, *server.Object, unsafe.Pointer)) func() {
				original := yellowStarShotCollideCall4E9E50
				yellowStarShotCollideCall4E9E50 = call
				return func() { yellowStarShotCollideCall4E9E50 = original }
			},
		},
		{
			name: "BoomCollide",
			install: func(call func(*server.Object, *server.Object, unsafe.Pointer)) func() {
				original := boomCollideCall4E9770
				boomCollideCall4E9770 = call
				return func() { boomCollideCall4E9770 = original }
			},
		},
		{
			name: "ChakramInMotionCollide",
			install: func(call func(*server.Object, *server.Object, unsafe.Pointer)) func() {
				original := chakramCollideCall4EAF00
				chakramCollideCall4EAF00 = call
				return func() { chakramCollideCall4EAF00 = original }
			},
		},
		{
			name: "ArrowCollide",
			size: unsafe.Sizeof(server.ArrowCollideData{}),
			install: func(call func(*server.Object, *server.Object, unsafe.Pointer)) func() {
				original := arrowCollideCall4EB490
				arrowCollideCall4EB490 = call
				return func() { arrowCollideCall4EB490 = original }
			},
		},
		{
			name: "MonsterArrowCollide",
			size: unsafe.Sizeof(server.MonsterArrowCollideData{}),
			install: func(call func(*server.Object, *server.Object, unsafe.Pointer)) func() {
				original := monsterArrowCollideCall4EB800
				monsterArrowCollideCall4EB800 = call
				return func() { monsterArrowCollideCall4EB800 = original }
			},
		},
		{
			name: "UndeadKillerCollide",
			size: unsafe.Sizeof(server.UndeadKillerCollideData{}),
			install: func(call func(*server.Object, *server.Object, unsafe.Pointer)) func() {
				original := undeadKillerCollideCall4EBD40
				undeadKillerCollideCall4EBD40 = call
				return func() { undeadKillerCollideCall4EBD40 = original }
			},
		},
		{
			name: "HarpoonCollide",
			size: unsafe.Sizeof(server.HarpoonCollideData{}),
			install: func(call func(*server.Object, *server.Object, unsafe.Pointer)) func() {
				original := harpoonCollideCall4EB6A0
				harpoonCollideCall4EB6A0 = call
				return func() { harpoonCollideCall4EB6A0 = original }
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
