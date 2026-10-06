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

// C-owned records exercise the production PlayerDamage adapter, its stock
// type-based block exclusions, DefaultDamage and real UnitSetHP. This is not
// stock GUI cloud placement. No replacement damage callback supplies HP.
func TestPlayerDamageNPCCloudPoisonNative4E17B0CRecords(t *testing.T) {
	srv := npcReflectLegacyServer4E17B0(t)
	srv.Audio.Init(srv)
	t.Cleanup(srv.Audio.Free)
	srv.FreeObjectTypes()
	t.Cleanup(srv.FreeObjectTypes)
	for _, name := range []string{"ToxicCloud", "SmallToxicCloud"} {
		if err := srv.Types.ReadObjectType(&things.Thing{Name: name}); err != nil {
			t.Fatal(err)
		}
	}
	if !srv.Objs.Init(600) {
		t.Fatal("NPC cloud allocator")
	}
	t.Cleanup(srv.Objs.FreeObjects)
	for _, name := range []string{"ToxicCloud", "SmallToxicCloud"} {
		for _, owner := range []string{"nil", "self", "imaginary", "Player", "NPC", "proxy-NPC"} {
			for _, immune := range []bool{false, true} {
				for _, raw := range []int32{0, 3, 10, 19} {
					t.Run(fmt.Sprintf("%s/%s/immune-%t/raw-%d", name, owner, immune, raw), func(t *testing.T) {
						v := srv.Objs.NewObject(&server.ObjectType{Damage: playerDamageMeleeCallbackNative4E17B0()})
						a, w, proxy, armor := srv.Objs.NewObject(&server.ObjectType{}), srv.Objs.NewObject(&server.ObjectType{}), srv.Objs.NewObject(&server.ObjectType{}), srv.Objs.NewObject(&server.ObjectType{})
						ud, freeUD := alloc.New(server.MonsterUpdateData{})
						npcUD, freeNPCUD := alloc.New(server.MonsterUpdateData{})
						player, freePlayer := alloc.New(server.Player{})
						playerUD, freePlayerUD := alloc.New(server.PlayerUpdateData{})
						health, freeHealth := alloc.New(server.HealthData{})
						armorHealth, freeArmorHealth := alloc.New(server.HealthData{})
						armorInit, freeArmorInit := alloc.New(server.ModifierInitData{})
						for _, free := range []func(){freeUD, freeNPCUD, freePlayer, freePlayerUD, freeHealth, freeArmorHealth, freeArmorInit} {
							t.Cleanup(free)
						}
						*ud = server.MonsterUpdateData{Field547: 99, Field546: 77, Field523_2: 88, Field1: math.Float32bits(0.125), Field518: math.Float32bits(0.9), ArmorEquipFlags: 0x3000000, WeaponEquipFlags: 0x400}
						ud.AIStack[0].Action = 21
						*npcUD = server.MonsterUpdateData{Field523_2: 66, Field1: math.Float32bits(0.25)}
						*player = server.Player{PlayerInd: 9}
						*playerUD = server.PlayerUpdateData{Player: player, State: server.PlayerState13}
						*health = server.HealthData{Cur: 2000, Field2: 2000, Max: 2000}
						*armorHealth = server.HealthData{Cur: 50, Field2: 50, Max: 50}
						v.TypeInd, v.ObjClass, v.ObjSubClass, v.Material = 71, object.ClassMonster, 0x11012, 0x4000
						if immune {
							v.ObjSubClass |= 0x200
						}
						v.UpdateData, v.HealthData, v.Frame134, v.Pos132 = unsafe.Pointer(ud), health, 77, types.Ptf(31, 47)
						v.Buffs = 1<<22 | 1<<26 | 1<<27
						armor.ObjClass, armor.ObjFlags, armor.HealthData = object.ClassArmor, object.FlagEquipped, armorHealth
						armor.InitData = unsafe.Pointer(armorInit)
						v.InvFirstItem = armor
						w.TypeInd, w.ObjClass, w.PrevPos = uint16(srv.Types.IndByID(name)), 0x190008, types.Ptf(44, 7)
						a.TypeInd = 1399
						source := a
						switch owner {
						case "nil":
							source = nil
						case "self":
							source = w
						case "imaginary":
							w.ObjOwner = a
						case "Player":
							a.ObjClass, a.UpdateData, w.ObjOwner = object.ClassPlayer, unsafe.Pointer(playerUD), a
						case "NPC", "proxy-NPC":
							a.ObjClass, a.ObjSubClass, a.UpdateData, w.ObjOwner = object.ClassMonster, 0x11012, unsafe.Pointer(npcUD), a
							if owner == "proxy-NPC" {
								proxy.ObjClass, proxy.ObjOwner, w.ObjOwner, source = object.ClassSimple, a, proxy, proxy
							}
						}
						for _, pointer := range []unsafe.Pointer{unsafe.Pointer(v), unsafe.Pointer(a), unsafe.Pointer(w), unsafe.Pointer(proxy), unsafe.Pointer(armor), unsafe.Pointer(ud), unsafe.Pointer(npcUD), unsafe.Pointer(player), unsafe.Pointer(playerUD), unsafe.Pointer(health), unsafe.Pointer(armorHealth)} {
							if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(pointer) <= math.MaxUint32 {
								t.Fatalf("NPC cloud pointer=%p, want >4 GiB", pointer)
							}
						}
						beforeSource, beforeCloud, beforeArmor, beforeArmorHealth, beforePlayer, beforePlayerUD := *a, *w, *armor, *armorHealth, *player, *playerUD
						if !objectDamageDispatchCallNative(v, source, w, raw, object.DamagePoison) {
							t.Fatal("native PlayerDamage NPC cloud rejected")
						}
						wantHP := uint16(2000 - raw)
						if immune {
							wantHP = 2000
						}
						if health.Cur != wantHP || health.Field2 != 2000 || health.Max != 2000 || ud.Field1 != math.Float32bits(0.125) || ud.Field518 != math.Float32bits(0.9) || ud.Field523_2 != 88 || *a != beforeSource || *w != beforeCloud || *armor != beforeArmor || *armorHealth != beforeArmorHealth || *player != beforePlayer || *playerUD != beforePlayerUD {
							t.Fatalf("native NPC cloud HP=%d want=%d armor=%d", health.Cur, wantHP, armorHealth.Cur)
						}
						if immune {
							if v.Obj130 != nil || v.Frame134 != 77 || v.Field131 != 0 || ud.Field547 != 0 || ud.StatusFlags != 0 {
								t.Fatal("cloud immunity tail changed")
							}
							wantType := uint32(w.TypeInd)
							if source == nil || source == w {
								wantType = 5
							}
							if ud.Field546 != wantType {
								t.Fatal("immune NPC lost executed PlayerDamage prefix")
							}
						} else {
							marker, markerType := uint32(1), uint32(w.TypeInd)
							wantPos := w.PrevPos
							if source == nil {
								// 004E0E93 -> 004E0F3D zeroes the hit position
								// for nil source, even with a non-nil weapon.
								wantPos = types.Pointf{}
							}
							if source == nil || source == w {
								marker, markerType = 2, 5
							}
							if v.Obj130 != w || v.Frame134 != 1400 || v.Field131 != 5 || v.Pos132 != wantPos || ud.Field547 != marker || ud.Field546 != markerType || !ud.StatusFlags.Has(object.MonStatusInjured) || ud.StatusFlags.Has(object.MonStatusOnFire) {
								t.Fatalf("native NPC cloud attribution=%p marker=%d/%d", v.Obj130, ud.Field547, ud.Field546)
							}
						}
						if (owner == "NPC" || owner == "proxy-NPC") && !immune && srv.IsEnemyTo(v, a) && npcUD.Field130 != 1400 {
							t.Fatal("NPC owner combat latch")
						}
						t.Logf("C->PlayerDamage NPC->DefaultDamage->UnitSetHP: target=%p update=%p source=%p cloud=%p name=%s owner=%s immune=%t raw=%d HP=2000->%d", v, ud, source, w, name, owner, immune, raw, wantHP)
					})
				}
			}
		}
	}
}
