package legacy

import (
	"math"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/strman"
	"github.com/opennox/libs/things"

	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/server"
)

func TestObjectUpdateParsersLoadNativeRecords(t *testing.T) {
	srv := server.NewHigh(nil, nil, strman.New())
	t.Cleanup(srv.Close)
	srv.FreeObjectTypes()
	t.Cleanup(srv.FreeObjectTypes)

	load := func(name string, args ...string) unsafe.Pointer {
		t.Helper()
		id := "UpdateParserFixture" + name
		if err := srv.Types.ReadObjectType(&things.Thing{
			Name: id, OnUpdate: &things.ProcFunc{Name: name, Args: args},
		}); err != nil {
			t.Fatalf("load %s: %v", name, err)
		}
		typ := srv.Types.ByID(id)
		if typ == nil || typ.UpdateData == nil {
			t.Fatalf("%s has no registered update data", name)
		}
		if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(typ.UpdateData) <= math.MaxUint32 {
			t.Fatalf("%s record pointer does not exercise native pointer width", name)
		}
		return typ.UpdateData
	}
	push := (*server.PushUpdateData53B030)(load("PushUpdate", "-0", "2.75"))
	if !math.Signbit(float64(push.Radius)) || !math.Signbit(float64(push.RadiusCopy)) || push.Force != 2.75 {
		t.Fatalf("push record = %+v", push)
	}
	damage := (*[3]int32)(load("LoopAndDamageUpdate", "\t-3+17", "19tail"))
	if *damage != [3]int32{-3, 17, 19} {
		t.Fatalf("loop-and-damage record = %v", *damage)
	}
	lifetime := (*server.LifetimeUpdateData53B8F0)(load("LifetimeUpdate", "  +280suffix"))
	if lifetime.Duration != 280 {
		t.Fatalf("lifetime duration = %d, want 280", lifetime.Duration)
	}
	for _, name := range []string{"TriggerUpdate", "ToggleUpdate"} {
		trigger := (*server.TriggerUpdateData)(load(name, "HasteOn", "HasteOff"))
		if trigger.SoundActivate != uint32(sound.ByName("HasteOn")) || trigger.SoundDeactivate != uint32(sound.ByName("HasteOff")) {
			t.Fatalf("%s parsed sound IDs = %d, %d", name, trigger.SoundActivate, trigger.SoundDeactivate)
		}
	}
	skull := (*server.SkullUpdateData)(load("SkullUpdate", "  Harpoon", "ignored"))
	if string(skull.ProjectileName[:8]) != "Harpoon\x00" || skull.ProjectileType != 0 {
		t.Fatalf("skull projectile = %q, %d", skull.ProjectileName, skull.ProjectileType)
	}
}

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
			name: "PhantomPlayerUpdate",
			size: unsafe.Sizeof(server.PhantomPlayerUpdateData{}),
			install: func(call func(*server.Object)) func() {
				old := phantomPlayerUpdateCall53B860
				phantomPlayerUpdateCall53B860 = call
				return func() { phantomPlayerUpdateCall53B860 = old }
			},
		},
		{
			name: "SkullUpdate",
			size: unsafe.Sizeof(server.SkullUpdateData{}),
			install: func(call func(*server.Object)) func() {
				old := skullUpdateCall54F9A0
				skullUpdateCall54F9A0 = call
				return func() { skullUpdateCall54F9A0 = old }
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
			name: "BlowUpdate",
			install: func(call func(*server.Object)) func() {
				old := blowUpdateCall53C160
				blowUpdateCall53C160 = call
				return func() { blowUpdateCall53C160 = old }
			},
		},
		{
			name: "SentryGlobeUpdate",
			size: unsafe.Sizeof(server.SentryUpdateData{}),
			install: func(call func(*server.Object)) func() {
				old := sentryGlobeUpdateCall510E60
				sentryGlobeUpdateCall510E60 = call
				return func() { sentryGlobeUpdateCall510E60 = old }
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
			name: "MeteorShowerUpdate",
			size: unsafe.Sizeof(server.MeteorUpdateData{}),
			install: func(call func(*server.Object)) func() {
				old := meteorShowerUpdateCall53D5A0
				meteorShowerUpdateCall53D5A0 = call
				return func() { meteorShowerUpdateCall53D5A0 = old }
			},
		},
		{
			name: "MeteorUpdate",
			size: unsafe.Sizeof(server.MeteorUpdateData{}),
			install: func(call func(*server.Object)) func() {
				old := meteorUpdateCall53D6E0
				meteorUpdateCall53D6E0 = call
				return func() { meteorUpdateCall53D6E0 = old }
			},
		},
		{
			name: "TelekinesisUpdate",
			install: func(call func(*server.Object)) func() {
				old := telekinesisUpdateCall53D330
				telekinesisUpdateCall53D330 = call
				return func() { telekinesisUpdateCall53D330 = old }
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
			name: "ToxicCloudUpdate",
			size: unsafe.Sizeof(server.ToxicCloudUpdateData{}),
			install: func(call func(*server.Object)) func() {
				old := toxicCloudUpdateCall53D850
				toxicCloudUpdateCall53D850 = call
				return func() { toxicCloudUpdateCall53D850 = old }
			},
		},
		{
			name: "SmallToxicCloudUpdate",
			size: unsafe.Sizeof(server.ToxicCloudUpdateData{}),
			install: func(call func(*server.Object)) func() {
				old := smallToxicCloudUpdateCall53D960
				smallToxicCloudUpdateCall53D960 = call
				return func() { smallToxicCloudUpdateCall53D960 = old }
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
		{
			name: "ChakramInMotionUpdate",
			size: unsafe.Sizeof(server.ChakramUpdateData{}),
			install: func(call func(*server.Object)) func() {
				old := chakramInMotionUpdateCall53DCC0
				chakramInMotionUpdateCall53DCC0 = call
				return func() { chakramInMotionUpdateCall53DCC0 = old }
			},
		},
		{
			name: "CrownUpdate",
			size: unsafe.Sizeof(server.CrownUpdateData{}),
			install: func(call func(*server.Object)) func() {
				old := crownUpdateCall53E1D0
				crownUpdateCall53E1D0 = call
				return func() { crownUpdateCall53E1D0 = old }
			},
		},
		{
			name: "HarpoonUpdate",
			size: 4,
			install: func(call func(*server.Object)) func() {
				old := Nox_xxx_updateHarpoon_54F380
				Nox_xxx_updateHarpoon_54F380 = call
				return func() { Nox_xxx_updateHarpoon_54F380 = old }
			},
		},
		{
			name: "MoverUpdate",
			size: unsafe.Sizeof(server.MoverUpdateData{}),
			install: func(call func(*server.Object)) func() {
				old := moverUpdateCall54F740
				moverUpdateCall54F740 = call
				return func() { moverUpdateCall54F740 = old }
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
