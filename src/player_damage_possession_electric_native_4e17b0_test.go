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

// Synthetic C-owned records, not stock-map/GUI possession. The registered C
// dispatcher, root ObserveClear/C status call, CameraUnlock, normal player
// update restoration and DefaultDamage/UnitSetHP are not replaced by mocks.
func TestPlayerDamagePossessionElectricNative4E17B0(t *testing.T) {
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
	if err := s.Types.ReadObjectType(&things.Thing{Name: "NativePossessionElectricPlayerDamage", OnDamage: &things.ProcFunc{Name: "PlayerDamage"}}); err != nil {
		t.Fatal(err)
	}
	if !s.Objs.Init(70) {
		t.Fatal("native allocator initialization")
	}
	t.Cleanup(s.Objs.FreeObjects)
	for _, playerSource := range []bool{false, true} {
		for _, selfWeapon := range []bool{false, true} {
			for _, typ := range []object.DamageType{object.DamageElectric, object.DamageAirborneElectric} {
				t.Run(fmt.Sprintf("player-source-%t/self-weapon-%t/type-%d", playerSource, selfWeapon, typ), func(t *testing.T) {
					newObject := func(damage unsafe.Pointer, ind uint16) *server.Object {
						obj := s.Objs.NewObject(&server.ObjectType{Damage: damage})
						obj.TypeInd, obj.ObjFlags = ind, 0
						return obj
					}
					target := newObject(s.Types.ByID("NativePossessionElectricPlayerDamage").Damage, 71)
					source := newObject(nil, 88)
					ud, freeUD := alloc.New(server.PlayerUpdateData{})
					player, freePlayer := alloc.New(server.Player{})
					sourceUD, freeSourceUD := alloc.New(server.MonsterUpdateData{})
					sourcePlayerUD, freeSourcePlayerUD := alloc.New(server.PlayerUpdateData{})
					sourcePlayer, freeSourcePlayer := alloc.New(server.Player{})
					health, freeHealth := alloc.New(server.HealthData{})
					for _, free := range []func(){freeUD, freePlayer, freeSourceUD, freeSourcePlayerUD, freeSourcePlayer, freeHealth} {
						t.Cleanup(free)
					}
					*player = server.Player{PlayerInd: 7, Field3680: 0x22, CameraFollowObj: source}
					*ud = server.PlayerUpdateData{Player: player, State: server.PlayerState13, Field21: math.Float32bits(0.5), Field76: 99, Field75: 77}
					*sourcePlayer = server.Player{PlayerInd: 8}
					*sourcePlayerUD = server.PlayerUpdateData{Player: sourcePlayer, State: server.PlayerState13}
					*health = server.HealthData{Cur: 60, Max: 60, Field2: 60}
					target.ObjClass, target.UpdateData, target.HealthData, target.Material = object.ClassPlayer, unsafe.Pointer(ud), health, 0x4000
					source.ObjClass, source.ObjSubClass, source.UpdateData = object.ClassMonster, 0x11012, unsafe.Pointer(sourceUD)
					if playerSource {
						source.ObjClass, source.UpdateData = object.ClassPlayer, unsafe.Pointer(sourcePlayerUD)
					}
					source.PrevPos = types.Ptf(44, 7)
					var weapon *server.Object
					if selfWeapon {
						weapon = source
					}
					if player.ObserveTarget() != source || target.Update != nil || legacy.Nox_xxx_playerObserveClear_4DDEF0 == nil {
						t.Fatal("native possession fixture/binding")
					}
					for _, pointer := range []unsafe.Pointer{unsafe.Pointer(target), unsafe.Pointer(source), unsafe.Pointer(ud), unsafe.Pointer(player), unsafe.Pointer(sourceUD), unsafe.Pointer(sourcePlayerUD), unsafe.Pointer(sourcePlayer), unsafe.Pointer(health)} {
						if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(pointer) <= math.MaxUint32 {
							t.Fatalf("pointer=%p, want above 4 GiB", pointer)
						}
					}
					beforeSource, beforeSourceUD, beforePlayerUD, beforePlayer := *source, *sourceUD, *sourcePlayerUD, *sourcePlayer
					if !playerSource {
						// The real DefaultDamage tail latches the hostile NPC's
						// first successful-hit frame at 00532880.
						beforeSourceUD.Field130 = 1400
					}
					// 1*5 + live carry .5 rounds to even 6; .0 protection leaves
					// that HP damage intact. Self-weapon electric has no marker 1.
					if !target.CallDamage(source, weapon, 5, typ) || health.Cur != 54 || player.Field3680 != 0x20 || player.CameraFollowObj != nil || player.ObserveTarget() != nil || target.Update != legacy.Get_nox_xxx_updatePlayer_4F8100() || ud.Field76 != 2 || ud.Field75 != uint32(typ) || ud.Field21 != math.Float32bits(-0.5) || target.Obj130 != source || target.Pos132 != source.PrevPos || target.Field131 != uint32(typ) || target.Frame134 != 1400 || *source != beforeSource || *sourceUD != beforeSourceUD || *sourcePlayerUD != beforePlayerUD || *sourcePlayer != beforePlayer {
						t.Fatalf("native electric possession: HP=%d status=%x camera=%p update=%p marker=%d/%d carry=%g attribution=%p position=%v", health.Cur, player.Field3680, player.CameraFollowObj, target.Update, ud.Field76, ud.Field75, math.Float32frombits(ud.Field21), target.Obj130, target.Pos132)
					}
					t.Logf("C->possessed electric: target=%p update=%p player=%p source=%p HP=60->54 status=0x22->0x20 camera cleared, normal update restored", target, ud, player, source)
				})
			}
		}
	}
}
