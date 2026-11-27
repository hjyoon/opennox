package legacy

import (
	"fmt"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/things"
	"github.com/opennox/opennox/v1/server"
)

func TestRawDefaultDamageRegisteredCallbacksKeepGoOwned(t *testing.T) {
	srv := npcReflectLegacyServer4E17B0(t)
	target := &server.Object{InvFirstItem: &server.Object{}}
	source := &server.Object{HealthData: &server.HealthData{Cur: 7}}
	weapon := &server.Object{InvHolder: source}
	objectArg := func(unit *server.Object) rawDefaultDamageNativeArgument4E0B30 {
		return rawDefaultDamageNativeArgument4E0B30{Kind: "object", Pointer: unsafe.Pointer(unit)}
	}

	t.Run("damage", func(t *testing.T) {
		damage := int32(-17)
		const kind = object.DamageType(0x10203040)
		calls := 0
		callback := objectDamageNativeProbePtr()
		server.RegisterObjectDamageGo(fmt.Sprintf("RawGoOwnedDamage%d", objectDamageNativeTestSequence.Add(1)), callback,
			func(a, b, c *server.Object, amount int32, typ object.DamageType) bool {
				calls++
				if a != target || b != source || c != weapon || amount != damage || typ != kind {
					t.Fatalf("damage arguments = %p/%p/%p/%d/%#x", a, b, c, amount, typ)
				}
				return true
			})
		reply := rawDefaultDamageInvokeService4E0B30(rawDefaultDamageNativeService4E0B30{
			Indirect: true, Site: 0x4e0d05, Callback: callback,
			Arguments: []rawDefaultDamageNativeArgument4E0B30{
				objectArg(target), objectArg(source), objectArg(weapon),
				{Kind: "word", Word: uint32(damage)}, {Kind: "word", Word: uint32(kind)},
			},
		})
		if calls != 1 || reply.Word != 0 {
			t.Fatalf("damage callback ran %d times and exposed return word %d", calls, reply.Word)
		}
	})

	runtime := defaultDamageWorldRuntime4E0B30(srv)
	for _, tc := range []struct {
		name     string
		site     uint32
		callback unsafe.Pointer
		bind     *func(*server.Object, *server.Object) int
		source   *server.Object
	}{
		{"default-sound", 0x4e1049, runtime.DefaultDamageSoundC, &Nox_xxx_soundDefaultDamageSound_532E20, source},
		{"player-sound", 0x4e104f, runtime.PlayerDamageSoundC, &Nox_xxx_soundPlayerDamageSound_5328B0, nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			old := *tc.bind
			t.Cleanup(func() { *tc.bind = old })
			calls := 0
			*tc.bind = func(a, b *server.Object) int {
				calls++
				if a != target || b != tc.source {
					t.Fatalf("sound arguments = %p/%p", a, b)
				}
				return 0x12345678
			}
			reply := rawDefaultDamageInvokeService4E0B30(rawDefaultDamageNativeService4E0B30{
				Indirect: true, Site: tc.site, Callback: tc.callback,
				Arguments: []rawDefaultDamageNativeArgument4E0B30{objectArg(target), objectArg(tc.source)},
			})
			if calls != 1 || reply.Word != 0 {
				t.Fatalf("sound callback ran %d times and exposed return word %d", calls, reply.Word)
			}
		})
	}
}

func TestRawDefaultDamageUsesOriginalChildEntries(t *testing.T) {
	srv := npcReflectLegacyServer4E17B0(t)
	srv.FreeObjectTypes()
	t.Cleanup(srv.FreeObjectTypes)
	target := unsafe.Pointer(&server.Object{InvFirstItem: &server.Object{}})
	source := unsafe.Pointer(&server.Object{HealthData: &server.HealthData{Cur: 7}})
	weapon := unsafe.Pointer(&server.Object{InvHolder: (*server.Object)(source)})
	invoke := func(t *testing.T, name string, attacker, item unsafe.Pointer, typ int32) {
		t.Helper()
		if err := srv.Types.ReadObjectType(&things.Thing{Name: "RawChild" + name, OnDamage: &things.ProcFunc{Name: name}}); err != nil {
			t.Fatal(err)
		}
		damage := int32(-17)
		reply := rawDefaultDamageInvokeService4E0B30(rawDefaultDamageNativeService4E0B30{
			Indirect: true, Site: 0x4e0d05, Callback: srv.Types.ByID("RawChild" + name).Damage,
			Arguments: []rawDefaultDamageNativeArgument4E0B30{
				{Kind: "object", Pointer: target}, {Kind: "object", Pointer: attacker}, {Kind: "object", Pointer: item},
				{Kind: "word", Word: uint32(damage)}, {Kind: "word", Word: uint32(typ)},
			},
		})
		if reply.Word != 0 {
			t.Fatalf("raw child return word reached parent: %d", reply.Word)
		}
	}

	for _, tc := range []struct {
		name string
		site uint32
		typ  int32
		want int32
	}{
		{"StoneDamage", 0x4e24c9, -11, -17},
		{"MechGolemDamage", 0x4e250b, 9, -34},
		{"FlammableDamage", 0x4e254d, 12, 9999999},
		{"BlackPowderDamage", 0x4e258f, 2, 999999},
	} {
		t.Run(tc.name, func(t *testing.T) {
			old := rawScalarDamageObserveParent4E24B0
			t.Cleanup(func() { rawScalarDamageObserveParent4E24B0 = old })
			calls := 0
			rawScalarDamageObserveParent4E24B0 = func(site uint32, a, b, c unsafe.Pointer, damage, typ int32) int32 {
				calls++
				if site != tc.site || a != target || b != source || c != weapon || damage != tc.want || typ != tc.typ {
					t.Fatalf("raw child parent call = %#x %p/%p/%p %d/%d", site, a, b, c, damage, typ)
				}
				return -1234
			}
			invoke(t, tc.name, source, weapon, tc.typ)
			if calls != 1 {
				t.Fatalf("raw child delegated %d times, want 1", calls)
			}
		})
	}

	t.Run("SkeletonDamage", func(t *testing.T) {
		calls := 0
		defer pushRawSkeletonObserver4E23C0(&rawSkeletonObserver4E23C0{
			Service: func(call rawSkeletonService4E23C0) (uint32, bool) {
				if call.Site == 0x4e23ce {
					return 0, true
				}
				calls++
				a := call.Arguments
				if call.Site != 0x4e249d || a[0].Pointer != target || a[1].Pointer != nil || a[2].Pointer != nil || int32(a[3].Word) != -17 || int32(a[4].Word) != -11 {
					t.Fatalf("raw Skeleton parent call = %+v", call)
				}
				return 1234, true
			},
		})()
		invoke(t, "SkeletonDamage", nil, nil, -11)
		if calls != 1 {
			t.Fatalf("raw Skeleton delegated %d times, want 1", calls)
		}
	})
}
