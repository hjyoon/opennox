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

// This is a C-owned API fixture, not a stock-map spell with negative balance
// data. CallDamage uses the registered C entry and real ObserveClear, electric
// scale/protection, DefaultDamage, UnitDamageClear and C UnitSetHP services.
func TestPlayerDamageSignedElectricNative4E17B0(t *testing.T) {
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
	if err := s.Types.ReadObjectType(&things.Thing{Name: "NativeSignedElectricPlayerDamage", OnDamage: &things.ProcFunc{Name: "PlayerDamage"}}); err != nil {
		t.Fatal(err)
	}
	if !s.Objs.Init(120) {
		t.Fatal("native allocator initialization")
	}
	t.Cleanup(s.Objs.FreeObjects)
	for _, playerSource := range []bool{false, true} {
		for _, selfWeapon := range []bool{false, true} {
			for _, typ := range []object.DamageType{object.DamageElectric, object.DamageAirborneElectric} {
				for _, observed := range []bool{false, true} {
					for _, tc := range []struct {
						raw      int32
						hp       uint16
						residual float32
					}{
						{-5, 64, -0.5}, // -5 + .5 -> -4; subtraction forwards low WORD without Max clamp.
						{0, 59, 0.5},   // 0 + .5 -> 0; DefaultDamage protection raises zero to 1.
						{1, 58, -0.5},  // Positive control: 1 + .5 -> even 2.
					} {
						t.Run(fmt.Sprintf("player-source-%t/self-weapon-%t/type-%d/observed-%t/raw-%d", playerSource, selfWeapon, typ, observed, tc.raw), func(t *testing.T) {
							newObject := func(damage unsafe.Pointer, ind uint16) *server.Object {
								obj := s.Objs.NewObject(&server.ObjectType{Damage: damage})
								obj.TypeInd, obj.ObjFlags = ind, 0
								return obj
							}
							target := newObject(s.Types.ByID("NativeSignedElectricPlayerDamage").Damage, 71)
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
							*player = server.Player{PlayerInd: 7, Field3680: 0x20}
							if observed {
								player.Field3680, player.CameraFollowObj = 0x22, source
							}
							*ud = server.PlayerUpdateData{Player: player, State: server.PlayerState13, Field21: math.Float32bits(0.5), Field76: 99, Field75: 77, Field40_0: 0x1122, Field40_1: 0xabcd}
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
							for _, pointer := range []unsafe.Pointer{unsafe.Pointer(target), unsafe.Pointer(source), unsafe.Pointer(ud), unsafe.Pointer(player), unsafe.Pointer(sourceUD), unsafe.Pointer(sourcePlayerUD), unsafe.Pointer(sourcePlayer), unsafe.Pointer(health)} {
								if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(pointer) <= math.MaxUint32 {
									t.Fatalf("pointer=%p, want above 4 GiB", pointer)
								}
							}
							beforeSource, beforeSourceUD, beforePlayerUD, beforePlayer := *source, *sourceUD, *sourcePlayerUD, *sourcePlayer
							if !playerSource {
								beforeSourceUD.Field130 = 1400
							}
							var wantUpdate unsafe.Pointer
							if observed {
								wantUpdate = legacy.Get_nox_xxx_updatePlayer_4F8100()
							}
							if !target.CallDamage(source, weapon, int(tc.raw), typ) || health.Cur != tc.hp || health.Max != 60 || health.Field2 != 60 || player.Field3680 != 0x20 || player.CameraFollowObj != nil || player.ObserveTarget() != nil || target.Update != wantUpdate || ud.Field76 != 2 || ud.Field75 != uint32(typ) || ud.Field21 != math.Float32bits(tc.residual) || ud.Field40_0 != 2 || ud.Field40_1 != 0xabcd || target.Obj130 != source || target.Pos132 != source.PrevPos || target.Field131 != uint32(typ) || target.Frame134 != 1400 || *source != beforeSource || *sourceUD != beforeSourceUD || *sourcePlayerUD != beforePlayerUD || *sourcePlayer != beforePlayer {
								t.Fatalf("native signed electric: raw=%d HP=%d status=%x camera=%p update=%p marker=%d/%d carry=%g attribution=%p", tc.raw, health.Cur, player.Field3680, player.CameraFollowObj, target.Update, ud.Field76, ud.Field75, math.Float32frombits(ud.Field21), target.Obj130)
							}
							t.Logf("C signed electric: target=%p source=%p raw=%d HP=60->%d carry=%g observed=%t", target, source, tc.raw, health.Cur, math.Float32frombits(ud.Field21), observed)
						})
					}
				}
			}
		}
	}
}
