package opennox

import (
	"bytes"
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

// Synthetic native-width records, not a stock-map assertion. Registered
// PlayerDamage -> real C armor lookup -> EquipDamage -> registered ArmorDamage
// -> DefaultDamage -> UnitSetHP run without substituted damage/HP callbacks.
// Only outgoing player item-health transport is captured.
func TestPlayerDamageNative4E17B0RegisteredOrderedArmorDamage(t *testing.T) {
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
	noxflags.SetGame(noxflags.GameModeCoop)
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
	for _, callback := range []string{"PlayerDamage", "ArmorDamage"} {
		if err := s.Types.ReadObjectType(&things.Thing{Name: "NativeOrderedArmor" + callback, OnDamage: &things.ProcFunc{Name: callback}}); err != nil {
			t.Fatal(err)
		}
	}
	if !s.Objs.Init(100) {
		t.Fatal("native object allocator")
	}
	t.Cleanup(s.Objs.FreeObjects)
	for _, playerTarget := range []bool{false, true} {
		for _, tc := range []struct {
			name                                string
			damage                              int
			typ                                 object.DamageType
			armor, hpCarry, itemCarry, armorDef float32
			wantHP, wantArmor                   uint16
			wantHPCarry, wantItemCarry          float32
			material                            object.Material
			god                                 bool
		}{
			{name: "cloth CRUSH", damage: 9, typ: object.DamageCrush, armor: 0.5, hpCarry: 0.4, itemCarry: 0.4, armorDef: 0.5, wantHP: 193, wantArmor: 23, wantHPCarry: 0.15, wantItemCarry: 0.4},
			{name: "metal CRUSH doubles wear", damage: 9, typ: object.DamageCrush, armor: 0.5, hpCarry: 0.4, itemCarry: 0.4, armorDef: 0.5, material: object.MaterialMetal, wantHP: 193, wantArmor: 21, wantHPCarry: 0.15, wantItemCarry: 0.4},
			{name: "metal CRUSH before GodMode", damage: 9, typ: object.DamageCrush, armor: 0.5, hpCarry: 0.4, itemCarry: 0.4, armorDef: 0.5, material: object.MaterialMetal, god: true, wantHP: 193, wantArmor: 21, wantHPCarry: 0.15, wantItemCarry: 0.4},
			{name: "zero flushes item carry", typ: object.DamageBlade, armor: 0.5, itemCarry: 0.75, armorDef: 0.5, wantHP: 200, wantArmor: 24, wantItemCarry: -0.25},
			{name: "negative updates item carry", damage: -3, typ: object.DamageBlade, armor: 0.5, itemCarry: 0.1, armorDef: 0.25, wantHP: 202, wantArmor: 25, wantHPCarry: 0.5, wantItemCarry: -0.4},
			{name: "zero armor preserves NaN", typ: object.DamageBlade, itemCarry: 0.75, armorDef: 0.25, wantHP: 200, wantArmor: 25, wantItemCarry: float32(math.NaN())},
			{name: "minimum HP follows armor", damage: 1, typ: object.DamageBlade, armor: 1, itemCarry: 0.75, armorDef: 0.25, wantHP: 199, wantArmor: 24},
		} {
			t.Run(fmt.Sprintf("player-target-%t/%s", playerTarget, tc.name), func(t *testing.T) {
				noxflags.UnsetEngine(noxflags.EngineGodMode)
				if tc.god {
					noxflags.SetEngine(noxflags.EngineGodMode)
				}
				newObject := func(callback string, ind uint16) *server.Object {
					def := &server.ObjectType{}
					if callback != "" {
						def.Damage = s.Types.ByID("NativeOrderedArmor" + callback).Damage
					}
					obj := s.Objs.NewObject(def)
					obj.TypeInd, obj.ObjFlags = ind, 0
					return obj
				}
				target, source, weapon, item := newObject("PlayerDamage", 71), newObject("", 72), newObject("", 777), newObject("ArmorDamage", 301)
				player, freePlayer := alloc.New(server.Player{})
				playerUD, freePlayerUD := alloc.New(server.PlayerUpdateData{})
				monsterUD, freeMonsterUD := alloc.New(server.MonsterUpdateData{})
				sourcePlayer, freeSourcePlayer := alloc.New(server.Player{})
				sourcePlayerUD, freeSourcePlayerUD := alloc.New(server.PlayerUpdateData{})
				sourceMonsterUD, freeSourceMonsterUD := alloc.New(server.MonsterUpdateData{})
				initData, freeInit := alloc.New(server.ModifierInitData{})
				weaponInit, freeWeaponInit := alloc.New(server.ModifierInitData{})
				itemUD, freeItemUD := alloc.New(server.WeaponArmorUpdateData{})
				hp, freeHP := alloc.New(server.HealthData{})
				itemHP, freeItemHP := alloc.New(server.HealthData{})
				armorDef, freeDef := alloc.New(server.Modifier{})
				for _, free := range []func(){freePlayer, freePlayerUD, freeMonsterUD, freeSourcePlayer, freeSourcePlayerUD, freeSourceMonsterUD, freeInit, freeWeaponInit, freeItemUD, freeHP, freeItemHP, freeDef} {
					t.Cleanup(free)
				}
				*player = server.Player{PlayerInd: 7, ArmorEquip: 0x405, WeaponEquip: 0x100}
				*playerUD = server.PlayerUpdateData{Player: player, State: server.PlayerState13, Field57: math.Float32bits(tc.armor), Field21: math.Float32bits(tc.hpCarry), Field76: 99, Field75: 77}
				*monsterUD = server.MonsterUpdateData{Field518: math.Float32bits(tc.armor), Field1: math.Float32bits(tc.hpCarry), Field547: 99, Field546: 77}
				*sourcePlayer = server.Player{PlayerInd: 8}
				*sourcePlayerUD = server.PlayerUpdateData{Player: sourcePlayer, State: server.PlayerState13}
				*sourceMonsterUD, *initData = server.MonsterUpdateData{}, server.ModifierInitData{}
				*itemUD = server.WeaponArmorUpdateData{Field0: math.Float32bits(tc.itemCarry)}
				*hp, *itemHP = server.HealthData{Cur: 200, Max: 200, Field2: 200}, server.HealthData{Cur: 25, Max: 25}
				*armorDef = server.Modifier{TypeInd: uint32(item.TypeInd), DamageCoeffOrArmor64: tc.armorDef}
				oldDef := s.Modif.Dword_5d4594_251608
				s.Modif.Dword_5d4594_251608 = armorDef
				t.Cleanup(func() { s.Modif.Dword_5d4594_251608 = oldDef })
				target.ObjClass, target.ObjSubClass, target.UpdateData = object.ClassMonster, 0x11012, unsafe.Pointer(monsterUD)
				source.ObjClass, source.UpdateData = object.ClassPlayer, unsafe.Pointer(sourcePlayerUD)
				if playerTarget {
					target.ObjClass, target.UpdateData = object.ClassPlayer, unsafe.Pointer(playerUD)
					source.ObjClass, source.ObjSubClass, source.UpdateData = object.ClassMonster, 0x11012, unsafe.Pointer(sourceMonsterUD)
				}
				target.HealthData, target.Material, source.PrevPos = hp, 0x4000, types.Ptf(44, 7)
				weapon.ObjClass, weapon.ObjSubClass, weapon.ObjOwner, weapon.PrevPos = object.ClassWeapon, object.SubClass(object.WeaponMace), source, types.Ptf(-20, 0)
				weapon.InitData = unsafe.Pointer(weaponInit)
				if tc.typ == object.DamageBlade {
					weapon.ObjSubClass = object.SubClass(object.WeaponSword)
				}
				item.ObjClass, item.ObjSubClass, item.ObjFlags = object.ClassArmor, 2, object.FlagEquipped
				item.UpdateData, item.InitData, item.HealthData, item.InvHolder = unsafe.Pointer(itemUD), unsafe.Pointer(initData), itemHP, target
				item.Material, target.InvFirstItem = uint16(tc.material), item
				var packets [][]byte
				oldSend := s.NetSendPacketXxx
				s.NetSendPacketXxx = func(recipient int, packet []byte, related *server.Object, remove, sequence int) int {
					if !playerTarget || recipient != 7 || related != nil || remove != 0 || sequence != 1 {
						t.Fatal("player item-health transport contract")
					}
					packets = append(packets, append([]byte(nil), packet...))
					return 1
				}
				t.Cleanup(func() { s.NetSendPacketXxx = oldSend })
				for _, pointer := range []unsafe.Pointer{unsafe.Pointer(target), unsafe.Pointer(source), unsafe.Pointer(weapon), unsafe.Pointer(item), target.UpdateData, source.UpdateData, weapon.InitData, item.UpdateData, item.InitData, unsafe.Pointer(player), unsafe.Pointer(sourcePlayer), unsafe.Pointer(hp), unsafe.Pointer(itemHP), unsafe.Pointer(armorDef)} {
					if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(pointer) <= math.MaxUint32 {
						t.Fatalf("pointer=%p, want above 4 GiB", pointer)
					}
				}
				if !s.IsEnemyTo(target, source) {
					t.Fatal("native melee fixture is not hostile")
				}
				before := *target
				if !target.CallDamage(source, weapon, tc.damage, tc.typ) {
					t.Fatal("registered PlayerDamage rejected")
				}
				wantHP := tc.wantHP
				if tc.god && playerTarget {
					wantHP = 200
				}
				marker, markerType, hpCarry := monsterUD.Field547, monsterUD.Field546, math.Float32frombits(monsterUD.Field1)
				if playerTarget {
					marker, markerType, hpCarry = playerUD.Field76, playerUD.Field75, math.Float32frombits(playerUD.Field21)
				}
				carry := math.Float32frombits(itemUD.Field0)
				carryOK := math.Abs(float64(carry-tc.wantItemCarry)) < 1e-6 || (math.IsNaN(float64(carry)) && math.IsNaN(float64(tc.wantItemCarry)))
				if hp.Cur != wantHP || itemHP.Cur != tc.wantArmor || !carryOK || math.Abs(float64(hpCarry-tc.wantHPCarry)) > 1e-6 || marker != 1 || markerType != 777 || target.InvFirstItem != item || item.InvHolder != target || s.Objs.DeletedList != nil {
					t.Fatalf("HP=%d armor=%d carry=%g/%g marker=%d/%d, want HP=%d armor=%d", hp.Cur, itemHP.Cur, hpCarry, carry, marker, markerType, wantHP, tc.wantArmor)
				}
				if tc.god && playerTarget {
					if target.Obj130 != before.Obj130 || target.Field131 != before.Field131 || target.Frame134 != before.Frame134 || target.Field38 != before.Field38 {
						t.Fatal("GodMode entered HP/default tail")
					}
				} else if target.Obj130 != weapon || target.Field131 != uint32(tc.typ) || target.Frame134 != 1400 || target.Pos132 != source.PrevPos || target.Field38 != math.MaxUint32 {
					t.Fatal("registered player/NPC HP attribution")
				}
				if itemHP.Cur != 25 {
					if item.Obj130 != weapon || item.Field131 != uint32(tc.typ) || item.Frame134 != 1400 || item.Pos132 != source.PrevPos || item.Field38 != math.MaxUint32 {
						t.Fatalf("registered ArmorDamage attribution: source=%p type=%d frame=%d pos=%v HP marker=%#x", item.Obj130, item.Field131, item.Frame134, item.Pos132, item.Field38)
					}
					if playerTarget {
						want := server.BuildShopItemHealthPacket4D87A0(item)
						if len(packets) != 1 || !bytes.Equal(packets[0], want[:]) {
							t.Fatal("real armor HP change was not reported")
						}
					}
				}
				if (!playerTarget || itemHP.Cur == 25) && len(packets) != 0 {
					t.Fatal("NPC/no-change item-health packet")
				}
				t.Logf("registered PlayerDamage -> ArmorDamage: target=%p item=%p HP=200->%d armor=25->%d carry=%g", target, item, hp.Cur, itemHP.Cur, carry)
			})
		}
	}
}
