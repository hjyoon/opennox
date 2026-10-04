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

// Synthetic C-owned records, not a stock-map or GUI possession assertion.
// The registered C dispatcher, ObserveClear/status C call, CameraUnlock,
// normal player update restoration and DefaultDamage/UnitSetHP are real.
// No damage, possession, camera, status or health callback is substituted.
func TestPlayerDamagePossessionMagicNative4E17B0(t *testing.T) {
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
	if err := s.Types.ReadObjectType(&things.Thing{Name: "NativePossessionMagicPlayerDamage", OnDamage: &things.ProcFunc{Name: "PlayerDamage"}}); err != nil {
		t.Fatal(err)
	}
	if !s.Objs.Init(40) {
		t.Fatal("native allocator initialization")
	}
	t.Cleanup(s.Objs.FreeObjects)
	for _, typ := range []object.DamageType{object.DamageFlame, object.DamageExplosion} {
		for _, playerSource := range []bool{false, true} {
			for _, splash := range []bool{false, true} {
				t.Run(fmt.Sprintf("type-%d/player-source-%t/splash-%t", typ, playerSource, splash), func(t *testing.T) {
					newObject := func(damage unsafe.Pointer, ind uint16) *server.Object {
						obj := s.Objs.NewObject(&server.ObjectType{Damage: damage})
						obj.TypeInd, obj.ObjFlags = ind, 0
						return obj
					}
					target := newObject(s.Types.ByID("NativePossessionMagicPlayerDamage").Damage, 71)
					source, missile := newObject(nil, 72), newObject(nil, 695)
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
					*ud = server.PlayerUpdateData{Player: player, State: server.PlayerState13, Field57: math.Float32bits(0.25), Field21: math.Float32bits(0.5), Field76: 88, Field75: 77}
					*sourceUD = server.MonsterUpdateData{}
					*sourcePlayer = server.Player{PlayerInd: 8}
					*sourcePlayerUD = server.PlayerUpdateData{Player: sourcePlayer, State: server.PlayerState13}
					*health = server.HealthData{Cur: 60, Max: 60, Field2: 60}
					target.ObjClass, target.UpdateData, target.HealthData, target.Material = object.ClassPlayer, unsafe.Pointer(ud), health, 0x4000
					source.ObjClass, source.UpdateData = object.ClassMonster, unsafe.Pointer(sourceUD)
					if playerSource {
						source.ObjClass, source.UpdateData = object.ClassPlayer, unsafe.Pointer(sourcePlayerUD)
					}
					source.PrevPos = types.Ptf(44, 7)
					missile.ObjClass, missile.ObjOwner = object.ClassMissile, source
					missile.PrevPos, missile.PosVec = types.Ptf(20, 0), types.Ptf(35, 0)
					if player.ObserveTarget() != source || target.Update != nil || legacy.Nox_xxx_playerObserveClear_4DDEF0 == nil {
						t.Fatal("native possession fixture/binding")
					}
					for _, pointer := range []unsafe.Pointer{unsafe.Pointer(target), unsafe.Pointer(source), unsafe.Pointer(missile), unsafe.Pointer(ud), unsafe.Pointer(player), source.UpdateData, unsafe.Pointer(sourcePlayerUD), unsafe.Pointer(sourcePlayer), unsafe.Pointer(health)} {
						if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(pointer) <= math.MaxUint32 {
							t.Fatalf("pointer=%p, want above 4 GiB", pointer)
						}
					}
					beforeSource, beforePlayerUD, beforePlayer, beforeMissile := *source, *sourcePlayerUD, *sourcePlayer, *missile
					attack, weapon := source, missile
					marker, markerType, wantHP, carry := uint32(1), uint32(missile.TypeInd), uint16(55), float32(0.5)
					if typ == object.DamageExplosion {
						wantHP, carry = 56, 0.25
					}
					if splash {
						attack, weapon, marker, markerType = missile, nil, 2, uint32(typ)
					}
					if !target.CallDamage(attack, weapon, 5, typ) || health.Cur != wantHP || player.Field3680 != 0x20 || player.CameraFollowObj != nil || player.ObserveTarget() != nil || target.Update != legacy.Get_nox_xxx_updatePlayer_4F8100() || ud.Field76 != marker || ud.Field75 != markerType || ud.Field21 != math.Float32bits(carry) || target.Obj130 != missile || target.Pos132 != missile.PrevPos || target.Field131 != uint32(typ) || target.Frame134 != 1400 || *source != beforeSource || *sourcePlayerUD != beforePlayerUD || *sourcePlayer != beforePlayer || *missile != beforeMissile {
						t.Fatalf("native magic possession: HP=%d status=%x camera=%p update=%p marker=%d/%d carry=%g", health.Cur, player.Field3680, player.CameraFollowObj, target.Update, ud.Field76, ud.Field75, math.Float32frombits(ud.Field21))
					}
					t.Logf("C->possessed player magic: target=%p update=%p player=%p source=%p missile=%p HP=60->%d status=0x22->0x20 camera cleared, normal player update restored", target, ud, player, source, missile, wantHP)
				})
			}
		}
	}
}
