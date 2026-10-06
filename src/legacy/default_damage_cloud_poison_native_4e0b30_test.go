package legacy

import (
	"fmt"
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/things"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

// Synthetic C-owned records exercise the real C dispatcher, registered
// DefaultDamage adapter and UnitDamageClear/UnitSetHP. They are not stock GUI
// cloud placement; no injected callback supplies HP or hit attribution.
func TestDefaultDamageCloudPoisonNative4E0B30CRecords(t *testing.T) {
	srv := npcReflectLegacyServer4E17B0(t)
	srv.Audio.Init(srv)
	t.Cleanup(srv.Audio.Free)
	srv.FreeObjectTypes()
	t.Cleanup(srv.FreeObjectTypes)
	if err := srv.Types.ReadObjectType(&things.Thing{Name: "NativeCloudPoisonDefault", OnDamage: &things.ProcFunc{Name: "DefaultDamage"}}); err != nil {
		t.Fatal(err)
	}
	if !srv.Objs.Init(300) {
		t.Fatal("cloud damage allocator")
	}
	t.Cleanup(srv.Objs.FreeObjects)
	for _, owner := range []string{"self", "imaginary", "Player", "NPC"} {
		for _, subclass := range []uint32{1, 2, 0x10, 0x10202, 0x400, 0x800, 0x200, 0xe00} {
			for _, raw := range []int32{3, 10, 19} {
				t.Run(fmt.Sprintf("%s/subclass-%x/raw-%d", owner, subclass, raw), func(t *testing.T) {
					v := srv.Objs.NewObject(&server.ObjectType{Damage: srv.Types.ByID("NativeCloudPoisonDefault").Damage})
					a, w := srv.Objs.NewObject(&server.ObjectType{}), srv.Objs.NewObject(&server.ObjectType{})
					ud, freeUD := alloc.New(server.MonsterUpdateData{})
					npcUD, freeNPCUD := alloc.New(server.MonsterUpdateData{})
					player, freePlayer := alloc.New(server.Player{})
					playerUD, freePlayerUD := alloc.New(server.PlayerUpdateData{})
					health, freeHealth := alloc.New(server.HealthData{})
					for _, free := range []func(){freeUD, freeNPCUD, freePlayer, freePlayerUD, freeHealth} {
						t.Cleanup(free)
					}
					*ud = server.MonsterUpdateData{Field547: 99, Field546: 77, Field523_2: 88, Field1: math.Float32bits(0.125)}
					*npcUD = server.MonsterUpdateData{Field523_2: 66, Field1: math.Float32bits(0.25)}
					*player = server.Player{PlayerInd: 9}
					*playerUD = server.PlayerUpdateData{Player: player, State: server.PlayerState13}
					*health = server.HealthData{Cur: 2000, Field2: 2000, Max: 2000}
					v.TypeInd, v.ObjClass, v.ObjSubClass, v.ObjFlags, v.Material = 71, object.ClassMonster, object.SubClass(subclass), 0, 0x4000
					v.UpdateData, v.HealthData, v.Frame134, v.Pos132 = unsafe.Pointer(ud), health, 77, types.Ptf(31, 47)
					v.Buffs = 1<<22 | 1<<26
					w.TypeInd, w.ObjClass, w.PrevPos = 1221, 0x190008, types.Ptf(44, 7)
					a.TypeInd = 1399
					switch owner {
					case "self":
						a = w
					case "imaginary":
						w.ObjOwner = a
					case "Player":
						a.ObjClass, a.UpdateData, w.ObjOwner = object.ClassPlayer, unsafe.Pointer(playerUD), a
					case "NPC":
						a.ObjClass, a.ObjSubClass, a.UpdateData, w.ObjOwner = object.ClassMonster, 0x11012, unsafe.Pointer(npcUD), a
					}
					for _, pointer := range []unsafe.Pointer{unsafe.Pointer(v), unsafe.Pointer(a), unsafe.Pointer(w), unsafe.Pointer(ud), unsafe.Pointer(npcUD), unsafe.Pointer(player), unsafe.Pointer(playerUD), unsafe.Pointer(health)} {
						if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(pointer) <= math.MaxUint32 {
							t.Fatalf("cloud damage pointer=%p, want >4 GiB", pointer)
						}
					}
					if w.FindOwnerChainPlayer() != a {
						t.Fatal("cloud terminal owner")
					}
					beforeSource, beforeCloud, beforePlayer, beforePlayerUD := *a, *w, *player, *playerUD
					beforeSync := v.Field38
					if !objectDamageDispatchCallNative(v, a, w, raw, object.DamagePoison) {
						t.Fatal("native poison return")
					}
					wantHP := uint16(2000 - raw)
					marker, markerType := uint32(1), uint32(1221)
					if owner == "self" {
						marker, markerType = 2, 5
					}
					if subclass&0x200 != 0 {
						wantHP = 2000
						if v.Obj130 != nil || v.Frame134 != 77 || v.Field131 != 0 || v.Pos132 != types.Ptf(31, 47) || v.Field38 != beforeSync || ud.Field547 != 0 || ud.Field546 != 77 || ud.StatusFlags != 0 {
							t.Fatal("native poison immunity prefix")
						}
					} else if v.Obj130 != w || v.Frame134 != 1400 || v.Field131 != 5 || v.Pos132 != w.PrevPos || v.Field38 != math.MaxUint32 || ud.Field547 != marker || ud.Field546 != markerType ||
						!ud.StatusFlags.Has(object.MonStatusInjured) || ud.StatusFlags.Has(object.MonStatusOnFire) {
						t.Fatalf("native poison tail attribution=%p frame=%d latch=%d/%d status=%x", v.Obj130, v.Frame134, ud.Field547, ud.Field546, ud.StatusFlags)
					}
					if health.Cur != wantHP || health.Field2 != 2000 || health.Max != 2000 || ud.Field1 != math.Float32bits(0.125) || ud.Field523_2 != 88 ||
						*a != beforeSource || *w != beforeCloud || *player != beforePlayer || *playerUD != beforePlayerUD || npcUD.Field1 != math.Float32bits(0.25) || npcUD.Field523_2 != 66 ||
						!v.HasEnchant(22) || !v.HasEnchant(26) {
						t.Fatalf("native cloud poison HP=%d want=%d", health.Cur, wantHP)
					}
					if owner == "NPC" && subclass&0x200 == 0 && srv.IsEnemyTo(v, a) && npcUD.Field130 != 1400 {
						t.Fatal("native source combat latch")
					}
					if (owner != "NPC" || subclass&0x200 != 0) && npcUD.Field130 != 0 {
						t.Fatal("unrelated source combat latch")
					}
					t.Logf("C->DefaultDamage->UnitSetHP: target=%p update=%p source=%p cloud=%p owner=%s subclass=%x raw=%d HP=2000->%d", v, ud, a, w, owner, subclass, raw, wantHP)
				})
			}
		}
	}
}
