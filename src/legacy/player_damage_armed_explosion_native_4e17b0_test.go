package legacy

import (
	"fmt"
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/things"
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

// Real C dispatcher -> registered PlayerDamage -> real armor lookup,
// EquipDamage/ArmorDamage -> DefaultDamage -> UnitSetHP. Neither unit damage,
// armor damage nor HP services are replaced. Every object and referenced
// record is C-owned above 4 GiB on a 64-bit host.
func TestPlayerDamageArmedExplosionNative4E17B0CProductionHP(t *testing.T) {
	srv := npcReflectLegacyServer4E17B0(t)
	srv.Audio.Init(srv)
	t.Cleanup(srv.Audio.Free)
	srv.FreeObjectTypes()
	t.Cleanup(srv.FreeObjectTypes)
	for _, proc := range []string{"PlayerDamage", "ArmorDamage"} {
		if err := srv.Types.ReadObjectType(&things.Thing{Name: "NativeArmedExplosion" + proc, OnDamage: &things.ProcFunc{Name: proc}}); err != nil {
			t.Fatal(err)
		}
	}
	if !srv.Objs.Init(256) {
		t.Fatal("C-owned object allocator")
	}
	t.Cleanup(srv.Objs.FreeObjects)
	for _, playerSource := range []bool{false, true} {
		for _, playerTarget := range []bool{false, true} {
			for _, attack := range []string{"Sword", "Wand", "Hammer"} {
				for _, gate := range []string{"ordinary", "invulnerable", "dead", "no-update"} {
					t.Run(fmt.Sprintf("player-source-%t/player-target-%t/%s/%s", playerSource, playerTarget, attack, gate), func(t *testing.T) {
						v, a, w, armor := srv.Objs.NewObject(&server.ObjectType{}), srv.Objs.NewObject(&server.ObjectType{}), srv.Objs.NewObject(&server.ObjectType{}), srv.Objs.NewObject(&server.ObjectType{})
						pu, freePU := alloc.New(server.PlayerUpdateData{})
						mu, freeMU := alloc.New(server.MonsterUpdateData{})
						apu, freeAPU := alloc.New(server.PlayerUpdateData{})
						amu, freeAMU := alloc.New(server.MonsterUpdateData{})
						info, freeInfo := alloc.New(server.Player{})
						ainfo, freeAInfo := alloc.New(server.Player{})
						hp, freeHP := alloc.New(server.HealthData{})
						ahp, freeAHP := alloc.New(server.HealthData{})
						init, freeInit := alloc.New(server.ModifierInitData{})
						ainit, freeAInit := alloc.New(server.ModifierInitData{})
						use, freeUse := alloc.New(server.WandUseData{})
						au, freeAU := alloc.New(server.WeaponArmorUpdateData{})
						def, freeDef := alloc.New(server.Modifier{})
						for _, free := range []func(){freePU, freeMU, freeAPU, freeAMU, freeInfo, freeAInfo, freeHP, freeAHP, freeInit, freeAInit, freeUse, freeAU, freeDef} {
							t.Cleanup(free)
						}
						*info, *ainfo = server.Player{PlayerInd: 9}, server.Player{PlayerInd: 10}
						*pu = server.PlayerUpdateData{Player: info, State: server.PlayerState13, Field57: math.Float32bits(.5), Field76: 99, Field75: 77}
						*mu = server.MonsterUpdateData{Field518: math.Float32bits(.5), Field547: 99, Field546: 77}
						*apu = server.PlayerUpdateData{Player: ainfo, State: server.PlayerState13}
						*hp, *ahp = server.HealthData{Cur: 2000, Max: 2000, Field2: 2000}, server.HealthData{Cur: 100, Max: 100, Field2: 100}
						v.TypeInd, v.ObjClass, v.ObjSubClass, v.ObjFlags, v.Material, v.UpdateData = 1343, object.Class(2097154), 69650, object.Flags(16777732), 0x4000, unsafe.Pointer(mu)
						if playerTarget {
							v.TypeInd, v.ObjClass, v.ObjSubClass, v.UpdateData = 713, object.Class(2621444), 0, unsafe.Pointer(pu)
						}
						v.HealthData, v.Damage, v.InvFirstItem = hp, srv.Types.ByID("NativeArmedExplosionPlayerDamage").Damage, armor
						a.TypeInd, a.ObjClass, a.ObjSubClass, a.ObjFlags, a.UpdateData = 1343, object.Class(2097154), 69778, object.Flags(16777732), unsafe.Pointer(amu)
						if playerSource {
							a.TypeInd, a.ObjClass, a.ObjSubClass, a.UpdateData = 713, object.Class(2621444), 0, unsafe.Pointer(apu)
						}
						a.PrevPos = types.Ptf(32.5, -18.25)
						w.TypeInd, w.ObjClass, w.ObjSubClass, w.ObjFlags, w.InitData = 1174, object.Class(85458944), 256, object.Flags(17301776), unsafe.Pointer(init)
						w.PrevPos = types.Ptf(20, 9)
						if attack == "Wand" {
							w.TypeInd, w.ObjClass, w.ObjSubClass, w.UseData.Ptr, use.Flags = 777, object.ClassWand, 0x10000, unsafe.Pointer(use), 2
						}
						if attack == "Hammer" {
							w.TypeInd, w.ObjClass, w.ObjSubClass = 777, object.ClassWeapon, 0x4000
						}
						armor.TypeInd, armor.ObjClass, armor.ObjFlags, armor.Material = 301, object.ClassArmor, object.FlagEquipped|object.FlagEnabled, 0x4000
						armor.InitData, armor.UpdateData, armor.HealthData, armor.InvHolder, armor.Damage = unsafe.Pointer(ainit), unsafe.Pointer(au), ahp, v, srv.Types.ByID("NativeArmedExplosionArmorDamage").Damage
						*def = server.Modifier{TypeInd: 301, DamageCoeffOrArmor64: .5}
						beforeDef := srv.Modif.Dword_5d4594_251608
						srv.Modif.Dword_5d4594_251608 = def
						defer func() { srv.Modif.Dword_5d4594_251608 = beforeDef }()
						v.TeamVal.ID, a.TeamVal.ID = 2, 1
						if !srv.IsEnemyTo(v, a) {
							t.Fatal("real hostile service")
						}
						switch gate {
						case "invulnerable":
							v.Buffs |= 1 << server.ENCHANT_INVULNERABLE
						case "dead":
							v.ObjFlags |= object.FlagDead
						case "no-update":
							v.ObjFlags |= object.FlagNoUpdate
						}
						for _, ptr := range []unsafe.Pointer{unsafe.Pointer(v), unsafe.Pointer(a), unsafe.Pointer(w), unsafe.Pointer(armor), unsafe.Pointer(pu), unsafe.Pointer(mu), unsafe.Pointer(apu), unsafe.Pointer(amu), unsafe.Pointer(info), unsafe.Pointer(ainfo), unsafe.Pointer(hp), unsafe.Pointer(ahp), unsafe.Pointer(init), unsafe.Pointer(ainit), unsafe.Pointer(use), unsafe.Pointer(au), unsafe.Pointer(def)} {
							if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(ptr) <= math.MaxUint32 {
								t.Fatalf("C-owned record=%p, want >4 GiB", ptr)
							}
						}
						var hpSteps, armorSteps [3]uint16
						for hit := range 3 {
							result := objectDamageDispatchCallNative(v, a, w, 3, object.DamageExplosion)
							marker, kind, carry := mu.Field547, mu.Field546, mu.Field1
							if playerTarget {
								marker, kind, carry = pu.Field76, pu.Field75, pu.Field21
							}
							if gate == "ordinary" {
								if !result || hp.Cur != []uint16{1998, 1997, 1995}[hit] || ahp.Cur != []uint16{99, 97, 96}[hit] ||
									carry != math.Float32bits([]float32{-.5, 0, -.5}[hit]) || au.Field0 != 0 || marker != 1 || kind != uint32(w.TypeInd) ||
									v.Obj130 != w || v.Pos132 != a.PrevPos || v.Field131 != 7 || v.Frame134 != 1400 || v.Field38 != math.MaxUint32 || (!playerSource && amu.Field130 != 1400) {
									t.Fatalf("real C hit=%d result=%t HP=%d armor=%d marker=%d/%d carry=%08x item-carry=%08x", hit, result, hp.Cur, ahp.Cur, marker, kind, carry, au.Field0)
								}
							} else if result != (gate == "invulnerable") || hp.Cur != 2000 || ahp.Cur != 100 || carry != 0 || marker != 99 || kind != 77 || v.Obj130 != nil || v.Frame134 != 0 || amu.Field130 != 0 {
								t.Fatalf("early %s changed HP/armor/prefix or return: %t %d/%d %d/%d", gate, result, hp.Cur, ahp.Cur, marker, kind)
							}
							hpSteps[hit], armorSteps[hit] = hp.Cur, ahp.Cur
						}
						t.Logf("C-owned >4 GiB C->registered PlayerDamage->real armor lookup/EquipDamage/ArmorDamage->DefaultDamage->UnitSetHP player-source=%t player-target=%t weapon=%s gate=%s actual damage=3/type=7 HP=2000->%v armor=100->%v", playerSource, playerTarget, attack, gate, hpSteps, armorSteps)
					})
				}
			}
		}
	}
}
