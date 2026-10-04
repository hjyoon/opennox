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

// C-owned high-address records drive the real C dispatcher, registered
// PlayerDamage NPC adapter, DefaultDamage and UnitDamageClear/UnitSetHP.
// No replacement callback supplies HP or attribution. Not a stock-map test.
func TestPlayerDamageNPCZapRayNativeCallback4E17B0CRecords(t *testing.T) {
	srv := npcReflectLegacyServer4E17B0(t)
	srv.Audio.Init(srv)
	t.Cleanup(srv.Audio.Free)
	if !srv.Objs.Init(70) {
		t.Fatal("native NPC ray allocator initialization")
	}
	t.Cleanup(srv.Objs.FreeObjects)
	for _, owner := range []string{"world", "Player", "NPC"} {
		for _, raw := range []int32{1, 19, 500} {
			t.Run(fmt.Sprintf("%s/raw-%d", owner, raw), func(t *testing.T) {
				target, unit, ray := srv.Objs.NewObject(&server.ObjectType{Damage: playerDamageMeleeCallbackNative4E17B0()}), srv.Objs.NewObject(&server.ObjectType{}), srv.Objs.NewObject(&server.ObjectType{})
				ud, freeUD := alloc.New(server.MonsterUpdateData{})
				npcUD, freeNPCUD := alloc.New(server.MonsterUpdateData{})
				playerUD, freePlayerUD := alloc.New(server.PlayerUpdateData{})
				player, freePlayer := alloc.New(server.Player{})
				health, freeHealth := alloc.New(server.HealthData{})
				for _, free := range []func(){freeUD, freeNPCUD, freePlayerUD, freePlayer, freeHealth} {
					t.Cleanup(free)
				}
				*health = server.HealthData{Cur: 2000, Field2: 2000, Max: 2000}
				*ud = server.MonsterUpdateData{Field547: 99, Field546: 77, Field523_2: 88, Field1: math.Float32bits(0.125), Field518: math.Float32bits(0.75), WeaponEquipFlags: 0x400}
				*player = server.Player{PlayerInd: 9}
				*playerUD = server.PlayerUpdateData{Player: player, State: server.PlayerState13, Field21: math.Float32bits(0.25)}
				*npcUD = server.MonsterUpdateData{Field523_2: 66, Field1: math.Float32bits(0.25)}
				target.TypeInd, target.ObjClass, target.ObjSubClass, target.ObjFlags, target.Material = 71, object.ClassMonster, 0x11012, 0, 0x4000
				target.UpdateData, target.HealthData = unsafe.Pointer(ud), health
				ray.TypeInd, ray.ObjClass, ray.ObjFlags, ray.PrevPos, ray.PosVec = 801, object.ClassLight|object.ClassSimple|object.ClassImmobile|object.ClassVisibleEnable, object.FlagAirborne, types.Ptf(44, 7), types.Ptf(10, 20)
				source := ray
				switch owner {
				case "Player":
					unit.TypeInd, unit.ObjClass, unit.UpdateData = 777, object.ClassPlayer, unsafe.Pointer(playerUD)
					source, ray.ObjOwner = unit, unit
				case "NPC":
					unit.TypeInd, unit.ObjClass, unit.ObjSubClass, unit.UpdateData = 778, object.ClassMonster, 0x11012, unsafe.Pointer(npcUD)
					source, ray.ObjOwner = unit, unit
				}
				for _, pointer := range []unsafe.Pointer{unsafe.Pointer(target), unsafe.Pointer(unit), unsafe.Pointer(ray), unsafe.Pointer(ud), unsafe.Pointer(npcUD), unsafe.Pointer(playerUD), unsafe.Pointer(player), unsafe.Pointer(health)} {
					if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(pointer) <= math.MaxUint32 {
						t.Fatalf("native NPC ray pointer=%p, want >4 GiB", pointer)
					}
				}
				if ray.FindOwnerChainPlayer() != source {
					t.Fatal("native NPC ray terminal owner")
				}
				beforeRay, beforeUnit, beforePlayerUD, beforePlayer := *ray, *unit, *playerUD, *player
				if !objectDamageDispatchCallNative(target, source, ray, raw, object.DamageZapRay) {
					t.Fatal("real native NPC ray dispatcher rejected damage")
				}
				wantMarker, wantKind := uint32(1), uint32(ray.TypeInd)
				if owner == "world" {
					wantMarker, wantKind = 2, 16
				}
				wantHP := uint16(2000 - raw)
				if health.Cur != wantHP || health.Field2 != 2000 || health.Max != 2000 || target.Obj130 != ray || target.Pos132 != ray.PrevPos || target.Field131 != 16 || target.Frame134 != 1400 || target.Field38 != math.MaxUint32 || ud.Field547 != wantMarker || ud.Field546 != wantKind || !ud.StatusFlags.Has(object.MonStatusInjured) || ud.StatusFlags.Has(object.MonStatusOnFire) || ud.Field1 != math.Float32bits(0.125) || ud.Field518 != math.Float32bits(0.75) || ud.Field523_2 != 88 || *ray != beforeRay || *unit != beforeUnit || *playerUD != beforePlayerUD || *player != beforePlayer || npcUD.Field1 != math.Float32bits(0.25) || npcUD.Field523_2 != 66 {
					t.Fatalf("native NPC ZAP_RAY HP=%d want=%d marker=%d/%d state=%d carry=%g", health.Cur, wantHP, ud.Field547, ud.Field546, ud.Field523_2, math.Float32frombits(ud.Field1))
				}
				if owner == "NPC" && srv.IsEnemyTo(target, source) && npcUD.Field130 != 1400 {
					t.Fatal("native NPC ray source first-hit latch")
				}
				t.Logf("C->PlayerDamage NPC->DefaultDamage->UnitSetHP: target=%p update=%p source=%p ray=%p owner=%s raw=%d HP=2000->%d carry=0.125 state=88 marker=%d/%d", target, ud, source, ray, owner, raw, wantHP, wantMarker, wantKind)
			})
		}
	}
}
