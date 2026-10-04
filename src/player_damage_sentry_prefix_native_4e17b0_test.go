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

// Synthetic C-owned records, not stock-map/GUI possession or ray input. The
// registered PlayerDamage dispatcher, ObserveClear/status/CameraUnlock,
// PlayerSetState and DamageClear/HP are real: no callback supplies the result.
func TestPlayerDamageSentryPrefixNative4E17B0(t *testing.T) {
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
	if err := s.Types.ReadObjectType(&things.Thing{Name: "NativeSentryPrefixPlayerDamage", OnDamage: &things.ProcFunc{Name: "PlayerDamage"}}); err != nil {
		t.Fatal(err)
	}
	if err := s.Types.ReadObjectType(&things.Thing{Name: "SentryGlobe"}); err != nil {
		t.Fatal(err)
	}
	if !s.Objs.Init(70) {
		t.Fatal("native allocator initialization")
	}
	t.Cleanup(s.Objs.FreeObjects)
	for _, observe := range []bool{false, true} {
		for _, raw := range []int{1, 19, 20, 500} {
			t.Run(fmt.Sprintf("observe-%t/raw-%d", observe, raw), func(t *testing.T) {
				target := s.Objs.NewObject(&server.ObjectType{Damage: s.Types.ByID("NativeSentryPrefixPlayerDamage").Damage})
				source := s.Objs.NewObject(&server.ObjectType{})
				weapon := s.Objs.NewObject(&server.ObjectType{})
				ud, freeUD := alloc.New(server.PlayerUpdateData{})
				player, freePlayer := alloc.New(server.Player{})
				sourceUD, freeSourceUD := alloc.New(server.PlayerUpdateData{})
				sourcePlayer, freeSourcePlayer := alloc.New(server.Player{})
				health, freeHealth := alloc.New(server.HealthData{})
				for _, free := range []func(){freeUD, freePlayer, freeSourceUD, freeSourcePlayer, freeHealth} {
					t.Cleanup(free)
				}
				*player = server.Player{PlayerInd: 7}
				if observe {
					player.Field3680, player.CameraFollowObj = 0x22, source
				}
				*sourcePlayer = server.Player{PlayerInd: 9}
				*sourceUD = server.PlayerUpdateData{Player: sourcePlayer, State: server.PlayerState13, Field76: 51, Field75: 53}
				*ud = server.PlayerUpdateData{Player: player, State: server.PlayerState13, Field57: math.Float32bits(0.25), Field21: math.Float32bits(0.125), Field76: 88, Field75: 77}
				*health = server.HealthData{Cur: 2000, Field2: 2000, Max: 2000}
				target.ObjClass, target.ObjFlags, target.UpdateData, target.HealthData, target.Material, target.TypeInd = object.ClassPlayer, 0, unsafe.Pointer(ud), health, 0x4000, 71
				source.ObjClass, source.ObjFlags, source.UpdateData, source.TypeInd = object.ClassPlayer, 0, unsafe.Pointer(sourceUD), 777
				weapon.ObjClass, weapon.ObjFlags, weapon.TypeInd, weapon.PrevPos, weapon.PosVec = object.ClassImmobile, 0, uint16(s.Types.ByID("SentryGlobe").Ind()), types.Ptf(44, 7), types.Ptf(-20, 0)
				for _, pointer := range []unsafe.Pointer{unsafe.Pointer(target), unsafe.Pointer(source), unsafe.Pointer(weapon), unsafe.Pointer(ud), unsafe.Pointer(player), unsafe.Pointer(sourceUD), unsafe.Pointer(sourcePlayer), unsafe.Pointer(health)} {
					if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(pointer) <= math.MaxUint32 {
						t.Fatalf("pointer=%p, want above 4 GiB", pointer)
					}
				}
				beforeSource, beforeSourceUD, beforeSourcePlayer, beforeWeapon := *source, *sourceUD, *sourcePlayer, *weapon
				if !target.CallDamage(source, weapon, raw, object.DamageZapRay) {
					t.Fatal("real native Sentry dispatcher rejected damage")
				}
				wantStatus, wantUpdate, wantState := uint32(0), unsafe.Pointer(nil), server.PlayerState13
				if observe {
					wantStatus, wantUpdate = 0x20, legacy.Get_nox_xxx_updatePlayer_4F8100()
				}
				if raw >= 20 {
					wantState = server.PlayerState30
				}
				wantHP := uint16(2000 - raw)
				if health.Cur != wantHP || health.Field2 != 2000 || health.Max != 2000 || ud.Field21 != math.Float32bits(0.125) || ud.Field76 != 1 || ud.Field75 != uint32(weapon.TypeInd) || ud.State != wantState || player.Field3680 != wantStatus || player.CameraFollowObj != nil || player.ObserveTarget() != nil || target.Update != wantUpdate || target.Obj130 != weapon || target.Pos132 != weapon.PrevPos || target.Field131 != uint32(object.DamageZapRay) || target.Frame134 != 1400 || *source != beforeSource || *sourceUD != beforeSourceUD || *sourcePlayer != beforeSourcePlayer || *weapon != beforeWeapon {
					t.Fatalf("native Sentry: HP=%d want=%d status=%x camera=%p update=%p state=%d want=%d marker=%d/%d carry=%g", health.Cur, wantHP, player.Field3680, player.CameraFollowObj, target.Update, ud.State, wantState, ud.Field76, ud.Field75, math.Float32frombits(ud.Field21))
				}
				t.Logf("real C->Sentry: target=%p update=%p player=%p source=%p weapon=%p raw=%d HP=2000->%d observe=%t carry=0.125 marker=1/%d status=%x state=%d", target, ud, player, source, weapon, raw, wantHP, observe, weapon.TypeInd, player.Field3680, ud.State)
			})
		}
	}
}
