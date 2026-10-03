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
	"github.com/opennox/opennox/v1/common/unit/ai"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

// This is a synthetic NPC/armor fixture, not a stock-map assertion. It uses
// the registered production PlayerDamage callback, actual C armor-definition
// lookup, native armor/HP damage and the real facing test. No damage, armor,
// ownership, health or durability callback is substituted. All records and
// links crossing those C boundaries are C-owned native-width allocations.
// The outgoing owner-HP transport alone is captured to verify the wire packet.
func TestPlayerDamageNPCPierceNative4E17B0HPAndDurability(t *testing.T) {
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
	// NPCs must take damage even when the player-only GodMode flag is set.
	noxflags.SetEngine(noxflags.EngineGodMode)
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
		if err := s.Types.ReadObjectType(&things.Thing{Name: "NativeNPCPierce" + callback, OnDamage: &things.ProcFunc{Name: callback}}); err != nil {
			t.Fatal(err)
		}
	}
	if !s.Objs.Init(80) {
		t.Fatal("native allocator initialization")
	}
	t.Cleanup(s.Objs.FreeObjects)
	for _, playerSource := range []bool{false, true} {
		for _, pure := range []bool{false, true} {
			for _, defense := range []string{"none", "ordinary rear", "Reflect Shield rear"} {
				t.Run(fmt.Sprintf("player-source-%t/pure-%t/%s", playerSource, pure, defense), func(t *testing.T) {
					newObject := func(callback string, ind uint16) *server.Object {
						def := &server.ObjectType{}
						if callback != "" {
							def.Damage = s.Types.ByID("NativeNPCPierce" + callback).Damage
						}
						obj := s.Objs.NewObject(def)
						obj.TypeInd, obj.ObjFlags = ind, 0
						return obj
					}
					target, source, arrow := newObject("PlayerDamage", 71), newObject("", 72), newObject("", 529)
					item, owner := newObject("ArmorDamage", 74), newObject("", 75)
					ud, freeUD := alloc.New(server.MonsterUpdateData{})
					sourceUD, freeSourceUD := alloc.New(server.MonsterUpdateData{})
					sourcePlayerUD, freeSourcePlayerUD := alloc.New(server.PlayerUpdateData{})
					sourcePlayer, freeSourcePlayer := alloc.New(server.Player{})
					playerUD, freePlayerUD := alloc.New(server.PlayerUpdateData{})
					player, freePlayer := alloc.New(server.Player{})
					initData, freeInit := alloc.New(server.ModifierInitData{})
					itemUD, freeItemUD := alloc.New(server.WeaponArmorUpdateData{})
					npcHP, freeHP := alloc.New(server.HealthData{})
					itemHP, freeItemHP := alloc.New(server.HealthData{})
					armorDef, freeDef := alloc.New(server.Modifier{})
					for _, free := range []func(){freeUD, freeSourceUD, freeSourcePlayerUD, freeSourcePlayer, freePlayerUD, freePlayer, freeInit, freeItemUD, freeHP, freeItemHP, freeDef} {
						t.Cleanup(free)
					}
					*ud = server.MonsterUpdateData{Field518: math.Float32bits(0.25), Field547: 99, Field546: 77, Field523_2: 0x34}
					*sourceUD, *playerUD, *player, *initData, *itemUD = server.MonsterUpdateData{}, server.PlayerUpdateData{}, server.Player{}, server.ModifierInitData{}, server.WeaponArmorUpdateData{}
					*sourcePlayer, *sourcePlayerUD = server.Player{PlayerInd: 8}, server.PlayerUpdateData{Player: sourcePlayer, State: server.PlayerState13, Field57: math.Float32bits(0.125)}
					*npcHP, *itemHP = server.HealthData{Cur: 60, Max: 60}, server.HealthData{Cur: 3, Max: 3}
					*armorDef = server.Modifier{TypeInd: uint32(item.TypeInd), DamageCoeffOrArmor64: 0.25}
					previousArmorDef := s.Modif.Dword_5d4594_251608
					s.Modif.Dword_5d4594_251608 = armorDef
					t.Cleanup(func() { s.Modif.Dword_5d4594_251608 = previousArmorDef })
					target.ObjClass, target.ObjSubClass, target.Material = object.ClassMonster, 0x11012, 0x4000
					target.UpdateData, target.HealthData = unsafe.Pointer(ud), npcHP
					source.ObjClass, source.UpdateData = object.ClassMonster, unsafe.Pointer(sourceUD)
					if playerSource {
						source.ObjClass, source.UpdateData = object.ClassPlayer, unsafe.Pointer(sourcePlayerUD)
					}
					source.PrevPos = types.Ptf(44, 7)
					owner.ObjClass, owner.UpdateData, playerUD.Player = object.ClassPlayer, unsafe.Pointer(playerUD), player
					// Model a player-owned NPC. GameplayFlag1 admits ranged damage
					// from either unit source; only a hostile MONSTER source has
					// the monster hit-time latch, never a PLAYER update.
					target.ObjOwner, owner.Field129 = owner, target
					player.PlayerInd, target.NetCode = 7, 0x1234
					ownerReports := 0
					previousSend := s.NetSendPacketXxx
					s.NetSendPacketXxx = func(recipient int, packet []byte, related *server.Object, remove, sequence int) int {
						wantPacket := [4]byte{65, 0x34, 0x12, byte(npcHP.Cur >> 1)}
						if recipient != 7 || len(packet) != 4 || [4]byte(packet) != wantPacket || related != nil || remove != 1 || sequence != 1 {
							t.Fatalf("native owner HP report: recipient=%d packet=%x related=%p remove=%d sequence=%d", recipient, packet, related, remove, sequence)
						}
						ownerReports++
						return 1
					}
					t.Cleanup(func() { s.NetSendPacketXxx = previousSend })
					arrow.ObjClass, arrow.ObjSubClass = object.Class(0x05200001), 0x10
					if pure {
						arrow.ObjClass = object.ClassMissile
					}
					arrow.ObjOwner = source
					arrow.PrevPos, arrow.PosVec = types.Ptf(-20, 0), types.Ptf(20, 0)
					arrow.VelVec, arrow.Direction1 = types.Ptf(4, 0), 0
					item.ObjClass, item.ObjSubClass, item.ObjFlags = object.ClassArmor, 2, object.FlagEquipped
					item.HealthData, item.UpdateData, item.InitData = itemHP, unsafe.Pointer(itemUD), unsafe.Pointer(initData)
					target.InvFirstItem, item.InvHolder = item, target
					ud.AIStack[0].Action = uint32(ai.ACTION_GUARD)
					if defense == "ordinary rear" {
						ud.ArmorEquipFlags, ud.AIStack[0].Action = 0x1000000, uint32(ai.ACTION_BLOCK_ATTACK)
						if legacy.Nox_server_testTwoPointsAndDirection_4E6E50(target.PosVec, int16(target.Direction1), arrow.PrevPos)&1 != 0 {
							t.Fatal("ordinary shield fixture is not a rear attack")
						}
					} else if defense == "Reflect Shield rear" {
						target.Buffs = 1 << server.ENCHANT_REFLECTIVE_SHIELD
						arrow.PosVec, arrow.PrevPos = types.Ptf(-20, 0), types.Ptf(20, 0)
						if legacy.Nox_server_testTwoPointsAndDirection_4E6E50(target.PosVec, int16(target.Direction1), arrow.PosVec)&1 != 0 {
							t.Fatal("Reflect Shield fixture is not a rear attack")
						}
					}
					if !playerSource && !s.IsEnemyTo(target, source) {
						t.Fatal("fixture monster must be hostile to the player-owned NPC")
					}
					for _, pointer := range []unsafe.Pointer{unsafe.Pointer(target), unsafe.Pointer(source), unsafe.Pointer(arrow), unsafe.Pointer(item), unsafe.Pointer(owner), target.UpdateData, source.UpdateData, item.UpdateData, item.InitData, owner.UpdateData, unsafe.Pointer(player), unsafe.Pointer(sourcePlayer), unsafe.Pointer(sourcePlayerUD), unsafe.Pointer(npcHP), unsafe.Pointer(itemHP), unsafe.Pointer(armorDef)} {
						if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(pointer) <= math.MaxUint32 {
							t.Fatalf("pointer=%p, want above 4 GiB", pointer)
						}
					}
					beforeSource, beforeSourcePlayerUD, beforeSourcePlayer := *source, *sourcePlayerUD, *sourcePlayer
					beforeArrow, beforeAction, beforeBuffs := *arrow, ud.AIStack[0], target.Buffs
					for hit, carry := range []float32{0.25, 0.5, -0.25} {
						if !target.CallDamage(source, arrow, 3, object.DamageImpale) {
							t.Fatalf("registered hit %d rejected", hit)
						}
						wantHP, wantArmor := []uint16{58, 56, 53}[hit], []uint16{2, 1, 1}[hit]
						wantPos := source.PrevPos
						if pure {
							wantPos = arrow.PrevPos
						}
						if npcHP.Cur != wantHP || itemHP.Cur != wantArmor || ud.Field1 != math.Float32bits(carry) || itemUD.Field0 != 0 ||
							ud.Field547 != 1 || ud.Field546 != uint32(arrow.TypeInd) || ud.Field523_2 != 0x34 || !ud.StatusFlags.Has(object.MonStatusInjured) ||
							target.Obj130 != arrow || target.Field131 != 3 || target.Frame134 != 1400 || target.Pos132 != wantPos || target.Field38 != math.MaxUint32 ||
							item.Obj130 != arrow || item.Field131 != 3 || item.Frame134 != 1400 || (!playerSource && sourceUD.Field130 != 1400) ||
							(playerSource && (*source != beforeSource || *sourcePlayerUD != beforeSourcePlayerUD || *sourcePlayer != beforeSourcePlayer)) ||
							*arrow != beforeArrow || ud.AIStack[0] != beforeAction || target.Buffs != beforeBuffs || target.InvFirstItem != item || item.InvHolder != target || target.ObjOwner != owner || owner.Field129 != target || ownerReports != hit+1 || s.Objs.DeletedList != nil {
							t.Fatalf("hit=%d HP=%d armor=%d carry=%g NPC marker=%d/%d item source=%p", hit, npcHP.Cur, itemHP.Cur, math.Float32frombits(ud.Field1), ud.Field547, ud.Field546, item.Obj130)
						}
					}
					t.Logf("registered NPC PIERCE player-source=%t: target=%p arrow=%p armor=%p native definition=%p HP=60->58->56->53 armor=3->2->1->1 carry=0.25,0.5,-0.25", playerSource, target, arrow, item, armorDef)
				})
			}
		}
	}
}
