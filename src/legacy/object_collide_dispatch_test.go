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
		{
			name: "BombCollide",
			size: unsafe.Sizeof(server.BombCollideData{}),
			install: func(call func(*server.Object, *server.Object, unsafe.Pointer)) func() {
				original := bombCollideCall4E96F0
				bombCollideCall4E96F0 = call
				return func() { bombCollideCall4E96F0 = original }
			},
		},
		{
			name: "ChestCollide",
			install: func(call func(*server.Object, *server.Object, unsafe.Pointer)) func() {
				original := chestCollideCall4E9C40
				chestCollideCall4E9C40 = call
				return func() { chestCollideCall4E9C40 = original }
			},
		},
		{
			name: "BearTrapCollide",
			install: func(call func(*server.Object, *server.Object, unsafe.Pointer)) func() {
				original := bearTrapCollideCall4EB890
				bearTrapCollideCall4EB890 = call
				return func() { bearTrapCollideCall4EB890 = original }
			},
		},
		{
			name: "PoisonGasTrapCollide",
			install: func(call func(*server.Object, *server.Object, unsafe.Pointer)) func() {
				original := poisonGasTrapCollideCall4EB910
				poisonGasTrapCollideCall4EB910 = call
				return func() { poisonGasTrapCollideCall4EB910 = original }
			},
		},
		{
			name: "TrapDoorCollide",
			size: unsafe.Sizeof(server.TrapDoorCollideData{}),
			install: func(call func(*server.Object, *server.Object, unsafe.Pointer)) func() {
				original := trapDoorCollideCall4EAB60
				trapDoorCollideCall4EAB60 = call
				return func() { trapDoorCollideCall4EAB60 = original }
			},
		},
		{
			name: "BallCollide",
			install: func(call func(*server.Object, *server.Object, unsafe.Pointer)) func() {
				original := ballCollideCall4EBA00
				ballCollideCall4EBA00 = call
				return func() { ballCollideCall4EBA00 = original }
			},
		},
		{
			name: "HomeBaseCollide",
			install: func(call func(*server.Object, *server.Object, unsafe.Pointer)) func() {
				original := homeBaseCollideCall4EBB80
				homeBaseCollideCall4EBB80 = func(first, second *server.Object, collision unsafe.Pointer) uint32 {
					call(first, second, collision)
					return 0
				}
				return func() { homeBaseCollideCall4EBB80 = original }
			},
		},
		{
			name: "ManaDrainCollide",
			size: unsafe.Sizeof(server.ManaDrainCollideData{}),
			install: func(call func(*server.Object, *server.Object, unsafe.Pointer)) func() {
				original := manaDrainCollideCall4E9490
				manaDrainCollideCall4E9490 = call
				return func() { manaDrainCollideCall4E9490 = original }
			},
		},
		{
			name: "AwardSpellCollide",
			size: unsafe.Sizeof(server.AwardSpellCollideData{}),
			install: func(call func(*server.Object, *server.Object, unsafe.Pointer)) func() {
				original := awardSpellCollideCall4EAD20
				awardSpellCollideCall4EAD20 = func(first, second *server.Object, collision unsafe.Pointer) int32 {
					call(first, second, collision)
					return 0
				}
				return func() { awardSpellCollideCall4EAD20 = original }
			},
		},
		{
			name: "GlyphCollide",
			install: func(call func(*server.Object, *server.Object, unsafe.Pointer)) func() {
				original := glyphCollideCall4E9A00
				glyphCollideCall4E9A00 = call
				return func() { glyphCollideCall4E9A00 = original }
			},
		},
		{
			name: "SpellProjectileCollide",
			install: func(call func(*server.Object, *server.Object, unsafe.Pointer)) func() {
				original := spellProjectileCollideCall4E9500
				spellProjectileCollideCall4E9500 = call
				return func() { spellProjectileCollideCall4E9500 = original }
			},
		},
		{
			name: "TeleportWakeCollide",
			size: unsafe.Sizeof(server.TeleportWakeCollideData{}),
			install: func(call func(*server.Object, *server.Object, unsafe.Pointer)) func() {
				original := teleportWakeCollideCall4EAE30
				teleportWakeCollideCall4EAE30 = call
				return func() { teleportWakeCollideCall4EAE30 = original }
			},
		},
		{
			name: "AnkhCollide",
			install: func(call func(*server.Object, *server.Object, unsafe.Pointer)) func() {
				original := ankhCollideCall4EBF40
				ankhCollideCall4EBF40 = call
				return func() { ankhCollideCall4EBF40 = original }
			},
		},
		{
			name: "ExitCollide",
			size: unsafe.Sizeof(server.ExitCollideData{}),
			install: func(call func(*server.Object, *server.Object, unsafe.Pointer)) func() {
				original := exitCollideCall4E9090
				exitCollideCall4E9090 = call
				return func() { exitCollideCall4E9090 = original }
			},
		},
		{
			name: "OwnCollide",
			install: func(call func(*server.Object, *server.Object, unsafe.Pointer)) func() {
				original := ownCollideCall4EA2C0
				ownCollideCall4EA2C0 = call
				return func() { ownCollideCall4EA2C0 = original }
			},
		},
		{
			name: "SparkCollide",
			size: 8,
			install: func(call func(*server.Object, *server.Object, unsafe.Pointer)) func() {
				original := sparkCollideCall4EA300
				sparkCollideCall4EA300 = call
				return func() { sparkCollideCall4EA300 = original }
			},
		},
		{
			name: "BarrelCollide",
			install: func(call func(*server.Object, *server.Object, unsafe.Pointer)) func() {
				original := barrelCollideCall4EAAA0
				barrelCollideCall4EAAA0 = call
				return func() { barrelCollideCall4EAAA0 = original }
			},
		},
		{
			name: "AudioEventCollide",
			size: unsafe.Sizeof(server.AudioEventCollideData{}),
			install: func(call func(*server.Object, *server.Object, unsafe.Pointer)) func() {
				original := audioEventCollideCall4EAAD0
				audioEventCollideCall4EAAD0 = call
				return func() { audioEventCollideCall4EAAD0 = original }
			},
		},
		{
			name: "TriggerCollide",
			install: func(call func(*server.Object, *server.Object, unsafe.Pointer)) func() {
				original := triggerCollideCall54FCD0
				triggerCollideCall54FCD0 = call
				return func() { triggerCollideCall54FCD0 = original }
			},
		},
		{
			name: "TeleportCollide",
			size: unsafe.Sizeof(server.TeleportCollideData{}),
			install: func(call func(*server.Object, *server.Object, unsafe.Pointer)) func() {
				original := teleportCollideCall4EACA0
				teleportCollideCall4EACA0 = call
				return func() { teleportCollideCall4EACA0 = original }
			},
		},
		{
			name: "DieCollide",
			install: func(call func(*server.Object, *server.Object, unsafe.Pointer)) func() {
				original := dieCollideCall4E99B0
				dieCollideCall4E99B0 = call
				return func() { dieCollideCall4E99B0 = original }
			},
		},
		{
			name: "SignCollide",
			install: func(call func(*server.Object, *server.Object, unsafe.Pointer)) func() {
				original := signCollideCall4EAB40
				signCollideCall4EAB40 = call
				return func() { signCollideCall4EAB40 = original }
			},
		},
		{
			name: "PentagramCollide",
			install: func(call func(*server.Object, *server.Object, unsafe.Pointer)) func() {
				original := pentagramCollideCall4EAB20
				pentagramCollideCall4EAB20 = call
				return func() { pentagramCollideCall4EAB20 = original }
			},
		},
		{
			name: "FlagCollide",
			install: func(call func(*server.Object, *server.Object, unsafe.Pointer)) func() {
				original := flagCollideCall4EA400
				flagCollideCall4EA400 = call
				return func() { flagCollideCall4EA400 = original }
			},
		},
		{
			name: "CrownCollide",
			install: func(call func(*server.Object, *server.Object, unsafe.Pointer)) func() {
				original := crownCollideCall4EBB50
				crownCollideCall4EBB50 = func(first, second *server.Object, collision unsafe.Pointer) uintptr {
					call(first, second, collision)
					return 0
				}
				return func() { crownCollideCall4EBB50 = original }
			},
		},
		{
			name: "MonsterGeneratorCollide",
			install: func(call func(*server.Object, *server.Object, unsafe.Pointer)) func() {
				original := monsterGeneratorCollideCall4EBE10
				monsterGeneratorCollideCall4EBE10 = call
				return func() { monsterGeneratorCollideCall4EBE10 = original }
			},
		},
		{
			name: "SoulGateCollide",
			size: unsafe.Sizeof(server.SoulGateCollideData{}),
			install: func(call func(*server.Object, *server.Object, unsafe.Pointer)) func() {
				original := soulGateCollideCall4EBE40
				soulGateCollideCall4EBE40 = call
				return func() { soulGateCollideCall4EBE40 = original }
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
