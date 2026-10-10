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

// Exercise the already-calculated DefaultDamage tail, not the separate
// PlayerDamage armor prefix. All object/data records are C-owned >4 GiB;
// neither the production damage binding nor UnitSetHP is replaced.
func TestDefaultDamagePlayerArmedExplosionNative4E0B30(t *testing.T) {
	srv := npcReflectLegacyServer4E17B0(t)
	srv.Audio.Init(srv)
	t.Cleanup(srv.Audio.Free)
	srv.FreeObjectTypes()
	t.Cleanup(srv.FreeObjectTypes)
	if err := srv.Types.ReadObjectType(&things.Thing{Name: "NativePlayerArmedExplosionDefault", OnDamage: &things.ProcFunc{Name: "DefaultDamage"}}); err != nil {
		t.Fatal(err)
	}
	if !srv.Objs.Init(64) {
		t.Fatal("C-owned allocator")
	}
	t.Cleanup(srv.Objs.FreeObjects)
	for _, playerSource := range []bool{false, true} {
		for _, attack := range []string{"Sword", "Wand", "Hammer"} {
			for _, invulnerable := range []bool{false, true} {
				t.Run(fmt.Sprintf("player-source-%t/%s/invulnerable-%t", playerSource, attack, invulnerable), func(t *testing.T) {
					v, a, w := srv.Objs.NewObject(&server.ObjectType{}), srv.Objs.NewObject(&server.ObjectType{}), srv.Objs.NewObject(&server.ObjectType{})
					u, freeU := alloc.New(server.PlayerUpdateData{})
					au, freeAU := alloc.New(server.MonsterUpdateData{})
					apu, freeAPU := alloc.New(server.PlayerUpdateData{})
					info, freeInfo := alloc.New(server.Player{})
					ainfo, freeAInfo := alloc.New(server.Player{})
					hp, freeHP := alloc.New(server.HealthData{})
					init, freeInit := alloc.New(server.ModifierInitData{})
					use, freeUse := alloc.New(server.WandUseData{})
					for _, free := range []func(){freeU, freeAU, freeAPU, freeInfo, freeAInfo, freeHP, freeInit, freeUse} {
						t.Cleanup(free)
					}
					*info, *ainfo = server.Player{PlayerInd: 9}, server.Player{PlayerInd: 10}
					*u = server.PlayerUpdateData{Player: info, State: server.PlayerState13, Field76: 99, Field75: 77, Field21: math.Float32bits(.125), Field57: math.Float32bits(.5)}
					*apu = server.PlayerUpdateData{Player: ainfo, State: server.PlayerState13}
					*hp = server.HealthData{Cur: 2000, Max: 2000, Field2: 2000}
					v.TypeInd, v.ObjClass, v.ObjFlags, v.Material = 713, object.Class(2621444), object.Flags(16777732), 0x4000
					v.UpdateData, v.HealthData, v.Damage = unsafe.Pointer(u), hp, srv.Types.ByID("NativePlayerArmedExplosionDefault").Damage
					a.TypeInd, a.ObjClass, a.ObjSubClass, a.ObjFlags, a.UpdateData = 1343, object.Class(2097154), 69778, object.Flags(16777732), unsafe.Pointer(au)
					if playerSource {
						a.TypeInd, a.ObjClass, a.ObjSubClass, a.UpdateData = 713, object.Class(2621444), 0, unsafe.Pointer(apu)
					}
					a.PrevPos = types.Ptf(32.5, -18.25)
					w.TypeInd, w.ObjClass, w.ObjSubClass, w.ObjFlags, w.InitData = 1174, object.Class(85458944), 256, object.Flags(17301776), unsafe.Pointer(init)
					w.PrevPos = types.Ptf(44, 7)
					if attack == "Wand" {
						w.TypeInd, w.ObjClass, w.ObjSubClass, w.UseData.Ptr, use.Flags = 777, object.ClassWand, 0x10000, unsafe.Pointer(use), 2
					}
					if attack == "Hammer" {
						w.TypeInd, w.ObjClass, w.ObjSubClass = 777, object.ClassWeapon, 0x4000
					}
					v.TeamVal.ID, a.TeamVal.ID = 2, 1
					if !srv.IsEnemyTo(v, a) {
						t.Fatal("hostile fixture rejected by real enemy service")
					}
					if invulnerable {
						v.Buffs |= 1 << server.ENCHANT_INVULNERABLE
					}
					for _, ptr := range []unsafe.Pointer{unsafe.Pointer(v), unsafe.Pointer(a), unsafe.Pointer(w), unsafe.Pointer(u), unsafe.Pointer(au), unsafe.Pointer(apu), unsafe.Pointer(info), unsafe.Pointer(ainfo), unsafe.Pointer(hp), unsafe.Pointer(init), unsafe.Pointer(use)} {
						if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(ptr) <= math.MaxUint32 {
							t.Fatalf("C-owned pointer=%p, want >4 GiB", ptr)
						}
					}
					before, beforeW, beforeAPU := *u, *w, *apu
					var steps [3]uint16
					for hit := range 3 {
						want := uint16(1997 - 3*hit)
						if invulnerable {
							want = 2000
						}
						if !objectDamageDispatchCallNative(v, a, w, 3, object.DamageExplosion) || hp.Cur != want || hp.Max != 2000 || hp.Field2 != 2000 || *u != before || *w != beforeW || *apu != beforeAPU {
							t.Fatalf("actual C EXPLOSION hit=%d HP=%d/%d or independent prefix changed", hit, hp.Cur, want)
						}
						if !invulnerable {
							if v.Obj130 != w || v.Pos132 != a.PrevPos || v.Field131 != 7 || v.Frame134 != 1400 || v.Field38 != math.MaxUint32 || (!playerSource && au.Field130 != 1400) {
								t.Fatal("actual C attribution/source hit latch/HP synchronization lost")
							}
						} else if v.Obj130 != nil || v.Frame134 != 0 || au.Field130 != 0 {
							t.Fatal("invulnerable player acquired damage metadata")
						}
						steps[hit] = hp.Cur
					}
					t.Logf("C-owned >4 GiB C->registered DefaultDamage->UnitSetHP target=%p NewPlayer source=%p source_type=%d weapon=%p %s damage=3 damage_type=7 invulnerable=%t HP=2000->%d->%d->%d; PlayerDamage armor/carry/marker prefix untouched", v, a, a.TypeInd, w, attack, invulnerable, steps[0], steps[1], steps[2])
				})
			}
		}
	}
}
