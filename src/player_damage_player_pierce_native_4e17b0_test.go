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

// Synthetic native-width player/armor records, not a stock-map assertion.
// Registered PlayerDamage, C armor lookup and facing, the native armor-wear
// tail, DefaultDamage and UnitSetHP run without replacement damage/HP callbacks.
func TestPlayerDamagePlayerPierceNative4E17B0HPAndDurability(t *testing.T) {
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
	for _, callback := range []string{"PlayerDamage", "ArmorDamage"} {
		if err := s.Types.ReadObjectType(&things.Thing{Name: "NativePlayerPierce" + callback, OnDamage: &things.ProcFunc{Name: callback}}); err != nil {
			t.Fatal(err)
		}
	}
	if !s.Objs.Init(160) {
		t.Fatal("native allocator initialization")
	}
	t.Cleanup(s.Objs.FreeObjects)
	for _, playerSource := range []bool{false, true} {
		for _, pure := range []bool{false, true} {
			for _, mode := range []string{"normal", "ordinary rear", "Reflect Shield rear", "GodMode", "zero", "negative", "zero with negative carry"} {
				t.Run(fmt.Sprintf("player-source-%t/pure-%t/%s", playerSource, pure, mode), func(t *testing.T) {
					noxflags.UnsetEngine(noxflags.EngineGodMode)
					if mode == "GodMode" {
						noxflags.SetEngine(noxflags.EngineGodMode)
					}
					newObject := func(callback string, ind uint16) *server.Object {
						def := &server.ObjectType{}
						if callback != "" {
							def.Damage = s.Types.ByID("NativePlayerPierce" + callback).Damage
						}
						obj := s.Objs.NewObject(def)
						obj.TypeInd, obj.ObjFlags = ind, 0
						return obj
					}
					target, source, arrow, item := newObject("PlayerDamage", 71), newObject("", 72), newObject("", 529), newObject("ArmorDamage", 74)
					ud, freeUD := alloc.New(server.PlayerUpdateData{})
					player, freePlayer := alloc.New(server.Player{})
					sourceUD, freeSourceUD := alloc.New(server.MonsterUpdateData{})
					sourcePlayerUD, freeSourcePlayerUD := alloc.New(server.PlayerUpdateData{})
					sourcePlayer, freeSourcePlayer := alloc.New(server.Player{})
					initData, freeInit := alloc.New(server.ModifierInitData{})
					itemUD, freeItemUD := alloc.New(server.WeaponArmorUpdateData{})
					hp, freeHP := alloc.New(server.HealthData{})
					itemHP, freeItemHP := alloc.New(server.HealthData{})
					armorDef, freeDef := alloc.New(server.Modifier{})
					for _, free := range []func(){freeUD, freePlayer, freeSourceUD, freeSourcePlayerUD, freeSourcePlayer, freeInit, freeItemUD, freeHP, freeItemHP, freeDef} {
						t.Cleanup(free)
					}
					*player = server.Player{PlayerInd: 7, ArmorEquip: 0x405, WeaponEquip: 0x100}
					*ud = server.PlayerUpdateData{Player: player, State: server.PlayerState13, Field57: math.Float32bits(0.25), Field76: 99, Field75: 77, Field40_0: 0x1234, Field40_1: 0xabcd}
					*sourceUD, *initData, *itemUD = server.MonsterUpdateData{}, server.ModifierInitData{}, server.WeaponArmorUpdateData{}
					*sourcePlayer = server.Player{PlayerInd: 8}
					*sourcePlayerUD = server.PlayerUpdateData{Player: sourcePlayer, State: server.PlayerState13, Field57: math.Float32bits(0.125)}
					*hp, *itemHP = server.HealthData{Cur: 60, Max: 60}, server.HealthData{Cur: 3, Max: 3}
					*armorDef = server.Modifier{TypeInd: uint32(item.TypeInd), DamageCoeffOrArmor64: 0.25}
					previousDef := s.Modif.Dword_5d4594_251608
					s.Modif.Dword_5d4594_251608 = armorDef
					t.Cleanup(func() { s.Modif.Dword_5d4594_251608 = previousDef })
					target.ObjClass, target.UpdateData, target.HealthData, target.Material = object.ClassPlayer, unsafe.Pointer(ud), hp, 0x4000
					source.ObjClass, source.UpdateData, source.PrevPos = object.ClassMonster, unsafe.Pointer(sourceUD), types.Ptf(44, 7)
					if playerSource {
						source.ObjClass, source.UpdateData = object.ClassPlayer, unsafe.Pointer(sourcePlayerUD)
					}
					arrow.ObjClass, arrow.ObjSubClass, arrow.ObjOwner = object.Class(0x05200001), 0x10, source
					if pure {
						arrow.ObjClass = object.ClassMissile
					}
					arrow.PosVec, arrow.PrevPos, arrow.VelVec = types.Ptf(20, 0), types.Ptf(-20, 0), types.Ptf(4, 0)
					item.ObjClass, item.ObjSubClass, item.ObjFlags = object.ClassArmor, 2, object.FlagEquipped
					item.UpdateData, item.InitData, item.HealthData, item.InvHolder = unsafe.Pointer(itemUD), unsafe.Pointer(initData), itemHP, target
					target.InvFirstItem = item
					if mode == "ordinary rear" {
						ud.State, player.ArmorEquip = server.PlayerState16, 0x1000000
						if legacy.Nox_server_testTwoPointsAndDirection_4E6E50(target.PosVec, int16(target.Direction1), arrow.PrevPos)&1 != 0 {
							t.Fatal("ordinary shield fixture is not a rear attack")
						}
					} else if mode == "Reflect Shield rear" {
						target.Buffs = 1 << server.ENCHANT_REFLECTIVE_SHIELD
						arrow.PosVec, arrow.PrevPos = types.Ptf(-20, 0), types.Ptf(20, 0)
						if legacy.Nox_server_testTwoPointsAndDirection_4E6E50(target.PosVec, int16(target.Direction1), arrow.PosVec)&1 != 0 {
							t.Fatal("Reflect Shield fixture is not a rear attack")
						}
					}
					for _, pointer := range []unsafe.Pointer{unsafe.Pointer(target), unsafe.Pointer(source), unsafe.Pointer(arrow), unsafe.Pointer(item), target.UpdateData, source.UpdateData, unsafe.Pointer(player), unsafe.Pointer(sourcePlayerUD), unsafe.Pointer(sourcePlayer), item.UpdateData, item.InitData, unsafe.Pointer(hp), unsafe.Pointer(itemHP), unsafe.Pointer(armorDef)} {
						if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(pointer) <= math.MaxUint32 {
							t.Fatalf("pointer=%p, want above 4 GiB", pointer)
						}
					}
					damages, carries, expectedHP, expectedArmor := []int{3, 3, 3}, []float32{0.25, 0.5, -0.25}, []uint16{58, 56, 53}, []uint16{2, 1, 1}
					switch mode {
					case "GodMode":
						expectedHP = []uint16{60, 60, 60}
					case "zero":
						damages, carries, expectedHP, expectedArmor = []int{0}, []float32{0}, []uint16{60}, []uint16{3}
					case "negative":
						damages, carries, expectedHP, expectedArmor = []int{-3}, []float32{-0.25}, []uint16{62}, []uint16{3}
					case "zero with negative carry":
						ud.Field21 = math.Float32bits(-0.75)
						damages, carries, expectedHP, expectedArmor = []int{0}, []float32{0.25}, []uint16{61}, []uint16{2}
					}
					beforeTarget, beforeSource, beforeSourcePlayerUD, beforeSourcePlayer, beforeArrow, beforePlayer, beforeBuffs, beforeState := *target, *source, *sourcePlayerUD, *sourcePlayer, *arrow, *player, target.Buffs, ud.State
					for hit, damage := range damages {
						if !target.CallDamage(source, arrow, damage, object.DamageImpale) {
							t.Fatalf("registered hit %d rejected", hit)
						}
						if hp.Cur != expectedHP[hit] || itemHP.Cur != expectedArmor[hit] || ud.Field21 != math.Float32bits(carries[hit]) || itemUD.Field0 != 0 || ud.Field76 != 1 || ud.Field75 != 529 ||
							ud.Field40_0 != 0x1234 || ud.Field40_1 != 0xabcd || ud.State != beforeState || *player != beforePlayer || target.Buffs != beforeBuffs ||
							target.InvFirstItem != item || item.InvHolder != target || *arrow != beforeArrow ||
							(playerSource && (*source != beforeSource || *sourcePlayerUD != beforeSourcePlayerUD || *sourcePlayer != beforeSourcePlayer)) || s.Objs.DeletedList != nil {
							t.Fatalf("hit=%d raw=%d HP=%d armor=%d carry=%g marker=%d/%d", hit, damage, hp.Cur, itemHP.Cur, math.Float32frombits(ud.Field21), ud.Field76, ud.Field75)
						}
						if mode == "GodMode" {
							// NewObject starts with a dirty sync mask; the GodMode
							// prefix must preserve it, not manufacture a clean mask.
							if target.Obj130 != beforeTarget.Obj130 || target.Field131 != beforeTarget.Field131 || target.Frame134 != beforeTarget.Frame134 || target.Field38 != beforeTarget.Field38 {
								t.Fatal("GodMode ran HP/default tail after armor wear")
							}
						} else {
							wantPos := source.PrevPos
							if pure {
								wantPos = arrow.PrevPos
							}
							if target.Obj130 != arrow || target.Field131 != 3 || target.Frame134 != 1400 || target.Pos132 != wantPos || target.Field38 != math.MaxUint32 {
								t.Fatal("native PIERCE attribution/synchronization")
							}
						}
						if itemHP.Cur < 3 && (item.Obj130 != arrow || item.Field131 != 3 || item.Frame134 != 1400) {
							t.Fatal("native armor attribution")
						}
					}
					t.Logf("registered player PIERCE: target=%p source=%p arrow=%p armor=%p mode=%s HP=%v armor=%v", target, source, arrow, item, mode, expectedHP, expectedArmor)
				})
			}
		}
	}
}
