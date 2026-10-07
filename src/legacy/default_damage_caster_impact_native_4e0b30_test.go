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

// These synthetic C-owned records run the real C dispatcher, the registered
// DefaultDamage Go export and UnitDamageClear/UnitSetHP. No replacement damage
// callback supplies HP or attribution. Actual stock Earthquake is exercised by
// the separate, unchanged public headless scenario.
func TestDefaultDamageCasterImpactNative4E0B30CRecords(t *testing.T) {
	srv := npcReflectLegacyServer4E17B0(t)
	srv.Audio.Init(srv)
	t.Cleanup(srv.Audio.Free)
	srv.FreeObjectTypes()
	t.Cleanup(srv.FreeObjectTypes)
	if err := srv.Types.ReadObjectType(&things.Thing{Name: "NativeCasterImpactDefault", OnDamage: &things.ProcFunc{Name: "DefaultDamage"}}); err != nil {
		t.Fatal(err)
	}
	if !srv.Objs.Init(400) {
		t.Fatal("caster IMPACT allocator")
	}
	t.Cleanup(srv.Objs.FreeObjects)
	for _, victim := range []string{"monster", "NPC", "immune-monster", "player", "world"} {
		for _, owner := range []string{"player-self", "player-owned-NPC", "imaginary-self", "imaginary-owned-NPC", "world-self", "nil"} {
			for _, raw := range []int32{3, 8, 19} {
				t.Run(fmt.Sprintf("%s/%s/raw-%d", victim, owner, raw), func(t *testing.T) {
					v := srv.Objs.NewObject(&server.ObjectType{Damage: srv.Types.ByID("NativeCasterImpactDefault").Damage})
					a, w := srv.Objs.NewObject(&server.ObjectType{}), srv.Objs.NewObject(&server.ObjectType{})
					ud, freeUD := alloc.New(server.MonsterUpdateData{})
					casterUD, freeCasterUD := alloc.New(server.MonsterUpdateData{})
					player, freePlayer := alloc.New(server.Player{})
					playerUD, freePlayerUD := alloc.New(server.PlayerUpdateData{})
					victimPlayer, freeVictimPlayer := alloc.New(server.Player{})
					victimPlayerUD, freeVictimPlayerUD := alloc.New(server.PlayerUpdateData{})
					health, freeHealth := alloc.New(server.HealthData{})
					for _, free := range []func(){freeUD, freeCasterUD, freePlayer, freePlayerUD, freeVictimPlayer, freeVictimPlayerUD, freeHealth} {
						t.Cleanup(free)
					}
					*ud = server.MonsterUpdateData{Field547: 99, Field546: 77, Field523_2: 88, Field1: math.Float32bits(0.125)}
					*casterUD = server.MonsterUpdateData{Field523_2: 66, Field1: math.Float32bits(0.375)}
					*player, *victimPlayer = server.Player{PlayerInd: 9}, server.Player{PlayerInd: 10}
					*playerUD = server.PlayerUpdateData{Player: player, State: server.PlayerState13}
					*victimPlayerUD = server.PlayerUpdateData{Player: victimPlayer, State: server.PlayerState13, Field21: math.Float32bits(0.25)}
					*health = server.HealthData{Cur: 2000, Field2: 2000, Max: 2000}
					v.TypeInd, v.ObjClass, v.ObjSubClass, v.ObjFlags, v.Material = 71, object.ClassMonster, 2, 0, 0x4000
					v.UpdateData, v.HealthData, v.Frame134, v.Pos132 = unsafe.Pointer(ud), health, 77, types.Ptf(31, 47)
					switch victim {
					case "NPC":
						v.ObjSubClass = 0x11012
					case "immune-monster":
						v.ObjSubClass = 0xe00 // No poison/fire/electric immunity applies to IMPACT.
					case "player":
						v.ObjClass, v.ObjSubClass, v.UpdateData = object.ClassPlayer, 0, unsafe.Pointer(victimPlayerUD)
					case "world":
						v.ObjClass, v.ObjSubClass = object.ClassObstacle, 0
					}
					a.TypeInd, a.PrevPos = 713, types.Ptf(101, 103)
					w.TypeInd, w.PrevPos = 1399, types.Ptf(17, 19)
					switch owner {
					case "player-self":
						a.ObjClass, a.UpdateData, w = object.ClassPlayer, unsafe.Pointer(playerUD), a
					case "player-owned-NPC":
						a.ObjClass, a.UpdateData, w.ObjOwner = object.ClassPlayer, unsafe.Pointer(playerUD), a
						w.ObjClass, w.ObjSubClass, w.UpdateData = object.ClassMonster, 0x11012, unsafe.Pointer(casterUD)
					case "imaginary-self":
						w = a // The actual script position-to-position caster has class 0.
					case "imaginary-owned-NPC":
						w.ObjOwner, w.ObjClass, w.ObjSubClass, w.UpdateData = a, object.ClassMonster, 0x11012, unsafe.Pointer(casterUD)
					case "world-self":
						w.ObjClass, a = object.ClassSimple, w
					case "nil":
						a = nil
					}
					for _, pointer := range []unsafe.Pointer{unsafe.Pointer(v), unsafe.Pointer(w), unsafe.Pointer(ud), unsafe.Pointer(casterUD), unsafe.Pointer(player), unsafe.Pointer(playerUD), unsafe.Pointer(victimPlayer), unsafe.Pointer(victimPlayerUD), unsafe.Pointer(health)} {
						if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(pointer) <= math.MaxUint32 {
							t.Fatalf("caster IMPACT pointer=%p, want >4 GiB", pointer)
						}
					}
					if a != nil && (!srv.IsEnemyTo(v, a) || w.FindOwnerChainPlayer() != a) {
						t.Fatal("native caster hostility/terminal owner fixture")
					}
					beforeCaster, beforePlayer, beforePlayerUD, beforeVictimPlayerUD := *w, *player, *playerUD, *victimPlayerUD
					beforeSource := server.Object{}
					if a != nil {
						beforeSource = *a
					}
					if !objectDamageDispatchCallNative(v, a, w, raw, object.DamageImpact) {
						t.Fatal("native caster IMPACT result")
					}
					position := w.PrevPos
					if a == nil {
						position = types.Pointf{}
					}
					if health.Cur != uint16(2000-raw) || health.Field2 != 2000 || health.Max != 2000 || v.Obj130 != w || v.Field131 != 11 || v.Frame134 != 1400 || v.Pos132 != position || v.Field38 != math.MaxUint32 {
						t.Fatalf("C->DefaultDamage->UnitSetHP: HP=%d want=%d attribution=%p want=%p frame=%d", health.Cur, 2000-raw, v.Obj130, w, v.Frame134)
					}
					if victim != "player" && victim != "world" {
						marker, markerType := uint32(1), uint32(w.TypeInd)
						if a == nil || a == w {
							marker, markerType = 2, 11
						}
						if ud.Field547 != marker || ud.Field546 != markerType || !ud.StatusFlags.Has(object.MonStatusInjured) || ud.StatusFlags.Has(object.MonStatusOnFire) || ud.Field523_2 != 88 || ud.Field1 != math.Float32bits(0.125) {
							t.Fatal("native caster IMPACT marker/status/carry")
						}
					}
					if *w != beforeCaster || (a != nil && *a != beforeSource) || *player != beforePlayer || *playerUD != beforePlayerUD || *victimPlayerUD != beforeVictimPlayerUD || casterUD.Field130 != 0 || casterUD.Field523_2 != 66 || casterUD.Field1 != math.Float32bits(0.375) {
						t.Fatal("native damage changed the caster/owner or armor carry")
					}
					// The caster's update and source class stay distinct: an owned
					// MONSTER weapon does not make its PLAYER/class-zero owner a
					// monster sound source or a source-side combat-latch record.
					if casterUD.StatusFlags != 0 || ud.StatusFlags.Has(object.MonStatusOnFire) {
						t.Fatal("native caster IMPACT fabricated elemental status")
					}
					t.Logf("C->DefaultDamage->UnitSetHP: target=%p update=%p source=%p caster=%p owner=%s victim=%s raw=%d HP=2000->%d", v, v.UpdateData, a, w, owner, victim, raw, health.Cur)
				})
			}
		}
	}
}
