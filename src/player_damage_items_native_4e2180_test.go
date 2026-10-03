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

// Synthetic native-width records, not a stock-map assertion. The actual C
// 004E2180 entry, C armor lookup, EquipDamage, registered ArmorDamage,
// DefaultDamage and UnitSetHP run without replacement damage/HP callbacks.
// Only outgoing player item-health transport is captured.
func TestPlayerDamageItemsNative4E2180RegisteredArmorDamage(t *testing.T) {
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
	if err := s.Types.ReadObjectType(&things.Thing{Name: "NativeArmorItems", OnDamage: &things.ProcFunc{Name: "ArmorDamage"}}); err != nil {
		t.Fatal(err)
	}
	if !s.Objs.Init(100) {
		t.Fatal("native object allocator")
	}
	t.Cleanup(s.Objs.FreeObjects)
	cases := []struct {
		name     string
		armor    float32
		carry    float32
		damages  []int32
		typ      object.DamageType
		material object.Material
		health   []uint16
		carries  []float32
	}{
		{name: "fractional wear", armor: 0.5, damages: []int32{3, 3}, typ: object.DamageImpale, health: []uint16{18, 17}, carries: []float32{-0.5, 0}},
		{name: "zero flushes carry", armor: 0.5, carry: 0.75, damages: []int32{0}, typ: object.DamageImpale, health: []uint16{19}, carries: []float32{-0.25}},
		{name: "negative carries forward", armor: 0.5, damages: []int32{-3, 3}, typ: object.DamageImpale, health: []uint16{20, 18}, carries: []float32{0.5, 0}},
		{name: "cloth crush", armor: 0.5, damages: []int32{3}, typ: object.DamageCrush, health: []uint16{18}, carries: []float32{-0.5}},
		{name: "metal crush doubles wear", armor: 0.5, damages: []int32{3}, typ: object.DamageCrush, material: object.MaterialMetal, health: []uint16{16}, carries: []float32{-0.5}},
		{name: "zero armor infinity", damages: []int32{3}, typ: object.DamageImpale, health: []uint16{20}, carries: []float32{float32(math.Inf(1))}},
		{name: "zero armor zero amount", damages: []int32{0}, typ: object.DamageImpale, health: []uint16{20}, carries: []float32{float32(math.NaN())}},
	}
	for _, playerOwner := range []bool{false, true} {
		for _, tc := range cases {
			t.Run(fmt.Sprintf("player-owner-%t/%s", playerOwner, tc.name), func(t *testing.T) {
				newObject := func(ind uint16) *server.Object {
					obj := s.Objs.NewObject(&server.ObjectType{})
					obj.TypeInd, obj.ObjFlags = ind, 0
					return obj
				}
				owner, source, effective, item := newObject(71), newObject(72), newObject(529), newObject(301)
				playerUD, freePlayerUD := alloc.New(server.PlayerUpdateData{})
				player, freePlayer := alloc.New(server.Player{})
				monsterUD, freeMonsterUD := alloc.New(server.MonsterUpdateData{})
				sourceUD, freeSourceUD := alloc.New(server.MonsterUpdateData{})
				initData, freeInit := alloc.New(server.ModifierInitData{})
				itemUD, freeItemUD := alloc.New(server.WeaponArmorUpdateData{})
				itemHP, freeHP := alloc.New(server.HealthData{})
				armorDef, freeDef := alloc.New(server.Modifier{})
				for _, free := range []func(){freePlayerUD, freePlayer, freeMonsterUD, freeSourceUD, freeInit, freeItemUD, freeHP, freeDef} {
					t.Cleanup(free)
				}
				*player = server.Player{PlayerInd: 7}
				*playerUD = server.PlayerUpdateData{Player: player, Field57: math.Float32bits(tc.armor)}
				*monsterUD = server.MonsterUpdateData{Field518: math.Float32bits(tc.armor)}
				*sourceUD, *initData = server.MonsterUpdateData{}, server.ModifierInitData{}
				*itemUD = server.WeaponArmorUpdateData{Field0: math.Float32bits(tc.carry)}
				*itemHP = server.HealthData{Cur: 20, Max: 20}
				*armorDef = server.Modifier{TypeInd: uint32(item.TypeInd), DamageCoeffOrArmor64: 0.25}
				oldDef := s.Modif.Dword_5d4594_251608
				s.Modif.Dword_5d4594_251608 = armorDef
				t.Cleanup(func() { s.Modif.Dword_5d4594_251608 = oldDef })
				owner.ObjClass, owner.ObjSubClass, owner.UpdateData = object.ClassMonster, 0x10, unsafe.Pointer(monsterUD)
				if playerOwner {
					owner.ObjClass, owner.UpdateData = object.ClassPlayer, unsafe.Pointer(playerUD)
				}
				source.ObjClass, source.UpdateData, source.PrevPos = object.ClassMonster, unsafe.Pointer(sourceUD), types.Ptf(44, 7)
				effective.ObjClass, effective.ObjOwner, effective.PrevPos = object.ClassMissile, source, types.Ptf(-20, 0)
				item.ObjClass, item.ObjSubClass, item.ObjFlags = object.ClassArmor, 2, object.FlagEquipped
				item.UpdateData, item.InitData, item.HealthData, item.InvHolder = unsafe.Pointer(itemUD), unsafe.Pointer(initData), itemHP, owner
				item.Material, item.Damage = uint16(tc.material), s.Types.ByID("NativeArmorItems").Damage
				owner.InvFirstItem = item
				var packets [][]byte
				oldSend := s.NetSendPacketXxx
				s.NetSendPacketXxx = func(recipient int, packet []byte, related *server.Object, remove, sequence int) int {
					if !playerOwner || recipient != 7 || related != nil || remove != 0 || sequence != 1 {
						t.Fatal("player item-health transport contract")
					}
					packets = append(packets, append([]byte(nil), packet...))
					return 1
				}
				t.Cleanup(func() { s.NetSendPacketXxx = oldSend })
				for _, pointer := range []unsafe.Pointer{unsafe.Pointer(owner), unsafe.Pointer(source), unsafe.Pointer(effective), unsafe.Pointer(item), owner.UpdateData, source.UpdateData, item.UpdateData, item.InitData, unsafe.Pointer(player), unsafe.Pointer(itemHP), unsafe.Pointer(armorDef)} {
					if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(pointer) <= math.MaxUint32 {
						t.Fatalf("native pointer=%p, want above 4 GiB", pointer)
					}
				}
				beforeOwner, beforeSource, beforeEffective, beforePlayerUD, beforeMonsterUD := *owner, *source, *effective, *playerUD, *monsterUD
				reports := 0
				for hit, damage := range tc.damages {
					beforeHP := itemHP.Cur
					legacy.Nox_xxx_playerDamageItems_4E2180(owner, source, effective, damage, tc.typ)
					carry, wantCarry := math.Float32frombits(itemUD.Field0), tc.carries[hit]
					carryOK := math.Float32bits(carry) == math.Float32bits(wantCarry) || (math.IsNaN(float64(carry)) && math.IsNaN(float64(wantCarry)))
					if itemHP.Cur != tc.health[hit] || !carryOK || *owner != beforeOwner || *source != beforeSource || *effective != beforeEffective || *playerUD != beforePlayerUD || *monsterUD != beforeMonsterUD || item.InvHolder != owner || s.Objs.DeletedList != nil {
						t.Fatalf("hit=%d raw=%d HP=%d carry=%g, want HP=%d carry=%g", hit, damage, itemHP.Cur, carry, tc.health[hit], wantCarry)
					}
					if itemHP.Cur != beforeHP {
						if item.Obj130 != effective || item.Field131 != uint32(tc.typ) || item.Frame134 != 1400 || item.Pos132 != effective.PrevPos || item.Field38 != math.MaxUint32 {
							t.Fatal("registered ArmorDamage/default HP attribution")
						}
						if playerOwner {
							reports++
							wantPacket := server.BuildShopItemHealthPacket4D87A0(item)
							if len(packets) != reports || !bytes.Equal(packets[reports-1], wantPacket[:]) {
								t.Fatal("registered wear was not reported after the real HP change")
							}
						}
					}
					if len(packets) != reports {
						t.Fatal("no-change/NPC wear emitted an item-health packet")
					}
				}
				t.Logf("C 004E2180 -> registered ArmorDamage: owner=%p item=%p HP=%v carry=%v", owner, item, tc.health, tc.carries)
			})
		}
	}
}
