package opennox

import (
	"fmt"
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/strman"
	"github.com/opennox/libs/things"
	"github.com/opennox/libs/types"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

// Synthetic native C records exercise C->PlayerDamage->DefaultDamage, real
// ObserveClear, PlayerSetState, DamageClear and UnitSetHP. There are no test HP,
// state or marker callbacks. This is not GUI ray-placement evidence.
func TestPlayerDamagePlayerZapRayNative4E17B0HPAndHurt(t *testing.T) {
	base := server.New(nil, nil, strman.New())
	t.Cleanup(base.Close)
	s := &Server{Server: base}
	oldServer, oldGetServer := noxServer, legacy.GetServer
	noxServer, legacy.GetServer = s, func() legacy.Server { return s }
	t.Cleanup(func() { noxServer, legacy.GetServer = oldServer, oldGetServer })
	oldGame, oldEngine, oldGameplay := noxflags.GetGame(), noxflags.GetEngine(), noxflags.GetGamePlay()
	noxflags.UnsetGame(oldGame)
	noxflags.UnsetEngine(oldEngine)
	noxflags.UnsetGamePlay(oldGameplay)
	noxflags.SetGamePlay(noxflags.GameplayFlag1)
	t.Cleanup(func() {
		noxflags.UnsetGame(noxflags.GetGame())
		noxflags.UnsetEngine(noxflags.GetEngine())
		noxflags.UnsetGamePlay(noxflags.GetGamePlay())
		noxflags.SetGame(oldGame)
		noxflags.SetEngine(oldEngine)
		noxflags.SetGamePlay(oldGameplay)
	})
	s.SetFrame(1400)
	s.Audio.Init(s.Server)
	t.Cleanup(s.Audio.Free)
	s.FreeObjectTypes()
	t.Cleanup(s.FreeObjectTypes)
	if err := s.Types.ReadObjectType(&things.Thing{Name: "NativePlayerZapRayEntry", OnDamage: &things.ProcFunc{Name: "PlayerDamage"}}); err != nil {
		t.Fatal(err)
	}
	if !s.Objs.Init(360) {
		t.Fatal("native ray allocator initialization")
	}
	t.Cleanup(s.Objs.FreeObjects)
	for _, owner := range []string{"world", "Player", "NPC"} {
		for _, observe := range []bool{false, true} {
			for _, state := range []server.PlayerState{server.PlayerState13, server.PlayerState1, server.PlayerState15} {
				for _, raw := range []int{-7, 0, 1, 19, 20, 500} {
					t.Run(fmt.Sprintf("%s/observe-%t/state-%d/raw-%d", owner, observe, state, raw), func(t *testing.T) {
						target := s.Objs.NewObject(&server.ObjectType{Damage: s.Types.ByID("NativePlayerZapRayEntry").Damage})
						unit, ray := s.Objs.NewObject(&server.ObjectType{}), s.Objs.NewObject(&server.ObjectType{})
						ud, freeUD := alloc.New(server.PlayerUpdateData{})
						player, freePlayer := alloc.New(server.Player{})
						sourceUD, freeSourceUD := alloc.New(server.PlayerUpdateData{})
						sourcePlayer, freeSourcePlayer := alloc.New(server.Player{})
						npcUD, freeNPCUD := alloc.New(server.MonsterUpdateData{})
						health, freeHealth := alloc.New(server.HealthData{})
						for _, free := range []func(){freeUD, freePlayer, freeSourceUD, freeSourcePlayer, freeNPCUD, freeHealth} {
							t.Cleanup(free)
						}
						*player = server.Player{PlayerInd: 7, ArmorEquip: 0x405, WeaponEquip: 0x400}
						*ud = server.PlayerUpdateData{Player: player, State: state, Field75: 77, Field76: 88, Field40_0: 0x1122, Field40_1: 0xabcd, Field57: math.Float32bits(0.75), Field21: math.Float32bits(0.125)}
						*sourcePlayer = server.Player{PlayerInd: 9}
						*sourceUD = server.PlayerUpdateData{Player: sourcePlayer, State: server.PlayerState13}
						*npcUD = server.MonsterUpdateData{Field523_2: 66, Field1: math.Float32bits(0.25)}
						*health = server.HealthData{Cur: 2000, Field2: 2000, Max: 2000}
						target.TypeInd, target.ObjClass, target.ObjFlags, target.Material, target.UpdateData, target.HealthData = 71, object.ClassPlayer, 0, 0x4000, unsafe.Pointer(ud), health
						ray.TypeInd, ray.ObjClass, ray.ObjFlags, ray.PrevPos, ray.PosVec = 801, object.ClassLight|object.ClassSimple|object.ClassImmobile|object.ClassVisibleEnable, object.FlagAirborne, types.Ptf(44, 7), types.Ptf(-20, 0)
						source := ray
						switch owner {
						case "Player":
							unit.TypeInd, unit.ObjClass, unit.UpdateData = 777, object.ClassPlayer, unsafe.Pointer(sourceUD)
							source, ray.ObjOwner = unit, unit
						case "NPC":
							unit.TypeInd, unit.ObjClass, unit.ObjSubClass, unit.UpdateData = 778, object.ClassMonster, 0x11012, unsafe.Pointer(npcUD)
							source, ray.ObjOwner = unit, unit
						}
						if observe {
							player.Field3680, player.CameraFollowObj = 0x22, source
						}
						for _, pointer := range []unsafe.Pointer{unsafe.Pointer(target), unsafe.Pointer(unit), unsafe.Pointer(ray), unsafe.Pointer(ud), unsafe.Pointer(player), unsafe.Pointer(sourceUD), unsafe.Pointer(sourcePlayer), unsafe.Pointer(npcUD), unsafe.Pointer(health)} {
							if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(pointer) <= math.MaxUint32 {
								t.Fatalf("native ray pointer=%p, want >4 GiB", pointer)
							}
						}
						if ray.FindOwnerChainPlayer() != source {
							t.Fatal("native ray terminal owner")
						}
						beforeUnit, beforeRay, beforeSourceUD, beforeSourcePlayer := *unit, *ray, *sourceUD, *sourcePlayer
						if !target.CallDamage(source, ray, raw, object.DamageZapRay) {
							t.Fatal("real C PlayerDamage ray dispatch returned false")
						}
						wantState, wantHP := state, uint16(2000-raw)
						if state == server.PlayerState13 && raw >= 20 {
							wantState = server.PlayerState30
						}
						marker, markerType := uint32(1), uint32(801)
						if source == ray {
							marker, markerType = 2, 16
						}
						if health.Cur != wantHP || health.Max != 2000 || health.Field2 != 2000 || target.Field38 != math.MaxUint32 || target.Obj130 != ray || target.Pos132 != ray.PrevPos || target.Field131 != 16 || target.Frame134 != 1400 || ud.State != wantState || ud.Field75 != markerType || ud.Field76 != marker || ud.Field40_0 != 0x1122 || ud.Field40_1 != 0xabcd || ud.Field57 != math.Float32bits(0.75) || ud.Field21 != math.Float32bits(0.125) || player.ArmorEquip != 0x405 || player.WeaponEquip != 0x400 || *unit != beforeUnit || *ray != beforeRay || *sourceUD != beforeSourceUD || *sourcePlayer != beforeSourcePlayer || npcUD.Field523_2 != 66 || npcUD.Field1 != math.Float32bits(0.25) {
							t.Fatalf("native Player ray HP=%d want=%d state=%d want=%d marker=%d/%d want=%d/%d", health.Cur, wantHP, ud.State, wantState, ud.Field76, ud.Field75, marker, markerType)
						}
						if observe && (player.Field3680 != 0x20 || player.ObserveTarget() != nil) {
							t.Fatal("real ObserveClear did not return possession")
						}
						if wantState == server.PlayerState30 && (ud.State2 != state || ud.Field41 != 1400 || target.Field34 != 1400) {
							t.Fatal("real PlayerSetState animation/frame stores")
						}
						if owner == "NPC" && s.IsEnemyTo(target, source) && npcUD.Field130 != 1400 {
							t.Fatal("native NPC source first-hit latch")
						}
						t.Logf("C->PlayerDamage->DefaultDamage->PlayerSetState/UnitSetHP: target=%p update=%p source=%p ray=%p owner=%s observe=%t raw=%d HP=2000->%d state=%d->%d markers=%d/%d armor=0.75 carry=0.125", target, ud, source, ray, owner, observe, raw, wantHP, state, wantState, marker, markerType)
					})
				}
			}
		}
	}
}
