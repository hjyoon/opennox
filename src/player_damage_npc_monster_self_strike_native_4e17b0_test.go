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

// Synthetic C-owned records exercise the registered production damage
// callback, real armor definition/durability and HP setters. Only the final
// owner-HP transport is captured; health, enemy and damage services are real.
func TestPlayerDamageNPCMonsterSelfStrikeNative4E17B0HPAndDurability(t *testing.T) {
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
	noxflags.SetEngine(noxflags.EngineGodMode) // Player-only immunity must not protect an NPC.
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
		if err := s.Types.ReadObjectType(&things.Thing{Name: "NativeNPCSelfStrike" + callback, OnDamage: &things.ProcFunc{Name: callback}}); err != nil {
			t.Fatal(err)
		}
	}
	if !s.Objs.Init(80) {
		t.Fatal("native allocator initialization")
	}
	t.Cleanup(s.Objs.FreeObjects)
	for _, typ := range []object.DamageType{object.DamageBlade, object.DamageCrush, object.DamageImpale, object.DamageDrain, object.DamageBite, object.DamageClaw} {
		for _, absorption := range []float32{0, 0.25, 1} {
			t.Run(fmt.Sprintf("type-%d/armor-%g", typ, absorption), func(t *testing.T) {
				newObject := func(callback string, ind uint16) *server.Object {
					def := &server.ObjectType{}
					if callback != "" {
						def.Damage = s.Types.ByID("NativeNPCSelfStrike" + callback).Damage
					}
					obj := s.Objs.NewObject(def)
					obj.TypeInd, obj.ObjFlags = ind, 0
					return obj
				}
				target, source, item, owner := newObject("PlayerDamage", 71), newObject("", 72), newObject("ArmorDamage", 74), newObject("", 75)
				ud, freeUD := alloc.New(server.MonsterUpdateData{})
				sourceUD, freeSourceUD := alloc.New(server.MonsterUpdateData{})
				playerUD, freePlayerUD := alloc.New(server.PlayerUpdateData{})
				player, freePlayer := alloc.New(server.Player{})
				initData, freeInit := alloc.New(server.ModifierInitData{})
				itemUD, freeItemUD := alloc.New(server.WeaponArmorUpdateData{})
				npcHP, freeHP := alloc.New(server.HealthData{})
				itemHP, freeItemHP := alloc.New(server.HealthData{})
				armorDef, freeDef := alloc.New(server.Modifier{})
				for _, free := range []func(){freeUD, freeSourceUD, freePlayerUD, freePlayer, freeInit, freeItemUD, freeHP, freeItemHP, freeDef} {
					t.Cleanup(free)
				}
				*ud = server.MonsterUpdateData{Field518: math.Float32bits(absorption), Field547: 99, Field546: 77, Field523_2: 0x34}
				*sourceUD, *initData, *itemUD = server.MonsterUpdateData{}, server.ModifierInitData{}, server.WeaponArmorUpdateData{}
				*player, *playerUD = server.Player{PlayerInd: 7}, server.PlayerUpdateData{Player: player, State: server.PlayerState13}
				*npcHP, *itemHP = server.HealthData{Cur: 60, Max: 60}, server.HealthData{Cur: 100, Max: 100}
				*armorDef = server.Modifier{TypeInd: uint32(item.TypeInd), DamageCoeffOrArmor64: absorption}
				previousArmorDef := s.Modif.Dword_5d4594_251608
				s.Modif.Dword_5d4594_251608 = armorDef
				t.Cleanup(func() { s.Modif.Dword_5d4594_251608 = previousArmorDef })
				target.ObjClass, target.ObjSubClass, target.Material = object.ClassMonster, 0x11012, 0x4000
				target.UpdateData, target.HealthData = unsafe.Pointer(ud), npcHP
				source.ObjClass, source.UpdateData, source.PrevPos = object.ClassMonster, unsafe.Pointer(sourceUD), types.Ptf(-20, 7)
				owner.ObjClass, owner.UpdateData = object.ClassPlayer, unsafe.Pointer(playerUD)
				target.ObjOwner, owner.Field129, target.NetCode = owner, target, 0x1234
				item.ObjClass, item.ObjSubClass, item.ObjFlags = object.ClassArmor, 2, object.FlagEquipped
				item.HealthData, item.UpdateData, item.InitData = itemHP, unsafe.Pointer(itemUD), unsafe.Pointer(initData)
				target.InvFirstItem, item.InvHolder = item, target
				// A zero-absorption case has no equipped armor; an equipped
				// zero-coefficient item would intentionally exercise 0/0 in
				// the original durability distribution, not bare NPC damage.
				var wantInventory *server.Object = item
				if absorption == 0 {
					target.InvFirstItem, item.InvHolder, item.ObjFlags = nil, nil, 0
					wantInventory = nil
				}
				ownerReports := 0
				previousSend := s.NetSendPacketXxx
				s.NetSendPacketXxx = func(recipient int, packet []byte, related *server.Object, remove, sequence int) int {
					want := [4]byte{65, 0x34, 0x12, byte(npcHP.Cur >> 1)}
					if recipient != 7 || len(packet) != 4 || [4]byte(packet) != want || related != nil || remove != 1 || sequence != 1 {
						t.Fatalf("owner HP report recipient=%d packet=%x", recipient, packet)
					}
					ownerReports++
					return 1
				}
				t.Cleanup(func() { s.NetSendPacketXxx = previousSend })
				if !s.IsEnemyTo(target, source) {
					t.Fatal("monster must be hostile to the player-owned NPC")
				}
				for _, pointer := range []unsafe.Pointer{unsafe.Pointer(target), unsafe.Pointer(source), unsafe.Pointer(item), unsafe.Pointer(owner), target.UpdateData, source.UpdateData, item.UpdateData, item.InitData, owner.UpdateData, unsafe.Pointer(player), unsafe.Pointer(npcHP), unsafe.Pointer(itemHP), unsafe.Pointer(armorDef)} {
					if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(pointer) <= math.MaxUint32 {
						t.Fatalf("pointer=%p, want above 4 GiB", pointer)
					}
				}
				wantHP, wantArmor, carry := uint16(60), uint16(100), float32(0)
				for hit := 0; hit < 3; hit++ {
					x := float64(absorption)
					if typ == object.DamageCrush {
						x *= 0.5
					}
					accumulated := float32((1-x)*3) + carry
					rounded := int32(math.RoundToEven(float64(accumulated)))
					effective := rounded
					carry = accumulated - float32(rounded)
					if typ == object.DamageDrain {
						effective, carry = 3, 0
					} else if absorption != 0 {
						wantArmor -= uint16(3 - rounded)
					}
					if effective == 0 {
						effective = 1
					}
					wantHP -= uint16(effective)
					if !target.CallDamage(source, source, 3, typ) {
						t.Fatalf("registered self-strike %d rejected", hit)
					}
					if npcHP.Cur != wantHP || itemHP.Cur != wantArmor || ud.Field1 != math.Float32bits(carry) || itemUD.Field0 != 0 || ud.Field547 != 2 || ud.Field546 != uint32(typ) || !ud.StatusFlags.Has(object.MonStatusInjured) || target.Obj130 != source || target.Frame134 != 1400 || target.Pos132 != source.PrevPos || sourceUD.Field130 != 1400 || ownerReports != hit+1 || target.InvFirstItem != wantInventory || wantInventory != nil && item.InvHolder != target || s.Objs.DeletedList != nil {
						t.Fatalf("hit=%d HP=%d/%d armor=%d/%d carry=%g/%g marker=%d/%d reports=%d", hit, npcHP.Cur, wantHP, itemHP.Cur, wantArmor, math.Float32frombits(ud.Field1), carry, ud.Field547, ud.Field546, ownerReports)
					}
				}
				t.Logf("registered NPC self-strike: target=%p source=weapon=%p type=%d HP=60->%d armor=100->%d carry=%g", target, source, typ, npcHP.Cur, itemHP.Cur, carry)
			})
		}
	}
}
