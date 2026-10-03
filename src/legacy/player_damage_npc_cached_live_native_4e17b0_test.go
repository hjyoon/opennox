package legacy

import (
	"fmt"
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

// Direct Go PlayerDamage body with C-owned >4GiB records. One test hook calls
// the production electric-armor service, then replaces UpdateData. The shared
// DefaultDamage runtime, protection, attribution and native UnitSetHP are real;
// this is not an unmodified C-dispatcher or natural stock-map mutation test.
func TestPlayerDamageNPCCachedLiveNative4E17B0ElectricCRecords(t *testing.T) {
	srv := npcReflectLegacyServer4E17B0(t)
	srv.Audio.Init(srv)
	t.Cleanup(srv.Audio.Free)
	if !srv.Objs.Init(4) {
		t.Fatal("native allocator")
	}
	t.Cleanup(srv.Objs.FreeObjects)
	for _, typ := range []object.DamageType{object.DamageElectric, object.DamageAirborneElectric} {
		t.Run(fmt.Sprint(typ), func(t *testing.T) {
			target, source := srv.Objs.NewObject(&server.ObjectType{}), srv.Objs.NewObject(&server.ObjectType{})
			old, freeOld := alloc.New(server.MonsterUpdateData{})
			live, freeLive := alloc.New(server.MonsterUpdateData{})
			playerUpdate, freePlayerUpdate := alloc.New(server.PlayerUpdateData{})
			player, freePlayer := alloc.New(server.Player{})
			health, freeHealth := alloc.New(server.HealthData{})
			for _, free := range []func(){freeOld, freeLive, freePlayerUpdate, freePlayer, freeHealth} {
				t.Cleanup(free)
			}
			*old = server.MonsterUpdateData{Field1: math.Float32bits(-0.5), Field547: 99, Field546: 77}
			*live = server.MonsterUpdateData{Field1: math.Float32bits(0.25), StatusFlags: 0x20000000}
			*player = server.Player{PlayerInd: 7}
			*playerUpdate = server.PlayerUpdateData{Player: player, State: server.PlayerState13}
			*health = server.HealthData{Cur: 20, Max: 20, Field2: 20}
			target.TypeInd, target.ObjClass, target.ObjSubClass, target.ObjFlags, target.Material = 71, object.ClassMonster, 0x11012, 0, 0x4000
			target.UpdateData, target.HealthData = unsafe.Pointer(old), health
			source.TypeInd, source.ObjClass, source.ObjFlags = 72, object.ClassPlayer, 0
			source.UpdateData, source.PrevPos = unsafe.Pointer(playerUpdate), types.Ptf(-3, 7)
			for _, ptr := range []unsafe.Pointer{unsafe.Pointer(target), unsafe.Pointer(source), unsafe.Pointer(old), unsafe.Pointer(live), unsafe.Pointer(playerUpdate), unsafe.Pointer(player), unsafe.Pointer(health)} {
				if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(ptr) <= math.MaxUint32 {
					t.Fatalf("native pointer=%p, want >4 GiB", ptr)
				}
			}
			beforeSource, beforePlayerUpdate, beforePlayer := *source, *playerUpdate, *player
			tail := defaultDamageWorldRuntime4E0B30(srv)
			calls := 0
			r := server.PlayerDamageRuntime4E17B0{
				Frame:            srv.Frame,
				QuestMode:        func() bool { return false },
				QuestDamageScale: func() float32 { t.Fatal("non-Quest scale"); return 0 },
				ElectricArmorScale: func(v *server.Object) float32 {
					if v != target || old.Field547 != 0 || old.Field546 != 77 {
						t.Fatal("electric service preceded cached marker reset")
					}
					scale := playerDamageElectricArmorScale4E2220(srv, v)
					if scale != 1 {
						t.Fatal("unarmored production electric scale")
					}
					calls++
					target.UpdateData = unsafe.Pointer(live)
					return scale
				},
				DefaultDamage: func(v, a, w *server.Object, d int32, gotType object.DamageType) bool {
					if d != 3 || old.Field547 != 2 || old.Field546 != uint32(typ) || live.Field1 != math.Float32bits(0.25) {
						t.Fatal("electric live carry or cached fallback marker")
					}
					return server.DefaultDamageWorld4E0B30(v, a, w, d, gotType, tail)
				},
				Unsupported: func(reason string, _, _, _ *server.Object, _ int32, _ object.DamageType) {
					t.Fatalf("native NPC electric: %s", reason)
				},
			}
			if h, result := server.PlayerDamageNative4E17B0(target, source, source, 3, typ, r); !h || !result || calls != 1 || health.Cur != 17 ||
				old.Field1 != math.Float32bits(-0.5) || live.Field1 != math.Float32bits(0.25) || live.Field547 != 2 || live.Field546 != uint32(typ) || live.Field523_2 != 2 ||
				target.Obj130 != source || target.Field131 != uint32(typ) || target.Frame134 != srv.Frame() || target.Field38 != math.MaxUint32 ||
				live.StatusFlags != 0x20000000|object.MonStatusInjured || *source != beforeSource || *playerUpdate != beforePlayerUpdate || *player != beforePlayer {
				t.Fatalf("C-owned NPC carry: handled/result=%t/%t HP=%d old carry=%g live carry=%g calls=%d", h, result, health.Cur, math.Float32frombits(old.Field1), math.Float32frombits(live.Field1), calls)
			}
			t.Logf("C-owned >4GiB NPC electric body/runtime: target=%p old=%p live=%p HP=20->17 source/self-weapon=%p type=%d", target, old, live, source, typ)
		})
	}
}
