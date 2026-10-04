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
func TestPlayerDamagePossessionMeleeNative4E17B0(t *testing.T) {
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
	if err := s.Types.ReadObjectType(&things.Thing{Name: "NativePossessionMeleePlayerDamage", OnDamage: &things.ProcFunc{Name: "PlayerDamage"}}); err != nil {
		t.Fatal(err)
	}
	if !s.Objs.Init(70) {
		t.Fatal("native allocator initialization")
	}
	t.Cleanup(s.Objs.FreeObjects)
	for _, playerSource := range []bool{false, true} {
		for _, kind := range []struct {
			name     string
			class    object.Class
			subclass object.SubClass
			typ      object.DamageType
		}{
			{"Sword", object.ClassWeapon, object.SubClass(object.WeaponSword), object.DamageBlade},
			{"MorningStar", object.ClassWeapon, object.SubClass(object.WeaponMace), object.DamageCrush},
			{"WarHammer", object.ClassWeapon, object.SubClass(object.WeaponHammer), object.DamageCrush},
			{"WoodenStaff", object.ClassWand, 0, object.DamageBlade},
			{"UnarmedClaw", 0, 0, object.DamageClaw},
			{"UnarmedCrush", 0, 0, object.DamageCrush},
			{"SimpleFist", object.ClassSimple, 0, object.DamageCrush},
		} {
			t.Run(fmt.Sprintf("player-source-%t/%s", playerSource, kind.name), func(t *testing.T) {
				newObject := func(damage unsafe.Pointer, ind uint16) *server.Object {
					obj := s.Objs.NewObject(&server.ObjectType{Damage: damage})
					obj.TypeInd, obj.ObjFlags = ind, 0
					return obj
				}
				target := newObject(s.Types.ByID("NativePossessionMeleePlayerDamage").Damage, 71)
				source, item := newObject(nil, 88), newObject(nil, 777)
				ud, freeUD := alloc.New(server.PlayerUpdateData{})
				player, freePlayer := alloc.New(server.Player{})
				sourceUD, freeSourceUD := alloc.New(server.MonsterUpdateData{})
				sourcePlayerUD, freeSourcePlayerUD := alloc.New(server.PlayerUpdateData{})
				sourcePlayer, freeSourcePlayer := alloc.New(server.Player{})
				health, freeHealth := alloc.New(server.HealthData{})
				init, freeInit := alloc.New(server.ModifierInitData{})
				for _, free := range []func(){freeUD, freePlayer, freeSourceUD, freeSourcePlayerUD, freeSourcePlayer, freeHealth, freeInit} {
					t.Cleanup(free)
				}
				*player = server.Player{PlayerInd: 7, Field3680: 0x22, CameraFollowObj: source}
				*ud = server.PlayerUpdateData{Player: player, State: server.PlayerState13, Field57: math.Float32bits(0.25), Field21: math.Float32bits(0.5), Field76: 88, Field75: 77}
				*sourcePlayer = server.Player{PlayerInd: 8}
				*sourcePlayerUD = server.PlayerUpdateData{Player: sourcePlayer, State: server.PlayerState13}
				*health = server.HealthData{Cur: 60, Max: 60, Field2: 60}
				target.ObjClass, target.UpdateData, target.HealthData, target.Material = object.ClassPlayer, unsafe.Pointer(ud), health, 0x4000
				source.ObjClass, source.ObjSubClass, source.UpdateData = object.ClassMonster, 0x11012, unsafe.Pointer(sourceUD)
				if playerSource {
					source.ObjClass, source.UpdateData = object.ClassPlayer, unsafe.Pointer(sourcePlayerUD)
				}
				source.PrevPos = types.Ptf(44, 7)
				item.ObjClass, item.ObjSubClass, item.InitData, item.PrevPos = kind.class, kind.subclass, unsafe.Pointer(init), types.Ptf(20, 0)
				weapon, attack := item, item
				if kind.class == 0 {
					weapon, attack = nil, source
				}
				if player.ObserveTarget() != source || target.Update != nil || legacy.Nox_xxx_playerObserveClear_4DDEF0 == nil {
					t.Fatal("native possession fixture/binding")
				}
				for _, pointer := range []unsafe.Pointer{unsafe.Pointer(target), unsafe.Pointer(source), unsafe.Pointer(item), unsafe.Pointer(ud), unsafe.Pointer(player), source.UpdateData, unsafe.Pointer(sourcePlayerUD), unsafe.Pointer(sourcePlayer), unsafe.Pointer(health), unsafe.Pointer(init)} {
					if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(pointer) <= math.MaxUint32 {
						t.Fatalf("pointer=%p, want above 4 GiB", pointer)
					}
				}
				beforeSource, beforeItem, beforePlayerUD, beforePlayer := *source, *item, *sourcePlayerUD, *sourcePlayer
				wantHP, carry := uint16(56), float32(0.25)
				if kind.typ == object.DamageCrush {
					wantHP, carry = 55, -0.125
				}
				wantPos := source.PrevPos
				if kind.class == object.ClassSimple {
					wantPos = item.PrevPos // 004E0EA7: SIMPLE records its own position.
				}
				if !target.CallDamage(source, weapon, 5, kind.typ) || health.Cur != wantHP || player.Field3680 != 0x20 || player.CameraFollowObj != nil || player.ObserveTarget() != nil || target.Update != legacy.Get_nox_xxx_updatePlayer_4F8100() || ud.Field76 != 1 || ud.Field75 != uint32(attack.TypeInd) || ud.Field21 != math.Float32bits(carry) || target.Obj130 != attack || target.Pos132 != wantPos || target.Field131 != uint32(kind.typ) || target.Frame134 != 1400 || *source != beforeSource || *item != beforeItem || *sourcePlayerUD != beforePlayerUD || *sourcePlayer != beforePlayer {
					t.Fatalf("native melee possession: HP=%d status=%x camera=%p update=%p marker=%d/%d carry=%g attribution=%p position=%v want=%v", health.Cur, player.Field3680, player.CameraFollowObj, target.Update, ud.Field76, ud.Field75, math.Float32frombits(ud.Field21), target.Obj130, target.Pos132, wantPos)
				}
				t.Logf("C->possessed melee: target=%p update=%p player=%p source=%p item=%p HP=60->%d status=0x22->0x20 camera cleared, normal update restored", target, ud, player, source, item, wantHP)
			})
		}
	}
}
